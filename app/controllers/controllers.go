package controllers

// import (
// 	"encoding/json"
// 	"errors"
// 	"math/rand"
// 	"net/http"
// 	"strconv"
// 	"time"

// 	"github.com/dipankarupd/zkp/app/db"
// 	"github.com/dipankarupd/zkp/app/models"
// 	"github.com/gorilla/mux"
// )

// var database *db.DB

// // Initialize the database connection for controllers
// func InitializeDB(db *db.DB) {
// 	database = db
// }

// // RegisterUser handles POST requests to register a new student
// func RegisterUser(w http.ResponseWriter, r *http.Request) {
// 	// Parse the request body
// 	var input struct {
// 		Name      string  `json:"name"`
// 		Latitude  float64 `json:"latitude"`
// 		Longitude float64 `json:"longitude"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&input)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Create table if not exists
// 	createTableQuery := `
// 		CREATE TABLE IF NOT EXISTS students (
// 			id SERIAL PRIMARY KEY,
// 			name VARCHAR(255) NOT NULL,
// 			latitude FLOAT NOT NULL,
// 			longitude FLOAT NOT NULL,
// 			login_token INTEGER NOT NULL UNIQUE,
// 			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 			last_checked TIMESTAMP,
// 			present_count INTEGER DEFAULT 0,
// 			absent_count INTEGER DEFAULT 0
// 		)
// 	`
// 	_, err = database.Exec(createTableQuery)
// 	if err != nil {
// 		http.Error(w, "Database error", http.StatusInternalServerError)
// 		return
// 	}

// 	// Generate a unique login token (6-digit number)
// 	token, err := generateUniqueToken()
// 	if err != nil {
// 		http.Error(w, "Failed to generate unique token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Insert student data into the database
// 	query := `
// 		INSERT INTO students (name, latitude, longitude, login_token, present_count, absent_count)
// 		VALUES ($1, $2, $3, $4, 0, 0)
// 		RETURNING id, created_at
// 	`

// 	var id string
// 	var createdAt time.Time
// 	err = database.QueryRow(
// 		query,
// 		input.Name,
// 		input.Latitude,
// 		input.Longitude,
// 		token,
// 	).Scan(&id, &createdAt)
// 	if err != nil {
// 		http.Error(w, "Failed to insert student data", http.StatusInternalServerError)
// 		return
// 	}

// 	// Create and return the student object
// 	student := models.Student{
// 		ID:   id,
// 		Name: input.Name,
// 		Location: models.Location{
// 			Latitude:  input.Latitude,
// 			Longitude: input.Longitude,
// 		},
// 		LoginToken:   token,
// 		CreatedAt:    createdAt,
// 		PresentCount: 0,
// 		AbsentCount:  0,
// 	}

// 	// Set content type and return the response
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(student)
// }

// // GetUserByToken handles GET requests to retrieve a student by token
// func GetUserByToken(w http.ResponseWriter, r *http.Request) {
// 	// Extract token from URL parameters
// 	params := mux.Vars(r)
// 	tokenStr := params["token"]

// 	// Convert token from string to int
// 	loginToken, err := strconv.Atoi(tokenStr)
// 	if err != nil {
// 		http.Error(w, "Invalid token format", http.StatusBadRequest)
// 		return
// 	}

// 	// Query the database for the student
// 	query := `
// 		SELECT id, name, latitude, longitude, created_at, last_checked, present_count, absent_count
// 		FROM students
// 		WHERE login_token = $1
// 	`

// 	var (
// 		id           string
// 		name         string
// 		latitude     float64
// 		longitude    float64
// 		createdAt    time.Time
// 		lastChecked  *time.Time
// 		presentCount int
// 		absentCount  int
// 	)

// 	err = database.QueryRow(query, loginToken).Scan(
// 		&id, &name, &latitude, &longitude, &createdAt,
// 		&lastChecked, &presentCount, &absentCount,
// 	)
// 	if err != nil {
// 		http.Error(w, "Student not found", http.StatusNotFound)
// 		return
// 	}

// 	// Create and return the student object
// 	student := models.Student{
// 		ID:   id,
// 		Name: name,
// 		Location: models.Location{
// 			Latitude:  latitude,
// 			Longitude: longitude,
// 		},
// 		LoginToken:   loginToken,
// 		CreatedAt:    createdAt,
// 		LastChecked:  lastChecked,
// 		PresentCount: presentCount,
// 		AbsentCount:  absentCount,
// 	}

// 	// Set content type and return the response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(student)
// }

// // UpdateAttendance handles PUT requests to update student attendance
// func UpdateAttendance(w http.ResponseWriter, r *http.Request) {
// 	// Extract token from URL parameters
// 	params := mux.Vars(r)
// 	tokenStr := params["token"]

// 	// Convert token from string to int
// 	loginToken, err := strconv.Atoi(tokenStr)
// 	if err != nil {
// 		http.Error(w, "Invalid token format", http.StatusBadRequest)
// 		return
// 	}

// 	// Parse the request body
// 	var input struct {
// 		LastChecked string `json:"last_checked"`
// 		IsPresent   bool   `json:"is_present"`
// 	}

// 	err = json.NewDecoder(r.Body).Decode(&input)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Parse the last_checked date
// 	lastChecked, err := time.Parse(time.RFC3339, input.LastChecked)
// 	if err != nil {
// 		http.Error(w, "Invalid date format for last_checked. Use RFC3339 format (e.g., 2025-03-17T15:04:05Z)", http.StatusBadRequest)
// 		return
// 	}

// 	// Prepare the update query
// 	var updateQuery string
// 	if input.IsPresent {
// 		updateQuery = `
// 			UPDATE students
// 			SET last_checked = $1, present_count = present_count + 1
// 			WHERE login_token = $2
// 			RETURNING id, name, latitude, longitude, created_at, last_checked, present_count, absent_count
// 		`
// 	} else {
// 		updateQuery = `
// 			UPDATE students
// 			SET last_checked = $1, absent_count = absent_count + 1
// 			WHERE login_token = $2
// 			RETURNING id, name, latitude, longitude, created_at, last_checked, present_count, absent_count
// 		`
// 	}

// 	// Execute the update
// 	var (
// 		id                 string
// 		name               string
// 		latitude           float64
// 		longitude          float64
// 		createdAt          time.Time
// 		updatedLastChecked time.Time
// 		presentCount       int
// 		absentCount        int
// 	)

// 	err = database.QueryRow(
// 		updateQuery,
// 		lastChecked,
// 		loginToken,
// 	).Scan(
// 		&id, &name, &latitude, &longitude, &createdAt,
// 		&updatedLastChecked, &presentCount, &absentCount,
// 	)
// 	if err != nil {
// 		http.Error(w, "Student not found or update failed", http.StatusNotFound)
// 		return
// 	}

// 	// Create and return the updated student object
// 	student := models.Student{
// 		ID:   id,
// 		Name: name,
// 		Location: models.Location{
// 			Latitude:  latitude,
// 			Longitude: longitude,
// 		},
// 		LoginToken:   loginToken,
// 		CreatedAt:    createdAt,
// 		LastChecked:  &updatedLastChecked,
// 		PresentCount: presentCount,
// 		AbsentCount:  absentCount,
// 	}

// 	// Set content type and return the response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(student)
// }

// // Helper function to generate a unique 6-digit token
// func generateUniqueToken() (int, error) {
// 	// Seed the random number generator
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))

// 	// Try up to 10 times to generate a unique token
// 	for i := 0; i < 10; i++ {
// 		// Generate a random 6-digit number
// 		token := r.Intn(900000) + 100000 // Range: 100000-999999

// 		// Check if this token is already in use
// 		var exists bool
// 		query := "SELECT EXISTS(SELECT 1 FROM students WHERE login_token = $1)"
// 		err := database.QueryRow(query, token).Scan(&exists)
// 		if err != nil {
// 			return 0, err
// 		}

// 		// If token doesn't exist, return it
// 		if !exists {
// 			return token, nil
// 		}
// 	}

// 	// If we couldn't generate a unique token after several attempts
// 	return 0, errors.New("failed to generate unique login token after multiple attempts")
// }
