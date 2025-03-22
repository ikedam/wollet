package wolbolt

import (
	"context"
	"io/ioutil"
	"net/http"

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
	data, err := ioutil.ReadFile(filePath)
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

		// Handle /ping logic
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/wol", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		// Handle /wol logic
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
