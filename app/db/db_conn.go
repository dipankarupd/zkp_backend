package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func ConnectDB() *DB {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	conn, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error opening the database: %v", err)
	}

	// verify with ping:
	err = conn.Ping()

	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	fmt.Println("✅ Database Connection successful!!!!")
	return &DB{
		conn: conn,
	}
}

func (db *DB) CloseDB() {
	if db.conn != nil {
		db.conn.Close()
		fmt.Println("Database connection closed.")
	}
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}
