package main

import (
	"flag"
	"log"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "migration direction (up/down)")
	flag.Parse()

	// Load config to get DSN
	cfg := config.MustLoad()

	// Initialize migration
	m, err := migrate.New(
		"file://db/migrations",
		cfg.DSN,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}

	log.Printf("Running migrations: %s", direction)

	if direction == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run up migrations: %v", err)
		}
		log.Println("Up migrations applied successfully")
	} else if direction == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run down migrations: %v", err)
		}
		log.Println("Down migrations applied successfully")
	} else {
		log.Fatalf("Invalid direction: %s", direction)
	}
}
