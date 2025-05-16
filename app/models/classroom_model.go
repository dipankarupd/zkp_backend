package models

import "time"

type Classroom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TeacherName string    `json:"teacher_name"`
	CreatedAt   time.Time `json:"created_at"`
}
