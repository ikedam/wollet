package wolbolt

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ikedam/wollet/pkg/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Secret       string  `yaml:"secret"`
	Port         int     `yaml:"port"`
	Count        int     `yaml:"count"`
	IntervalSecs float64 `yaml:"interval_secs"`
	PingFile     string  `yaml:"pingfile"`
	LogFile      string  `yaml:"logfile"`
}

func LoadConfig(filePath string) (*Config, error) {
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

	if config.PingFile == "" {
		config.PingFile = "ping.txt"
	}
	if config.LogFile == "" {
		config.LogFile = "ping.log"
	}

	return &config, nil
}

func Run(config *Config) error {
	log.Info(context.Background(), "Starting WOLbolt")

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		// Parse the secret from the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(r.Context(), "Failed to read request body", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if string(body) != config.Secret {
			log.Warn(r.Context(), "Invalid secret", log.String("remote_addr", r.RemoteAddr))
			w.Write([]byte("OK"))
			return
		}

		// Record the IP address and timestamp
		ip := r.RemoteAddr
		timestamp := time.Now().UTC().Format(time.RFC3339)
		tempFile := config.PingFile + ".tmp"
		file, err := os.Create(tempFile)
		if err != nil {
			log.Error(r.Context(), "Failed to create temp file", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = file.WriteString(ip + "\n" + timestamp + "\n")
		if err != nil {
			log.Error(r.Context(), "Failed to write to temp file", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := os.Rename(tempFile, config.PingFile); err != nil {
			log.Error(r.Context(), "Failed to rename temp file", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info(r.Context(), "Ping recorded", log.String("ip", ip), log.String("timestamp", timestamp))
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/wol", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		// Read the recorded IP address from the ping file
		data, err := os.ReadFile(config.PingFile)
		if err != nil {
			log.Error(r.Context(), "Failed to read ping file", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		lines := strings.Split(string(data), "\n")
		if len(lines) < 1 || lines[0] == "" {
			log.Warn(r.Context(), "Ping file is empty or invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ip := lines[0]

		// Send UDP packets to the recorded IP address
		addr := fmt.Sprintf("%s:%d", ip, config.Port)
		conn, err := net.Dial("udp", addr)
		if err != nil {
			log.Error(r.Context(), "Failed to dial UDP", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		for i := 0; i < config.Count; i++ {
			_, err := conn.Write([]byte(config.Secret))
			if err != nil {
				log.Error(r.Context(), "Failed to send UDP packet", log.WithError(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			time.Sleep(time.Duration(config.IntervalSecs * float64(time.Second)))
		}

		log.Info(r.Context(), "UDP packets sent", log.String("ip", ip), log.Int("count", config.Count))
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte("OK"))
	})

	return http.ListenAndServe(":8080", nil)
}
