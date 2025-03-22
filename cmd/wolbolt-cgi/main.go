package main

import (
	"log"
	"os"

	"github.com/ikedam/wollet/pkg/wolbolt"
)

func main() {
	// Load configuration
	configFile := "wolbolt.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := wolbolt.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Run the main logic
	if err := wolbolt.Run(config); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
