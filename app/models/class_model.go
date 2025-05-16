package models

import "time"

type Class struct {
	ID        string    `json:"id"`
	ClassId   string    `json:"classroom_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	MeetLink  string    `json:"meet_link"`
	CreatedAt time.Time `json:"created_at"`
}
