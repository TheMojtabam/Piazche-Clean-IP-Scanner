package webui

import (
	_ "embed"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"piyazche/config"
	"piyazche/scanner"
)

// ── Disk Persistence ─────────────────────────────────────────────────────────

type persistedState struct {
	ProxyConfig          string                         `json:"proxyConfig"`
	ScanConfig           string                         `json:"scanConfig"`
	Templates            []config.ConfigTemplate        `json:"templates,omitempty"`
	RawURL               string                         `json:"rawUrl,omitempty"`
	HealthEntries        map[string]*config.HealthEntry `json:"healthEntries,omitempty"`
	HealthEnabled        *bool                          `json:"healthEnabled,omitempty"`
	HealthIntervalMins   *int                           `json:"healthIntervalMins,omitempty"`
	TrafficDetectEnabled *bool                          `json:"trafficDetect,omitempty"`
	Sessions             []ScanSession                  `json:"sessions,omitempty"`
	SavedRanges          string                         `json:"savedRanges,omitempty"`
	SubnetStats          []config.SubnetStat            `json:"subnetStats,omitempty"`
}

// configPersistPath returns the path for UI config.
// Checks ./piyazche_ui.json first, falls back to ~/.piyazche/ui.json
func configPersistPath() string {
	local := "piyazche_ui.json"
	// Try to create a temp file to check writability
	if f, err := os.OpenFile(local, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		f.Close()
		return local
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return local
	}
	dir := filepath.Join(home, ".piyazche")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "ui.json")
}

func saveStateToDisk(proxyJSON, scanJSON, rawURL string, templates []config.ConfigTemplate, healthEntries map[string]*config.HealthEntry, healthEnabled bool, healthIntervalMins int, trafficDetect bool, sessions []ScanSession, savedRanges string, subnetStats []config.SubnetStat) {
	// HealthEntries رو deep copy کن قبل از persist
	heCopy := make(map[string]*config.HealthEntry, len(healthEntries))
	for k, v := range healthEntries {
		cp := *v
		heCopy[k] = &cp
	}
	data, _ := json.MarshalIndent(persistedState{
		ProxyConfig:          proxyJSON,
		ScanConfig:           scanJSON,
		Templates:            templates,
		RawURL:               rawURL,
		HealthEntries:        heCopy,
		HealthEnabled:        &healthEnabled,
		HealthIntervalMins:   &healthIntervalMins,
		TrafficDetectEnabled: &trafficDetect,
		Sessions:             sessions,
		SavedRanges:          savedRanges,
		SubnetStats:          subnetStats,
	}, "", "  ")
	os.WriteFile(configPersistPath(), data, 0644)
}

// saveStateToDiskFromServer — helper که state رو از server می‌خونه و ذخیره میکنه
func (s *Server) saveStateToDiskNow() {
	s.state.mu.RLock()
	proxyJSON := s.state.SavedProxyConfig
	scanJSON := s.state.SavedScanConfig
	rawURL := s.state.SavedRawURL
	templates := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates, s.state.Templates)
	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	heCopy := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries {
		cp := *v
		heCopy[k] = &cp
	}
	sessions := make([]ScanSession, len(s.state.Sessions))
	copy(sessions, s.state.Sessions)
	savedRanges := s.state.SavedRanges
	subnetStats := make([]config.SubnetStat, len(s.state.SubnetStats))
	copy(subnetStats, s.state.SubnetStats)
	s.state.mu.RUnlock()
	saveStateToDisk(proxyJSON, scanJSON, rawURL, templates, heCopy, healthEnabled, healthIntervalMins, trafficDetect, sessions, savedRanges, subnetStats)
}

func loadStateFromDisk() (proxyJSON, scanJSON, rawURL string, templates []config.ConfigTemplate, healthEntries map[string]*config.HealthEntry, healthEnabled *bool, healthIntervalMins *int, trafficDetect *bool, sessions []ScanSession, savedRanges string, subnetStats []config.SubnetStat) {
	data, err := os.ReadFile(configPersistPath())
	if err != nil {
		return "", "", "", nil, nil, nil, nil, nil, nil, "", nil
	}
	var ps persistedState
	if json.Unmarshal(data, &ps) != nil {
		return "", "", "", nil, nil, nil, nil, nil, nil, "", nil
	}
	return ps.ProxyConfig, ps.ScanConfig, ps.RawURL, ps.Templates, ps.HealthEntries, ps.HealthEnabled, ps.HealthIntervalMins, ps.TrafficDetectEnabled, ps.Sessions, ps.SavedRanges, ps.SubnetStats
}

// ── Server ────────────────────────────────────────────────────────────────────

// Server — Web UI HTTP server
type Server struct {
	port    int
	state   *AppState
	hub     *WSHub
	srv     *http.Server
	mu      sync.Mutex
}

// AppState وضعیت کلی app — اینجا همه چیز نگه داشته میشه
type AppState struct {
	mu sync.RWMutex

	// scan state
	ScanStatus   string          // "idle", "scanning", "paused", "done"
	ScanPhase    string          // "phase1", "phase2"
	Progress     ScanProgress
	P2Progress   P2ScanProgress  // وضعیت فاز ۲
	Results      []scanner.Result
	Phase2Results []scanner.Phase2Result

	// history
	Sessions []ScanSession

	// current config
	CurrentConfig *config.Config
	ConfigRaw     string

	// scan control
	cancelFn      context.CancelFunc
	phase2CancelFn context.CancelFunc // جداگانه برای فاز 2
	scannerRef    *scanner.Scanner
	CurrentIP     string

	// saved config
	SavedProxyConfig string
	SavedScanConfig  string
	SavedRawURL      string // لینک اصلی (vless:// vmess:// trojan://) برای copy-with-IP
	SavedRanges      string // IP ranges ذخیره شده

	// config templates
	Templates []config.ConfigTemplate

	// subnet intelligence
	SubnetStats []config.SubnetStat

	// health monitor
	HealthEntries        map[string]*config.HealthEntry
	healthStop           chan struct{}
	healthOnce           sync.Once
	healthTicker         *time.Ticker
	HealthIntervalMins   int  // default: 3
	HealthEnabled        bool // مانیتور فعال/غیرفعال
	TrafficDetectEnabled bool // تشخیص ترافیک بدون speed test

	// TUI log
	TUILog []string
}

// P2ScanProgress پروگرس فاز ۲
type P2ScanProgress struct {
	Total     int
	Done      int
	Passed    int
	StartTime time.Time
	ETA       string
	Rate      float64
}

type ScanProgress struct {
	Total     int
	Done      int
	Succeeded int
	Failed    int
	Rate      float64 // IPs/sec
	StartTime time.Time
	ETA       string
}

type ScanSession struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"startedAt"`
	Duration  string    `json:"duration"`
	TotalIPs  int       `json:"totalIPs"`
	Passed    int       `json:"passed"`
	Config    string    `json:"config"` // config name
	Results   []scanner.Phase2Result `json:"results"`
}

// NewServer یه server جدید می‌سازه
func NewServer(port int) *Server {
	// Load persisted UI config from disk
	proxyJSON, scanJSON, rawURL, savedTemplates, savedHealthEntries, savedHealthEnabled, savedHealthInterval, savedTrafficDetect, savedSessions, savedRanges, savedSubnetStats := loadStateFromDisk()

	if savedTemplates == nil {
		savedTemplates = []config.ConfigTemplate{}
	}
	if savedHealthEntries == nil {
		savedHealthEntries = make(map[string]*config.HealthEntry)
	}
	if savedSessions == nil {
		savedSessions = []ScanSession{}
	}
	if savedSubnetStats == nil {
		savedSubnetStats = []config.SubnetStat{}
	}

	// مقادیر پیش‌فرض monitor — بعد از لود از دیسک override میشن
	healthEnabled := true
	healthIntervalMins := 3
	trafficDetect := false
	if savedHealthEnabled != nil {
		healthEnabled = *savedHealthEnabled
	}
	if savedHealthInterval != nil && *savedHealthInterval > 0 {
		healthIntervalMins = *savedHealthInterval
	}
	if savedTrafficDetect != nil {
		trafficDetect = *savedTrafficDetect
	}

	state := &AppState{
		ScanStatus:           "idle",
		Sessions:             savedSessions,
		SavedProxyConfig:     proxyJSON,
		SavedScanConfig:      scanJSON,
		SavedRawURL:          rawURL,
		SavedRanges:          savedRanges,
		Templates:            savedTemplates,
		HealthEntries:        savedHealthEntries,
		healthStop:           make(chan struct{}),
		HealthIntervalMins:   healthIntervalMins,
		HealthEnabled:        healthEnabled,
		TrafficDetectEnabled: trafficDetect,
		SubnetStats:          savedSubnetStats,
	}

	hub := NewWSHub()

	mux := http.NewServeMux()
	s := &Server{port: port, state: state, hub: hub}
	s.registerRoutes(mux)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: mux,
	}

	return s
}

//go:embed assets/index.html
var indexHTMLContent string
