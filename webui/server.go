package webui

import (
	"context"
	"encoding/base64"
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
	"piyazche/optimizer"
	"piyazche/scanner"
	"piyazche/shodan"
	"piyazche/utils"
	"piyazche/xray"
)

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
	mux.HandleFunc("/api/config/build-link", s.handleBuildLink)
	mux.HandleFunc("/api/config/multi-parse", s.handleMultiParse)
	mux.HandleFunc("/api/results", s.handleResults)
	mux.HandleFunc("/api/results/export", s.handleExport)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/shodan/harvest", s.handleShodanHarvest)
	mux.HandleFunc("/api/ips/expand", s.handleIPExpand)
	mux.HandleFunc("/api/config/save", s.handleConfigSave)
	mux.HandleFunc("/api/config/load", s.handleConfigLoad)
	mux.HandleFunc("/api/geoip", s.handleGeoIP)
	mux.HandleFunc("/api/config/active", s.handleConfigActive)
	mux.HandleFunc("/api/tui/stream", s.handleTUIStream)
	// Templates
	mux.HandleFunc("/api/templates", s.handleTemplates)
	mux.HandleFunc("/api/templates/save", s.handleTemplateSave)
	mux.HandleFunc("/api/templates/delete", s.handleTemplateDelete)
	// Subnet stats
	mux.HandleFunc("/api/subnets", s.handleSubnets)
	// Subscription import
	mux.HandleFunc("/api/subscription/fetch", s.handleSubscriptionFetch)
	// Sessions persistence
	mux.HandleFunc("/api/sessions/save", s.handleSessionsSave)
	// Health monitor
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/health/add", s.handleHealthAdd)
	mux.HandleFunc("/api/health/remove", s.handleHealthRemove)
	mux.HandleFunc("/api/health/settings", s.handleHealthSettings)
	mux.HandleFunc("/api/health/check-now", s.handleHealthCheckNow)
	// Phase 3 speed test
	mux.HandleFunc("/api/phase3/run", s.handlePhase3Run)
	mux.HandleFunc("/api/fragment/auto", s.handleFragmentAuto)
	// Quick test (live config test)
	mux.HandleFunc("/api/quicktest", s.handleQuickTest)
	// System info
	mux.HandleFunc("/api/sysinfo", s.handleSysInfo)
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
	phase2Cancel := s.state.phase2CancelFn
	s.state.mu.Unlock()

	// اگه paused بود اول resume کن تا goroutineها آزاد بشن
	if scnr != nil && scnr.IsPaused() {
		scnr.Resume()
	}
	if scnr != nil {
		scnr.Stop()
	}
	if phase2Cancel != nil {
		phase2Cancel()
	}
	if cancelFn != nil {
		cancelFn()
	}

	s.state.mu.Lock()
	s.state.ScanStatus = "idle"
	s.state.ScanPhase = ""
	s.state.scannerRef = nil
	s.state.phase2CancelFn = nil
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
	phase := s.state.ScanPhase
	scnr := s.state.scannerRef
	s.state.mu.Unlock()

	if status == "scanning" && phase == "phase1" && scnr != nil {
		scnr.Pause()
		s.state.mu.Lock()
		s.state.ScanStatus = "paused"
		s.state.mu.Unlock()
		s.hub.Broadcast("status", map[string]string{"status": "paused", "phase": s.state.ScanPhase})
		jsonOK(w, "paused")
	} else if status == "paused" && phase == "phase1" && scnr != nil {
		scnr.Resume()
		s.state.mu.Lock()
		s.state.ScanStatus = "scanning"
		s.state.mu.Unlock()
		s.hub.Broadcast("status", map[string]string{"status": "scanning", "phase": s.state.ScanPhase})
		jsonOK(w, "resumed")
	} else if status == "scanning" && phase == "phase2" {
		// فاز 2 pause ندارد — stop میشه
		jsonError(w, "phase2 cannot be paused, use stop", 400)
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

	// rawURL رو ذخیره کن برای build-link بعداً
	input := strings.TrimSpace(req.Input)
	if strings.HasPrefix(input, "vless://") || strings.HasPrefix(input, "vmess://") || strings.HasPrefix(input, "trojan://") {
		s.state.mu.Lock()
		s.state.SavedRawURL = input
		rawURL := input
		proxyJSON := s.state.SavedProxyConfig
		_ = proxyJSON
		scanJSON := s.state.SavedScanConfig
		_ = rawURL
		_ = proxyJSON
		_ = scanJSON
		templates := make([]config.ConfigTemplate, len(s.state.Templates))
		copy(templates, s.state.Templates)
		_ = templates
		s.state.mu.Unlock()
		heCopyForSave := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
		for k, v := range s.state.HealthEntries { cp := *v; heCopyForSave[k] = &cp }
		_ = heCopyForSave
		healthEnabled := s.state.HealthEnabled
		healthIntervalMins := s.state.HealthIntervalMins
		trafficDetect := s.state.TrafficDetectEnabled
		_ = healthEnabled
		_ = healthIntervalMins
		_ = trafficDetect
		go s.saveStateToDiskNow()
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

// handleBuildLink یه IP رو با config ذخیره‌شده ترکیب می‌کنه و لینک میده
func (s *Server) handleBuildLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IP == "" {
		jsonError(w, "ip required", 400)
		return
	}

	s.state.mu.RLock()
	rawURL := s.state.SavedRawURL
	_ = rawURL
	s.state.mu.RUnlock()

	if rawURL == "" {
		jsonError(w, "no raw proxy link saved — import a vless/vmess/trojan link first", 400)
		return
	}

	cfg, err := ParseProxyURL(rawURL)
	if err != nil {
		jsonError(w, "could not re-parse saved link: "+err.Error(), 500)
		return
	}

	link, err := BuildProxyURL(cfg, req.IP, rawURL)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"link": link, "ip": req.IP})
}

// handleMultiParse چند لینک رو یکجا parse می‌کنه
func (s *Server) handleMultiParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	results, errs := ParseMultiProxy(req.Input)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"results": results,
		"errors":  errs,
		"count":   len(results),
	})
}

// handleGeoIP اطلاعات جغرافیایی یه IP رو از طریق proxy برمیگردونه
func (s *Server) handleGeoIP(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		jsonError(w, "ip required", 400)
		return
	}

	// اگه قبلاً fetch شده از cache برگردون
	s.state.mu.RLock()
	if entry, ok := s.state.HealthEntries[ip]; ok && entry.GeoInfo != nil {
		geo := entry.GeoInfo
		s.state.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(geo)
		return
	}
	s.state.mu.RUnlock()

	// بدون proxy مستقیم fetch کن (GeoIP API ها معمولاً block نیستن)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/" + ip + "?fields=country,countryCode,city,isp,as")
	if err != nil {
		jsonError(w, "geoip fetch failed: "+err.Error(), 502)
		return
	}
	defer resp.Body.Close()

	var apiResp struct {
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
		AS          string `json:"as"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		jsonError(w, "geoip parse failed", 502)
		return
	}

	geo := &config.GeoInfo{
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		City:        apiResp.City,
		ISP:         apiResp.ISP,
		ASN:         apiResp.AS,
		FetchedAt:   time.Now().UnixMilli(),
	}

	// cache کن توی health entry اگه وجود داره
	s.state.mu.Lock()
	if entry, ok := s.state.HealthEntries[ip]; ok {
		entry.GeoInfo = geo
	}
	s.state.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(geo)
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
	results := s.state.Phase2Results
	rawURL := s.state.SavedRawURL
	_ = rawURL
	s.state.mu.RUnlock()

	format := r.URL.Query().Get("format")

	// جمع IP های passed
	var passedIPs []string
	for _, res := range results {
		if res.Passed {
			passedIPs = append(passedIPs, res.IP)
		}
	}

	switch format {
	case "txt":
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", `attachment; filename="ips.txt"`)
		for _, ip := range passedIPs {
			fmt.Fprintf(w, "%s\n", ip)
		}

	case "links":
		// هر IP رو با config ذخیره‌شده به لینک تبدیل کن
		if rawURL == "" {
			jsonError(w, "no proxy link saved — import a link first", 400)
			return
		}
		cfg, err := ParseProxyURL(rawURL)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", `attachment; filename="links.txt"`)
		for _, ip := range passedIPs {
			link, err := BuildProxyURL(cfg, ip, rawURL)
			if err == nil {
				fmt.Fprintf(w, "%s\n", link)
			}
		}

	case "clash":
		if rawURL == "" {
			jsonError(w, "no proxy link saved — import a link first", 400)
			return
		}
		cfg, err := ParseProxyURL(rawURL)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/yaml")
		w.Header().Set("Content-Disposition", `attachment; filename="piyazche-clash.yaml"`)
		fmt.Fprint(w, BuildClashProxies(cfg, passedIPs, rawURL))

	case "singbox":
		if rawURL == "" {
			jsonError(w, "no proxy link saved — import a link first", 400)
			return
		}
		cfg, err := ParseProxyURL(rawURL)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="piyazche-singbox.json"`)
		fmt.Fprint(w, BuildSingboxOutbounds(cfg, passedIPs))

	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="results.json"`)
		s.state.mu.RLock()
		json.NewEncoder(w).Encode(s.state.Phase2Results)
		s.state.mu.RUnlock()
	}
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
		// "" = embedded CF subnets (fallback اگه ipv4.txt نبود)
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
		// context جداگانه برای فاز 2 — قابل کنسل مستقل
		p2Ctx, p2Cancel := context.WithCancel(context.Background())
		s.state.mu.Lock()
		s.state.phase2CancelFn = p2Cancel
		s.state.mu.Unlock()
		defer p2Cancel()

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

		p2results := scanner.RunPhase2WithCallback(p2Ctx, cfg, results, onP2Progress)

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
	if len(s.state.Sessions) > 50 {
		s.state.Sessions = s.state.Sessions[:50]
	}
	s.state.mu.Unlock()

	// Persist sessions to disk
	go s.saveStateToDiskNow()

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
			status := s.state.ScanStatus
			phase := s.state.ScanPhase
			s.state.mu.RUnlock()

			// وقتی فاز 2 شروع شد، این goroutine کارش تموم شده
			if status == "done" || status == "idle" || phase == "phase2" {
				return
			}

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
			cfg2, err := s.buildMergedConfig("")
			if err != nil {
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
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
	rawURL := s.state.SavedRawURL
	_ = rawURL
	_ = proxyJSON
	_ = scanJSON
	templates := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates, s.state.Templates)
	_ = templates
	s.state.mu.Unlock()

	// Persist to disk so config survives restarts
	heCopyForSave := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries { cp := *v; heCopyForSave[k] = &cp }
	_ = heCopyForSave
	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	_ = healthEnabled
	_ = healthIntervalMins
	_ = trafficDetect
	go s.saveStateToDiskNow()

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
		"rawUrl":      s.state.SavedRawURL,
	})
}

// buildMergedConfig — saved config کامل رو لود میکنه + quick override اعمال میکنه
func (s *Server) buildMergedConfig(quickOverrideJSON string) (*config.Config, error) {
	cfg := config.DefaultConfig()

	s.state.mu.RLock()
	proxyJSON := s.state.SavedProxyConfig
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
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
			Phase3   *config.Phase3Config   `json:"phase3"`
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
			// Phase3 — DownloadURL/UploadURL رو به Scan هم اعمال کن
			if saved.Phase3 != nil {
				cfg.Phase3 = *saved.Phase3
				if saved.Phase3.DownloadURL != "" {
					cfg.Scan.DownloadURL = saved.Phase3.DownloadURL
				}
				if saved.Phase3.UploadURL != "" {
					cfg.Scan.UploadURL = saved.Phase3.UploadURL
				}
				if saved.Phase3.MinDLMbps > 0 {
					cfg.Scan.MinDownloadMbps = saved.Phase3.MinDLMbps
				}
				if saved.Phase3.MinULMbps > 0 {
					cfg.Scan.MinUploadMbps = saved.Phase3.MinULMbps
				}
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
	proxyJSON := s.state.SavedProxyConfig
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
	rawURL := s.state.SavedRawURL
	_ = rawURL
	templates := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates, s.state.Templates)
	_ = templates
	s.state.mu.Unlock()

	// سیو روی دیسک
	heCopyForSave := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries { cp := *v; heCopyForSave[k] = &cp }
	_ = heCopyForSave
	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	_ = healthEnabled
	_ = healthIntervalMins
	_ = trafficDetect
	go s.saveStateToDiskNow()

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
	proxyJSON := s.state.SavedProxyConfig
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
	rawURL2 := s.state.SavedRawURL
	templates2 := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates2, s.state.Templates)
	s.state.mu.Unlock()

	// سیو روی دیسک
	heCopyForSave2 := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries { cp := *v; heCopyForSave2[k] = &cp }
	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	_ = healthEnabled
	_ = healthIntervalMins
	_ = trafficDetect
	go s.saveStateToDiskNow()

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

// handleHealthCheckNow همه IP های monitor رو فوری چک می‌کنه
func (s *Server) handleHealthCheckNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	go s.runHealthChecks()
	jsonOK(w, "checks started")
}

// handleHealthAdd یه IP به monitor اضافه می‌کنه
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
	proxyJSON := s.state.SavedProxyConfig
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
	rawURL := s.state.SavedRawURL
	_ = rawURL
	templates := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates, s.state.Templates)
	_ = templates
	heCopy := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries { cp := *v; heCopy[k] = &cp }
	s.state.mu.Unlock()

	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	_ = healthEnabled
	_ = healthIntervalMins
	_ = trafficDetect
	go s.saveStateToDiskNow()

	// Health monitor goroutine شروع کن (اگه قبلاً نبوده)
	s.startHealthMonitor()

	// بلافاصله یه چک اولیه بزن در background
	go func() {
		cfg, err := s.buildMergedConfig("")
		if err != nil || cfg.Proxy.UUID == "" {
			s.hub.Broadcast("health_update", map[string]interface{}{
				"ip":     req.IP,
				"status": "unknown",
				"error":  "no proxy config",
			})
			return
		}
		testURL := cfg.Scan.TestURL
		if testURL == "" {
			testURL = "https://www.gstatic.com/generate_204"
		}
		s.checkOneIP(cfg, req.IP, testURL)
	}()

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
	proxyJSON := s.state.SavedProxyConfig
	_ = proxyJSON
	scanJSON := s.state.SavedScanConfig
	_ = scanJSON
	rawURL := s.state.SavedRawURL
	_ = rawURL
	templates := make([]config.ConfigTemplate, len(s.state.Templates))
	copy(templates, s.state.Templates)
	_ = templates
	heCopy := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
	for k, v := range s.state.HealthEntries { cp := *v; heCopy[k] = &cp }
	s.state.mu.Unlock()
	healthEnabled := s.state.HealthEnabled
	healthIntervalMins := s.state.HealthIntervalMins
	trafficDetect := s.state.TrafficDetectEnabled
	_ = healthEnabled
	_ = healthIntervalMins
	_ = trafficDetect
	go s.saveStateToDiskNow()
	jsonOK(w, "removed")
}

// startHealthMonitor یه goroutine شروع می‌کنه که هر N دقیقه IP ها رو ping می‌کنه
func (s *Server) startHealthMonitor() {
	s.state.healthOnce.Do(func() {
		go func() {
			s.state.mu.RLock()
			interval := s.state.HealthIntervalMins
			s.state.mu.RUnlock()
			if interval <= 0 {
				interval = 3
			}
			ticker := time.NewTicker(time.Duration(interval) * time.Minute)
			s.state.mu.Lock()
			s.state.healthTicker = ticker
			s.state.mu.Unlock()
			defer ticker.Stop()
			for {
				select {
				case <-s.state.healthStop:
					return
				case <-ticker.C:
					s.state.mu.RLock()
					enabled := s.state.HealthEnabled
					s.state.mu.RUnlock()
					if enabled {
						s.runHealthChecks()
					}
				}
			}
		}()
	})
}

// handleHealthSettings تنظیمات مانیتور رو آپدیت می‌کنه
func (s *Server) handleHealthSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.state.mu.RLock()
		defer s.state.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled":           s.state.HealthEnabled,
			"intervalMins":      s.state.HealthIntervalMins,
			"trafficDetect":     s.state.TrafficDetectEnabled,
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "GET or POST only", 405)
		return
	}
	var req struct {
		Enabled       *bool `json:"enabled"`
		IntervalMins  *int  `json:"intervalMins"`
		TrafficDetect *bool `json:"trafficDetect"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	s.state.mu.Lock()
	if req.Enabled != nil {
		s.state.HealthEnabled = *req.Enabled
	}
	if req.IntervalMins != nil && *req.IntervalMins > 0 {
		s.state.HealthIntervalMins = *req.IntervalMins
		// ticker رو ریست کن
		if s.state.healthTicker != nil {
			s.state.healthTicker.Reset(time.Duration(*req.IntervalMins) * time.Minute)
		}
	}
	if req.TrafficDetect != nil {
		s.state.TrafficDetectEnabled = *req.TrafficDetect
	}
	s.state.mu.Unlock()
	// تنظیمات رو روی دیسک ذخیره کن
	go s.saveStateToDiskNow()
	jsonOK(w, "settings updated")
}

// handlePhase3Run تست سرعت جداگانه (Phase 3) روی IP های موفق فاز 2 اجرا می‌کنه
func (s *Server) handlePhase3Run(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		IPs         []string `json:"ips"`
		DownloadURL string   `json:"downloadUrl"`
		UploadURL   string   `json:"uploadUrl"`
		TestUpload  bool     `json:"testUpload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IPs) == 0 {
		jsonError(w, "ips required", 400)
		return
	}
	if req.DownloadURL == "" {
		req.DownloadURL = "https://speed.cloudflare.com/__down?bytes=5000000"
	}
	if req.UploadURL == "" {
		req.UploadURL = "https://speed.cloudflare.com/__up"
	}

	cfg, err := s.buildMergedConfig("")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// SpeedTest و BandwidthMode رو فعال کن
	cfg.Scan.SpeedTest = true
	cfg.Scan.BandwidthMode = config.BandwidthSpeedTest
	if req.DownloadURL != "" {
		cfg.Scan.DownloadURL = req.DownloadURL
	}
	if cfg.Scan.DownloadURL == "" {
		cfg.Scan.DownloadURL = "https://speed.cloudflare.com/__down?bytes=5000000"
	}
	if req.UploadURL != "" {
		cfg.Scan.UploadURL = req.UploadURL
	}
	if cfg.Scan.UploadURL == "" {
		cfg.Scan.UploadURL = "https://speed.cloudflare.com/__up"
	}
	cfg.Scan.StabilityRounds = 1 // یه دور کافیه
	cfg.Scan.PacketLossCount = 1  // phase3 فقط سرعت مهمه، packet loss نه

	// Phase1 results مصنوعی بساز
	var phase1Results []scanner.Result
	for _, ip := range req.IPs {
		phase1Results = append(phase1Results, scanner.Result{IP: ip, Success: true})
	}

	go func() {
		s.hub.Broadcast("phase3_start", map[string]int{"count": len(req.IPs)})
		var done3 int
		total3 := len(req.IPs)
		p3results := scanner.RunPhase2WithCallback(context.Background(), cfg, phase1Results, func(r scanner.Phase2Result) {
			done3++
			pct := done3 * 100 / total3
			s.hub.Broadcast("phase3_progress", map[string]interface{}{
				"ip":    r.IP,
				"done":  done3,
				"total": total3,
				"pct":   pct,
				"dl":    r.DownloadMbps,
				"ul":    r.UploadMbps,
			})
		})
		s.hub.Broadcast("phase3_done", map[string]interface{}{"results": p3results})
		s.tuiLog(fmt.Sprintf("✓ Phase 3 (Speed Test) تموم شد — %d IP", len(p3results)), "ok")
	}()

	jsonOK(w, "phase3 started")
}

func (s *Server) runHealthChecks() {
	// config رو از saved state بساز — CurrentConfig همیشه nil هست
	cfg, err := s.buildMergedConfig("")
	if err != nil || cfg.Proxy.UUID == "" {
		// proxy config نداریم — health check ممکن نیست
		s.hub.Broadcast("health_error", map[string]string{
			"message": "no proxy config — import a proxy link first",
		})
		return
	}

	s.state.mu.RLock()
	entries := make(map[string]*config.HealthEntry)
	for k, v := range s.state.HealthEntries {
		cp := *v
		entries[k] = &cp
	}
	s.state.mu.RUnlock()

	if len(entries) == 0 {
		return
	}

	testURL := cfg.Scan.TestURL
	if testURL == "" {
		testURL = "https://www.gstatic.com/generate_204"
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 4) // حداکثر ۴ چک موازی

	for ip := range entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			s.checkOneIP(cfg, ip, testURL)
		}(ip)
	}
	wg.Wait()
}

// checkOneIP یه IP رو با xray تست می‌کنه و نتیجه رو آپدیت می‌کنه
func (s *Server) checkOneIP(cfg *config.Config, ip string, testURL string) {
	port := utils.AcquirePort()
	defer utils.ReleasePort(port)

	cfgCopy := *cfg
	cfgCopy.Xray.LogLevel = "none"

	xrayCfg, err := config.GenerateXrayConfig(&cfgCopy, ip, port)
	if err != nil {
		s.updateHealthEntry(ip, false, 0, "config error: "+err.Error())
		return
	}

	mgr := xray.NewManagerWithDebug(false)
	if err := mgr.Start(xrayCfg, port); err != nil {
		s.updateHealthEntry(ip, false, 0, "xray start error: "+err.Error())
		return
	}
	defer mgr.Stop()

	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()
	if err := mgr.WaitForReadyWithContext(startCtx, 9*time.Second); err != nil {
		s.updateHealthEntry(ip, false, 0, "xray not ready")
		return
	}

	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()
	result := xray.TestConnectivityWithContext(testCtx, port, testURL, 9*time.Second)

	latency := float64(0)
	if result.Success {
		latency = float64(result.Latency.Milliseconds())
	}
	errMsg := ""
	if result.Error != nil {
		errMsg = result.Error.Error()
	}
	s.updateHealthEntry(ip, result.Success, latency, errMsg)
}

// updateHealthEntry نتیجه تست رو در state ذخیره می‌کنه و broadcast میکنه
func (s *Server) updateHealthEntry(ip string, success bool, latencyMs float64, errMsg string) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	he, ok := s.state.HealthEntries[ip]
	if !ok {
		return
	}

	now := time.Now().UnixMilli()
	he.TotalChecks++
	he.LastCheck = now

	// latency history — آخرین ۵۰ چک
	latVal := int64(latencyMs)
	if !success {
		latVal = 0 // 0 = fail در graph
	}
	he.LatencyHistory = append(he.LatencyHistory, latVal)
	he.CheckTimes = append(he.CheckTimes, now)
	if len(he.LatencyHistory) > 50 {
		he.LatencyHistory = he.LatencyHistory[len(he.LatencyHistory)-50:]
		he.CheckTimes = he.CheckTimes[len(he.CheckTimes)-50:]
	}

	if success {
		prevStatus := he.Status
		he.TotalAlive++
		he.LatencyMs = latencyMs
		he.LastSeen = now
		he.ConsecFails = 0
		if prevStatus == config.HealthDead || prevStatus == config.HealthUnknown {
			he.Status = config.HealthRecovered
		} else {
			he.Status = config.HealthAlive
		}
	} else {
		he.ConsecFails++
		if he.ConsecFails >= 2 {
			he.Status = config.HealthDead
		} else if he.Status == config.HealthUnknown {
			he.Status = config.HealthDead
		}
	}

	if he.TotalChecks > 0 {
		he.UptimePct = float64(he.TotalAlive) / float64(he.TotalChecks) * 100
	}

	s.hub.Broadcast("health_update", map[string]interface{}{
		"ip":             ip,
		"status":         string(he.Status),
		"latencyMs":      he.LatencyMs,
		"uptimePct":      he.UptimePct,
		"lastCheck":      he.LastCheck,
		"consecFails":    he.ConsecFails,
		"error":          errMsg,
		"latencyHistory": he.LatencyHistory,
		"checkTimes":     he.CheckTimes,
	})
}

// ── Quick Test (live config test) ─────────────────────────────────────────────

// handleFragmentAuto — fragment auto optimizer رو از WebUI اجرا می‌کنه
// از همون optimizer که --fragment-mode auto در CLI استفاده می‌کنه
func (s *Server) handleFragmentAuto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	var req struct {
		TestIP string `json:"testIp"` // optional — اگه خالی باشه از اولین IP فایل
	}
	json.NewDecoder(r.Body).Decode(&req)

	cfg, err := s.buildMergedConfig("")
	if err != nil || cfg.Proxy.UUID == "" {
		jsonError(w, "no proxy config — import a link first", 400)
		return
	}

	// اگه testIP نداده باشه از healthEntries استفاده کن
	testIP := req.TestIP
	if testIP == "" {
		s.state.mu.RLock()
		for ip := range s.state.HealthEntries {
			testIP = ip
			break
		}
		s.state.mu.RUnlock()
	}
	if testIP == "" {
		jsonError(w, "testIp required (or add an IP to monitor first)", 400)
		return
	}

	// در goroutine اجرا کن و progress رو broadcast کن
	go func() {
		s.hub.Broadcast("fragment_auto_start", map[string]string{"testIp": testIP})
		s.tuiLog("🔍 Fragment auto-optimization started for "+testIP, "info")

		tester := optimizer.NewFragmentTester(cfg, testIP)

		finderConfig := optimizer.FinderConfig{
			MaxTriesPerZone:   20,
			SuccessThreshold:  0.5,
			MinRangeWidth:     5,
			EnableCorrelation: true,
		}
		if cfg.Fragment.Auto.MaxTests > 0 {
			finderConfig.MaxTriesPerZone = cfg.Fragment.Auto.MaxTests
		}
		if cfg.Fragment.Auto.SuccessThreshold > 0 {
			finderConfig.SuccessThreshold = cfg.Fragment.Auto.SuccessThreshold
		}

		opt := optimizer.NewOptimizer(finderConfig, tester.CreateTesterFunc())

		sizeRange := optimizer.Range{Min: 10, Max: 60}
		intervalRange := optimizer.Range{Min: 10, Max: 32}
		if cfg.Fragment.Auto.LengthRange.Min > 0 {
			sizeRange = optimizer.Range{
				Min: cfg.Fragment.Auto.LengthRange.Min,
				Max: cfg.Fragment.Auto.LengthRange.Max,
			}
		}
		if cfg.Fragment.Auto.IntervalRange.Min > 0 {
			intervalRange = optimizer.Range{
				Min: cfg.Fragment.Auto.IntervalRange.Min,
				Max: cfg.Fragment.Auto.IntervalRange.Max,
			}
		}

		results, err := opt.FindOptimalRanges(context.Background(), sizeRange, intervalRange)
		if err != nil {
			s.tuiLog("Fragment auto error: "+err.Error(), "err")
			s.hub.Broadcast("fragment_auto_done", map[string]interface{}{"error": err.Error()})
			return
		}

		best := optimizer.GetBestResult(results)

		// نتایج رو برای UI آماده کن
		zonesOut := make([]map[string]interface{}, 0, len(results))
		for _, r := range results {
			zonesOut = append(zonesOut, map[string]interface{}{
				"zone":         r.Zone,
				"success":      r.Success,
				"sizeRange":    r.SizeRange.String(),
				"intervalRange": r.IntervalRange.String(),
				"latencyMs":    r.Latency.Milliseconds(),
				"successCount": r.SuccessCount,
				"totalTests":   r.TotalTests,
			})
		}

		payload := map[string]interface{}{
			"zones": zonesOut,
		}

		if best != nil {
			s.tuiLog(fmt.Sprintf("✓ Best fragment: zone=%s size=%s interval=%s latency=%dms",
				best.Zone, best.SizeRange, best.IntervalRange, best.Latency.Milliseconds()), "ok")

			payload["best"] = map[string]interface{}{
				"zone":          best.Zone,
				"sizeRange":     best.SizeRange.String(),
				"intervalRange": best.IntervalRange.String(),
				"latencyMs":     best.Latency.Milliseconds(),
			}

			// auto-apply: scanConfig رو آپدیت کن
			s.state.mu.Lock()
			var saved struct {
				Scan     *config.ScanConfig     `json:"scan"`
				Fragment *config.FragmentConfig `json:"fragment"`
				Xray     *config.XrayConfig     `json:"xray"`
				Phase3   *config.Phase3Config   `json:"phase3"`
			}
			if s.state.SavedScanConfig != "" {
				json.Unmarshal([]byte(s.state.SavedScanConfig), &saved)
			}
			if saved.Fragment == nil {
				saved.Fragment = &config.FragmentConfig{}
			}
			saved.Fragment.Mode = "manual"
			saved.Fragment.Packets = best.Zone
			saved.Fragment.Manual.Length = best.SizeRange.String()
			saved.Fragment.Manual.Interval = best.IntervalRange.String()
			b, _ := json.Marshal(saved)
			s.state.SavedScanConfig = string(b)
			proxyJSON := s.state.SavedProxyConfig
			_ = proxyJSON
			rawURL := s.state.SavedRawURL
			_ = rawURL
			templates := make([]config.ConfigTemplate, len(s.state.Templates))
			copy(templates, s.state.Templates)
			_ = templates
			heCopy := make(map[string]*config.HealthEntry, len(s.state.HealthEntries))
			for k, v := range s.state.HealthEntries { cp := *v; heCopy[k] = &cp }
			s.state.mu.Unlock()
			healthEnabled := s.state.HealthEnabled
			healthIntervalMins := s.state.HealthIntervalMins
			trafficDetect := s.state.TrafficDetectEnabled
			_ = healthEnabled
			_ = healthIntervalMins
			_ = trafficDetect
			go s.saveStateToDiskNow()

			payload["applied"] = true
		} else {
			s.tuiLog("✗ Fragment auto: no working configuration found", "warn")
			payload["applied"] = false
		}

		s.hub.Broadcast("fragment_auto_done", payload)
	}()

	jsonOK(w, "fragment optimization started")
}

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

// ── Subscription Fetch ────────────────────────────────────────────────────────

var serverStartTime = time.Now()

// handleSubscriptionFetch یه subscription URL رو fetch می‌کنه و لینک‌ها رو برمیگردونه
func (s *Server) handleSubscriptionFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		jsonError(w, "url required", 400)
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(req.URL)
	if err != nil {
		jsonError(w, "fetch failed: "+err.Error(), 502)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		jsonError(w, fmt.Sprintf("subscription server returned %d", resp.StatusCode), 502)
		return
	}

	bodyBytes := make([]byte, 0, 512*1024)
	buf := make([]byte, 4096)
	total := 0
	for total < 512*1024 {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
			total += n
		}
		if readErr != nil {
			break
		}
	}

	bodyStr := strings.TrimSpace(string(bodyBytes))
	if decoded := tryBase64Decode(bodyStr); decoded != "" {
		bodyStr = decoded
	}

	results, errs := ParseMultiProxy(bodyStr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"results": results,
		"errors":  errs,
		"count":   len(results),
	})
}

// tryBase64Decode سعی می‌کنه یه string رو base64 decode کنه
func tryBase64Decode(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s) < 10 {
		return ""
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		decoded, err := enc.DecodeString(s)
		if err == nil {
			result := strings.TrimSpace(string(decoded))
			if strings.Contains(result, "://") {
				return result
			}
		}
	}
	return ""
}

// ── Sessions Save ─────────────────────────────────────────────────────────────

// handleSessionsSave به client اجازه می‌ده sessions رو روی server ذخیره کنه
func (s *Server) handleSessionsSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		Sessions []ScanSession `json:"sessions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", 400)
		return
	}
	s.state.mu.Lock()
	serverIDs := map[string]bool{}
	for _, ss := range s.state.Sessions {
		serverIDs[ss.ID] = true
	}
	for _, cs := range req.Sessions {
		if !serverIDs[cs.ID] {
			s.state.Sessions = append(s.state.Sessions, cs)
		}
	}
	if len(s.state.Sessions) > 50 {
		s.state.Sessions = s.state.Sessions[:50]
	}
	s.state.mu.Unlock()
	go s.saveStateToDiskNow()
	jsonOK(w, "saved")
}

// ── System Info ───────────────────────────────────────────────────────────────

func (s *Server) handleSysInfo(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.buildMergedConfig("")
	threads := 0
	if cfg != nil {
		threads = cfg.Scan.Threads
	}
	uptime := time.Since(serverStartTime)
	h := int(uptime.Hours())
	m := int(uptime.Minutes()) % 60
	sec := int(uptime.Seconds()) % 60
	uptimeStr := ""
	if h > 0 {
		uptimeStr = fmt.Sprintf("%dh %dm %ds", h, m, sec)
	} else if m > 0 {
		uptimeStr = fmt.Sprintf("%dm %ds", m, sec)
	} else {
		uptimeStr = fmt.Sprintf("%ds", sec)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"uptime":      uptimeStr,
		"threads":     threads,
		"persistPath": configPersistPath(),
	})
}
