package wolbolt

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	baseDir := filepath.Dir(execPath)
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(baseDir, filePath)
	}

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
	if !filepath.IsAbs(config.PingFile) {
		config.PingFile = filepath.Join(baseDir, config.PingFile)
	}
	if config.LogFile == "" {
		config.LogFile = "ping.log"
	}
	if !filepath.IsAbs(config.LogFile) {
		config.LogFile = filepath.Join(baseDir, config.LogFile)
	}

	return &config, nil
}

type Server struct {
	mux *http.ServeMux
}

func NewServerForCGI(config *Config) *Server {
	return newServer(
		os.Getenv("SCRIPT_NAME"),
		config,
	)
}

func newServer(prefix string, config *Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("POST %v/ping", prefix), func(w http.ResponseWriter, r *http.Request) {
		// Parse the secret from the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(r.Context(), "Failed to read request body", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		bodyStr := string(body)
		bodyStr = strings.TrimRight(bodyStr, "\n")
		if bodyStr != config.Secret {
			log.Warn(r.Context(), "Invalid secret", log.String("remote_addr", r.RemoteAddr))
			recordBadRequest(r.Context(), config.LogFile, getRemoteAddr(r))
			w.Write([]byte("OK"))
			return
		}

		// Read current config
		currentPingResult, err := ReadPingResult(r.Context(), config.PingFile)
		if err != nil {
			log.Error(r.Context(), "Failed to read ping file: ignored", log.WithError(err))
		}

		// Record the IP address and timestamp
		pingResult := &PingResult{
			IP:         getRemoteAddr(r),
			UpdateTime: time.Now().UTC(),
		}
		err = pingResult.WriteTo(r.Context(), config.PingFile)
		if err != nil {
			log.Error(r.Context(), "Failed to write ping result", log.WithError(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if currentPingResult == nil || currentPingResult.IP != pingResult.IP {
			recordPingResult(
				r.Context(),
				config.LogFile,
				currentPingResult,
				pingResult,
			)
		}
		w.Write([]byte("OK"))
	})

	mux.HandleFunc(fmt.Sprintf("POST %v/wol", prefix), func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc(fmt.Sprintf("POST %v/{$}", prefix), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := log.CtxWithLogger(
		r.Context(),
		log.LoggerString(
			"uri",
			getPath(r),
		),
		log.LoggerString(
			"method",
			r.Method,
		),
		log.LoggerString(
			"remote_addr",
			getRemoteAddr(r),
		),
	)
	s.mux.ServeHTTP(w, r.WithContext(ctx))
}
