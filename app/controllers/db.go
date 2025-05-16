// app/controllers/db.go
package controllers

import "github.com/dipankarupd/zkp/app/db"

var DB *db.DB

func InitializeDB(database *db.DB) {
	DB = database
}
