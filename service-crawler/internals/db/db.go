package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB() *sqlx.DB {
	dsn := fmt.Sprintf(
		`host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`,
		os.Getenv("DB_CRAWLER_HOST"),
		os.Getenv("DB_CRAWLER_PORT"),
		os.Getenv("DB_CRAWLER_USER"),
		os.Getenv("DB_CRAWLER_PASSWORD"),
		os.Getenv("DB_CRAWLER_DBNAME"),
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect db: %v \n", err)
	}

	schema, err := os.ReadFile("./internals/db/schema.sql")
	if err != nil {
		log.Fatalf("Unable to find schema file: %v\n", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalf("Unable to run schema: %v\n", err)
	}

	log.Printf("DB is ready to use!")
	return db
}
