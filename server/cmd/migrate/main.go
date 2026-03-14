package main

import (
	"context"
	"log"
	"os"

	"server/internal/config"
	"server/internal/pkg/db"

	_ "server/migrations"

	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.LoadConfig()

	gormDB, err := db.NewPostgresConn(cfg.Database)
	if err != nil {
		log.Fatalf("failed to open db connection: %v\n", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v\n", err)
	}
	defer sqlDB.Close()

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
	if err := goose.RunContext(context.Background(), command, sqlDB, "."); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
	log.Println("Migration completed!")
}
