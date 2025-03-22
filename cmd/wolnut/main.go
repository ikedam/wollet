package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ikedam/wollet/pkg/wolnut"
)

func main() {
	// Determine the base directory of the executable
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	baseDir := filepath.Dir(execPath)

	// Load configuration
	configFile := filepath.Join(baseDir, "wolnut.yaml")
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := wolnut.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Run the main logic
	if err := wolnut.Run(config); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
