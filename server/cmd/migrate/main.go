package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "server/migrations"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	dbString := "user=postgres password=postgres dbname=app_db host=localhost port=5432 sslmode=disable"

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		log.Fatalf("failed to open db connection: %v\n", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v\n", err)
	}

	// Default to "up" if no argument is provided, otherwise use the provided argument
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Run the goose command
	log.Printf("Running goose %s...", command)
	if err := goose.RunContext(context.Background(), command, db, "."); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
	log.Println("Migration completed!")
}
