package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"server/internal/config"
	"server/internal/domain/board"
	"server/internal/pkg/db"
	"server/internal/pkg/logger"
	"server/internal/scheduler"
	"server/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.New(cfg.App.Env)
	log.Info().Str("app", cfg.App.Name).Str("env", string(cfg.App.Env)).Msg("Starting application... ->")

	database, err := db.NewPostgresConn(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Database connected successfully!")

	boardRepo := board.NewRepository(database)

	sched, err := scheduler.New(boardRepo, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create scheduler")
	}
	sched.Start()

	app := server.New(cfg, log, database)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info().Msg("shutting down server")

		if err := sched.Shutdown(); err != nil {
			log.Error().Err(err).Msg("scheduler forced to shutdown")
		}

		if err := app.Shutdown(); err != nil {
			log.Error().Err(err).Msg("server forced to shutdown")
		}

		sqlDB, _ := database.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	addr := fmt.Sprintf(":%s", cfg.HTTP.Port)
	log.Info().Str("addr", addr).Msg("Listening server on ->")

	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
