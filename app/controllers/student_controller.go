package controllers

import (
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
func MarkAttendance(w http.ResponseWriter, r *http.Request) {}

func ViewHistory(w http.ResponseWriter, r *http.Request) {}

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
