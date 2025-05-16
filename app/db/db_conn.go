// package db

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"

// 	_ "github.com/lib/pq"
// )

// type DB struct {
// 	conn *sql.DB
// }

// func ConnectDB() *DB {
// 	dbHost := os.Getenv("DB_HOST")
// 	dbPort := os.Getenv("DB_PORT")
// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")
// 	dbName := os.Getenv("DB_NAME")
// 	connStr := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName,
// 	)

// 	conn, err := sql.Open("postgres", connStr)

// 	if err != nil {
// 		log.Fatalf("Error opening the database: %v", err)
// 	}

// 	// verify with ping:
// 	err = conn.Ping()

// 	if err != nil {
// 		log.Fatalf("Database connection failed: %v", err)
// 	}

// 	fmt.Println("✅ Database Connection successful!!!!")
// 	return &DB{
// 		conn: conn,
// 	}
// }

// func (db *DB) CloseDB() {
// 	if db.conn != nil {
// 		db.conn.Close()
// 		fmt.Println("Database connection closed.")
// 	}
// }

// func (db *DB) getConn() *sql.DB {
// 	return db.conn
// }

// // Exec executes a query without returning any rows
// func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
// 	return db.conn.Exec(query, args...)
// }

// // QueryRow executes a query that is expected to return at most one row
// func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
// 	return db.conn.QueryRow(query, args...)
// }

// // Query executes a query that returns rows
// func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
// 	return db.conn.Query(query, args...)
// }

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

	// Log connection parameters for debugging (mask password in production)
	log.Printf("Connecting to database at %s:%s as %s (database: %s)",
		dbHost, dbPort, dbUser, dbName)

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	// Set connection timeout for Docker environment
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return nil
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Set a short timeout for Ping to avoid hanging
	conn.SetConnMaxIdleTime(10 * time.Second)

	// Verify with ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = conn.PingContext(ctx)
	if err != nil {
		log.Printf("Database ping failed: %v", err)
		return nil
	}

	log.Println("✅ Database Connection successful!!!!")
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

func (db *DB) GetConn() *sql.DB {
	return db.conn
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
