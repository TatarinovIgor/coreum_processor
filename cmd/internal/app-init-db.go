package internal

import (
	"database/sql"
	"fmt"
	"log"
)

// DBConnect inits database connection
func DBConnect() *sql.DB {
	// db connection env variables
	var (
		dbHost     = MustString("DATABASE_HOST")
		dbPort     = MustInt("DATABASE_PORT")
		dbName     = MustString("DATABASE_NAME")
		dbUsername = MustString("DATABASE_USER")
		dbPassword = MustString("DATABASE_PASS")
		dbTimeout  = GetInt("DATABASE_CONNECT_TIMEOUT", 15)
		dbSSLMode  = GetString("DATABASE_SSL_MODE", "disable")
	)

	dbConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		dbHost, dbPort, dbUsername, dbPassword, dbName, dbSSLMode, dbTimeout,
	)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(0)

	if err := db.Ping(); err != nil {
		log.Fatalf("database is not available right now: %v", err)
	}

	return db
}
