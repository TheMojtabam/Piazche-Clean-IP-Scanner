package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"piyazche/config"
	"piyazche/scanner"
	"piyazche/shodan"
	"piyazche/utils"
)

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
	Results      []scanner.Result
	Phase2Results []scanner.Phase2Result

	// history
	Sessions []ScanSession

	// current config
	CurrentConfig *config.Config
	ConfigRaw     string // JSON string of current config

	// scan control
	cancelFn    context.CancelFunc
	scannerRef  *scanner.Scanner
	CurrentIP   string // live IP being scanned

	// saved config (server-side persistent)
	SavedProxyConfig string // JSON of proxy part from parsed link
	SavedScanConfig  string // JSON of scan settings

	// TUI log lines broadcast to SSE clients
	TUILog []string
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
	state := &AppState{
		ScanStatus: "idle",
		Sessions:   []ScanSession{},
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
		QuickSettings string `json:"quickSettings"` // فقط override های سریع
		IPRanges      string `json:"ipRanges"`
		MaxIPs        int    `json:"maxIPs"`
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

	// بیلد config: saved config کامل + quick override
	cfg, err := s.buildMergedConfig(req.QuickSettings)
	if err != nil {
		jsonError(w, "config parse error: "+err.Error(), 400)
		return
	}

	go s.runScan(cfg, req.IPRanges, req.MaxIPs)
	jsonOK(w, "scan started")
}

func (s *Server) handleScanStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if s.state.cancelFn != nil {
		s.state.cancelFn()
		s.state.ScanStatus = "idle"
	}

	jsonOK(w, "stopped")
}

func (s *Server) handleScanPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if s.state.ScanStatus == "scanning" {
		s.state.ScanStatus = "paused"
		jsonOK(w, "paused")
	} else if s.state.ScanStatus == "paused" {
		s.state.ScanStatus = "scanning"
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
			s.hub.Broadcast("phase2_progress", map[string]interface{}{
				"ip":         r.IP,
				"done":       done2,
				"total":      total2,
				"passed":     r.Passed,
				"latency":    r.AvgLatencyMs,
				"jitter":     r.JitterMs,
				"loss":       r.PacketLossPct,
				"dl":         dlStr,
				"ul":         ulStr,
				"score":      r.StabilityScore,
				"failReason": r.FailReason,
			})
			icon := "✓"
			if !r.Passed {
				icon = "✗"
			}
			s.tuiLog(fmt.Sprintf("[%d/%d] %s %s  lat:%.0fms  loss:%.0f%%  ↓%s  ↑%s",
				done2, total2, icon, r.IP, r.AvgLatencyMs, r.PacketLossPct, dlStr, ulStr),
				map[bool]string{true: "ok", false: "err"}[r.Passed])
		}

		p2results := scanner.RunPhase2WithCallback(ctx, cfg, results, onP2Progress)

		s.state.mu.Lock()
		s.state.Phase2Results = p2results
		s.state.mu.Unlock()

		s.hub.Broadcast("phase2_done", map[string]interface{}{
			"results": p2results,
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
				"Total":     progress.Total,
				"Done":      progress.Done,
				"Succeeded": progress.Succeeded,
				"Failed":    progress.Failed,
				"Rate":      progress.Rate,
				"ETA":       progress.ETA,
				"CurrentIP": currentIP,
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
		ProxyConfig string `json:"proxyConfig"` // from parsed link
		ScanConfig  string `json:"scanConfig"`  // from settings UI
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
	s.state.mu.Unlock()
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
			cfg.Fragment = proxyCfg.Fragment
		}
	}

	// ۲. saved scan config کامل (از صفحه تنظیمات)
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
			if saved.Fragment != nil && proxyJSON == "" {
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
