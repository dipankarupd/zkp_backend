package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dipankarupd/zkp/app/controllers"
	"github.com/dipankarupd/zkp/app/db"
	"github.com/dipankarupd/zkp/app/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Connect to the database
	database := db.ConnectDB()
	defer database.CloseDB()

	// Initialize the database connection for controllers
	controllers.InitializeDB(database)

	// Create a new router
	router := mux.NewRouter()

	// Register routes
	routes.RegisterNewRoutes(router)

	// Get the port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the HTTP server
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
