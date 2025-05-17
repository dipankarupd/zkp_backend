package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dipankarupd/zkp/app/models"
	"github.com/dipankarupd/zkp/app/utils"
	"github.com/gorilla/mux"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	type RegisterRequest struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid request payload",
			Data:    nil,
		})
		return
	}

	token, err := generateUniqueToken()
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to generate unique token",
			Data:    nil,
		})
		return
	}

	var studentID int
	var registeredAt time.Time
	err = DB.QueryRow(`
		INSERT INTO students (name, latitude, longitude, token)
		VALUES ($1, $2, $3, $4)
		RETURNING id, registered_at
	`, req.Name, req.Latitude, req.Longitude, token).Scan(&studentID, &registeredAt)

	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to register student",
			Data:    nil,
		})
		return
	}

	student := &models.Student{
		ID:           fmt.Sprint(studentID),
		Name:         req.Name,
		Location:     models.Location{Latitude: req.Latitude, Longitude: req.Longitude},
		LoginToken:   token,
		RegisteredAt: registeredAt,
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse[models.Student]{
		Status:  "success",
		Message: "Student registered successfully",
		Data:    student,
	})
}

func GetUserByToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from URL
	vars := mux.Vars(r)
	tokenStr := vars["token"]

	// Convert token string to integer
	token, err := strconv.Atoi(tokenStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid token format",
			Data:    nil,
		})
		return
	}

	// Query the student from database
	var student models.Student
	var latitude, longitude float64
	err = DB.QueryRow(`
		SELECT id, name, latitude, longitude, token, registered_at
		FROM students
		WHERE token = $1
	`, token).Scan(
		&student.ID,
		&student.Name,
		&latitude,
		&longitude,
		&student.LoginToken,
		&student.RegisteredAt,
	)

	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not found",
			Data:    nil,
		})
		return
	}

	student.Location = models.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Respond with student data
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[models.Student]{
		Status:  "success",
		Message: "Student retrieved successfully",
		Data:    &student,
	})
}

func JoinClassroom(w http.ResponseWriter, r *http.Request) {
	// Parse classroom ID from URL
	params := mux.Vars(r)
	classroomIDStr := params["id"]
	classroomID, err := strconv.Atoi(classroomIDStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid classroom ID",
			Data:    nil,
		})
		return
	}

	// Parse request body
	type joinRequest struct {
		Token int `json:"token"`
	}
	var req joinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid request payload",
			Data:    nil,
		})
		return
	}

	// Find student by token
	var studentID int
	query := `SELECT id FROM students WHERE token = $1`
	err = DB.QueryRow(query, req.Token).Scan(&studentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not found",
			Data:    nil,
		})
		return
	}

	// Check if already enrolled
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM student_classroom_enrollment
			WHERE student_id = $1 AND classroom_id = $2
		)`
	err = DB.QueryRow(checkQuery, studentID, classroomID).Scan(&exists)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to check enrollment",
			Data:    nil,
		})
		return
	}

	if exists {
		utils.RespondWithJSON(w, http.StatusConflict, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student already enrolled in this classroom",
			Data:    nil,
		})
		return
	}

	// Check if the first class has already started
	var firstClassStart time.Time
	firstClassQuery := `
		SELECT MIN(start_time) FROM classes 
		WHERE classroom_id = $1`
	err = DB.QueryRow(firstClassQuery, classroomID).Scan(&firstClassStart)
	if err != nil && err != sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to check class schedule",
			Data:    nil,
		})
		return
	}

	// If there is at least one class and it has already started
	if !firstClassStart.IsZero() && firstClassStart.Before(time.Now()) {
		utils.RespondWithJSON(w, http.StatusForbidden, utils.APIResponse[any]{
			Status:  "error",
			Message: "Classes already started. Cannot enroll now.",
			Data:    nil,
		})
		return
	}

	// Enroll the student
	var enrollmentID int
	joinedAt := time.Now()
	insertQuery := `
		INSERT INTO student_classroom_enrollment (student_id, classroom_id, joined_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	err = DB.QueryRow(insertQuery, studentID, classroomID, joinedAt).Scan(&enrollmentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to enroll student",
			Data:    nil,
		})
		return
	}

	// Return the Enrollment model
	enrollment := models.EnrollmentModels{
		ID:          fmt.Sprint(enrollmentID),
		StudentID:   fmt.Sprint(studentID),
		ClassroomID: fmt.Sprint(classroomID),
		JoinedAt:    joinedAt,
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[models.EnrollmentModels]{
		Status:  "success",
		Message: "Student enrolled in classroom successfully",
		Data:    &enrollment,
	})
}

func MarkAttendance(w http.ResponseWriter, r *http.Request) {

	type AttendanceRequest struct {
		CheckedInTime time.Time `json:"checked_in_time"`
		IsPresent     bool      `json:"is_present"`
	}

	params := mux.Vars(r)
	tokenStr := params["token"]
	classIDStr := params["id"]

	token, err := strconv.Atoi(tokenStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid student token",
			Data:    nil,
		})
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid class ID",
			Data:    nil,
		})
		return
	}

	var input AttendanceRequest
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid JSON payload",
			Data:    nil,
		})
		return
	}

	// Step 1: Get student ID from token
	var studentID int
	err = DB.QueryRow(`SELECT id FROM students WHERE token = $1`, token).Scan(&studentID)
	if err == sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not found",
			Data:    nil,
		})
		return
	} else if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error finding student",
			Data:    nil,
		})
		return
	}

	// Step 2: Get classroom ID and meet link from class ID
	var classroomID int
	var meetLink string
	err = DB.QueryRow(`SELECT classroom_id, link FROM classes WHERE id = $1`, classID).Scan(&classroomID, &meetLink)
	if err == sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Class not found",
			Data:    nil,
		})
		return
	} else if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error finding class",
			Data:    nil,
		})
		return
	}

	// Step 3: Determine attendance status
	status := "absent"
	if input.IsPresent {
		status = "present"
	}

	// Step 4: Insert into attendance table
	_, err = DB.Exec(`
		INSERT INTO attendance (student_id, class_id, attended_at, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (student_id, class_id) DO UPDATE 
		SET status = EXCLUDED.status, attended_at = EXCLUDED.attended_at
	`, studentID, classID, input.CheckedInTime, status)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to mark attendance",
			Data:    nil,
		})
		return
	}

	// Step 5: Update or insert into record table
	if status == "present" {
		_, err = DB.Exec(`
			INSERT INTO record (student_id, classroom_id, present_count, last_attended)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (student_id, classroom_id) DO UPDATE 
			SET present_count = record.present_count + 1, last_attended = EXCLUDED.last_attended
		`, studentID, classroomID, input.CheckedInTime)
	} else {
		_, err = DB.Exec(`
			INSERT INTO record (student_id, classroom_id, absent_count, last_attended)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (student_id, classroom_id) DO UPDATE 
			SET absent_count = record.absent_count + 1, last_attended = EXCLUDED.last_attended
		`, studentID, classroomID, input.CheckedInTime)
	}

	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to update record table",
			Data:    nil,
		})
		return
	}

	data := any(map[string]string{"meet_link": meetLink})

	// Step 6: Return meet link
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[any]{
		Status:  "success",
		Message: "Attendance marked successfully",
		Data:    &data,
	})
}
func ViewHistory(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tokenStr := params["token"]
	classroomIDStr := params["classroom_id"]

	// Parse classroom ID
	classroomID, err := strconv.Atoi(classroomIDStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid classroom ID",
			Data:    nil,
		})
		return
	}

	// Parse token as int (assuming token is integer)
	token, err := strconv.Atoi(tokenStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid token",
			Data:    nil,
		})
		return
	}

	// Step 1: Get student ID and name from token
	var studentID int
	var studentName string
	err = DB.QueryRow(`SELECT id, name FROM students WHERE token = $1`, token).Scan(&studentID, &studentName)
	if err == sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not found",
			Data:    nil,
		})
		return
	} else if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error",
			Data:    nil,
		})
		return
	}

	// Step 2: Get attendance summary (present_count, absent_count)
	var presentCount, absentCount int
	err = DB.QueryRow(`
		SELECT present_count, absent_count 
		FROM record 
		WHERE student_id = $1 AND classroom_id = $2
	`, studentID, classroomID).Scan(&presentCount, &absentCount)

	if err == sql.ErrNoRows {
		// If no record found, assume 0 attendance
		presentCount = 0
		absentCount = 0
	} else if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to fetch attendance record",
			Data:    nil,
		})
		return
	}

	totalAttendance := presentCount + absentCount
	var attendancePercentage float64
	if totalAttendance > 0 {
		attendancePercentage = (float64(presentCount) / float64(totalAttendance)) * 100
	} else {
		attendancePercentage = 0.0
	}

	responseData := map[string]any{
		"id":                    strconv.Itoa(studentID),
		"name":                  studentName,
		"total_class_taken":     totalAttendance,
		"present_count":         presentCount,
		"absent_count":          absentCount,
		"attendance_percentage": attendancePercentage,
	}

	resp := utils.APIResponse[map[string]any]{
		Status:  "success",
		Message: "Student attendance progress retrieved",
		Data:    &responseData,
	}

	utils.RespondWithJSON(w, http.StatusOK, resp)
}

func GetClassroomsOfStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tokenStr := params["token"]

	// Convert token to int
	token, err := strconv.Atoi(tokenStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid student token",
			Data:    nil,
		})
		return
	}

	// Get student ID from token
	var studentID int
	err = DB.QueryRow(`SELECT id FROM students WHERE token = $1`, token).Scan(&studentID)
	if err == sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not found for provided token",
			Data:    nil,
		})
		return
	} else if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error while retrieving student",
			Data:    nil,
		})
		return
	}

	// Fetch classrooms student is enrolled in
	rows, err := DB.Query(`
		SELECT c.id, c.name, c.description, c.teacher_name, c.created_at
		FROM classrooms c
		INNER JOIN student_classroom_enrollment e ON c.id = e.classroom_id
		WHERE e.student_id = $1
		ORDER BY c.created_at ASC
	`, studentID)

	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error while fetching classrooms",
			Data:    nil,
		})
		return
	}
	defer rows.Close()

	var classrooms []map[string]any
	for rows.Next() {
		var id int
		var name, description, teacherName string
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &description, &teacherName, &createdAt); err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
				Status:  "error",
				Message: "Error parsing classroom row",
				Data:    nil,
			})
			return
		}

		classroom := map[string]any{
			"id":           strconv.Itoa(id),
			"name":         name,
			"description":  description,
			"teacher_name": teacherName,
			"created_at":   createdAt.Format(time.RFC3339),
		}
		classrooms = append(classrooms, classroom)
	}

	res := any(classrooms)
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[any]{
		Status:  "success",
		Message: "Classrooms retrieved successfully",
		Data:    &res,
	})
}

func generateUniqueToken() (int, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 10; i++ {
		token := r.Intn(900000) + 100000 // 6-digit

		var exists bool
		err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM students WHERE token = $1)`, token).Scan(&exists)
		if err != nil {
			return 0, err
		}

		if !exists {
			return token, nil
		}
	}

	return 0, errors.New("failed to generate unique login token after multiple attempts")
}
