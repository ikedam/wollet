package wolnut

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ikedam/wollet/pkg/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Secret string `yaml:"secret"`
	Target string `yaml:"target"`
	Port   int    `yaml:"port"`
	Ping   struct {
		URL          string  `yaml:"url"`
		IntervalSecs float64 `yaml:"interval_secs"`
		BasicUser    string  `yaml:"basic_user"`
		BasicPass    string  `yaml:"basic_pass"`
	} `yaml:"ping"`
}

func LoadConfig(filePath string) (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	baseDir := filepath.Dir(execPath)
	filePath = filepath.Join(baseDir, filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Run(config *Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Handle signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		cancel()
	}()

	// Start periodic pinging
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(config.Ping.IntervalSecs * float64(time.Second))):
				// Perform ping logic here
				log.Info(ctx, "Pinging WOLbolt", log.String("url", config.Ping.URL))
			}
		}
	}()

	// Start UDP listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		addr := fmt.Sprintf(":%d", config.Port)
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			log.Error(ctx, "Failed to start UDP listener", log.WithError(err))
			cancel()
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, addr, err := conn.ReadFrom(buf)
				if err != nil {
					log.Warn(ctx, "Error reading UDP packet", log.WithError(err))
					continue
				}

				if string(buf[:n]) == config.Secret {
					log.Info(ctx, "Received valid magic packet", log.String("from", addr.String()))
					// Send Wake-on-LAN magic packet to target
				} else {
					log.Warn(ctx, "Invalid magic packet received", log.String("from", addr.String()))
				}
			}
		}
	}()

	wg.Wait()
	return nil
}
