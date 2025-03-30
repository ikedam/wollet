package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ikedam/wollet/pkg/log"
	"github.com/ikedam/wollet/pkg/wolnut"

	"github.com/spf13/cobra"
)

func main() {
	var pidFile string

	var rootCmd = &cobra.Command{
		Use:   "wolnut [config]",
		Short: "wolnut pings wolbolt and relays WOL command from wolbolt",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := "wolnut.yaml"
			if len(args) > 0 {
				configFile = args[0]
			}

			return mainImpl(configFile, pidFile)
		},
	}

	// フラグの追加
	rootCmd.Flags().StringVarP(&pidFile, "pidfile", "p", "", "Path to the PID file")

	// コマンドの実行
	if err := rootCmd.Execute(); err != nil {
		log.Error(context.Background(), "Failed to load configuration", log.WithError(err))
		os.Exit(1)
	}
	os.Exit(0)
}

func mainImpl(configFile string, pidFile string) error {
	ctx := context.Background()

	if pidFile != "" {
		// PIDファイルを作成
		if err := writePIDFile(ctx, pidFile); err != nil {
			return err
		}
		defer removePIDFile(ctx, pidFile)
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

	return nil
}

func writePIDFile(ctx context.Context, pidFile string) error {
	pid := os.Getpid()
	file, err := os.Create(pidFile)
	if err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", pid)
	if err != nil {
		return fmt.Errorf("failed to write to PID file: %w", err)
	}

	log.Info(ctx, "PID file created", log.String("pidfile", pidFile), log.Int("pid", pid))
	return nil
}

func removePIDFile(ctx context.Context, pidFile string) {
	err := os.Remove(pidFile)
	if err != nil {
		log.Info(ctx, "Failed to remove PID file", log.WithError(err))
	} else {
		log.Info(ctx, "PID file removed", log.String("pidfile", pidFile))
	}
}
