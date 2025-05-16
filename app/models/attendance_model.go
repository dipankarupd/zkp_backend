package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// AttendanceStatus is a custom type for attendance.
type AttendanceStatus string

const (
	StatusPresent AttendanceStatus = "present"
	StatusAbsent  AttendanceStatus = "absent"
)

// Attendance is your model.
type Attendance struct {
	ID         int              `json:"id"`
	StudentID  int              `json:"student_id"`
	ClassID    int              `json:"class_id"`
	Status     AttendanceStatus `json:"status" validate:"oneof=present absent"`
	AttendedAt time.Time        `json:"attended_at"`
}

// Validate runs struct‐level validation.
func (a *Attendance) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
