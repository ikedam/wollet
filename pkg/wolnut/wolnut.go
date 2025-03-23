package wolnut

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ikedam/wollet/pkg/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Secret string `yaml:"secret"`
	Target string `yaml:"target"`
	Iface  string `yaml:"iface"`
	Port   int    `yaml:"port"`
	Ping   struct {
		URL          string  `yaml:"url"`
		IntervalSecs float64 `yaml:"interval_secs"`
		BasicUser    string  `yaml:"basic_user"`
		BasicPass    string  `yaml:"basic_pass"`
	} `yaml:"ping"`
}

func LoadConfig(filePath string) (*Config, error) {
	if !filepath.IsAbs(filePath) {
		execPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		baseDir := filepath.Dir(execPath)
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

	return &config, nil
}

func Run(ctx context.Context, config *Config) error {
	// Parse the MAC address from the target string
	mac, err := net.ParseMAC(config.Target)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}

	// ネットワークインターフェースを取得
	iface, err := net.InterfaceByName(config.Iface)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %w", config.Iface, err)
	}

	// RAWソケットを作成
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(syscall.ETH_P_ALL))
	if err != nil {
		return fmt.Errorf("failed to create raw socket: %w", err)
	}
	defer syscall.Close(fd)

	// UDP の待ち受け
	addr := fmt.Sprintf(":%d", config.Port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	defer conn.Close()
	log.Info(ctx, "Listing for UDP packets", log.String("address", addr))

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Handle signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		cancel()
		conn.Close()
	}()

	// Start periodic pinging
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err := doPing(ctx, config)
			if err != nil {
				if err == context.Canceled {
					log.Info(ctx, "Stopping pinging")
					return
				}
				log.Error(ctx, "Failed to ping...continue", log.WithError(err))
			}

			select {
			case <-ctx.Done():
				log.Info(ctx, "Stopping pinging")
				return
			case <-time.After(time.Duration(config.Ping.IntervalSecs * float64(time.Second))):
				continue
			}
		}
	}()

	// Start UDP listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, addr, err := conn.ReadFrom(buf)
				if err != nil {
					log.Warn(ctx, "Error reading UDP packet", log.WithError(err))
					if err == net.ErrClosed {
						return
					}
					continue
				}

				secret := string(buf[:n])
				secret = strings.TrimRight(secret, "\n")
				if secret != config.Secret {
					log.Warn(ctx, "Invalid magic packet received", log.String("from", addr.String()))
					continue
				}
				err = doWOL(ctx, mac, iface, fd)
				if err != nil {
					log.Error(ctx, "Failed to send WOL packet", log.WithError(err))
					continue
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

func doPing(ctx context.Context, config *Config) error {
	log.Info(ctx, "Pinging WOLbolt", log.String("url", config.Ping.URL))
	body := bytes.NewBuffer([]byte(config.Secret))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.Ping.URL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if config.Ping.BasicUser != "" && config.Ping.BasicPass != "" {
		req.SetBasicAuth(config.Ping.BasicUser, config.Ping.BasicPass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func doWOL(ctx context.Context, mac net.HardwareAddr, iface *net.Interface, fd int) error {
	log.Info(ctx, "Sending WOL packet", log.String("target", mac.String()))

	// Create the magic packet
	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], mac)
	}

	// Send the magic packet as an Ethernet frame
	ethFrame := append(mac, iface.HardwareAddr...) // 宛先MAC + 送信元MAC
	ethFrame = append(ethFrame, packet...)
	// RAWソケットで送信
	err := syscall.Sendto(fd, ethFrame, 0, &syscall.SockaddrLinklayer{
		Ifindex:  iface.Index,
		Protocol: syscall.ETH_P_ALL,
	})
	if err != nil {
		return fmt.Errorf("failed to send magic packet: %w", err)
	}
	return nil
}
