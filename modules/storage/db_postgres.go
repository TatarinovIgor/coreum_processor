package storage

import (
	"database/sql"
	"fmt"
	"log"
)

func DbPSQLConnector() *sql.DB {
	dbUrl := "postgresql://postgres:local-postgres0!@localhost:5433/postgres"
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")
	return db
}
