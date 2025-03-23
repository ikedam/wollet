package main

import (
	"context"
	"os"

	"github.com/ikedam/wollet/pkg/log"
	"github.com/ikedam/wollet/pkg/wolnut"
)

func main() {
	ctx := context.Background()

	// Load configuration
	configFile := "wolnut.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := wolnut.LoadConfig(configFile)
	if err != nil {
		log.Error(ctx, "Failed to load configuration", log.WithError(err))
		os.Exit(1)
	}

	// Run the main logic
	if err := wolnut.Run(ctx, config); err != nil {
		log.Error(ctx, "Application error", log.WithError(err))
		os.Exit(1)
	}
}
