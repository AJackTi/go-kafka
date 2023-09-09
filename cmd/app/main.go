package main

import (
	"log"

	"github.com/AJackTi/go-kafka/config"
	"github.com/AJackTi/go-kafka/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
