package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/AJackTi/go-kafka/internal/migration"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("MYSQL_URL"))
	sourcePath := strings.TrimSpace(os.Getenv("MIGRATIONS_PATH"))
	if sourcePath == "" {
		sourcePath = "migrations"
	}

	flag.StringVar(&databaseURL, "database-url", databaseURL, "MySQL DSN or mysql:// URL")
	flag.StringVar(&sourcePath, "path", sourcePath, "migration source directory")
	flag.Parse()

	if err := migration.Run(databaseURL, sourcePath); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")
}
