package models

import "time"

type Student struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Location     Location  `json:"location"`
	LoginToken   int       `json:"token"`
	RegisteredAt time.Time `json:"registered_at"`
}
