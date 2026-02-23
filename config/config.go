package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"piyazche/utils"
)

// Config represents the main configuration file
type Config struct {
	Proxy    ProxyConfig    `json:"proxy"`
	Fragment FragmentConfig `json:"fragment"`
	Scan     ScanConfig     `json:"scan"`
	Output   OutputConfig   `json:"output"`
	Xray     XrayConfig     `json:"xray"`
}

// XrayConfig represents xray-specific settings
type XrayConfig struct {
	LogLevel string    `json:"logLevel"` // none, error, warning, info, debug
	Mux      MuxConfig `json:"mux"`
}

// MuxConfig represents mux settings for xray
type MuxConfig struct {
	Enabled         bool   `json:"enabled"`
	Concurrency     int    `json:"concurrency"`     // -1 to disable, 8 is typical
	XudpConcurrency int    `json:"xudpConcurrency"` // typically 4
	XudpProxyUDP443 string `json:"xudpProxyUDP443"` // "reject", "allow", etc.
}

// ProxyConfig represents the proxy settings
type ProxyConfig struct {
	UUID    string         `json:"uuid"`
	Address string         `json:"address"` // IP or domain for vnext
	Port    int            `json:"port"`
	Method  string         `json:"method"` // tls, reality
	Type    string         `json:"type"`   // ws, xhttp, grpc, tcp, httpupgrade
	TLS     *TlsConfig     `json:"tls,omitempty"`
	Reality *RealityConfig `json:"reality,omitempty"`
	WS      *WsConfig      `json:"ws,omitempty"`
	Grpc    *GrpcConfig    `json:"grpc,omitempty"`
	Xhttp   *XhttpConfig   `json:"xhttp,omitempty"`
}

// TlsConfig represents TLS security settings
type TlsConfig struct {
	SNI           string   `json:"sni"`
	ALPN          []string `json:"alpn"`
	Fingerprint   string   `json:"fingerprint"`
	AllowInsecure bool     `json:"allowInsecure"`
}

// RealityConfig represents reality security settings
type RealityConfig struct {
	PublicKey   string `json:"publicKey"`
	ShortId     string `json:"shortId"`
	SpiderX     string `json:"spiderX"`
	Fingerprint string `json:"fingerprint"`
	ServerName  string `json:"serverName"`
}

// WsConfig represents WebSocket transport settings
type WsConfig struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

// GrpcConfig represents gRPC transport settings
type GrpcConfig struct {
	ServiceName        string `json:"serviceName"`
	Authority          string `json:"authority"`
	MultiMode          bool   `json:"multiMode"`
	IdleTimeout        int    `json:"idleTimeout"`
	HealthCheckTimeout int    `json:"healthCheckTimeout"`
}

// XhttpConfig represents xhttp transport settings
type XhttpConfig struct {
	Host string `json:"host"`
	Path string `json:"path"`
	Mode string `json:"mode"` // auto, stream, etc.
}

// FragmentConfig represents TLS fragment settings
type FragmentConfig struct {
	Enabled bool               `json:"enabled"`
	Mode    string             `json:"mode"` // "manual", "auto", or "off"
	Packets string             `json:"packets"`
	Manual  ManualFragment     `json:"manual"`
	Auto    AutoFragmentConfig `json:"auto"`
}

// ManualFragment represents manual fragment settings
type ManualFragment struct {
	Length   string `json:"length"`   // e.g., "10-20"
	Interval string `json:"interval"` // e.g., "10-20" (ms)
}

// AutoFragmentConfig represents auto-discovery settings
type AutoFragmentConfig struct {
	LengthRange      Range   `json:"lengthRange"`
	IntervalRange    Range   `json:"intervalRange"`
	MaxTests         int     `json:"maxTests"`
	SuccessThreshold float64 `json:"successThreshold"`
	TestIP           string  `json:"testIP"` // Custom IP for testing, uses random from subnet list if empty
}

// Range represents a min-max range
type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// ScanConfig represents scanner settings
type ScanConfig struct {
	Threads         int    `json:"threads"`
	Timeout         int    `json:"timeout"` // seconds
	TestURL         string `json:"testUrl"`
	MaxLatency      int     `json:"maxLatency"`      // ms
	MaxPacketLoss   float64 `json:"maxPacketLoss"`   // percent, 0=disabled
	Retries         int    `json:"retries"`
	MaxIPs          int    `json:"maxIPs"`
	Shuffle         bool   `json:"shuffle"`
	SampleSize      int    `json:"sampleSize"` // IPs per subnet
	SpeedTest       bool   `json:"speedTest"`
	DownloadURL     string `json:"downloadUrl"`
	UploadURL       string `json:"uploadUrl"`
	PacketLossCount    int     `json:"packetLossCount"`    // number of pings for packet loss
	StabilityRounds    int     `json:"stabilityRounds"`    // phase-2 rounds (0=disabled)
	StabilityInterval  int     `json:"stabilityInterval"`  // seconds between rounds
	JitterTest         bool    `json:"jitterTest"`         // measure latency jitter
	MinDownloadMbps    float64 `json:"minDownloadMbps"`    // filter: 0=disabled
	MinUploadMbps      float64 `json:"minUploadMbps"`      // filter: 0=disabled
	MaxPacketLossPct   float64 `json:"maxPacketLossPct"`   // filter: -1=disabled 0=strict
}

// OutputConfig represents output settings
type OutputConfig struct {
	Format                  string `json:"format"` // csv, json
	Directory               string `json:"directory"`
	SaveIntermediateResults bool   `json:"saveIntermediateResults"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Proxy: ProxyConfig{
			UUID:   "",
			Port:   443,
			Method: "tls",
			Type:   "ws",
			TLS: &TlsConfig{
				SNI:           "",
				ALPN:          []string{"http/1.1", "h2"},
				Fingerprint:   "chrome",
				AllowInsecure: false,
			},
			WS: &WsConfig{
				Host: "",
				Path: "/",
			},
		},
		Fragment: FragmentConfig{
			Enabled: true,
			Mode:    "manual",
			Packets: "tlshello",
			Manual: ManualFragment{
				Length:   "10-20",
				Interval: "10-20",
			},
			Auto: AutoFragmentConfig{
				LengthRange:      Range{Min: 1, Max: 100},
				IntervalRange:    Range{Min: 1, Max: 100},
				MaxTests:         200,
				SuccessThreshold: 0.6,
			},
		},
		Scan: ScanConfig{
			Threads:         1,
			Timeout:         10,
			TestURL:         "https://www.gstatic.com/generate_204",
			MaxLatency:      2500,
			Retries:         2,
			MaxIPs:          0,
			Shuffle:         true,
			SampleSize:      1,
			SpeedTest:       false,
			DownloadURL:     "https://speed.cloudflare.com/__down?bytes=1000000",
			UploadURL:       "https://speed.cloudflare.com/__up",
			PacketLossCount:    5,
			StabilityRounds:    0,
			StabilityInterval:  5,
			JitterTest:         false,
			MinDownloadMbps:    0,
			MinUploadMbps:      0,
			MaxPacketLossPct:   -1,
		},
		Output: OutputConfig{
			Format:                  "csv",
			Directory:               "results",
			SaveIntermediateResults: true,
		},
		Xray: XrayConfig{
			LogLevel: "none",
			Mux: MuxConfig{
				Enabled:         false,
				Concurrency:     -1,
				XudpConcurrency: 4,
				XudpProxyUDP443: "reject",
			},
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Proxy.UUID == "" {
		return fmt.Errorf("proxy.uuid is required")
	}

	// Validate method
	method := c.Proxy.Method
	if method == "" {
		method = "tls"
		c.Proxy.Method = method
	}
	if method != "tls" && method != "reality" {
		return fmt.Errorf("invalid proxy.method: %s (must be 'tls' or 'reality')", method)
	}

	// Validate TLS config
	if method == "tls" {
		if c.Proxy.TLS == nil {
			return fmt.Errorf("proxy.tls is required when method is 'tls'")
		}
		if c.Proxy.TLS.SNI == "" {
			return fmt.Errorf("proxy.tls.sni is required")
		}
	}

	// Validate Reality config
	if method == "reality" {
		if c.Proxy.Reality == nil {
			return fmt.Errorf("proxy.reality is required when method is 'reality'")
		}
		if c.Proxy.Reality.PublicKey == "" {
			return fmt.Errorf("proxy.reality.publicKey is required")
		}
	}

	if c.Proxy.Port <= 0 {
		c.Proxy.Port = 443
	}

	validTypes := map[string]bool{"ws": true, "xhttp": true, "grpc": true, "tcp": true, "httpupgrade": true}
	if !validTypes[c.Proxy.Type] {
		return fmt.Errorf("invalid proxy.type: %s", c.Proxy.Type)
	}

	if c.Fragment.Mode == "" {
		c.Fragment.Mode = "manual"
	}
	if c.Fragment.Mode != "manual" && c.Fragment.Mode != "auto" && c.Fragment.Mode != "off" {
		return fmt.Errorf("invalid fragment.mode: %s", c.Fragment.Mode)
	}

	if c.Scan.Threads <= 0 {
		c.Scan.Threads = 1
	}
	if c.Scan.Timeout <= 0 {
		c.Scan.Timeout = 10
	}

	return nil
}

// GetTimeout returns the timeout as a duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Scan.Timeout) * time.Second
}

// GetMaxLatency returns the max latency as a duration
func (c *Config) GetMaxLatency() time.Duration {
	return time.Duration(c.Scan.MaxLatency) * time.Millisecond
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// PrintConfigInfo prints configuration information in a colorized format
func (c *Config) PrintConfigInfo() {
	line := "═══════════════════════════════════════════════════════════"
	fmt.Printf("%s%s%s\n", utils.Cyan, line, utils.Reset)
	fmt.Printf("%s%s  Piyazche Configuration%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("%s%s%s\n\n", utils.Cyan, line, utils.Reset)

	fmt.Printf("%s%s▸ Proxy Settings%s\n", utils.Bold, utils.Yellow, utils.Reset)

	if c.Proxy.Address != "" {
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Address:", utils.Reset, utils.Cyan, c.Proxy.Address, utils.Reset)
	}

	method := c.Proxy.Method
	if method == "" {
		method = "tls"
	}

	if method == "tls" && c.Proxy.TLS != nil {
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "SNI:", utils.Reset, utils.Cyan, c.Proxy.TLS.SNI, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Fingerprint:", utils.Reset, utils.Magenta, c.Proxy.TLS.Fingerprint, utils.Reset)
		alpnStr := "none"
		if len(c.Proxy.TLS.ALPN) > 0 {
			alpnStr = ""
			for i, a := range c.Proxy.TLS.ALPN {
				if i > 0 {
					alpnStr += ", "
				}
				alpnStr += a
			}
		}
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "ALPN:", utils.Reset, utils.White, alpnStr, utils.Reset)
	}
	fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Port:", utils.Reset, utils.White, c.Proxy.Port, utils.Reset)

	switch c.Proxy.Type {
	case "ws":
		if c.Proxy.WS != nil {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Host:", utils.Reset, utils.White, c.Proxy.WS.Host, utils.Reset)
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Path:", utils.Reset, utils.White, c.Proxy.WS.Path, utils.Reset)
		}
	case "xhttp":
		if c.Proxy.Xhttp != nil {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Host:", utils.Reset, utils.White, c.Proxy.Xhttp.Host, utils.Reset)
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Path:", utils.Reset, utils.White, c.Proxy.Xhttp.Path, utils.Reset)
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Mode:", utils.Reset, utils.White, c.Proxy.Xhttp.Mode, utils.Reset)
		}
	case "grpc":
		if c.Proxy.Grpc != nil {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Service Name:", utils.Reset, utils.White, c.Proxy.Grpc.ServiceName, utils.Reset)
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Authority:", utils.Reset, utils.White, c.Proxy.Grpc.Authority, utils.Reset)
			fmt.Printf("  %s%-18s%s %s%t%s\n", utils.Gray, "Multi Mode:", utils.Reset, utils.White, c.Proxy.Grpc.MultiMode, utils.Reset)
		}
	}

	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Network:", utils.Reset, utils.Green, c.Proxy.Type, utils.Reset)

	methodColor := utils.Green
	if method == "reality" {
		methodColor = utils.Magenta
	}
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Method:", utils.Reset, methodColor, method, utils.Reset)

	if method == "reality" && c.Proxy.Reality != nil {
		pubKey := c.Proxy.Reality.PublicKey
		if len(pubKey) > 16 {
			pubKey = pubKey[:16] + "..."
		}
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Public Key:", utils.Reset, utils.Dim, pubKey, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Short ID:", utils.Reset, utils.White, c.Proxy.Reality.ShortId, utils.Reset)
		if c.Proxy.Reality.ServerName != "" {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Server Name:", utils.Reset, utils.Cyan, c.Proxy.Reality.ServerName, utils.Reset)
		}
	}

	fmt.Printf("\n%s%s▸ Fragment Settings%s\n", utils.Bold, utils.Yellow, utils.Reset)

	modeColor := utils.White
	switch c.Fragment.Mode {
	case "auto":
		modeColor = utils.Green
	case "manual":
		modeColor = utils.Cyan
	case "off":
		modeColor = utils.Red
	}
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Mode:", utils.Reset, modeColor, c.Fragment.Mode, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Packets:", utils.Reset, utils.Magenta, c.Fragment.Packets, utils.Reset)

	if c.Fragment.Mode == "manual" || c.Fragment.Mode == "" {
		fmt.Printf("  %s%-18s%s %s%s%s bytes\n", utils.Gray, "Length:", utils.Reset, utils.Green, c.Fragment.Manual.Length, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%s%s ms\n", utils.Gray, "Interval:", utils.Reset, utils.Green, c.Fragment.Manual.Interval, utils.Reset)
	}

	if c.Fragment.Mode == "auto" {
		fmt.Printf("  %s%-18s%s %s%d-%d%s bytes\n", utils.Gray, "Length Range:", utils.Reset, utils.Green,
			c.Fragment.Auto.LengthRange.Min, c.Fragment.Auto.LengthRange.Max, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%d-%d%s ms\n", utils.Gray, "Interval Range:", utils.Reset, utils.Green,
			c.Fragment.Auto.IntervalRange.Min, c.Fragment.Auto.IntervalRange.Max, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Max Tests:", utils.Reset, utils.White, c.Fragment.Auto.MaxTests, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%.0f%%%s\n", utils.Gray, "Success Threshold:", utils.Reset, utils.White,
			c.Fragment.Auto.SuccessThreshold*100, utils.Reset)
		if c.Fragment.Auto.TestIP != "" {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Test IP:", utils.Reset, utils.Cyan, c.Fragment.Auto.TestIP, utils.Reset)
		} else {
			fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Test IP:", utils.Reset, utils.Dim, "(random)", utils.Reset)
		}
	}

	fmt.Printf("\n%s%s▸ Scan Settings%s\n", utils.Bold, utils.Yellow, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Threads:", utils.Reset, utils.Cyan, c.Scan.Threads, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%ds%s\n", utils.Gray, "Timeout:", utils.Reset, utils.White, c.Scan.Timeout, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%dms%s\n", utils.Gray, "Max Latency:", utils.Reset, utils.White, c.Scan.MaxLatency, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Retries:", utils.Reset, utils.White, c.Scan.Retries, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Test URL:", utils.Reset, utils.Dim, c.Scan.TestURL, utils.Reset)

	speedColor := utils.Red
	speedStatus := "disabled"
	if c.Scan.SpeedTest {
		speedColor = utils.Green
		speedStatus = "enabled"
	}
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Speed Test:", utils.Reset, speedColor, speedStatus, utils.Reset)
	if c.Scan.SpeedTest {
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Download URL:", utils.Reset, utils.Dim, c.Scan.DownloadURL, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Upload URL:", utils.Reset, utils.Dim, c.Scan.UploadURL, utils.Reset)
	}
	plCount := c.Scan.PacketLossCount
	if plCount <= 0 {
		plCount = 5
	}
	fmt.Printf("  %s%-18s%s %s%d pings%s\n", utils.Gray, "Packet Loss:", utils.Reset, utils.White, plCount, utils.Reset)

	jitterColor := utils.Red
	jitterStatus := "disabled"
	if c.Scan.JitterTest {
		jitterColor = utils.Green
		jitterStatus = "enabled"
	}
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Jitter Test:", utils.Reset, jitterColor, jitterStatus, utils.Reset)

	if c.Scan.StabilityRounds > 0 {
		fmt.Printf("\n%s%s▸ Phase-2 Stability%s\n", utils.Bold, utils.Yellow, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%d rounds%s\n", utils.Gray, "Rounds:", utils.Reset, utils.Green, c.Scan.StabilityRounds, utils.Reset)
		fmt.Printf("  %s%-18s%s %s%ds interval%s\n", utils.Gray, "Interval:", utils.Reset, utils.White, c.Scan.StabilityInterval, utils.Reset)
	}

	fmt.Printf("\n%s%s▸ Filters%s\n", utils.Bold, utils.Yellow, utils.Reset)
	if c.Scan.MaxPacketLossPct < 0 {
		fmt.Printf("  %s%-18s%s %sdisabled%s\n", utils.Gray, "Max Pkt Loss:", utils.Reset, utils.Dim, utils.Reset)
	} else {
		fmt.Printf("  %s%-18s%s %s%.0f%%%s\n", utils.Gray, "Max Pkt Loss:", utils.Reset, utils.Green, c.Scan.MaxPacketLossPct, utils.Reset)
	}
	if c.Scan.MinDownloadMbps > 0 {
		fmt.Printf("  %s%-18s%s %s%.1f Mbps%s\n", utils.Gray, "Min Download:", utils.Reset, utils.Green, c.Scan.MinDownloadMbps, utils.Reset)
	}
	if c.Scan.MinUploadMbps > 0 {
		fmt.Printf("  %s%-18s%s %s%.1f Mbps%s\n", utils.Gray, "Min Upload:", utils.Reset, utils.Green, c.Scan.MinUploadMbps, utils.Reset)
	}

	fmt.Printf("\n%s%s▸ Xray Settings%s\n", utils.Bold, utils.Yellow, utils.Reset)
	muxColor := utils.Red
	muxStatus := "disabled"
	if c.Xray.Mux.Enabled {
		muxColor = utils.Green
		muxStatus = fmt.Sprintf("enabled (concurrency: %d)", c.Xray.Mux.Concurrency)
	}
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Mux:", utils.Reset, muxColor, muxStatus, utils.Reset)

	fmt.Println()
}
