package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"piyazche/config"
	"piyazche/scanner"
	"piyazche/shodan"
	"piyazche/utils"
	"piyazche/xray"
)

// ── Disk Persistence ─────────────────────────────────────────────────────────

type persistedState struct {
	ProxyConfig string `json:"proxyConfig"`
	ScanConfig  string `json:"scanConfig"`
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

func saveStateToDisk(proxyJSON, scanJSON string) {
	data, _ := json.MarshalIndent(persistedState{ProxyConfig: proxyJSON, ScanConfig: scanJSON}, "", "  ")
	os.WriteFile(configPersistPath(), data, 0644)
}

func loadStateFromDisk() (proxyJSON, scanJSON string) {
	data, err := os.ReadFile(configPersistPath())
	if err != nil {
		return "", ""
	}
	var ps persistedState
	if json.Unmarshal(data, &ps) != nil {
		return "", ""
	}
	return ps.ProxyConfig, ps.ScanConfig
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
	cancelFn    context.CancelFunc
	scannerRef  *scanner.Scanner
	CurrentIP   string

	// saved config
	SavedProxyConfig string
	SavedScanConfig  string

	// config templates
	Templates []config.ConfigTemplate

	// subnet intelligence
	SubnetStats []config.SubnetStat

	// health monitor
	HealthEntries map[string]*config.HealthEntry
	healthStop    chan struct{}
	healthOnce    sync.Once

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
	proxyJSON, scanJSON := loadStateFromDisk()

	state := &AppState{
		ScanStatus:       "idle",
		Sessions:         []ScanSession{},
		SavedProxyConfig: proxyJSON,
		SavedScanConfig:  scanJSON,
		Templates:        []config.ConfigTemplate{},
		HealthEntries:    make(map[string]*config.HealthEntry),
		healthStop:       make(chan struct{}),
	}

	hub := NewWSHub()

	mux := http.NewServeMux()
	s := &Server{port: port, state: state, hub: hub}
	s.registerRoutes(mux)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

// Start شروع HTTP server
func (s *Server) Start() error {
	go s.hub.Run()
	fmt.Printf("  Web UI: http://localhost:%d\n", s.port)
	return s.srv.ListenAndServe()
}

// Stop خاموش کردن server
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.srv.Shutdown(ctx)
}

// registerRoutes ثبت همه route ها
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static files (embed)
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/static/", s.handleStatic)

	// WebSocket
	mux.HandleFunc("/ws", s.hub.HandleWS)

	// API endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/scan/start", s.handleScanStart)
	mux.HandleFunc("/api/scan/stop", s.handleScanStop)
	mux.HandleFunc("/api/scan/pause", s.handleScanPause)
	mux.HandleFunc("/api/config/parse", s.handleConfigParse)
	mux.HandleFunc("/api/results", s.handleResults)
	mux.HandleFunc("/api/results/export", s.handleExport)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/shodan/harvest", s.handleShodanHarvest)
	mux.HandleFunc("/api/ips/expand", s.handleIPExpand)
	mux.HandleFunc("/api/config/save", s.handleConfigSave)
	mux.HandleFunc("/api/config/load", s.handleConfigLoad)
	mux.HandleFunc("/api/config/active", s.handleConfigActive)
	mux.HandleFunc("/api/tui/stream", s.handleTUIStream)
	// Templates
	mux.HandleFunc("/api/templates", s.handleTemplates)
	mux.HandleFunc("/api/templates/save", s.handleTemplateSave)
	mux.HandleFunc("/api/templates/delete", s.handleTemplateDelete)
	// Subnet stats
	mux.HandleFunc("/api/subnets", s.handleSubnets)
	// Health monitor
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/health/add", s.handleHealthAdd)
	mux.HandleFunc("/api/health/remove", s.handleHealthRemove)
	// Quick test (live config test)
	mux.HandleFunc("/api/quicktest", s.handleQuickTest)
}

// --- API Handlers ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   s.state.ScanStatus,
		"phase":    s.state.ScanPhase,
		"progress": s.state.Progress,
	})
}

func (s *Server) handleScanStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	var req struct {
		QuickSettings json.RawMessage `json:"quickSettings"` // فقط override های سریع
		IPRanges      string          `json:"ipRanges"`
		MaxIPs        int             `json:"maxIPs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request: "+err.Error(), 400)
		return
	}

	s.state.mu.Lock()
	if s.state.ScanStatus == "scanning" {
		s.state.mu.Unlock()
		jsonError(w, "scan already running", 409)
		return
	}
	s.state.mu.Unlock()

	// تبدیل quickSettings به string برای buildMergedConfig
	quickSettingsStr := ""
	if len(req.QuickSettings) > 0 && string(req.QuickSettings) != "null" {
		quickSettingsStr = string(req.QuickSettings)
	}

	// بیلد config: saved config کامل + quick override
	cfg, err := s.buildMergedConfig(quickSettingsStr)
	if err != nil {
		jsonError(w, "config parse error: "+err.Error(), 400)
		return
	}

	// IP count رو برای پاسخ سریع به JS محاسبه کنیم
	totalCount := 0
	if req.IPRanges != "" {
		sampleSize := cfg.Scan.SampleSize
		if sampleSize <= 0 { sampleSize = 1 }
		ips := parseIPInputWithSample(req.IPRanges, sampleSize)
		if req.MaxIPs > 0 && req.MaxIPs < len(ips) {
			ips = ips[:req.MaxIPs]
		}
		totalCount = len(ips)
	}

	go s.runScan(cfg, req.IPRanges, req.MaxIPs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "scan started",
		"total":   totalCount,
	})
}

func (s *Server) handleScanStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	s.state.mu.Lock()
	scnr := s.state.scannerRef
	cancelFn := s.state.cancelFn
	s.state.mu.Unlock()

	// اگه paused بود اول resume کن تا goroutineها آزاد بشن
	if scnr != nil && scnr.IsPaused() {
		scnr.Resume()
	}
	if scnr != nil {
		scnr.Stop()
	}
	if cancelFn != nil {
		cancelFn()
	}

	s.state.mu.Lock()
	s.state.ScanStatus = "idle"
	s.state.mu.Unlock()

	s.hub.Broadcast("status", map[string]string{"status": "idle", "phase": ""})
	jsonOK(w, "stopped")
}

func (s *Server) handleScanPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	s.state.mu.Lock()
	status := s.state.ScanStatus
	scnr := s.state.scannerRef
	s.state.mu.Unlock()

	if status == "scanning" && scnr != nil {
		scnr.Pause()
		s.state.mu.Lock()
		s.state.ScanStatus = "paused"
		s.state.mu.Unlock()
		s.hub.Broadcast("status", map[string]string{"status": "paused", "phase": s.state.ScanPhase})
		jsonOK(w, "paused")
	} else if status == "paused" && scnr != nil {
		scnr.Resume()
		s.state.mu.Lock()
		s.state.ScanStatus = "scanning"
		s.state.mu.Unlock()
		s.hub.Broadcast("status", map[string]string{"status": "scanning", "phase": s.state.ScanPhase})
		jsonOK(w, "resumed")
	} else {
		jsonError(w, "not scanning", 400)
	}
}

func (s *Server) handleConfigParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	var req struct {
		Input string `json:"input"` // vless:// vmess:// یا JSON
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}

	cfg, err := ParseProxyURL(req.Input)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}

	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	parsed := map[string]interface{}{
		"uuid":    cfg.Proxy.UUID,
		"address": cfg.Proxy.Address,
		"port":    cfg.Proxy.Port,
		"type":    cfg.Proxy.Type,
		"method":  cfg.Proxy.Method,
	}
	if cfg.Proxy.TLS != nil {
		parsed["sni"] = cfg.Proxy.TLS.SNI
		parsed["fp"] = cfg.Proxy.TLS.Fingerprint
		parsed["alpn"] = cfg.Proxy.TLS.ALPN
	}
	if cfg.Proxy.WS != nil {
		parsed["path"] = cfg.Proxy.WS.Path
		parsed["wsHost"] = cfg.Proxy.WS.Host
	}
	if cfg.Proxy.Grpc != nil {
		parsed["serviceName"] = cfg.Proxy.Grpc.ServiceName
	}
	if cfg.Proxy.Reality != nil {
		parsed["sni"] = cfg.Proxy.Reality.ServerName
		parsed["pbk"] = cfg.Proxy.Reality.PublicKey[:min16(len(cfg.Proxy.Reality.PublicKey))] + "..."
		parsed["sid"] = cfg.Proxy.Reality.ShortId
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"config": string(cfgJSON),
		"parsed": parsed,
	})
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"phase1":  s.state.Results,
		"phase2":  s.state.Phase2Results,
		"status":  s.state.ScanStatus,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()

	format := r.URL.Query().Get("format")

	if format == "txt" {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", `attachment; filename="ips.txt"`)
		for _, r := range s.state.Phase2Results {
			if r.Passed {
				fmt.Fprintf(w, "%s\n", r.IP)
			}
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="results.json"`)
	json.NewEncoder(w).Encode(s.state.Phase2Results)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()
	json.NewEncoder(w).Encode(s.state.Sessions)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// serve static assets embedded
	http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))).ServeHTTP(w, r)
}

// --- Scan Runner ---

func (s *Server) runScan(cfg *config.Config, ipRanges string, maxIPs int) {
	ctx, cancel := context.WithCancel(context.Background())

	s.state.mu.Lock()
	s.state.cancelFn = cancel
	s.state.ScanStatus = "scanning"
	s.state.ScanPhase = "phase1"
	s.state.Progress = ScanProgress{StartTime: time.Now()}
	s.state.Results = nil
	s.state.Phase2Results = nil
	s.state.mu.Unlock()

	s.hub.Broadcast("status", map[string]string{"status": "scanning", "phase": "phase1"})

	scnr := scanner.NewScannerWithDebug(cfg, false)
	s.tuiLog("▶ اسکن شروع شد — "+fmt.Sprintf("%d IP", scnr.IPCount()), "info")

	// Live IP tracking callback — fires when IP is dispatched to worker
	var lastResultCount int64
	scnr.OnIPStart = func(ip string) {
		s.state.mu.Lock()
		s.state.CurrentIP = ip
		s.state.mu.Unlock()
		s.hub.Broadcast("live_ip", map[string]string{"ip": ip})

		// also broadcast any newly completed results
		results := scnr.GetResults()
		count := int64(results.Count())
		if count > lastResultCount {
			all := results.All()
			for i := lastResultCount; i < count && i < int64(len(all)); i++ {
				r := all[i]
				s.hub.Broadcast("ip_result", map[string]interface{}{
					"ip":      r.IP,
					"success": r.Success,
					"latency": r.LatencyMs,
				})
			}
			lastResultCount = count
		}
	}

	// Load IPs from input
	if ipRanges != "" {
		sampleSize := cfg.Scan.SampleSize
		if sampleSize <= 0 { sampleSize = 1 }
		ips := parseIPInputWithSample(ipRanges, sampleSize)
		if maxIPs > 0 && maxIPs < len(ips) {
			ips = ips[:maxIPs]
		}
		if cfg.Scan.Shuffle {
			utils.ShuffleIPs(ips)
		}
		scnr.LoadIPsFromList(ips, 0, false)
	} else {
		scnr.LoadIPs("ipv4.txt", maxIPs, cfg.Scan.Shuffle)
	}

	s.state.mu.Lock()
	s.state.scannerRef = scnr
	s.state.Progress.Total = scnr.IPCount()
	s.state.mu.Unlock()

	// Progress broadcaster
	go s.broadcastProgress(ctx, scnr)

	_ = ctx // scanner uses its own context via Stop()

	s.tuiLog(fmt.Sprintf("⚡ Phase 1 — %d IP در صف اسکن", scnr.IPCount()), "info")
	if err := scnr.Run(); err != nil {
		s.hub.Broadcast("error", map[string]string{"message": err.Error()})
		s.tuiLog("✗ خطا: "+err.Error(), "err")
	}

	// Collect phase1 results
	results := scnr.GetResults().GetSuccessful()
	s.state.mu.Lock()
	s.state.ScanPhase = "phase2"
	for _, r := range scnr.GetResults().All() {
		s.state.Results = append(s.state.Results, r)
	}
	s.state.mu.Unlock()

	s.hub.Broadcast("phase2_start", map[string]int{"count": len(results)})
	s.tuiLog(fmt.Sprintf("🔬 Phase 2 شروع شد — %d IP", len(results)), "phase2")

	// Phase 2
	if len(results) > 0 && cfg.Scan.StabilityRounds > 0 {
		total2 := len(results)
		var done2 int
		p2StartTime := time.Now()

		// P2 progress state init
		s.state.mu.Lock()
		s.state.P2Progress = P2ScanProgress{Total: total2, StartTime: p2StartTime}
		s.state.mu.Unlock()

		onP2Progress := func(r scanner.Phase2Result) {
			done2++
			dlStr := "—"
			if r.DownloadMbps > 0 {
				dlStr = fmt.Sprintf("%.1fM", r.DownloadMbps)
			}
			ulStr := "—"
			if r.UploadMbps > 0 {
				ulStr = fmt.Sprintf("%.1fM", r.UploadMbps)
			}

			// ETA محاسبه
			elapsed := time.Since(p2StartTime).Seconds()
			rate2 := 0.0
			eta2 := "—"
			if elapsed > 0 && done2 > 0 {
				rate2 = float64(done2) / elapsed
				remaining := float64(total2-done2) / rate2
				if remaining > 0 {
					d := time.Duration(remaining) * time.Second
					if d < time.Minute {
						eta2 = fmt.Sprintf("%ds", int(d.Seconds()))
					} else {
						eta2 = fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
					}
				}
			}

			// Grade
			grade := scoreToGrade(r.StabilityScore)

			s.state.mu.Lock()
			s.state.P2Progress.Done = done2
			if r.Passed { s.state.P2Progress.Passed++ }
			s.state.P2Progress.Rate = rate2
			s.state.P2Progress.ETA = eta2
			s.state.mu.Unlock()

			pct2 := 0
			if total2 > 0 { pct2 = done2 * 100 / total2 }

			s.hub.Broadcast("phase2_progress", map[string]interface{}{
				"ip":         r.IP,
				"done":       done2,
				"total":      total2,
				"pct":        pct2,
				"passed":     r.Passed,
				"latency":    r.AvgLatencyMs,
				"jitter":     r.JitterMs,
				"loss":       r.PacketLossPct,
				"dl":         dlStr,
				"ul":         ulStr,
				"score":      r.StabilityScore,
				"grade":      grade,
				"failReason": r.FailReason,
				"eta":        eta2,
				"rate":       rate2,
			})
			icon := "✓"
			if !r.Passed {
				icon = "✗"
			}
			s.tuiLog(fmt.Sprintf("[%d/%d] %s %s  lat:%.0fms  loss:%.0f%%  score:%.0f(%s)  ↓%s  ↑%s",
				done2, total2, icon, r.IP, r.AvgLatencyMs, r.PacketLossPct, r.StabilityScore, grade, dlStr, ulStr),
				map[bool]string{true: "ok", false: "err"}[r.Passed])
		}

		p2results := scanner.RunPhase2WithCallback(ctx, cfg, results, onP2Progress)

		// Subnet stats از نتایج phase1
		subnetMap := map[string]*config.SubnetStat{}
		for _, r := range scnr.GetResults().All() {
			parts := strings.Split(r.IP, ".")
			if len(parts) >= 3 {
				subnet := parts[0]+"."+parts[1]+"."+parts[2]+".0/24"
				if _, ok := subnetMap[subnet]; !ok {
					subnetMap[subnet] = &config.SubnetStat{Subnet: subnet}
				}
				subnetMap[subnet].Total++
				if r.Success {
					subnetMap[subnet].Passed++
					subnetMap[subnet].AvgLatMs = (subnetMap[subnet].AvgLatMs*float64(subnetMap[subnet].Passed-1) + float64(r.LatencyMs)) / float64(subnetMap[subnet].Passed)
				}
			}
		}
		var subnetList []config.SubnetStat
		for _, v := range subnetMap {
			if v.Total > 0 {
				v.PassRate = float64(v.Passed) / float64(v.Total) * 100
			}
			subnetList = append(subnetList, *v)
		}
		// sort by pass rate desc
		for i := 1; i < len(subnetList); i++ {
			for j := i; j > 0 && subnetList[j].PassRate > subnetList[j-1].PassRate; j-- {
				subnetList[j], subnetList[j-1] = subnetList[j-1], subnetList[j]
			}
		}

		s.state.mu.Lock()
		s.state.Phase2Results = p2results
		s.state.SubnetStats = subnetList
		s.state.mu.Unlock()

		s.hub.Broadcast("phase2_done", map[string]interface{}{
			"results": p2results,
			"subnets": subnetList,
		})
	}

	// Save session
	s.state.mu.Lock()
	s.state.ScanStatus = "done"
	duration := time.Since(s.state.Progress.StartTime)
	passed := 0
	for _, r := range s.state.Phase2Results {
		if r.Passed {
			passed++
		}
	}
	session := ScanSession{
		ID:        fmt.Sprintf("%d", time.Now().Unix()),
		StartedAt: s.state.Progress.StartTime,
		Duration:  duration.Round(time.Second).String(),
		TotalIPs:  s.state.Progress.Total,
		Passed:    passed,
		Results:   s.state.Phase2Results,
	}
	s.state.Sessions = append([]ScanSession{session}, s.state.Sessions...)
	if len(s.state.Sessions) > 20 {
		s.state.Sessions = s.state.Sessions[:20]
	}
	s.state.mu.Unlock()

	s.hub.Broadcast("scan_done", map[string]interface{}{
		"duration": duration.Round(time.Second).String(),
		"passed":   passed,
	})
	s.tuiLog(fmt.Sprintf("✓ اسکن تموم شد — %d موفق — %s", passed, duration.Round(time.Second)), "ok")
}

func (s *Server) broadcastProgress(ctx context.Context, scnr *scanner.Scanner) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.state.mu.RLock()
			if s.state.ScanStatus == "done" || s.state.ScanStatus == "idle" {
				s.state.mu.RUnlock()
				return
			}
			s.state.mu.RUnlock()

			stats := scnr.GetResults()
			done := stats.Count()
			succeeded := stats.SuccessCount()

			s.state.mu.Lock()
			s.state.Progress.Done = done
			s.state.Progress.Succeeded = succeeded
			s.state.Progress.Failed = done - succeeded
			elapsed := time.Since(s.state.Progress.StartTime).Seconds()
			if elapsed > 0 {
				s.state.Progress.Rate = float64(done) / elapsed
				remaining := s.state.Progress.Total - done
				if s.state.Progress.Rate > 0 {
					eta := time.Duration(float64(remaining)/s.state.Progress.Rate) * time.Second
					s.state.Progress.ETA = eta.Round(time.Second).String()
				}
			}
			progress := s.state.Progress
			s.state.mu.Unlock()

			s.state.mu.RLock()
			currentIP := s.state.CurrentIP
			s.state.mu.RUnlock()

			s.hub.Broadcast("progress", map[string]interface{}{
				"total":     progress.Total,
				"done":      progress.Done,
				"succeeded": progress.Succeeded,
				"failed":    progress.Failed,
				"rate":      progress.Rate,
				"eta":       progress.ETA,
				"currentIP": currentIP,
			})
		}
	}
}

// --- IP Expand Handler ---
// میزان IP هایی که از CIDR expand میشه رو قبل از اسکن نشون میده

func (s *Server) handleIPExpand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		IPRanges string `json:"ipRanges"`
		MaxIPs   int    `json:"maxIPs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	ips := parseIPInput(req.IPRanges)
	if req.MaxIPs > 0 && len(ips) > req.MaxIPs {
		ips = ips[:req.MaxIPs]
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(ips),
		"preview": func() []string {
			if len(ips) > 5 { return ips[:5] }
			return ips
		}(),
	})
}

// --- Shodan Handler ---

func (s *Server) handleShodanHarvest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	var req struct {
		APIKey      string `json:"apiKey"`
		Query       string `json:"query"`
		Pages       int    `json:"pages"`
		ExcludeCF   bool   `json:"excludeCF"`
		AutoScan    bool   `json:"autoScan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request: "+err.Error(), 400)
		return
	}
	if req.APIKey == "" {
		jsonError(w, "apiKey is required", 400)
		return
	}
	if req.Pages <= 0 {
		req.Pages = 1
	}

	go func() {
		s.hub.Broadcast("shodan_status", map[string]string{"status": "harvesting"})

		cfg := shodan.HarvestConfig{
			APIKey:          req.APIKey,
			Query:           req.Query,
			UseDefaultQuery: req.Query == "",
			Pages:           req.Pages,
			ExcludeCFRanges: req.ExcludeCF,
		}
		h := shodan.NewHarvester(cfg)
		result, err := h.Harvest(context.Background())
		if err != nil {
			s.hub.Broadcast("shodan_error", map[string]string{"message": err.Error()})
			return
		}

		s.hub.Broadcast("shodan_done", map[string]interface{}{
			"ips":   result.IPs,
			"total": result.TotalFound,
			"count": len(result.IPs),
		})

		if req.AutoScan && len(result.IPs) > 0 {
			s.state.mu.RLock()
			cfg2 := s.state.CurrentConfig
			s.state.mu.RUnlock()
			if cfg2 == nil {
				cfg2 = config.DefaultConfig()
			}
			go s.runScan(cfg2, joinLines(result.IPs), 0)
		}
	}()

	jsonOK(w, "harvest started")
}

func min16(n int) int {
	if n < 16 {
		return n
	}
	return 16
}

func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

// --- Helpers ---

func jsonOK(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": msg})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": msg})
}

func parseConfigFromJSON(raw string) (*config.Config, error) {
	if raw == "" {
		return config.DefaultConfig(), nil
	}
	cfg := config.DefaultConfig()
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parseIPInput(input string) []string {
	return parseIPInputWithSample(input, 1)
}

func parseIPInputWithSample(input string, sampleSize int) []string {
	var ips []string
	if sampleSize <= 0 {
		sampleSize = 1
	}
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "/") {
			// CIDR — expand with sampleSize
			expanded, err := utils.ExpandCIDR(line, sampleSize)
			if err != nil {
				fmt.Printf("warning: invalid CIDR %s: %v\n", line, err)
				continue
			}
			ips = append(ips, expanded...)
		} else if net.ParseIP(line) != nil {
			ips = append(ips, line)
		}
	}
	return ips
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// --- Config Save/Load Handlers ---

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		ProxyConfig string `json:"proxyConfig"`
		ScanConfig  string `json:"scanConfig"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	s.state.mu.Lock()
	if req.ProxyConfig != "" {
		s.state.SavedProxyConfig = req.ProxyConfig
	}
	if req.ScanConfig != "" {
		s.state.SavedScanConfig = req.ScanConfig
	}
	proxyJSON := s.state.SavedProxyConfig
	scanJSON := s.state.SavedScanConfig
	s.state.mu.Unlock()

	// Persist to disk so config survives restarts
	go saveStateToDisk(proxyJSON, scanJSON)

	jsonOK(w, "saved")
}

func (s *Server) handleConfigActive(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.buildMergedConfig("")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"threads":         cfg.Scan.Threads,
		"timeout":         cfg.Scan.Timeout,
		"maxLatency":      cfg.Scan.MaxLatency,
		"stabilityRounds": cfg.Scan.StabilityRounds,
		"stabilityInterval": cfg.Scan.StabilityInterval,
		"packetLossCount": cfg.Scan.PacketLossCount,
		"speedTest":       cfg.Scan.SpeedTest,
		"jitterTest":      cfg.Scan.JitterTest,
		"maxPacketLossPct": cfg.Scan.MaxPacketLossPct,
		"minDownloadMbps": cfg.Scan.MinDownloadMbps,
		"minUploadMbps":   cfg.Scan.MinUploadMbps,
		"testUrl":         cfg.Scan.TestURL,
		"fragmentMode":    cfg.Fragment.Mode,
		"proxy":           cfg.Proxy.Address,
	})
}

func (s *Server) handleConfigLoad(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"proxyConfig": s.state.SavedProxyConfig,
		"scanConfig":  s.state.SavedScanConfig,
		"hasProxy":    s.state.SavedProxyConfig != "",
	})
}

// buildMergedConfig — saved config کامل رو لود میکنه + quick override اعمال میکنه
func (s *Server) buildMergedConfig(quickOverrideJSON string) (*config.Config, error) {
	cfg := config.DefaultConfig()

	s.state.mu.RLock()
	proxyJSON := s.state.SavedProxyConfig
	scanJSON := s.state.SavedScanConfig
	s.state.mu.RUnlock()

	// ۱. proxy config از لینک parse شده
	if proxyJSON != "" {
		var proxyCfg config.Config
		if err := json.Unmarshal([]byte(proxyJSON), &proxyCfg); err == nil {
			cfg.Proxy = proxyCfg.Proxy
			// Fragment from proxy link is a fallback only; settings UI overrides it below
			if proxyCfg.Fragment.Mode != "" {
				cfg.Fragment = proxyCfg.Fragment
			}
		}
	}

	// ۲. saved scan config (from settings UI) — always overrides proxy fragment
	if scanJSON != "" {
		var saved struct {
			Scan     *config.ScanConfig     `json:"scan"`
			Fragment *config.FragmentConfig `json:"fragment"`
			Xray     *config.XrayConfig     `json:"xray"`
			Shodan   *config.ShodanConfig   `json:"shodan"`
		}
		if err := json.Unmarshal([]byte(scanJSON), &saved); err == nil {
			if saved.Scan != nil {
				cfg.Scan = *saved.Scan
			}
			if saved.Fragment != nil {
				// Settings UI fragment ALWAYS wins, even when proxy link is set
				cfg.Fragment = *saved.Fragment
			}
			if saved.Xray != nil {
				cfg.Xray = *saved.Xray
			}
			if saved.Shodan != nil {
				cfg.Shodan = *saved.Shodan
			}
		}
	}

	// ۳. quick override از دکمه شروع (فقط فیلدهایی که صریحاً داده شدن)
	if quickOverrideJSON != "" {
		var q struct {
			Threads         *int     `json:"threads"`
			Timeout         *int     `json:"timeout"`
			MaxLatency      *int     `json:"maxLatency"`
			StabilityRounds *int     `json:"stabilityRounds"`
			SampleSize      *int     `json:"sampleSize"`
		}
		if err := json.Unmarshal([]byte(quickOverrideJSON), &q); err == nil {
			if q.Threads != nil && *q.Threads > 0 { cfg.Scan.Threads = *q.Threads }
			if q.Timeout != nil && *q.Timeout > 0 { cfg.Scan.Timeout = *q.Timeout }
			if q.MaxLatency != nil && *q.MaxLatency > 0 { cfg.Scan.MaxLatency = *q.MaxLatency }
			if q.StabilityRounds != nil { cfg.Scan.StabilityRounds = *q.StabilityRounds }
			if q.SampleSize != nil && *q.SampleSize > 0 { cfg.Scan.SampleSize = *q.SampleSize }
		}
	}

	return cfg, nil
}

// --- TUI SSE Stream ---

func (s *Server) tuiLog(msg string, level string) {
	line := fmt.Sprintf(`{"t":"%s","l":"%s","m":%s}`,
		time.Now().Format("15:04:05"),
		level,
		jsonStr(msg),
	)
	s.state.mu.Lock()
	s.state.TUILog = append(s.state.TUILog, line)
	if len(s.state.TUILog) > 500 {
		s.state.TUILog = s.state.TUILog[len(s.state.TUILog)-500:]
	}
	s.state.mu.Unlock()
	s.hub.Broadcast("tui", map[string]string{"t": time.Now().Format("15:04:05"), "l": level, "m": msg})
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *Server) handleTUIStream(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	logs := make([]string, len(s.state.TUILog))
	copy(logs, s.state.TUILog)
	s.state.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lines": logs})
}

// ── scoreToGrade ──────────────────────────────────────────────────────────────
func scoreToGrade(score float64) string {
	switch {
	case score >= 92:
		return "A+"
	case score >= 85:
		return "A"
	case score >= 75:
		return "B+"
	case score >= 65:
		return "B"
	case score >= 55:
		return "C+"
	case score >= 45:
		return "C"
	case score >= 35:
		return "D"
	default:
		return "F"
	}
}

// ── Config Templates ──────────────────────────────────────────────────────────

func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"templates": s.state.Templates})
}

func (s *Server) handleTemplateSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		Name   string `json:"name"`
		RawURL string `json:"rawUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.RawURL == "" {
		jsonError(w, "name and rawUrl required", 400)
		return
	}
	cfg, err := ParseProxyURL(req.RawURL)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	tmpl := config.ConfigTemplate{
		ID:         fmt.Sprintf("%d", time.Now().UnixMilli()),
		Name:       req.Name,
		RawURL:     req.RawURL,
		ConfigJSON: string(cfgJSON),
		CreatedAt:  time.Now().Unix(),
	}
	s.state.mu.Lock()
	s.state.Templates = append(s.state.Templates, tmpl)
	s.state.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func (s *Server) handleTemplateDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct{ ID string `json:"id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	s.state.mu.Lock()
	out := s.state.Templates[:0]
	for _, t := range s.state.Templates {
		if t.ID != req.ID {
			out = append(out, t)
		}
	}
	s.state.Templates = out
	s.state.mu.Unlock()
	jsonOK(w, "deleted")
}

// ── Subnet Intelligence ───────────────────────────────────────────────────────

func (s *Server) handleSubnets(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"subnets": s.state.SubnetStats})
}

// ── Health Monitor ────────────────────────────────────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.state.mu.RLock()
	entries := make([]*config.HealthEntry, 0, len(s.state.HealthEntries))
	for _, e := range s.state.HealthEntries {
		entries = append(entries, e)
	}
	s.state.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": entries})
}

func (s *Server) handleHealthAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		IP            string  `json:"ip"`
		BaseLatencyMs float64 `json:"baseLatencyMs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IP == "" {
		jsonError(w, "ip required", 400)
		return
	}
	s.state.mu.Lock()
	if _, exists := s.state.HealthEntries[req.IP]; !exists {
		s.state.HealthEntries[req.IP] = &config.HealthEntry{
			IP:            req.IP,
			Status:        config.HealthUnknown,
			BaseLatencyMs: req.BaseLatencyMs,
			LastCheck:     time.Now().UnixMilli(),
		}
	}
	s.state.mu.Unlock()
	// Health monitor loop شروع کن
	s.startHealthMonitor()
	jsonOK(w, "added")
}

func (s *Server) handleHealthRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct{ IP string `json:"ip"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	s.state.mu.Lock()
	delete(s.state.HealthEntries, req.IP)
	s.state.mu.Unlock()
	jsonOK(w, "removed")
}

// startHealthMonitor یه goroutine شروع می‌کنه که هر ۳ دقیقه IP ها رو ping می‌کنه
func (s *Server) startHealthMonitor() {
	s.state.healthOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(3 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-s.state.healthStop:
					return
				case <-ticker.C:
					s.runHealthChecks()
				}
			}
		}()
	})
}

func (s *Server) runHealthChecks() {
	s.state.mu.RLock()
	cfg := s.state.CurrentConfig
	entries := make(map[string]*config.HealthEntry)
	for k, v := range s.state.HealthEntries {
		cp := *v
		entries[k] = &cp
	}
	s.state.mu.RUnlock()

	if cfg == nil {
		return
	}

	for ip, entry := range entries {
		func(ip string, e *config.HealthEntry) {
			port := utils.AcquirePort()
			defer utils.ReleasePort(port)

			cfgCopy := *cfg
			cfgCopy.Xray.LogLevel = "none"
			xrayCfg, err := config.GenerateXrayConfig(&cfgCopy, ip, port)
			if err != nil {
				return
			}
			mgr := xray.NewManagerWithDebug(false)
			if err := mgr.Start(xrayCfg, port); err != nil {
				return
			}
			defer mgr.Stop()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := mgr.WaitForReadyWithContext(ctx, 8*time.Second); err != nil {
				return
			}

			testCtx, testCancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer testCancel()
			result := xray.TestConnectivityWithContext(testCtx, port, cfg.Scan.TestURL, 7*time.Second)

			s.state.mu.Lock()
			if he, ok := s.state.HealthEntries[ip]; ok {
				he.TotalChecks++
				he.LastCheck = time.Now().UnixMilli()
				if result.Success {
					prevStatus := he.Status
					he.TotalAlive++
					he.LatencyMs = float64(result.Latency.Milliseconds())
					he.LastSeen = time.Now().UnixMilli()
					he.ConsecFails = 0
					if prevStatus == config.HealthDead || prevStatus == config.HealthUnknown {
						he.Status = config.HealthRecovered
					} else {
						he.Status = config.HealthAlive
					}
				} else {
					he.ConsecFails++
					if he.ConsecFails >= 3 {
						he.Status = config.HealthDead
					}
				}
				if he.TotalChecks > 0 {
					he.UptimePct = float64(he.TotalAlive) / float64(he.TotalChecks) * 100
				}
				s.hub.Broadcast("health_update", map[string]interface{}{
					"ip":        ip,
					"status":    string(he.Status),
					"latencyMs": he.LatencyMs,
					"uptimePct": he.UptimePct,
					"lastCheck": he.LastCheck,
				})
			}
			s.state.mu.Unlock()
		}(ip, entry)
	}
}

// ── Quick Test (live config test) ─────────────────────────────────────────────

func (s *Server) handleQuickTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		RawURL string `json:"rawUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RawURL == "" {
		jsonError(w, "rawUrl required", 400)
		return
	}
	cfg, err := ParseProxyURL(req.RawURL)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}

	port := utils.AcquirePort()
	defer utils.ReleasePort(port)

	cfgCopy := *cfg
	cfgCopy.Xray.LogLevel = "none"
	xrayCfg, err := config.GenerateXrayConfig(&cfgCopy, "", port)
	if err != nil {
		jsonError(w, "config error: "+err.Error(), 500)
		return
	}
	mgr := xray.NewManagerWithDebug(false)
	if err := mgr.Start(xrayCfg, port); err != nil {
		jsonError(w, "xray start failed: "+err.Error(), 500)
		return
	}
	defer mgr.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	if err := mgr.WaitForReadyWithContext(ctx, 10*time.Second); err != nil {
		jsonError(w, "xray not ready", 500)
		return
	}

	testURL := cfg.Scan.TestURL
	if testURL == "" {
		testURL = "https://www.gstatic.com/generate_204"
	}
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()
	result := xray.TestConnectivityWithContext(testCtx, port, testURL, 9*time.Second)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   result.Success,
		"latencyMs": result.Latency.Milliseconds(),
		"error":     fmt.Sprintf("%v", result.Error),
	})
}
