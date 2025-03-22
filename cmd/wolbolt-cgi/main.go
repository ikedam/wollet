package main

import (
	"context"
	"net/http/cgi"
	"os"

	"github.com/ikedam/wollet/pkg/log"
	"github.com/ikedam/wollet/pkg/wolbolt"
)

func main() {
	// Load configuration
	configFile := "wolbolt.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	ctx := context.Background()

	config, err := wolbolt.LoadConfig(configFile)
	if err != nil {
		log.Error(ctx, "Failed to load configuration", log.WithError(err))
		os.Exit(1)
	}

	// Run the main logic
	s := wolbolt.NewServerForCGI(config)
	if err := cgi.Serve(s); err != nil {
		log.Error(ctx, "Application error", log.WithError(err))
		os.Exit(1)
	}
	os.Exit(0)
}
