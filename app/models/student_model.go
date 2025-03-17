package models

import "time"

type Student struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Location     Location   `json:"location"`
	LoginToken   int        `json:"login_token"`
	CreatedAt    time.Time  `json:"created_at"`
	LastChecked  *time.Time `json:"last_checked,omitempty"`
	PresentCount int        `json:"present_count"`
	AbsentCount  int        `json:"absent_count"`
}
