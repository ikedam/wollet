package wolbolt

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ikedam/wollet/pkg/log"
)

func recordBadRequest(ctx context.Context, logfile string, remoteAddress string) {
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(ctx, "Failed to open log file", log.WithError(err))
		return
	}
	defer file.Close()
	logLine := fmt.Sprintf("%s: Bad request from %s\n", time.Now().Format(time.RFC3339), remoteAddress)
	_, err = file.WriteString(logLine)
	if err != nil {
		log.Error(ctx, "Failed to write to log file", log.WithError(err))
		return
	}
}

func recordPingResult(ctx context.Context, logfile string, currentPingResult *PingResult, newPingResult *PingResult) {
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(ctx, "Failed to open log file", log.WithError(err))
		return
	}
	defer file.Close()
	var logLine string
	if currentPingResult != nil {
		logLine = fmt.Sprintf("%s: Ping recorded from %s (was %s)\n", time.Now().Format(time.RFC3339), newPingResult.IP, currentPingResult.IP)
	} else {
		logLine = fmt.Sprintf("%s: Ping recorded from %s\n", time.Now().Format(time.RFC3339), newPingResult.IP)
	}
	_, err = file.WriteString(logLine)
	if err != nil {
		log.Error(ctx, "Failed to write to log file", log.WithError(err))
		return
	}
}
