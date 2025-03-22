package wolbolt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type PingResult struct {
	IP         string
	UpdateTime time.Time
}

func (p *PingResult) WriteTo(ctx context.Context, filename string) error {
	tempFile := fmt.Sprintf("%v.%v.tmp", filename, os.Getpid())
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	timestamp := p.UpdateTime.Format(time.RFC3339)
	_, err = file.WriteString(p.IP + "\n" + timestamp + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

func ReadPingResult(ctx context.Context, filename string) (*PingResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read ping file: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid ping file format")
	}
	ip := lines[0]
	timestamp, err := time.Parse(time.RFC3339, lines[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	return &PingResult{
		IP:         ip,
		UpdateTime: timestamp,
	}, nil
}
