package controllers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dipankarupd/zkp/app/models"
	"github.com/dipankarupd/zkp/app/utils"
	"github.com/gorilla/mux"
)

func CreateClassroom(w http.ResponseWriter, r *http.Request) {
	// Define request structure
	type ClassroomRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Teacher     string `json:"teacher_name"`
	}

	var req ClassroomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Teacher == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid request payload. 'name' and 'teacher_name' are required.",
			Data:    nil,
		})
		return
	}

	// Insert classroom into the DB
	var classroomID int
	var createdAt time.Time
	query := `
		INSERT INTO classrooms (name, description, teacher_name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := DB.QueryRow(query, req.Name, req.Description, req.Teacher).Scan(&classroomID, &createdAt)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to create classroom",
			Data:    nil,
		})
		return
	}

	// Build classroom model for response
	classroom := models.Classroom{
		ID:          fmt.Sprint(classroomID),
		Name:        req.Name,
		Description: req.Description,
		TeacherName: req.Teacher,
		CreatedAt:   createdAt,
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse[models.Classroom]{
		Status:  "success",
		Message: "Classroom created successfully",
		Data:    &classroom,
	})
}

func GetAllClassrooms(w http.ResponseWriter, r *http.Request) {
	// Query to fetch all classrooms
	query := `SELECT id, name, description, teacher_name, created_at FROM classrooms`

	rows, err := DB.Query(query)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to fetch classrooms",
			Data:    nil,
		})
		return
	}
	defer rows.Close()

	classrooms := []models.Classroom{}

	for rows.Next() {
		var c models.Classroom
		var id int
		err := rows.Scan(&id, &c.Name, &c.Description, &c.TeacherName, &c.CreatedAt)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
				Status:  "error",
				Message: "Failed to parse classroom data",
				Data:    nil,
			})
			return
		}
		c.ID = fmt.Sprint(id)
		classrooms = append(classrooms, c)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Error reading classrooms",
			Data:    nil,
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[[]models.Classroom]{
		Status:  "success",
		Message: "Classrooms retrieved successfully",
		Data:    &classrooms,
	})
}

func CreateClass(w http.ResponseWriter, r *http.Request) {
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
	type ClassRequest struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Link      string `json:"link"`
	}

	var req ClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid request body",
			Data:    nil,
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid start_time format. Use RFC3339",
			Data:    nil,
		})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil || endTime.Before(startTime) {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid end_time. It must be after start_time",
			Data:    nil,
		})
		return
	}

	if req.Link == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Meeting link is required",
			Data:    nil,
		})
		return
	}

	// Insert the new class
	var classID int
	var createdAt time.Time
	err = DB.QueryRow(`
		INSERT INTO classes (classroom_id, start_time, end_time, link)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, classroomID, startTime, endTime, req.Link).Scan(&classID, &createdAt)

	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to create class",
			Data:    nil,
		})
		return
	}

	// Respond with the class details
	class := models.Class{
		ID:        strconv.Itoa(classID),
		ClassId:   strconv.Itoa(classroomID),
		StartTime: startTime,
		EndTime:   endTime,
		MeetLink:  req.Link,
		CreatedAt: createdAt,
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse[models.Class]{
		Status:  "success",
		Message: "Class created successfully",
		Data:    &class,
	})
}

func GetClassroomDetails(w http.ResponseWriter, r *http.Request) {
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

	// Step 1: Fetch classroom info
	var classroom models.Classroom
	query := `SELECT id, name, description, teacher_name, created_at FROM classrooms WHERE id = $1`
	err = DB.QueryRow(query, classroomID).Scan(
		&classroom.ID,
		&classroom.Name,
		&classroom.Description,
		&classroom.TeacherName,
		&classroom.CreatedAt,
	)
	if err == sql.ErrNoRows {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Classroom not found",
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

	// Step 2: Count total students
	var totalStudents int
	err = DB.QueryRow(`SELECT COUNT(*) FROM student_classroom_enrollment WHERE classroom_id = $1`, classroomID).Scan(&totalStudents)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to count students",
			Data:    nil,
		})
		return
	}

	// Step 3: Count total classes
	var totalClasses int
	err = DB.QueryRow(`SELECT COUNT(*) FROM classes WHERE classroom_id = $1`, classroomID).Scan(&totalClasses)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to count classes",
			Data:    nil,
		})
		return
	}

	// Step 4: Get student list with attendance summary
	rows, err := DB.Query(`
		SELECT s.id, s.name, COALESCE(r.present_count, 0), COALESCE(r.absent_count, 0)
		FROM students s
		INNER JOIN student_classroom_enrollment e ON s.id = e.student_id
		LEFT JOIN record r ON s.id = r.student_id AND r.classroom_id = $1
		WHERE e.classroom_id = $1
	`, classroomID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to fetch student data",
			Data:    nil,
		})
		return
	}
	defer rows.Close()

	type StudentSummary struct {
		ID                   string  `json:"id"`
		Name                 string  `json:"name"`
		PresentCount         int     `json:"present_count"`
		AbsentCount          int     `json:"absent_count"`
		AttendancePercentage float64 `json:"attendance_percentage"`
	}

	var students []StudentSummary
	for rows.Next() {
		var s StudentSummary
		var present, absent int
		err := rows.Scan(&s.ID, &s.Name, &present, &absent)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
				Status:  "error",
				Message: "Failed to parse student row",
				Data:    nil,
			})
			return
		}

		s.PresentCount = present
		s.AbsentCount = absent
		total := present + absent
		if total > 0 {
			s.AttendancePercentage = float64(present) / float64(total) * 100
		} else {
			s.AttendancePercentage = 0.0
		}
		students = append(students, s)
	}

	// Final JSON response
	response := map[string]any{
		"id":                classroom.ID,
		"name":              classroom.Name,
		"description":       classroom.Description,
		"teacher_name":      classroom.TeacherName,
		"created_at":        classroom.CreatedAt.Format(time.RFC3339),
		"total_students":    totalStudents,
		"classes_conducted": totalClasses,
		"students":          students,
	}
	respData := any(response)
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[any]{
		Status:  "success",
		Message: "Classroom details retrieved",
		Data:    &respData,
	})
}

// func GetAllStudents(w http.ResponseWriter, r *http.Request) {}

func GetIndividualStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	classroomIDStr := params["id"]
	studentIDStr := params["student_id"]

	classroomID, err := strconv.Atoi(classroomIDStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid classroom ID",
			Data:    nil,
		})
		return
	}

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse[any]{
			Status:  "error",
			Message: "Invalid student ID",
			Data:    nil,
		})
		return
	}

	// Fetch student name
	var studentName string
	err = DB.QueryRow(`SELECT name FROM students WHERE id = $1`, studentID).Scan(&studentName)
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
			Message: "Database error fetching student",
			Data:    nil,
		})
		return
	}

	// Check if student enrolled in classroom
	var enrollmentExists bool
	err = DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM student_classroom_enrollment WHERE student_id=$1 AND classroom_id=$2)`, studentID, classroomID).Scan(&enrollmentExists)
	if err != nil || !enrollmentExists {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.APIResponse[any]{
			Status:  "error",
			Message: "Student not enrolled in this classroom",
			Data:    nil,
		})
		return
	}

	// Fetch present_count and absent_count from record table
	var presentCount, absentCount int
	err = DB.QueryRow(`SELECT present_count, absent_count FROM record WHERE student_id=$1 AND classroom_id=$2`, studentID, classroomID).Scan(&presentCount, &absentCount)
	if err == sql.ErrNoRows {
		// No attendance record yet; set counts to zero
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

	// Calculate attendance percentage
	totalAttendance := presentCount + absentCount
	var attendancePercentage float64
	if totalAttendance > 0 {
		attendancePercentage = (float64(presentCount) / float64(totalAttendance)) * 100
	} else {
		attendancePercentage = 0.0
	}

	// Build response data
	responseData := map[string]any{
		"id":                    strconv.Itoa(studentID),
		"name":                  studentName,
		"total_class_taken":     totalAttendance, // sum of present + absent
		"present_count":         presentCount,
		"absent_count":          absentCount,
		"attendance_percentage": attendancePercentage,
	}
	res := any(responseData)
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[any]{
		Status:  "success",
		Message: "Student attendance details retrieved",
		Data:    &res,
	})
}

func GetAllClasses(w http.ResponseWriter, r *http.Request) {
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

	rows, err := DB.Query(`
		SELECT id, classroom_id, start_time, end_time, link, created_at
		FROM classes
		WHERE classroom_id = $1
		ORDER BY start_time ASC
	`, classroomID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Database error while fetching classes",
			Data:    nil,
		})
		return
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var class models.Class
		err := rows.Scan(
			&class.ID,
			&class.ClassId,
			&class.StartTime,
			&class.EndTime,
			&class.MeetLink,
			&class.CreatedAt,
		)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
				Status:  "error",
				Message: "Error parsing class row",
				Data:    nil,
			})
			return
		}
		classes = append(classes, class)
	}

	response := map[string]any{
		"classroom_id": classroomIDStr,
		"classes":      classes,
	}

	res := any(response)
	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse[any]{
		Status:  "success",
		Message: "Classes retrieved successfully",
		Data:    &res,
	})
}

func MakeCSV(w http.ResponseWriter, r *http.Request) {
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

	// Query student summary: id, name, present_count, absent_count
	rows, err := DB.Query(`
		SELECT s.id, s.name, COALESCE(r.present_count, 0), COALESCE(r.absent_count, 0)
		FROM students s
		INNER JOIN student_classroom_enrollment e ON s.id = e.student_id
		LEFT JOIN record r ON s.id = r.student_id AND r.classroom_id = $1
		WHERE e.classroom_id = $1
	`, classroomID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to fetch student data",
			Data:    nil,
		})
		return
	}
	defer rows.Close()

	var csvData [][]string
	// Header row
	csvData = append(csvData, []string{"Student ID", "Name", "Total Classes", "Present Count", "Absent Count", "Attendance Percentage"})

	for rows.Next() {
		var id, name string
		var present, absent int
		if err := rows.Scan(&id, &name, &present, &absent); err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
				Status:  "error",
				Message: "Failed to scan row",
				Data:    nil,
			})
			return
		}

		total := present + absent
		percentage := 0.0
		if total > 0 {
			percentage = float64(present) / float64(total) * 100
		}

		csvData = append(csvData, []string{
			id,
			name,
			strconv.Itoa(total),
			strconv.Itoa(present),
			strconv.Itoa(absent),
			fmt.Sprintf("%.2f", percentage),
		})
	}

	// Create buffer to write CSV in-memory
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	err = writer.WriteAll(csvData)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.APIResponse[any]{
			Status:  "error",
			Message: "Failed to write CSV data",
			Data:    nil,
		})
		return
	}
	writer.Flush()

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=attendance_report.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
