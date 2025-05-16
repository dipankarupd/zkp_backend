package models

import "time"

type EnrollmentModels struct {
	ID          string    `json:"id"`
	StudentID   string    `json:"student_id"`
	ClassroomID string    `json:"classroom_id"`
	JoinedAt    time.Time `json:"joined_at"`
}
