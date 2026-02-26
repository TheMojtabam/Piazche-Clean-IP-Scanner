package webui

import (
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

func saveStateToDisk(proxyJSON, scanJSON, rawURL string, templates []config.ConfigTemplate, healthEntries map[string]*config.HealthEntry, healthEnabled bool, healthIntervalMins int, trafficDetect bool, sessions []ScanSession) {
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
	s.state.mu.RUnlock()
	saveStateToDisk(proxyJSON, scanJSON, rawURL, templates, heCopy, healthEnabled, healthIntervalMins, trafficDetect, sessions)
}

func loadStateFromDisk() (proxyJSON, scanJSON, rawURL string, templates []config.ConfigTemplate, healthEntries map[string]*config.HealthEntry, healthEnabled *bool, healthIntervalMins *int, trafficDetect *bool, sessions []ScanSession) {
	data, err := os.ReadFile(configPersistPath())
	if err != nil {
		return "", "", "", nil, nil, nil, nil, nil, nil
	}
	var ps persistedState
	if json.Unmarshal(data, &ps) != nil {
		return "", "", "", nil, nil, nil, nil, nil, nil
	}
	return ps.ProxyConfig, ps.ScanConfig, ps.RawURL, ps.Templates, ps.HealthEntries, ps.HealthEnabled, ps.HealthIntervalMins, ps.TrafficDetectEnabled, ps.Sessions
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
	proxyJSON, scanJSON, rawURL, savedTemplates, savedHealthEntries, savedHealthEnabled, savedHealthInterval, savedTrafficDetect, savedSessions := loadStateFromDisk()

	if savedTemplates == nil {
		savedTemplates = []config.ConfigTemplate{}
	}
	if savedHealthEntries == nil {
		savedHealthEntries = make(map[string]*config.HealthEntry)
	}
	if savedSessions == nil {
		savedSessions = []ScanSession{}
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
		Templates:            savedTemplates,
		HealthEntries:        savedHealthEntries,
		healthStop:           make(chan struct{}),
		HealthIntervalMins:   healthIntervalMins,
		HealthEnabled:        healthEnabled,
		TrafficDetectEnabled: trafficDetect,
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

const indexHTMLContent = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Piyazche Scanner</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&family=Space+Mono:wght@400;700&family=Bebas+Neue&family=Orbitron:wght@400;700;900&display=swap" rel="stylesheet">
<style>
/* ══════════════════════════════════════════════════
   THEME SYSTEM — 3 themes × day/night = 6 combos
══════════════════════════════════════════════════ */

/* ── BASE SHARED ── */
:root{
  --rad:10px;--rad-sm:6px;--rad-xs:4px;
  --font-head:'Space Grotesk',sans-serif;
  --font-mono:'Space Mono',monospace;
}

/* ══ NEON NIGHT (default) ══ */
:root,
[data-theme="neon-night"]{
  --bg:#04060a;--bg2:#07090f;--bg3:#0d1420;--bg4:#121a28;
  --bd:#1a2840;--bd2:#243350;--bd3:#2e4060;
  --tx:#e8f0ff;--tx2:#8aaad0;--dim:#4a6490;
  --g:#00ffaa;--gd:rgba(0,255,170,.08);--g2:#00cc88;
  --c:#38bfff;--cd:rgba(56,191,255,.08);--c2:#0099dd;
  --y:#ffd700;--yd:rgba(255,215,0,.08);
  --r:#ff3d75;--rd:rgba(255,61,117,.08);
  --p:#c060ff;--pd:rgba(192,96,255,.08);
  --o:#ff8800;
  --glow-g:0 0 20px rgba(0,255,170,.4),0 0 6px rgba(0,255,170,.2);
  --glow-c:0 0 20px rgba(56,191,255,.4),0 0 6px rgba(56,191,255,.2);
  --glow-r:0 0 20px rgba(255,61,117,.4),0 0 6px rgba(255,61,117,.2);
  --glow-p:0 0 20px rgba(192,96,255,.4);
  --shadow:0 4px 24px rgba(0,0,0,.6);
  --logo-filter:drop-shadow(0 0 8px rgba(56,191,255,.4));
}

/* ══ NEON DAY ══ */
[data-theme="neon-day"]{
  --bg:#eef2fa;--bg2:#ffffff;--bg3:#e4eaf6;--bg4:#d8e2f0;
  --bd:#bfcfe6;--bd2:#aabcd8;--bd3:#90a8c8;
  --tx:#0f1e38;--tx2:#3a5580;--dim:#7090b8;
  --g:#007a40;--gd:rgba(0,122,64,.1);--g2:#006030;
  --c:#0066cc;--cd:rgba(0,102,204,.1);--c2:#004499;
  --y:#aa7000;--yd:rgba(170,112,0,.1);
  --r:#cc0033;--rd:rgba(204,0,51,.1);
  --p:#6600bb;--pd:rgba(102,0,187,.1);
  --o:#bb4400;
  --glow-g:0 2px 10px rgba(0,122,64,.25);
  --glow-c:0 2px 10px rgba(0,102,204,.25);
  --glow-r:0 2px 10px rgba(204,0,51,.25);
  --glow-p:0 2px 10px rgba(102,0,187,.25);
  --shadow:0 2px 14px rgba(0,0,0,.12);
  --logo-filter:drop-shadow(0 1px 3px rgba(0,119,221,.3));
}

/* ══ NAVY NIGHT ══ */
[data-theme="navy-night"]{
  --bg:#0a0e1a;--bg2:#0f1526;--bg3:#141c30;--bg4:#1a2540;
  --bd:#1e2d4a;--bd2:#253660;--bd3:#2e4070;
  --tx:#e8eaf6;--tx2:#7986cb;--dim:#3f51b5;
  --g:#00e676;--gd:rgba(0,230,118,.08);--g2:#00c853;
  --c:#4d8fff;--cd:rgba(77,143,255,.1);--c2:#1565c0;
  --y:#ffab40;--yd:rgba(255,171,64,.08);
  --r:#ff5252;--rd:rgba(255,82,82,.08);
  --p:#ce93d8;--pd:rgba(206,147,216,.08);
  --o:#ff6d00;
  --glow-g:0 0 20px rgba(0,230,118,.35),0 0 6px rgba(0,230,118,.15);
  --glow-c:0 0 20px rgba(77,143,255,.4),0 0 6px rgba(77,143,255,.2);
  --glow-r:0 0 20px rgba(255,82,82,.4),0 0 6px rgba(255,82,82,.2);
  --glow-p:0 0 20px rgba(206,147,216,.35);
  --shadow:0 4px 24px rgba(0,0,0,.6);
  --logo-filter:drop-shadow(0 0 10px rgba(77,143,255,.4));
}

/* ══ NAVY DAY ══ */
[data-theme="navy-day"]{
  --bg:#f0f4fb;--bg2:#ffffff;--bg3:#e6edf8;--bg4:#d8e4f4;
  --bd:#c0d0e8;--bd2:#a8bedd;--bd3:#8aaacf;
  --tx:#0d1b3e;--tx2:#2a4a80;--dim:#6080b0;
  --g:#1b7a3e;--gd:rgba(27,122,62,.1);--g2:#155f30;
  --c:#1565c0;--cd:rgba(21,101,192,.1);--c2:#0d47a1;
  --y:#b06000;--yd:rgba(176,96,0,.1);
  --r:#b71c1c;--rd:rgba(183,28,28,.1);
  --p:#6a1b9a;--pd:rgba(106,27,154,.1);
  --o:#bf360c;
  --glow-g:0 2px 10px rgba(27,122,62,.25);
  --glow-c:0 2px 10px rgba(21,101,192,.25);
  --glow-r:0 2px 10px rgba(183,28,28,.25);
  --glow-p:0 2px 10px rgba(106,27,154,.25);
  --shadow:0 2px 14px rgba(0,0,0,.12);
  --logo-filter:drop-shadow(0 1px 4px rgba(21,101,192,.3));
}

/* ══ WARM NIGHT ══ */
[data-theme="warm-night"]{
  --bg:#0e0b08;--bg2:#161109;--bg3:#1e160c;--bg4:#261d12;
  --bd:#332619;--bd2:#44341f;--bd3:#55422a;
  --tx:#f5e6cc;--tx2:#a08060;--dim:#6b4e2a;
  --g:#7ec850;--gd:rgba(126,200,80,.08);--g2:#60a030;
  --c:#5bc8ff;--cd:rgba(91,200,255,.08);--c2:#3090cc;
  --y:#ffd966;--yd:rgba(255,217,102,.08);
  --r:#ff4d4d;--rd:rgba(255,77,77,.08);
  --p:#cc88ff;--pd:rgba(204,136,255,.08);
  --o:#ff8c00;
  --glow-g:0 0 20px rgba(126,200,80,.35),0 0 6px rgba(126,200,80,.15);
  --glow-c:0 0 20px rgba(91,200,255,.35),0 0 6px rgba(91,200,255,.15);
  --glow-r:0 0 20px rgba(255,77,77,.4),0 0 6px rgba(255,77,77,.2);
  --glow-p:0 0 20px rgba(204,136,255,.35);
  --shadow:0 4px 24px rgba(0,0,0,.7);
  --logo-filter:drop-shadow(0 0 10px rgba(255,140,0,.35));
}

/* ══ WARM DAY ══ */
[data-theme="warm-day"]{
  --bg:#fdf6ec;--bg2:#ffffff;--bg3:#f5ebe0;--bg4:#ecddd0;
  --bd:#ddc8b0;--bd2:#ccb098;--bd3:#b89880;
  --tx:#2e1a08;--tx2:#7a4820;--dim:#a8703a;
  --g:#3d7a1a;--gd:rgba(61,122,26,.1);--g2:#2d600f;
  --c:#005588;--cd:rgba(0,85,136,.1);--c2:#003d66;
  --y:#8a6000;--yd:rgba(138,96,0,.1);
  --r:#aa1818;--rd:rgba(170,24,24,.1);
  --p:#661a88;--pd:rgba(102,26,136,.1);
  --o:#c05000;
  --glow-g:0 2px 10px rgba(61,122,26,.25);
  --glow-c:0 2px 10px rgba(0,85,136,.25);
  --glow-r:0 2px 10px rgba(170,24,24,.25);
  --glow-p:0 2px 10px rgba(102,26,136,.25);
  --shadow:0 2px 14px rgba(0,0,0,.1);
  --logo-filter:drop-shadow(0 1px 4px rgba(192,80,0,.3));
}

/* ── logo glow via variable ── */
.logo{ filter:var(--logo-filter); }

/* ── Theme picker modal ── */
.theme-picker-overlay{
  display:none;position:fixed;inset:0;
  background:rgba(0,0,0,.6);backdrop-filter:blur(4px);
  z-index:10000;align-items:center;justify-content:center;
}
.theme-picker-overlay.open{display:flex;}
.theme-picker{
  background:var(--bg2);border:1px solid var(--bd2);
  border-radius:14px;padding:24px;width:480px;max-width:95vw;
  box-shadow:0 20px 60px rgba(0,0,0,.5);
}
.tp-title{
  font-family:'Orbitron',monospace;font-size:14px;font-weight:700;
  letter-spacing:2px;color:var(--c);margin-bottom:18px;
  display:flex;align-items:center;justify-content:space-between;
}
.tp-close{
  background:none;border:none;color:var(--dim);cursor:pointer;
  font-size:18px;line-height:1;padding:2px 6px;
  transition:color .15s;
}
.tp-close:hover{color:var(--tx)}
.tp-section{margin-bottom:18px;}
.tp-label{font-size:9px;color:var(--dim);letter-spacing:2px;text-transform:uppercase;font-family:var(--font-mono);margin-bottom:10px;}
.tp-themes{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;}
.tp-theme-btn{
  background:var(--bg3);border:2px solid var(--bd2);
  border-radius:8px;padding:10px 8px;cursor:pointer;
  transition:all .15s;text-align:center;
  font-family:var(--font-mono);font-size:10px;color:var(--tx2);
}
.tp-theme-btn:hover{border-color:var(--c);color:var(--tx)}
.tp-theme-btn.active{border-color:var(--c);color:var(--c);background:var(--cd);}
.tp-theme-btn .tp-swatch{
  display:flex;gap:3px;justify-content:center;margin-bottom:6px;
}
.tp-swatch span{width:12px;height:12px;border-radius:50%;}
.tp-mode{display:grid;grid-template-columns:1fr 1fr;gap:8px;}
.tp-mode-btn{
  background:var(--bg3);border:2px solid var(--bd2);
  border-radius:8px;padding:10px;cursor:pointer;
  transition:all .15s;font-family:var(--font-mono);
  font-size:11px;color:var(--tx2);display:flex;
  align-items:center;justify-content:center;gap:8px;
}
.tp-mode-btn:hover{border-color:var(--c);color:var(--tx)}
.tp-mode-btn.active{border-color:var(--c);color:var(--c);background:var(--cd);}
.tp-mode-icon{font-size:18px;}
*{margin:0;padding:0;box-sizing:border-box}
html{height:100%}

/* ══ UI MODE TOGGLE ══ */
.ui-mode-toggle{display:flex;align-items:center;gap:4px;background:var(--bg2);border:1px solid var(--bd2);border-radius:var(--rad);padding:3px;margin-right:6px;}
.ui-mode-btn{padding:3px 10px;border-radius:calc(var(--rad) - 2px);border:none;background:transparent;color:var(--dim);font-family:var(--font-head);font-size:11px;cursor:pointer;transition:.2s;}
.ui-mode-btn.active{background:var(--c);color:var(--bg);font-weight:700;}
.ui-mode-btn:hover:not(.active){color:var(--tx);background:var(--cd);}

/* compact mode hides non-essential elements */
body.compact-mode .card-hd .card-hd-extra,
body.compact-mode .phd p,
body.compact-mode .nav-group{display:none!important;}
body.compact-mode .sidebar{width:52px;}
body.compact-mode .nav-item .nav-icon{margin-right:0;}
body.compact-mode .nav-item span:not(.nav-icon):not(.nav-badge){display:none;}
body.compact-mode .main{margin-left:52px;}
body.compact-mode .nav-item{justify-content:center;padding:8px 0;}
body.compact-mode .tui-body{height:calc(100vh - 170px);}
body.compact-mode .live-feed-body{height:110px;}
body{font-family:var(--font-head);background:var(--bg);color:var(--tx);height:100%;font-size:15px;line-height:1.6;overflow:hidden;transition:background .3s,color .3s}
.app{display:grid;grid-template-columns:200px 1fr;grid-template-rows:56px 1fr;height:100vh}

/* ══ TOPBAR ══ */
.topbar{
  grid-column:1/-1;
  background:var(--bg2);
  border-bottom:2px solid var(--bd2);
  display:flex;align-items:center;
  padding:0 18px;gap:14px;
  position:relative;z-index:100;
  box-shadow:var(--shadow);
}
.logo{
  font-family:'Orbitron',monospace;
  font-size:18px;letter-spacing:4px;font-weight:900;
  user-select:none;
  background:linear-gradient(90deg,var(--c),var(--g),var(--p));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
}
.status-pill{
  display:flex;align-items:center;gap:6px;
  padding:4px 12px;border-radius:20px;
  font-size:12px;background:var(--bg3);
  border:1px solid var(--bd2);
  font-family:var(--font-mono);
}
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0;transition:all .3s}
.dot-idle{background:var(--dim)}
.dot-scan{background:var(--g);box-shadow:var(--glow-g);animation:pulse 1.2s ease-in-out infinite}
.dot-warn{background:var(--y)}
.dot-done{background:var(--c)}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:.3;transform:scale(.7)}}
.tb-right{margin-left:auto;display:flex;align-items:center;gap:8px}
.proxy-chip{
  display:inline-flex;align-items:center;gap:5px;
  padding:3px 10px;border-radius:5px;
  font-size:11px;font-family:var(--font-mono);
  background:var(--gd);border:1px solid rgba(0,255,170,.3);
  color:var(--g);cursor:pointer;transition:all .15s;
}
.proxy-chip:hover{background:rgba(0,255,170,.12)}
.theme-btn{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx2);padding:5px 13px;border-radius:20px;
  font-size:11px;cursor:pointer;font-family:var(--font-mono);
  transition:all .2s;letter-spacing:.5px;
  display:flex;align-items:center;gap:5px;
}
.theme-btn:hover{border-color:var(--c);color:var(--c);background:var(--cd)}
.theme-icon{font-size:13px}

/* ══ SIDEBAR ══ */
.sidebar{
  background:var(--bg2);border-right:1px solid var(--bd);
  display:flex;flex-direction:column;overflow-y:auto;
}
.nav-group{
  padding:14px 14px 4px;
  font-size:9px;letter-spacing:2.5px;text-transform:uppercase;
  color:var(--dim);font-family:var(--font-mono);
}
.nav-item{
  display:flex;align-items:center;gap:9px;
  padding:9px 16px;cursor:pointer;
  transition:all .12s;color:var(--tx2);
  font-size:13px;border:none;background:none;
  width:100%;text-align:left;
  border-left:2px solid transparent;
  font-family:var(--font-head);font-weight:500;
}
.nav-item:hover{background:var(--bg3);color:var(--tx)}
.nav-item.active{
  background:linear-gradient(90deg,var(--cd),transparent);
  color:var(--c);border-left-color:var(--c);
}
.nav-icon{font-size:14px;min-width:18px;text-align:center}
.nav-badge{
  margin-left:auto;background:var(--bd2);
  color:var(--dim);font-size:9px;padding:1px 6px;
  border-radius:8px;font-family:var(--font-mono);
}
.nav-badge.live{background:var(--gd);color:var(--g)}

/* ══ MAIN ══ */
.main{overflow-y:auto;overflow-x:hidden;background:var(--bg)}
.main::-webkit-scrollbar{width:4px}
.main::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
.page{display:none;padding:22px 24px 32px;min-height:100%}
.page.active{display:block}

/* ══ PAGE HEADER ══ */
.phd{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:18px;gap:12px}
.phd-l h2{
  font-family:'Orbitron',monospace;font-size:18px;font-weight:700;
  letter-spacing:2px;
  background:linear-gradient(90deg,var(--tx),var(--c));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
}
.phd-l p{font-size:12px;color:var(--tx2);margin-top:3px}
.phd-r{display:flex;gap:7px;align-items:center;flex-shrink:0}

/* ══ STATS ROW ══ */
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:10px;margin-bottom:16px}
.stat-card{
  background:var(--bg2);border:1px solid var(--bd);
  border-radius:var(--rad-sm);padding:14px 16px;
  position:relative;overflow:hidden;
  transition:border-color .2s,transform .1s;
  cursor:default;
}
.stat-card:hover{transform:translateY(-1px);border-color:var(--bd3)}
.stat-card::before{
  content:'';position:absolute;top:0;left:0;right:0;height:2px;
  background:linear-gradient(90deg,var(--c),var(--g));
  opacity:0;transition:opacity .2s;
}
.stat-card:hover::before{opacity:1}
.stat-v{
  font-family:'Orbitron',monospace;font-weight:700;
  line-height:1;transition:color .3s,text-shadow .3s;
  font-size:28px;
  letter-spacing:-1px;
}
.stat-v.active{text-shadow:0 0 20px currentColor}
.stat-l{font-size:9px;color:var(--dim);margin-top:8px;letter-spacing:2px;text-transform:uppercase}

/* ══ PROGRESS ══ */
.prog-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:14px;box-shadow:var(--shadow)}
.prog-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 16px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono)}
.prog-hd-l{display:flex;align-items:center;gap:7px}
.prog-bd{padding:14px 16px}
.prog-meta{display:flex;justify-content:space-between;font-size:11px;color:var(--dim);margin-bottom:6px;font-family:var(--font-mono)}
.prog-pct{color:var(--c);font-weight:700;font-size:14px}
.prog-wrap{background:var(--bg);border-radius:3px;height:8px;overflow:hidden;margin-bottom:10px;border:1px solid var(--bd)}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:3px;transition:width .5s cubic-bezier(.4,0,.2,1);width:0%;box-shadow:var(--glow-c);position:relative}
.prog-bar::after{
  content:'';position:absolute;top:0;left:0;right:0;bottom:0;
  background:linear-gradient(90deg,transparent,rgba(255,255,255,.3),transparent);
  animation:shimmer 1.5s infinite;
}
@keyframes shimmer{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
.prog-bar.p2{background:linear-gradient(90deg,var(--p),var(--c));box-shadow:var(--glow-p)}

/* ══ LIVE FEED ══ */
.live-feed{background:var(--bg);border:1px solid var(--bd);border-radius:var(--rad-sm);overflow:hidden}
.live-feed-hd{padding:7px 12px;border-bottom:1px solid var(--bd);display:flex;align-items:center;gap:8px;font-size:10px;color:var(--dim);font-family:var(--font-mono)}
.live-feed-body{height:130px;overflow-y:auto;padding:8px 12px;display:flex;flex-direction:column-reverse;gap:2px}
.live-feed-body::-webkit-scrollbar{width:3px}
.live-feed-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.live-row{display:flex;align-items:center;gap:8px;font-family:var(--font-mono);font-size:12px;padding:2px 0;animation:fadeIn .2s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(-3px)}to{opacity:1;transform:none}}
.live-row-ok{color:var(--g)}.live-row-fail{color:var(--dim)}
.live-row-scan{color:var(--tx2)}.live-row-p2{color:var(--p)}
.live-ip{color:var(--c);min-width:128px;font-weight:700}
.live-lat{color:var(--y);min-width:62px}
.live-tag{font-size:9px;padding:1px 5px;border-radius:3px;flex-shrink:0}
.tag-ok{background:var(--gd);color:var(--g)}.tag-fail{background:var(--rd);color:var(--r)}.tag-p2{background:var(--pd);color:var(--p)}
.spin{width:12px;height:12px;border:1.5px solid var(--bd3);border-top-color:var(--c);border-radius:50%;animation:spin .6s linear infinite;flex-shrink:0}
@keyframes spin{to{transform:rotate(360deg)}}

/* ══ CARDS ══ */
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:12px;box-shadow:var(--shadow)}
.card-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:9px 14px;font-size:11px;color:var(--tx2);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono);letter-spacing:.3px}
.card-bd{padding:14px}

/* ══ FORMS ══ */
input,select,textarea{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx);border-radius:var(--rad-xs);
  padding:7px 10px;font-family:var(--font-mono);
  font-size:13px;width:100%;outline:none;
  transition:border-color .15s,box-shadow .15s;
}
input:focus,select:focus,textarea:focus{border-color:var(--c);box-shadow:0 0 0 2px var(--cd)}
textarea{resize:vertical;min-height:60px}
select option{background:var(--bg3)}
label{display:block;font-size:11px;color:var(--tx2);margin-bottom:4px;letter-spacing:.3px;text-transform:uppercase;font-family:var(--font-mono)}

/* ══ BUTTONS ══ */
.btn{
  display:inline-flex;align-items:center;gap:5px;
  padding:7px 16px;border-radius:var(--rad-sm);
  font-family:var(--font-mono);font-size:11px;
  cursor:pointer;border:1px solid transparent;
  font-weight:700;letter-spacing:.5px;
  transition:all .15s;
}
.btn-success{background:var(--gd);border-color:var(--g2);color:var(--g);box-shadow:0 0 12px rgba(0,255,170,.15)}
.btn-success:hover{background:rgba(0,255,170,.15);box-shadow:var(--glow-g)}
.btn-danger{background:var(--rd);border-color:var(--r);color:var(--r)}
.btn-danger:hover{background:rgba(255,45,107,.15);box-shadow:var(--glow-r)}
.btn-primary{background:var(--cd);border-color:var(--c2);color:var(--c);box-shadow:0 0 12px rgba(56,191,255,.15)}
.btn-primary:hover{background:rgba(56,191,255,.15);box-shadow:var(--glow-c)}
.btn-warn{background:var(--yd);border-color:var(--y);color:var(--y)}
.btn{background:var(--bg3);border-color:var(--bd2);color:var(--tx2)}
.btn:hover{border-color:var(--bd3);color:var(--tx)}
.btn-success,.btn-danger,.btn-primary,.btn-warn{background:var(--gd)}.btn-success,.btn-success:hover{background:var(--gd);border-color:var(--g2);color:var(--g)}
.btn-sm{padding:5px 11px;font-size:11px}
.btn:disabled{opacity:.4;cursor:not-allowed}
.copy-btn{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx2);border-radius:3px;padding:2px 7px;
  font-size:11px;cursor:pointer;transition:all .12s;
}
.copy-btn:hover{border-color:var(--c);color:var(--c)}

/* Explicit button color overrides (must come after .btn) */
.btn-success-real{background:var(--gd)!important;border-color:var(--g2)!important;color:var(--g)!important}
.btn-success-real:hover{background:rgba(0,255,170,.15)!important;box-shadow:var(--glow-g)!important}
.btn-danger-real{background:var(--rd)!important;border-color:var(--r)!important;color:var(--r)!important}
.btn-primary-real{background:var(--cd)!important;border-color:var(--c2)!important;color:var(--c)!important}

/* ══ TABLE ══ */
.tbl-wrap{overflow-x:auto}
.tbl{width:100%;border-collapse:collapse;font-size:13px}
.tbl th{
  padding:9px 12px;text-align:left;
  font-size:9px;letter-spacing:1.5px;text-transform:uppercase;
  color:var(--dim);font-family:var(--font-mono);
  background:var(--bg3);border-bottom:1px solid var(--bd2);
  position:sticky;top:0;
}
.tbl td{padding:8px 12px;border-bottom:1px solid var(--bd);vertical-align:middle}
.tbl tr:hover td{background:var(--bg3)}
.tbl tr.pass-row:hover td{background:rgba(0,255,170,.04)}
.tbl tr.fail-row td{opacity:.7}
.tbl tr.p1-row:hover td{background:rgba(56,191,255,.04)}

/* ══ BADGES ══ */
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:3px;font-size:10px;font-family:var(--font-mono);font-weight:700;letter-spacing:.5px}
.bg{background:var(--gd);color:var(--g);border:1px solid rgba(0,255,170,.3)}
.br{background:var(--rd);color:var(--r);border:1px solid rgba(255,45,107,.3)}
.by{background:var(--yd);color:var(--y);border:1px solid rgba(255,215,0,.3)}

/* ══ IP CHIPS ══ */
.ip-chips{display:flex;flex-wrap:wrap;gap:6px}
.ip-chip{
  display:inline-flex;align-items:center;gap:6px;
  background:var(--cd);border:1px solid var(--c2);
  color:var(--c);border-radius:4px;padding:4px 10px;
  font-family:var(--font-mono);font-size:11px;
  cursor:pointer;transition:all .12s;
}
.ip-chip:hover{background:rgba(56,191,255,.15);box-shadow:var(--glow-c)}
.ip-chip .lat{color:var(--y);font-size:11px}

/* ══ TABS ══ */
.tab-bar{display:flex;gap:1px;margin-bottom:0;background:var(--bd);border-radius:var(--rad) var(--rad) 0 0;overflow:hidden}
.tab{
  flex:1;padding:9px 14px;border:none;background:var(--bg3);
  color:var(--tx2);font-family:var(--font-mono);font-size:11px;
  cursor:pointer;transition:all .12s;
}
.tab:hover{background:var(--bg2);color:var(--tx)}
.tab.active{background:var(--bg2);color:var(--c);border-bottom:2px solid var(--c)}

/* ══ FORM GRID ══ */
.f-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.f-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.f-row{display:flex;flex-direction:column;gap:4px;margin-bottom:10px}
.chk-row{display:flex;align-items:center;gap:6px;font-size:11px;font-family:var(--font-mono);color:var(--tx2);cursor:pointer;margin-bottom:8px}
.chk-row input{width:auto}

/* ══ TUI ══ */
.tui-wrap{background:var(--bg2);border:1px solid var(--bd2);border-radius:var(--rad);overflow:hidden;box-shadow:var(--shadow)}
.tui-hd{
  background:var(--bg3);padding:8px 14px;
  display:flex;align-items:center;gap:8px;
  border-bottom:1px solid var(--bd2);
}
.tui-dots{display:flex;gap:5px}
.tui-dot{width:10px;height:10px;border-radius:50%}
.tui-body{height:calc(100vh - 200px);overflow-y:auto;padding:12px 16px;font-family:var(--font-mono);font-size:12px}
.tui-body::-webkit-scrollbar{width:3px}
.tui-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.tui-line{display:flex;gap:10px;padding:1px 0;animation:fadeIn .15s}
.tui-t{color:var(--dim);flex-shrink:0;font-size:10px}
.tui-ok{color:var(--g)}.tui-err{color:var(--r)}.tui-warn{color:var(--y)}
.tui-p2{color:var(--p)}.tui-scan{color:var(--c)}.tui-info{color:var(--tx2)}
.cursor{display:inline-block;width:7px;height:12px;background:var(--c);margin-left:2px;animation:blink 1s step-end infinite;vertical-align:text-bottom}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}

/* ══ HISTORY ══ */
.hist-item{
  display:flex;align-items:center;gap:16px;
  padding:14px 18px;background:var(--bg2);
  border:1px solid var(--bd);border-radius:var(--rad);
  margin-bottom:8px;cursor:pointer;
  transition:all .12s;
}
.hist-item:hover{border-color:var(--c);background:var(--bg3);transform:translateX(2px)}
.hist-n{font-family:'Orbitron',monospace;font-size:28px;font-weight:700;min-width:50px;text-align:center}
.hist-info{flex:1}
.hist-date{font-size:10px;color:var(--dim);font-family:var(--font-mono);margin-top:2px}

/* ══ SESSION BANNER ══ */
.session-banner{
  display:flex;justify-content:space-between;align-items:center;
  padding:9px 14px;background:var(--yd);border:1px solid var(--y);
  border-radius:var(--rad-sm);margin-bottom:12px;
  font-family:var(--font-mono);font-size:11px;color:var(--y);
}

/* ══ PARSED BOX ══ */
.parsed-box{font-family:var(--font-mono);font-size:12px;line-height:1.8}
.parsed-box .k{color:var(--dim)}.parsed-box .v{color:var(--c)}

/* ══ SCROLLBAR GLOBAL ══ */
::-webkit-scrollbar{width:4px;height:4px}
::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}

/* ══ RESPONSIVE ══ */
@media(max-width:768px){.app{grid-template-columns:1fr}.sidebar{display:none}.stats-row{grid-template-columns:repeat(2,1fr)}.f-grid,.f-grid-3{grid-template-columns:1fr}}
</style>
</head>
<body>
<div class="app">

<!-- TOPBAR -->
<div class="topbar">
  <div class="logo">PIYAZCHE</div>
  <div class="status-pill">
    <div class="dot dot-idle" id="sDot"></div>
    <span id="sTxt" style="font-family:var(--font-mono);font-size:11px">idle</span>
    <span id="sPhase" style="color:var(--dim);font-size:10px"></span>
  </div>
  <div id="proxyChip" style="display:none" class="proxy-chip" onclick="nav('import')" title="Active config — click to change">
    <span>⬡</span><span id="proxyChipTxt"></span>
  </div>
  <div class="tb-right">
    <span id="tbProgress" style="font-family:var(--font-mono);font-size:11px;color:var(--dim)"></span>
    <div class="ui-mode-toggle" title="Switch UI mode">
      <button class="ui-mode-btn active" id="uiModeFull" onclick="setUIMode('full')">Full</button>
      <button class="ui-mode-btn" id="uiModeCompact" onclick="setUIMode('compact')">Compact</button>
    </div>
    <button class="theme-btn" onclick="openThemePicker()" id="themeBtn" title="Change theme">
      <span class="theme-icon" id="themeIcon">🌙</span>
      <span id="themeTxt">NEON</span>
    </button>
  </div>
</div>

<!-- SIDEBAR -->
<div class="sidebar">
  <div class="nav-group">Scanner</div>
  <button class="nav-item active" data-page="scan" onclick="nav('scan',this)">
    <span class="nav-icon">⚡</span>Scan
    <span class="nav-badge live" id="nbScan" style="display:none">LIVE</span>
  </button>
  <button class="nav-item" data-page="results" onclick="nav('results',this)">
    <span class="nav-icon">◈</span>Results
    <span class="nav-badge" id="nbResults">0</span>
  </button>
  <button class="nav-item" data-page="subnets" onclick="nav('subnets',this);loadSubnets()">
    <span class="nav-icon">▦</span>Subnets
  </button>
  <button class="nav-item" data-page="monitor" onclick="nav('monitor',this);loadHealth();loadMonitorSettings()">
    <span class="nav-icon">♡</span>Monitor
    <span class="nav-badge" id="nbMonitor" style="display:none">0</span>
  </button>
  <button class="nav-item" data-page="history" onclick="nav('history',this)">
    <span class="nav-icon">◷</span>History
    <span class="nav-badge" id="nbHistory">0</span>
  </button>
  <div class="nav-group">Config</div>
  <button class="nav-item" data-page="templates" onclick="nav('templates',this);loadTemplates()">
    <span class="nav-icon">⬡</span>Templates
    <span class="nav-badge" id="nbTemplates">0</span>
  </button>
  <button class="nav-item" data-page="config" onclick="nav('config',this)">
    <span class="nav-icon">⚙</span>Settings
  </button>
  <button class="nav-item" data-page="import" onclick="nav('import',this)">
    <span class="nav-icon">↑</span>Import Link
  </button>
  <div class="nav-group">Tools</div>
  <button class="nav-item" data-page="tui" onclick="nav('tui',this)">
    <span class="nav-icon">▸</span>Live Log
  </button>
  <button class="nav-item" data-page="sysinfo" onclick="nav('sysinfo',this);loadSysInfo()">
    <span class="nav-icon">⬡</span>System
  </button>
</div>

<!-- MAIN -->
<div class="main">

<!-- ══ SCAN PAGE ══ -->
<div id="page-scan" class="page active">
  <div class="phd">
    <div class="phd-l"><h2>Scan</h2><p id="configSummary" style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">No config — import a proxy link first</p></div>
    <div class="phd-r">
      <button class="btn btn-success-real" id="btnStart" onclick="startScan()">▶ Start</button>
      <button class="btn btn-danger-real" id="btnStop" onclick="stopScan()" style="display:none">■ Stop</button>
    </div>
  </div>

  <!-- Stats -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-v" id="stTotal" style="color:var(--tx2)">—</div>
      <div class="stat-l">Total IPs</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stDone" style="color:var(--c)">—</div>
      <div class="stat-l">Scanned</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stPass" style="color:var(--g)">0</div>
      <div class="stat-l">Passed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stFail" style="color:var(--r)">0</div>
      <div class="stat-l">Failed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stETA" style="color:var(--y);font-size:20px">—</div>
      <div class="stat-l">ETA</div>
    </div>
  </div>

  <!-- Progress -->
  <div class="prog-card">
    <div class="prog-hd">
      <div class="prog-hd-l">
        <div class="dot dot-idle" id="pDot"></div>
        <span id="progLabel">Ready</span>
      </div>
      <span id="progRate" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="prog-bd">
      <div class="prog-meta">
        <span id="progTxt" style="font-family:var(--font-mono)">0 / 0</span>
        <span class="prog-pct" id="progPct">0%</span>
      </div>
      <div class="prog-wrap"><div class="prog-bar" id="progBar"></div></div>
      <!-- Quick settings -->
      <div style="display:grid;grid-template-columns:repeat(5,1fr);gap:8px;margin-top:4px">
        <div><label>Threads</label><input type="number" id="qThreads" value="200" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Timeout (s)</label><input type="number" id="qTimeout" value="8" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Max Lat (ms)</label><input type="number" id="qMaxLat" value="3500" style="font-size:11px;padding:4px 7px"></div>
        <div><label>P2 Rounds</label><input type="number" id="qRounds" value="3" min="0" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Sample/Subnet</label><input type="number" id="sampleSize" value="1" min="1" style="font-size:11px;padding:4px 7px"></div>
      </div>
    </div>
  </div>

  <!-- IP Input + Feed -->
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
    <div class="card">
      <div class="card-hd"><div>IP Ranges</div><span id="feedCount" style="color:var(--dim)"></span></div>
      <div class="card-bd" style="padding:10px">
        <textarea id="ipInput" rows="7" placeholder="104.16.0.0/12&#10;162.158.0.0/15&#10;Or single IPs..."></textarea>
        <div style="display:flex;justify-content:space-between;align-items:center;margin-top:6px">
          <span style="font-size:10px;color:var(--dim);font-family:var(--font-mono)">CIDR or plain IPs</span>
          <input type="number" id="maxIPInput" placeholder="Max IPs (0=all)" style="width:140px;font-size:11px;padding:3px 7px">
        </div>
        <!-- Quick Range Selector -->
        <div style="margin-top:8px">
          <div style="font-family:var(--font-mono);font-size:9px;color:var(--dim);margin-bottom:5px;letter-spacing:1px" id="rangeLabel">QUICK RANGES</div>
          <div style="display:flex;flex-wrap:wrap;gap:4px" id="quickRanges"></div>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="card-hd">
        <div class="live-feed-hd" style="padding:0;border:none;background:none">
          <div class="dot dot-idle" id="feedDot"></div>
          <span>Live Feed</span>
        </div>
        <span id="feedCount2" style="color:var(--dim);font-size:10px"></span>
      </div>
      <div class="live-feed-body" id="liveFeed" style="height:160px">
        <div class="live-row live-row-scan"><span style="color:var(--dim)">›</span><span>Waiting to start...</span></div>
      </div>
    </div>
  </div>
</div>

<!-- ══ RESULTS PAGE ══ -->
<div id="page-results" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Results</h2><p id="resSummary" style="font-family:var(--font-mono);font-size:10px">No results yet</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="exportResults('txt')">↓ IPs</button>
      <button class="btn btn-sm" onclick="exportResults('links')">↓ Links</button>
      <button class="btn btn-sm" onclick="exportResults('clash')">↓ Clash</button>
      <button class="btn btn-sm" onclick="exportResults('singbox')">↓ Sing-box</button>
      <button class="btn btn-sm" onclick="exportResults('json')">↓ JSON</button>
      <button class="btn btn-sm" onclick="copyAllPassed()">⎘ Copy All</button>
      <button class="btn btn-sm btn-primary-real" id="btnP3" onclick="runPhase3()">🚀 Speed Test (Phase 3)</button>
    </div>
  </div>

  <!-- IP Chips -->
  <div class="card" style="margin-bottom:12px">
    <div class="card-hd">
      <div>✓ Passed IPs</div>
      <span id="passedBadge" class="badge bg">0</span>
    </div>
    <div class="card-bd"><div class="ip-chips" id="ipChips"><span style="color:var(--dim);font-size:12px">No results</span></div></div>
  </div>

  <!-- Tabs -->
  <div class="tab-bar">
    <button class="tab active" onclick="switchTab('p2',this)">▶ Phase 2 — Deep Test</button>
    <button class="tab" onclick="switchTab('p1',this)">⚡ Phase 1 — Initial Scan</button>
  </div>

  <!-- Phase 2 table -->
  <div id="tab-p2" class="card" style="border-radius:0 0 var(--rad) var(--rad)">
    <div class="card-hd">
      <div>Phase 2 — Stability & Speed</div>
      <span id="p2CountBadge" style="color:var(--dim);font-size:10px"></span>
    </div>
    <!-- Filter & Sort Controls -->
    <div style="padding:8px 12px;display:flex;gap:8px;flex-wrap:wrap;align-items:center;border-bottom:1px solid var(--b2)">
      <input type="text" id="filterIP" placeholder="Filter IP..." style="width:130px;font-size:11px;padding:3px 7px;font-family:var(--font-mono)" oninput="applyP2Filter()">
      <select id="filterStatus" style="font-size:11px;padding:3px 6px" onchange="applyP2Filter()">
        <option value="all">All</option>
        <option value="pass">PASS only</option>
        <option value="fail">FAIL only</option>
      </select>
      <select id="sortP2" style="font-size:11px;padding:3px 6px" onchange="applyP2Filter()">
        <option value="score">Sort: Score ↓</option>
        <option value="latency">Sort: Latency ↑</option>
        <option value="dl">Sort: Download ↓</option>
        <option value="loss">Sort: Pkt Loss ↑</option>
        <option value="jitter">Sort: Jitter ↑</option>
      </select>
      <button class="btn btn-sm" onclick="applyP2Filter()" style="padding:2px 8px">↺</button>
    </div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP Address</th><th>Score</th><th>Latency</th><th>Jitter</th><th>Pkt Loss</th><th>Download</th><th>Upload</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody id="p2Tbody"><tr><td colspan="10" style="text-align:center;color:var(--dim);padding:32px;font-family:var(--font-mono)">No Phase 2 results yet</td></tr></tbody>
      </table>
    </div>
  </div>

  <!-- Phase 1 table -->
  <div id="tab-p1" class="card" style="display:none;border-radius:0 0 var(--rad) var(--rad)">
    <div class="card-hd">
      <div>Phase 1 — Initial Scan (passed only)</div>
      <span id="p1CountBadge" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP Address</th><th>Latency</th><th>Pkt Loss</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody id="p1Tbody"><tr><td colspan="6" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<!-- ══ HISTORY PAGE ══ -->
<div id="page-history" class="page">
  <div class="phd">
    <div class="phd-l"><h2>History</h2><p>Previous scan sessions — saved locally</p></div>
    <div class="phd-r">
      <button class="btn btn-sm btn-danger-real" onclick="clearHistory()">✕ Clear History</button>
    </div>
  </div>
  <div id="histList"><p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p></div>
</div>

<!-- ══ CONFIG PAGE ══ -->
<div id="page-config" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Settings</h2><p>Saved automatically to disk on save</p></div>
    <div class="phd-r">
      <button class="btn btn-success-real" onclick="saveConfig()">⬡ Save Settings</button>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>⚡ PHASE 1 — Initial Scan</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Threads</label><input type="number" id="cfgThreads" value="200" min="1"></div>
        <div class="f-row"><label>Timeout (seconds)</label><input type="number" id="cfgTimeout" value="8" min="1"></div>
        <div class="f-row"><label>Max Latency (ms)</label><input type="number" id="cfgMaxLat" value="3500"></div>
        <div class="f-row"><label>Retries</label><input type="number" id="cfgRetries" value="2" min="0"></div>
        <div class="f-row"><label>Max IPs (0 = all)</label><input type="number" id="cfgMaxIPs" value="0" min="0"></div>
        <div class="f-row"><label>Sample per Subnet</label><input type="number" id="cfgSampleSize" value="1" min="1"></div>
      </div>
      <div class="f-row"><label>Test URL</label><input type="text" id="cfgTestURL" value="https://www.gstatic.com/generate_204"></div>
      <label class="chk-row"><input type="checkbox" id="cfgShuffle" checked> Shuffle IPs before scan</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>◈ PHASE 2 — Deep Stability Test</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Rounds (0 = disabled)</label><input type="number" id="cfgRounds" value="3" min="0"></div>
        <div class="f-row"><label>Interval (seconds)</label><input type="number" id="cfgInterval" value="5" min="1"></div>
        <div class="f-row"><label>Ping count (packet loss)</label><input type="number" id="cfgPLCount" value="5" min="1"></div>
        <div class="f-row"><label>Max Packet Loss % (-1 = off)</label><input type="number" id="cfgMaxPL" value="-1" min="-1" max="100"></div>
      </div>
      <label class="chk-row"><input type="checkbox" id="cfgJitter"> Measure Jitter (RFC 3550)</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd">
      <div>🚀 PHASE 3 — Speed Test (جداگانه)</div>
      <label style="display:flex;align-items:center;gap:6px;cursor:pointer">
        <input type="checkbox" id="cfgP3Enabled" onchange="toggleP3Settings(this.checked)"> فعال
      </label>
    </div>
    <div class="card-bd" id="p3Settings" style="display:none">
      <div class="f-grid-3">
        <div class="f-row"><label>Min Download Mbps (0 = off)</label><input type="number" id="cfgMinDL" value="0" min="0" step="0.1"></div>
        <div class="f-row"><label>Min Upload Mbps (0 = off)</label><input type="number" id="cfgMinUL" value="0" min="0" step="0.1"></div>
        <div></div>
      </div>
      <label class="chk-row"><input type="checkbox" id="cfgP3Upload"> تست آپلود هم بزن</label>
      <div style="margin-top:10px;font-size:10px;color:var(--dim)">
        سرور: <span style="color:var(--c);font-family:var(--font-mono)">speed.cloudflare.com</span>
        &nbsp;·&nbsp; Phase 3 از نتایج Phase 2 خودکار اجرا میشه یا می‌تونی از دکمه Results دستی اجرا کنی
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>⬡ FRAGMENT</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Mode</label>
          <select id="cfgFragMode" onchange="onFragModeChange(this.value)">
            <option value="manual">manual</option>
            <option value="auto">auto (اسکن خودکار بهترین حالت)</option>
            <option value="off">off</option>
          </select>
        </div>
        <div class="f-row" id="fragPktsRow"><label>Packets (manual)</label><input type="text" id="cfgFragPkts" value="tlshello"></div>
        <div></div>
        <div class="f-row" id="fragLenRow"><label>Length (manual)</label><input type="text" id="cfgFragLen" value="10-20"></div>
        <div class="f-row" id="fragIntRow"><label>Interval ms (manual)</label><input type="text" id="cfgFragInt" value="10-20"></div>
        <div></div>
      </div>
      <div id="fragAutoInfo" style="display:none;background:var(--gd);border:1px solid var(--g2);border-radius:var(--rad-xs);padding:10px;font-size:12px;color:var(--tx);font-family:var(--font-mono);margin-top:6px">
        <div style="color:var(--g);margin-bottom:6px">✦ Auto Mode — تست همه ۵ zone به صورت خودکار</div>
        <div style="color:var(--tx2);font-size:11px">zones: tlshello · 1-3 · 1-5 · 1-10 · random</div>
        <div style="margin-top:8px;display:flex;gap:8px;align-items:center;flex-wrap:wrap">
          <input type="text" id="fragAutoTestIP" placeholder="Test IP (اختیاری)" style="width:160px;font-size:11px">
          <button class="btn btn-sm" id="btnFragAuto" onclick="runFragmentAuto()">⚡ Run Auto Optimizer</button>
        </div>
        <div id="fragAutoResult" style="display:none;margin-top:8px;padding:8px;background:var(--bg3);border-radius:4px;font-size:11px"></div>
      </div>
      <div style="font-size:11px;color:var(--tx2);font-family:var(--font-mono);margin-top:4px" id="fragManualInfo">
        Manual: مقادیر رو خودت تنظیم کن. Auto: اسکنر خودش بهترین رو پیدا میکنه و apply می‌کنه.
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>▸ XRAY</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Log Level</label>
          <select id="cfgXrayLog">
            <option value="none">none</option>
            <option value="error">error</option>
            <option value="warning">warning</option>
            <option value="info">info</option>
            <option value="debug">debug</option>
          </select>
        </div>
        <div class="f-row"><label>Mux Concurrency (-1 = off)</label><input type="number" id="cfgMuxConc" value="-1"></div>
        <div style="display:flex;align-items:flex-end;padding-bottom:11px"><label class="chk-row"><input type="checkbox" id="cfgMuxEnabled"> Enable Mux</label></div>
      </div>
    </div>
  </div>
</div>

<!-- ══ IMPORT PAGE ══ -->
<div id="page-import" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Import Config</h2><p>یه یا چند تا vless/vmess/trojan لینک paste کن</p></div>
    <div class="phd-r" id="clearProxyBtn" style="display:none">
      <button class="btn btn-danger-real btn-sm" onclick="clearSavedProxy()">✕ Remove Config</button>
    </div>
  </div>
  <!-- Subscription Import -->
  <div class="card" style="margin-bottom:12px">
    <div class="card-hd">
      <div>🔗 Subscription URL</div>
      <span style="font-size:10px;color:var(--dim)">لینک سابسکریپشن رو وارد کن</span>
    </div>
    <div class="card-bd" style="padding:10px;display:flex;flex-direction:column;gap:8px">
      <div style="display:flex;gap:8px">
        <input type="text" id="subUrlInput" placeholder="https://sub.example.com/..." style="flex:1;font-family:var(--font-mono);font-size:11px">
        <button class="btn btn-primary-real" onclick="fetchSubscription()" id="btnFetchSub">↓ Fetch</button>
      </div>
      <div id="subStatus" style="font-size:10px;color:var(--dim);font-family:var(--font-mono);display:none"></div>
      <div id="subResults" style="display:none;max-height:260px;overflow-y:auto;display:flex;flex-direction:column;gap:6px"></div>
    </div>
  </div>
  <div class="card">
    <div class="card-hd">
      <div>⬡ Proxy Link(s)</div>
      <div style="display:flex;gap:8px;align-items:center">
        <label style="font-size:10px;color:var(--dim)"><input type="checkbox" id="multiImportMode" onchange="toggleMultiMode(this.checked)"> Multi-Import</label>
      </div>
    </div>
    <div class="card-bd">
      <div class="f-row">
        <label id="linkInputLabel">vless:// or vmess:// or trojan://</label>
        <textarea id="linkInput" rows="3" placeholder="vless://uuid@domain:443?..."></textarea>
      </div>
      <div style="display:flex;gap:8px;flex-wrap:wrap">
        <button class="btn btn-primary-real" onclick="parseLink()">▶ Parse & Save</button>
        <button class="btn btn-sm" id="btnMultiParse" style="display:none" onclick="parseMultiLinks()">⬡ Parse All Links</button>
      </div>
    </div>
  </div>
  <div id="parsedResult" style="display:none" class="card">
    <div class="card-hd"><div>✓ Config Parsed</div></div>
    <div class="card-bd">
      <div class="parsed-box" id="parsedBox"></div>
    </div>
  </div>
  <!-- Multi-import results -->
  <div id="multiResults" style="display:none">
    <div class="phd" style="padding:0 0 8px"><div class="phd-l"><h3 style="font-size:13px">Parsed Configs</h3></div></div>
    <div id="multiResultsList"></div>
  </div>
</div>

<!-- ══ SUBNETS PAGE ══ -->
<div id="page-subnets" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Subnet Intelligence</h2><p style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">بهترین subnet‌ها بر اساس نتایج اسکن</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="loadSubnets()">↺ Refresh</button>
    </div>
  </div>
  <div id="subnetList" style="display:flex;flex-direction:column;gap:8px;padding:16px"></div>
</div>

<!-- ══ MONITOR PAGE ══ -->
<div id="page-monitor" class="page">
  <div class="phd">
    <div class="phd-l"><h2>IP Health Monitor</h2><p style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">وضعیت live + latency graph</p></div>
    <div class="phd-r" style="gap:6px">
      <button class="btn btn-sm" onclick="checkAllNow()">⚡ Check Now</button>
      <button class="btn btn-sm" onclick="loadHealth()">↺ Refresh</button>
    </div>
  </div>
  <!-- Monitor Settings -->
  <div class="card" style="margin:0 16px 12px">
    <div class="card-hd">⚙ تنظیمات Monitor</div>
    <div class="card-bd" style="padding:12px">
      <div style="display:flex;gap:16px;flex-wrap:wrap;align-items:center">
        <label class="chk-row"><input type="checkbox" id="monitorEnabled" checked onchange="saveMonitorSettings()"> فعال</label>
        <div style="display:flex;align-items:center;gap:6px">
          <span style="font-size:11px;color:var(--dim);font-family:var(--font-mono)">هر</span>
          <input type="number" id="monitorInterval" value="3" min="1" max="60" style="width:60px" onchange="saveMonitorSettings()">
          <span style="font-size:11px;color:var(--dim);font-family:var(--font-mono)">دقیقه چک کن</span>
        </div>
        <label class="chk-row"><input type="checkbox" id="monitorTrafficDetect" onchange="saveMonitorSettings()"> Traffic Detect</label>
        <label class="chk-row"><input type="checkbox" id="monitorShowGraph" checked onchange="renderHealthList()"> نمایش Graph</label>
      </div>
    </div>
  </div>
  <div id="healthList" style="display:flex;flex-direction:column;gap:8px;padding:16px"></div>
  <div style="padding:0 16px 16px">
    <div class="card">
      <div class="card-hd">اضافه کردن IP به Monitor</div>
      <div class="card-bd" style="padding:10px;display:flex;gap:8px">
        <input type="text" id="monitorIPInput" placeholder="104.18.x.x" style="flex:1">
        <button class="btn" onclick="addToMonitor()">+ Add</button>
      </div>
    </div>
  </div>
</div>

<!-- ══ TEMPLATES PAGE ══ -->
<div id="page-templates" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Config Templates</h2><p style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">کانفیگ‌های ذخیره‌شده</p></div>
  </div>
  <div style="padding:16px">
    <div class="card" style="margin-bottom:14px">
      <div class="card-hd">ذخیره کانفیگ جدید</div>
      <div class="card-bd" style="padding:10px;display:flex;flex-direction:column;gap:8px">
        <input type="text" id="tmplName" placeholder="اسم کانفیگ (مثلاً: Fastly-DE)" style="width:100%">
        <textarea id="tmplURL" rows="2" placeholder="vless:// vmess:// trojan://"></textarea>
        <div style="display:flex;gap:8px">
          <button class="btn" style="flex:1" onclick="saveTemplate()">💾 Save Template</button>
        </div>
      </div>
    </div>
    <div id="templateList" style="display:flex;flex-direction:column;gap:8px"></div>
  </div>
</div>

<!-- ══ SYSTEM INFO PAGE ══ -->
<div id="page-sysinfo" class="page">
  <div class="phd">
    <div class="phd-l"><h2>System</h2><p style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">وضعیت سیستم و برنامه</p></div>
    <div class="phd-r"><button class="btn btn-sm" onclick="loadSysInfo()">↺ Refresh</button></div>
  </div>
  <div style="display:grid;grid-template-columns:repeat(3,1fr);gap:12px;padding:16px">
    <div class="card"><div class="card-hd">🧵 Goroutines</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siGoroutines" style="font-size:28px;font-weight:700;font-family:var(--font-mono);color:var(--c)">—</div></div></div>
    <div class="card"><div class="card-hd">🧠 Memory (MB)</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siMemMB" style="font-size:28px;font-weight:700;font-family:var(--font-mono);color:var(--g)">—</div><div id="siMemSys" style="font-size:10px;color:var(--dim);font-family:var(--font-mono)">sys: —</div></div></div>
    <div class="card"><div class="card-hd">♻ GC Cycles</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siGC" style="font-size:28px;font-weight:700;font-family:var(--font-mono);color:var(--y)">—</div></div></div>
    <div class="card"><div class="card-hd">⏱ Uptime</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siUptime" style="font-size:22px;font-weight:700;font-family:var(--font-mono);color:var(--tx)">—</div></div></div>
    <div class="card"><div class="card-hd">🔌 Active Ports</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siPorts" style="font-size:28px;font-weight:700;font-family:var(--font-mono);color:var(--o)">—</div></div></div>
    <div class="card"><div class="card-hd">📡 Scan Threads</div><div class="card-bd" style="padding:14px;text-align:center"><div id="siThreads" style="font-size:28px;font-weight:700;font-family:var(--font-mono);color:var(--c)">—</div></div></div>
  </div>
  <div style="padding:0 16px">
    <div class="card"><div class="card-hd">Build Info</div>
      <div class="card-bd" style="padding:12px;font-family:var(--font-mono);font-size:11px;color:var(--dim);display:flex;flex-direction:column;gap:4px">
        <div>Version: <span id="siVersion" style="color:var(--tx)">—</span></div>
        <div>Go Runtime: <span id="siGoVer" style="color:var(--tx)">—</span></div>
        <div>OS/Arch: <span id="siOS" style="color:var(--tx)">—</span></div>
        <div>Persist File: <span id="siPersistPath" style="color:var(--tx)">—</span></div>
      </div>
    </div>
  </div>
</div>

<!-- ══ TUI LOG ══ -->
<div id="page-tui" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Live Log</h2><p>All scanner events</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="clearTUI()">✕ Clear</button>
      <button class="btn btn-sm" id="btnAS" onclick="toggleAS()">↓ Auto-scroll</button>
    </div>
  </div>
  <div class="tui-wrap">
    <div class="tui-hd">
      <div class="tui-dots">
        <div class="tui-dot" style="background:#ff3366"></div>
        <div class="tui-dot" style="background:#ffd700"></div>
        <div class="tui-dot" style="background:#00ffaa"></div>
      </div>
      <span style="margin-left:8px;font-size:11px;color:var(--tx2)">piyazche — scanner log</span>
      <span id="tuiStatus" style="margin-left:auto;color:var(--dim);font-size:10px">idle</span>
    </div>
    <div class="tui-body" id="tuiBody">
      <div class="tui-line"><span class="tui-t">--:--:--</span><span class="tui-info">Piyazche ready<span class="cursor"></span></span></div>
    </div>
  </div>
</div>

</div><!-- /main -->
</div><!-- /app -->

<!-- ══ THEME PICKER MODAL ══ -->
<div class="theme-picker-overlay" id="themePickerOverlay" onclick="closeThemePickerOutside(event)">
  <div class="theme-picker">
    <div class="tp-title">
      <span>⬡ APPEARANCE</span>
      <button class="tp-close" onclick="closeThemePicker()">✕</button>
    </div>

    <div class="tp-section">
      <div class="tp-label">Theme</div>
      <div class="tp-themes">
        <button class="tp-theme-btn active" data-base="neon" onclick="selectThemeBase('neon',this)">
          <div class="tp-swatch">
            <span style="background:#38bfff"></span>
            <span style="background:#00ffaa"></span>
            <span style="background:#c060ff"></span>
          </div>
          NEON
        </button>
        <button class="tp-theme-btn" data-base="navy" onclick="selectThemeBase('navy',this)">
          <div class="tp-swatch">
            <span style="background:#4d8fff"></span>
            <span style="background:#00e676"></span>
            <span style="background:#ce93d8"></span>
          </div>
          NAVY
        </button>
        <button class="tp-theme-btn" data-base="warm" onclick="selectThemeBase('warm',this)">
          <div class="tp-swatch">
            <span style="background:#ff8c00"></span>
            <span style="background:#ffd966"></span>
            <span style="background:#5bc8ff"></span>
          </div>
          WARM
        </button>
      </div>
    </div>

    <div class="tp-section">
      <div class="tp-label">Mode</div>
      <div class="tp-mode">
        <button class="tp-mode-btn" data-mode="night" onclick="selectThemeMode('night',this)">
          <span class="tp-mode-icon">🌙</span>
          <span>NIGHT</span>
        </button>
        <button class="tp-mode-btn active" data-mode="day" onclick="selectThemeMode('day',this)">
          <span class="tp-mode-icon">☀️</span>
          <span>DAY</span>
        </button>
      </div>
    </div>
  </div>
</div>

<script>
// ══ STATE ══
let ws=null,p1Results=[],p2Results=[],shodanIPs=[],tuiAS=true,viewingSession=false;
let feedRows=[],maxFeedRows=100,currentTab='p2';
// localStorage key for history
const LS_HISTORY='pyz_history_v2';

// ══ THEME SYSTEM ══
let currentThemeBase = 'neon';  // 'neon' | 'navy' | 'warm'
let currentThemeMode = 'night'; // 'night' | 'day'
const LS_THEME = 'pyz_theme_v3';

const THEME_META = {
  neon:  { label:'NEON',  icon:{ night:'🌙', day:'☀️' } },
  navy:  { label:'NAVY',  icon:{ night:'🌙', day:'☀️' } },
  warm:  { label:'WARM',  icon:{ night:'🌙', day:'☀️' } },
};

function buildThemeId(base, mode){ return base + '-' + mode; }

function applyTheme(){
  const tid = buildThemeId(currentThemeBase, currentThemeMode);
  document.documentElement.setAttribute('data-theme', tid);
  const meta = THEME_META[currentThemeBase];
  const icon = currentThemeMode === 'night' ? '🌙' : '☀️';
  const el = document.getElementById('themeIcon');
  const tx = document.getElementById('themeTxt');
  if(el) el.textContent = icon;
  if(tx) tx.textContent = meta.label;
  localStorage.setItem(LS_THEME, JSON.stringify({ base: currentThemeBase, mode: currentThemeMode }));
  // sync picker buttons
  document.querySelectorAll('.tp-theme-btn').forEach(b => {
    b.classList.toggle('active', b.dataset.base === currentThemeBase);
  });
  document.querySelectorAll('.tp-mode-btn').forEach(b => {
    b.classList.toggle('active', b.dataset.mode === currentThemeMode);
  });
}

function selectThemeBase(base, btn){
  currentThemeBase = base;
  document.querySelectorAll('.tp-theme-btn').forEach(b => b.classList.remove('active'));
  if(btn) btn.classList.add('active');
  applyTheme();
}

function selectThemeMode(mode, btn){
  currentThemeMode = mode;
  document.querySelectorAll('.tp-mode-btn').forEach(b => b.classList.remove('active'));
  if(btn) btn.classList.add('active');
  applyTheme();
}

function openThemePicker(){
  document.getElementById('themePickerOverlay').classList.add('open');
}
function closeThemePicker(){
  document.getElementById('themePickerOverlay').classList.remove('open');
}
function closeThemePickerOutside(e){
  if(e.target === document.getElementById('themePickerOverlay')) closeThemePicker();
}

// load saved theme
(function(){
  try {
    const saved = JSON.parse(localStorage.getItem(LS_THEME));
    if(saved && saved.base) currentThemeBase = saved.base;
    if(saved && saved.mode) currentThemeMode = saved.mode;
  } catch(e){}
  // legacy migration
  const oldTheme = localStorage.getItem('pyz_theme_v2');
  if(oldTheme && !localStorage.getItem(LS_THEME)){
    currentThemeMode = oldTheme === 'day' ? 'day' : 'night';
  }
  applyTheme();
})();

// ══ NAV ══
function nav(page,btn){
  document.querySelectorAll('.page').forEach(p=>p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(b=>b.classList.remove('active'));
  document.getElementById('page-'+page).classList.add('active');
  if(btn) btn.classList.add('active');
  else{const b=document.querySelector('[data-page="'+page+'"]');if(b)b.classList.add('active');}
  if(page==='results') refreshResults();
  if(page==='history') refreshHistory();
}

// ══ TABS ══
function switchTab(tab,btn){
  currentTab=tab;
  document.querySelectorAll('.tab').forEach(t=>t.classList.remove('active'));
  if(btn) btn.classList.add('active');
  document.getElementById('tab-p2').style.display=tab==='p2'?'':'none';
  document.getElementById('tab-p1').style.display=tab==='p1'?'':'none';
}

// ══ WS ══
function connectWS(){
  const proto=location.protocol==='https:'?'wss':'ws';
  ws=new WebSocket(proto+'://'+location.host+'/ws');
  ws.onmessage=e=>{try{handleWS(JSON.parse(e.data));}catch(err){}};
  ws.onclose=()=>setTimeout(connectWS,2000);
}

function handleWS(msg){
  const{type,payload}=msg;
  switch(type){
    case 'status': setStatus(payload.status,payload.phase); break;
    case 'progress': onProgress(payload); break;
    case 'live_ip': addFeedRow(payload.ip,'scan'); break;
    case 'ip_result':
      if(payload.success) addFeedRow(payload.ip+' · '+payload.latency+'ms','ok');
      break;
    case 'tui': appendTUI(payload); break;
    case 'phase2_start':
      setStatus('scanning','phase2');
      addFeedRow('◈ Phase 2 starting — '+payload.count+' IPs','p2');
      // Progress bar reset for phase2
      document.getElementById('progBar').classList.add('p2');
      document.getElementById('progBar').style.width='0%';
      document.getElementById('progPct').textContent='0%';
      document.getElementById('progTxt').textContent='0 / '+payload.count+' (P2)';
      setStatValue('stTotal',payload.count,'var(--p)');
      setStatValue('stDone',0,'var(--c)');
      setStatValue('stPass',0,'var(--g)');
      break;
    case 'phase2_progress':
      onPhase2Progress(payload);
      if(!viewingSession){renderP2();updatePassedChips();}
      break;
    case 'phase2_done':
      handlePhase2Done(payload);
      break;
    case 'scan_done':
      setStatus('done','');
      addFeedRow('✓ Scan complete — '+payload.passed+' passed','ok');
      if(!viewingSession){refreshResults();}
      saveSessionToHistory(payload);
      refreshHistory();
      setTimeout(syncHistoryToServer, 1500); // sync to server after results fetched
      break;
    case 'health_update':
      handleHealthUpdate(payload);
      break;
    case 'health_error':
      handleHealthError(payload);
      break;
    case 'phase3_start':
      addFeedRow('🚀 Phase 3 (Speed Test) شروع شد — '+payload.count+' IP','p2');
      break;
    case 'phase3_progress':
      addFeedRow('⚡ '+payload.ip+' ↓'+(payload.dl||0).toFixed(1)+'M','ok');
      break;
    case 'phase3_done':
      if(payload.results){
        payload.results.forEach(r=>{
          const ex=p2Results.findIndex(x=>x.IP===r.IP);
          if(ex>=0){p2Results[ex].DownloadMbps=r.DownloadMbps;p2Results[ex].UploadMbps=r.UploadMbps;}
        });
        renderP2();
      }
      addFeedRow('✓ Phase 3 تموم شد','ok');
      break;
    case 'fragment_auto_start':
      addFeedRow('🔍 Fragment auto-optimizer started for '+payload.testIp,'info');
      break;
    case 'fragment_auto_done':{
      const btn=document.getElementById('btnFragAuto');
      const res=document.getElementById('fragAutoResult');
      if(btn){btn.disabled=false;btn.textContent='⚡ Run Auto Optimizer';}
      if(payload.error){
        if(res){res.style.color='var(--r)';res.textContent='✗ '+payload.error;}
        addFeedRow('✗ Fragment auto failed: '+payload.error,'err');
      } else if(payload.best){
        const b=payload.best;
        const txt='✓ Best: zone='+b.zone+' size='+b.sizeRange+' interval='+b.intervalRange+' ('+b.latencyMs+'ms)';
        if(res){res.style.color='var(--g)';res.textContent=txt+' — applied!';}
        addFeedRow(txt,'ok');
        // Update UI fields
        if(document.getElementById('cfgFragMode')){
          document.getElementById('cfgFragMode').value='manual';
          onFragModeChange('manual');
        }
        if(document.getElementById('cfgFragPkts')) document.getElementById('cfgFragPkts').value=b.zone;
        if(document.getElementById('cfgFragLen')) document.getElementById('cfgFragLen').value=b.sizeRange;
        if(document.getElementById('cfgFragInt')) document.getElementById('cfgFragInt').value=b.intervalRange;
      } else {
        if(res){res.style.color='var(--r)';res.textContent='✗ هیچ تنظیم fragment ای کار نکرد';}
        addFeedRow('✗ Fragment auto: no working config found','warn');
      }
      break;}
    case 'error': appendTUI({t:now(),l:'err',m:payload.message}); break;
    case 'shodan_done':
      shodanIPs=payload.ips||[];
      appendTUI({t:now(),l:'ok',m:'Shodan: '+shodanIPs.length+' IPs found'});
      break;
  }
}

function now(){return new Date().toLocaleTimeString('en-US',{hour12:false});}

// ══ LIVE FEED ══
function addFeedRow(txt,type){
  const body=document.getElementById('liveFeed');
  const div=document.createElement('div');
  div.className='live-row live-row-'+type;
  div.innerHTML='<span style="color:var(--dim);font-size:9px;font-family:var(--font-mono);flex-shrink:0">'+now()+'</span><span>'+txt+'</span>';
  body.insertBefore(div,body.firstChild);
  feedRows.push(div);
  if(feedRows.length>maxFeedRows){const old=feedRows.shift();if(old.parentNode)old.parentNode.removeChild(old);}
  const c=document.getElementById('feedCount2');
  if(c) c.textContent=feedRows.length+' events';
}

// ══ TUI ══
function appendTUI(entry){
  const body=document.getElementById('tuiBody');
  const div=document.createElement('div');div.className='tui-line';
  const cls='tui-'+(entry.l==='ok'?'ok':entry.l==='err'?'err':entry.l==='warn'?'warn':entry.l==='p2'?'p2':entry.l==='scan'?'scan':'info');
  div.innerHTML='<span class="tui-t">'+entry.t+'</span><span class="'+cls+'">'+entry.m+'</span>';
  body.appendChild(div);
  if(tuiAS) body.scrollTop=body.scrollHeight;
  const st=document.getElementById('tuiStatus');if(st)st.textContent=entry.m.substring(0,50);
}
function clearTUI(){document.getElementById('tuiBody').innerHTML='';}
function toggleAS(){tuiAS=!tuiAS;document.getElementById('btnAS').textContent=tuiAS?'↓ Auto-scroll':'— Manual';}

// ══ STATUS ══
function setStatus(st,phase){
  const dot=document.getElementById('sDot'),txt=document.getElementById('sTxt'),ph=document.getElementById('sPhase');
  const pdot=document.getElementById('pDot');
  const scan=document.getElementById('nbScan');
  dot.className='dot dot-'+(st==='scanning'?'scan':st==='done'?'done':st==='paused'?'warn':'idle');
  if(pdot) pdot.className='dot dot-'+(st==='scanning'?'scan':st==='done'?'done':'idle');
  txt.textContent=st;
  ph.textContent=phase?'· '+phase:'';
  if(scan) scan.style.display=st==='scanning'?'':'none';
  document.getElementById('btnStart').style.display=st==='scanning'||st==='paused'?'none':'';
  document.getElementById('btnStop').style.display=st==='scanning'||st==='paused'?'':'none';
  if(st==='idle'){
    document.getElementById('progBar').style.width='0%';
    document.getElementById('progPct').textContent='0%';
    document.getElementById('progTxt').textContent='0 / 0';
    document.getElementById('tbProgress').textContent='';
  }
  if(st==='scanning'&&phase==='phase2') document.getElementById('progBar').classList.add('p2');
  else if(st!=='scanning') document.getElementById('progBar').classList.remove('p2');

  // Animate stat numbers when active
  ['stTotal','stDone','stPass','stFail'].forEach(id=>{
    const el=document.getElementById(id);
    if(el) el.classList.toggle('active',st==='scanning');
  });
}

// ══ PROGRESS ══
function onProgress(p){
  // Update charts
  const pct=p.total>0?Math.round(p.done/p.total*100):0;
  document.getElementById('progBar').style.width=pct+'%';
  document.getElementById('progPct').textContent=pct+'%';
  document.getElementById('progTxt').textContent=(p.done||0)+' / '+(p.total||0);
  // FIX: properly update stat cards with numbers
  setStatValue('stTotal',p.total||'—','var(--tx2)');
  setStatValue('stDone',p.done||0,'var(--c)');
  setStatValue('stPass',p.succeeded||p.passed||0,'var(--g)');
  setStatValue('stFail',p.failed||0,'var(--r)');
  if(p.eta) setStatValue('stETA',p.eta,'var(--y)');
  document.getElementById('tbProgress').textContent=(p.done||0)+'/'+(p.total||0)+' · '+pct+'%';
  if(p.rate>0) document.getElementById('progRate').textContent=(p.rate||0).toFixed(1)+' IP/s';
}

function setStatValue(id,val,color){
  const el=document.getElementById(id);
  if(!el) return;
  el.textContent=val;
  if(color) el.style.color=color;
}

// ══ SCAN ══
async function startScan(){
  const ipInput=document.getElementById('ipInput').value.trim();
  if(!ipInput){appendTUI({t:now(),l:'warn',m:'No IP ranges entered'});return;}
  const maxIPs=parseInt(document.getElementById('maxIPInput').value)||0;
  const quickSettings={
    threads:parseInt(document.getElementById('qThreads').value)||200,
    timeout:parseInt(document.getElementById('qTimeout').value)||8,
    maxLatency:parseInt(document.getElementById('qMaxLat').value)||3500,
    stabilityRounds:parseInt(document.getElementById('qRounds').value)||3,
    sampleSize:parseInt(document.getElementById('sampleSize').value)||1,
  };
  const btn=document.getElementById('btnStart');
  btn.disabled=true;
  viewingSession=false;
  const b=document.getElementById('sessionBanner');if(b)b.remove();
  p1Results=[];p2Results=[];
  feedRows=[];
  document.getElementById('liveFeed').innerHTML='<div class="live-row live-row-scan"><span style="color:var(--dim)">›</span><span>Scan started...</span></div>';
  document.getElementById('progBar').classList.remove('p2');
  // Reset stat cards
  setStatValue('stTotal','—','var(--tx2)');
  setStatValue('stDone',0,'var(--c)');
  setStatValue('stPass',0,'var(--g)');
  setStatValue('stFail',0,'var(--r)');
  setStatValue('stETA','—','var(--y)');
  const res=await fetch('/api/scan/start',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({quickSettings,ipRanges:ipInput,maxIPs})});
  const data=await res.json();
  btn.disabled=false;
  if(!data.ok){appendTUI({t:now(),l:'err',m:'Error: '+data.error});return;}
  setStatus('scanning','phase1');
  appendTUI({t:now(),l:'ok',m:'▶ Scan started — '+data.total+' IPs'});
}

async function stopScan(){
  await fetch('/api/scan/stop',{method:'POST'});
  setStatus('idle','');
}

async function pauseScan(){
  const res=await fetch('/api/scan/pause',{method:'POST'});
  const d=await res.json();
  if(!d.ok){appendTUI({t:now(),l:'warn',m:d.error||'cannot pause now'});return;}
  if(d.message==='paused') {setStatus('paused','phase1');appendTUI({t:now(),l:'warn',m:'⏸ Scan paused — press Resume to continue'});}
  else if(d.message==='resumed') {setStatus('scanning','phase1');appendTUI({t:now(),l:'ok',m:'▶ Scan resumed'});}
}

// ══ RESULTS ══
function refreshResults(){
  if(viewingSession) return;
  fetch('/api/results').then(r=>r.json()).then(data=>{
    p1Results=data.phase1||[];
    p2Results=data.phase2||[];
    renderP1();renderP2();updatePassedChips();
  });
}

function updatePassedChips(){
  const passed=(p2Results||[]).filter(r=>r.Passed);
  document.getElementById('resSummary').textContent=passed.length+' passed out of '+(p2Results||[]).length+' tested';
  document.getElementById('passedBadge').textContent=passed.length;
  document.getElementById('nbResults').textContent=passed.length;
  const chips=document.getElementById('ipChips');
  if(!passed.length){chips.innerHTML='<span style="color:var(--dim);font-size:12px">No results</span>';return;}
  chips.innerHTML=passed.filter(r=>r.Passed||r.passed||r.success).map(r=>{
    const grade=scoreToGrade(r.StabilityScore||0);
    const gc=gradeColor(grade);
    const dl=r.DownloadMbps>0?' ↓'+r.DownloadMbps.toFixed(1):'';
    return '<div class="ip-chip" data-ip="'+r.IP+'" data-action="copyvless" title="Copy vless link">'+
      '<span style="color:'+gc+';font-family:var(--font-mono);font-weight:700;font-size:9px;margin-right:4px">'+grade+'</span>'+
      r.IP+'<span class="lat">'+Math.round(r.AvgLatencyMs)+'ms'+dl+'</span></div>';
  }).join('');
  // event delegation
  const chipsEl=document.getElementById('ipChips');
  if(chipsEl) chipsEl.onclick=function(e){
    const chip=e.target.closest('[data-action]');
    if(!chip) return;
    copyWithIP(chip.dataset.ip);
  };
}

function scoreToGrade(s){
  if(s>=92)return'A+';if(s>=85)return'A';if(s>=75)return'B+';
  if(s>=65)return'B';if(s>=55)return'C+';if(s>=45)return'C';
  if(s>=35)return'D';return'F';
}
function gradeColor(g){
  if(g==='A+'||g==='A')return'var(--g)';
  if(g==='B+'||g==='B')return'var(--c)';
  if(g==='C+'||g==='C')return'var(--y)';
  if(g==='D')return'var(--o)';
  return'var(--r)';
}

function renderP2(){
  // Delegate to filter/sort which also handles rendering
  applyP2Filter();
}

function renderP1(){
  const tbody=document.getElementById('p1Tbody');
  const succ=(p1Results||[]).filter(r=>r.success||r.Success);
  document.getElementById('p1CountBadge').textContent=(p1Results||[]).length+' IPs scanned · '+succ.length+' passed';
  if(!succ.length){
    tbody.innerHTML='<tr><td colspan="6" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr>';
    return;
  }
  tbody.innerHTML=succ.map((r,i)=>{
    const ip=r.ip||r.IP||'';const lat=r.latency_ms||r.LatencyMs||0;
    const lc=lat<=500?'var(--g)':lat<=1500?'var(--y)':'var(--r)';
    // Phase 1 packet loss (may not be available, show — if 0)
    const pl=r.packet_loss_pct||r.PacketLossPct||0;
    const plTxt=pl>0?pl.toFixed(0)+'%':'—';
    const plc=pl<=0?'var(--dim)':pl<=5?'var(--g)':pl<=20?'var(--y)':'var(--r)';
    return '<tr class="p1-row">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700;font-family:var(--font-mono)">'+ip+'</td>'+
      '<td style="color:'+lc+';font-family:var(--font-mono)">'+Math.round(lat)+'ms</td>'+
      '<td style="color:'+plc+';font-family:var(--font-mono)">'+plTxt+'</td>'+
      '<td><span class="badge bg">OK</span></td>'+
      '<td><button class="copy-btn" onclick="copyIP(\''+ip+'\')" title="Copy IP">⎘</button></td>'+
    '</tr>';
  }).join('');
  if(p1Results.length>succ.length) tbody.innerHTML+='<tr><td colspan="6" style="text-align:center;color:var(--dim);padding:10px;font-size:10px">'+(p1Results.length-succ.length)+' failed IPs hidden</td></tr>';
}

function copyAllPassed(){
  const passed=(p2Results||[]).filter(r=>r.Passed).map(r=>r.IP);
  if(!passed.length) return;
  navigator.clipboard.writeText(passed.join('\n')).then(()=>appendTUI({t:now(),l:'ok',m:'⎘ '+passed.length+' IPs copied'}));
}

// ══ HISTORY — localStorage based ══
function saveSessionToHistory(payload){
  // Also fetch current results to store
  fetch('/api/results').then(r=>r.json()).then(data=>{
    let history=loadHistory();
    const session={
      id:Date.now().toString(),
      startedAt:new Date().toISOString(),
      totalIPs:payload.total||0,
      passed:payload.passed||0,
      duration:payload.duration||'',
      results:data.phase2||[],
      p1Results:data.phase1||[]
    };
    history.unshift(session);
    // Keep last 50 sessions
    if(history.length>50) history=history.slice(0,50);
    try{localStorage.setItem(LS_HISTORY,JSON.stringify(history));}catch(e){console.warn('History save error',e);}
    updateHistoryBadge(history.length);
  }).catch(()=>{
    // Save minimal info if results fetch fails
    let history=loadHistory();
    const session={
      id:Date.now().toString(),
      startedAt:new Date().toISOString(),
      totalIPs:payload.total||0,
      passed:payload.passed||0,
      duration:payload.duration||'',
      results:[],p1Results:[]
    };
    history.unshift(session);
    if(history.length>50) history=history.slice(0,50);
    try{localStorage.setItem(LS_HISTORY,JSON.stringify(history));}catch(e){}
    updateHistoryBadge(history.length);
  });
}

function loadHistory(){
  try{
    const raw=localStorage.getItem(LS_HISTORY);
    return raw?JSON.parse(raw):[];
  }catch(e){return [];}
}

function updateHistoryBadge(count){
  const el=document.getElementById('nbHistory');
  if(el) el.textContent=count>0?count:'';
}

function refreshHistory(){
  // Merge server sessions (disk-persistent) + localStorage sessions
  fetch('/api/sessions').then(r=>r.json()).then(srvSessions=>{
    let localHistory=loadHistory();
    // Merge: add any server sessions not in local (from previous runs)
    const localIDs=new Set(localHistory.map(s=>s.id));
    const merged=[...localHistory];
    for(const ss of(srvSessions||[])){
      if(!localIDs.has(ss.id)){
        merged.push(ss);
      }
    }
    // Sort by date desc
    merged.sort((a,b)=>new Date(b.startedAt)-new Date(a.startedAt));
    // Update localStorage with merged
    if(merged.length>localHistory.length){
      try{localStorage.setItem(LS_HISTORY,JSON.stringify(merged.slice(0,50)));}catch(e){}
    }
    renderHistoryList(merged);
  }).catch(()=>{
    renderHistoryList(loadHistory());
  });
}

function renderHistoryList(sessions){
  const el=document.getElementById('histList');
  updateHistoryBadge(sessions.length);
  if(!sessions||!sessions.length){
    el.innerHTML='<p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p>';
    return;
  }
  el.innerHTML=sessions.map(s=>{
    const passed=s.passed||(s.results||[]).filter(r=>r.Passed).length;
    const d=new Date(s.startedAt);
    const total=s.totalIPs||(s.results||[]).length;
    return '<div class="hist-item" onclick="showSession(\''+s.id+'\',\'local\')">'+
      '<div class="hist-n" style="color:'+(passed>0?'var(--g)':'var(--dim)')+'">'+passed+'</div>'+
      '<div class="hist-info">'+
        '<div style="color:var(--tx);font-weight:600;font-size:12px">'+total+' IPs · '+passed+' passed</div>'+
        '<div class="hist-date">'+d.toLocaleString()+(s.duration?' · '+s.duration:'')+'</div>'+
      '</div>'+
      '<div style="color:var(--dim);font-size:10px">▶</div>'+
    '</div>';
  }).join('');
}

function showSession(id,source){
  let sessions=loadHistory();
  let s=sessions.find(x=>x.id===id);
  if(!s){
    // Try server
    fetch('/api/sessions').then(r=>r.json()).then(srvSessions=>{
      const ss=srvSessions.find(x=>x.id===id);
      if(ss) _showSessionData(ss);
    });
    return;
  }
  _showSessionData(s);
}

function _showSessionData(s){
  viewingSession=true;
  p2Results=s.results||[];p1Results=s.p1Results||[];
  renderP2();renderP1();updatePassedChips();
  nav('results');
  const existing=document.getElementById('sessionBanner');if(existing)existing.remove();
  const banner=document.createElement('div');
  banner.id='sessionBanner';banner.className='session-banner';
  const d=new Date(s.startedAt);
  const passed=s.passed||(s.results||[]).filter(r=>r.Passed).length;
  banner.innerHTML='<span>📂 Viewing: '+d.toLocaleString()+' — '+passed+' passed</span>'+
    '<button onclick="clearSession()" style="background:var(--rd);border:1px solid var(--r);color:var(--r);padding:3px 10px;cursor:pointer;border-radius:3px;font-size:11px;font-family:var(--font-mono)">✕ Back to live</button>';
  document.getElementById('page-results').insertBefore(banner,document.getElementById('page-results').firstChild);
}

function clearSession(){
  viewingSession=false;
  const b=document.getElementById('sessionBanner');if(b)b.remove();
  refreshResults();
}

function clearHistory(){
  if(!confirm('Clear all history?')) return;
  localStorage.removeItem(LS_HISTORY);
  updateHistoryBadge(0);
  refreshHistory();
}

// ══ CONFIG ══
function onFragModeChange(val){
  const isAuto=val==='auto';
  const isOff=val==='off';
  document.getElementById('fragPktsRow').style.opacity=(isAuto||isOff)?'.4':'1';
  document.getElementById('fragLenRow').style.opacity=(isAuto||isOff)?'.4':'1';
  document.getElementById('fragIntRow').style.opacity=(isAuto||isOff)?'.4':'1';
  document.getElementById('fragAutoInfo').style.display=isAuto?'':'none';
  document.getElementById('fragManualInfo').style.display=isAuto?'none':'';
}

async function runFragmentAuto(){
  const testIP=document.getElementById('fragAutoTestIP')?.value?.trim()||'';
  const btn=document.getElementById('btnFragAuto');
  if(btn){btn.disabled=true;btn.textContent='⏳ Running...';}
  const res=document.getElementById('fragAutoResult');
  if(res){res.style.display='';res.style.color='var(--y)';res.textContent='⏳ Testing all zones (tlshello · 1-3 · 1-5 · 1-10 · random)...';}
  try{
    const r=await fetch('/api/fragment/auto',{method:'POST',headers:{'Content-Type':'application/json'},
      body:JSON.stringify({testIp:testIP})});
    const d=await r.json();
    if(!d.ok){
      if(res){res.style.color='var(--r)';res.textContent='✗ '+d.error;}
    } else {
      if(res){res.textContent='⏳ Optimizing fragments across all zones...';}
    }
  }catch(e){
    if(res){res.style.color='var(--r)';res.textContent='✗ Error: '+e.message;}
    if(btn){btn.disabled=false;btn.textContent='⚡ Run Auto Optimizer';}
  }
}

function saveConfig(){
  const scanCfg={
    scan:{
      threads:parseInt(document.getElementById('cfgThreads').value)||200,
      timeout:parseInt(document.getElementById('cfgTimeout').value)||8,
      maxLatency:parseInt(document.getElementById('cfgMaxLat').value)||3500,
      retries:parseInt(document.getElementById('cfgRetries').value)||2,
      maxIPs:parseInt(document.getElementById('cfgMaxIPs').value)||0,
      sampleSize:parseInt(document.getElementById('cfgSampleSize').value)||1,
      testUrl:document.getElementById('cfgTestURL').value,
      shuffle:document.getElementById('cfgShuffle').checked,
      stabilityRounds:parseInt(document.getElementById('cfgRounds').value)||3,
      stabilityInterval:parseInt(document.getElementById('cfgInterval').value)||5,
      packetLossCount:parseInt(document.getElementById('cfgPLCount').value)||5,
      maxPacketLossPct:parseFloat(document.getElementById('cfgMaxPL').value),
      minDownloadMbps:parseFloat(document.getElementById('cfgMinDL').value)||0,
      minUploadMbps:parseFloat(document.getElementById('cfgMinUL').value)||0,
      jitterTest:document.getElementById('cfgJitter').checked,
    },
    fragment:{
      mode:document.getElementById('cfgFragMode').value,
      packets:document.getElementById('cfgFragPkts').value,
      manual:{length:document.getElementById('cfgFragLen').value,interval:document.getElementById('cfgFragInt').value}
    },
    xray:{
      logLevel:document.getElementById('cfgXrayLog').value,
      mux:{enabled:document.getElementById('cfgMuxEnabled').checked,concurrency:parseInt(document.getElementById('cfgMuxConc').value)||-1}
    },
    phase3:{
      enabled:document.getElementById('cfgP3Enabled').checked,
      downloadUrl:document.getElementById('cfgDLURL').value,
      uploadUrl:document.getElementById('cfgULURL').value,
      testUpload:document.getElementById('cfgP3Upload').checked,
      minDlMbps:parseFloat(document.getElementById('cfgMinDL').value)||0,
      minUlMbps:parseFloat(document.getElementById('cfgMinUL').value)||0,
    },
  };
  // Sync quick panel
  document.getElementById('qThreads').value=scanCfg.scan.threads;
  document.getElementById('qTimeout').value=scanCfg.scan.timeout;
  document.getElementById('qMaxLat').value=scanCfg.scan.maxLatency;
  document.getElementById('qRounds').value=scanCfg.scan.stabilityRounds;
  document.getElementById('sampleSize').value=scanCfg.scan.sampleSize;
  fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({scanConfig:JSON.stringify(scanCfg)})}).then(()=>{
    appendTUI({t:now(),l:'ok',m:'✓ Settings saved to disk'});
    updateConfigSummary(scanCfg.scan,scanCfg.fragment);
    nav('scan');
  });
}

function updateConfigSummary(s,f){
  const el=document.getElementById('configSummary');
  const frag=f?(' · frag:'+f.mode):'';
  el.innerHTML='threads:'+s.threads+' · timeout:'+s.timeout+'s · maxLat:'+s.maxLatency+'ms · rounds:'+s.stabilityRounds+frag+(s.speedTest?' · speed:ON':'');
}

// ══ IMPORT ══
async function parseLink(){
  const input=document.getElementById('linkInput').value.trim();if(!input)return;
  const res=await fetch('/api/config/parse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})});
  const data=await res.json();
  if(!data.ok){appendTUI({t:now(),l:'err',m:'Error: '+data.error});return;}
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:data.config})});
  const p=data.parsed;
  updateProxyChip(p.address,p.method,p.type);
  document.getElementById('parsedBox').innerHTML=
    '<span class="k">uuid: </span><span class="v">'+maskUUID(p.uuid)+'</span><br>'+
    '<span class="k">address: </span><span class="v">'+p.address+'</span><br>'+
    '<span class="k">port: </span><span class="v">'+p.port+'</span><br>'+
    '<span class="k">type: </span><span class="v">'+p.type+'</span><br>'+
    '<span class="k">method: </span><span class="v">'+p.method+'</span>'+
    (p.sni?'<br><span class="k">sni: </span><span class="v">'+p.sni+'</span>':'')+
    (p.path?'<br><span class="k">path: </span><span class="v">'+p.path+'</span>':'')+
    (p.fp?'<br><span class="k">fingerprint: </span><span class="v">'+p.fp+'</span>':'');
  document.getElementById('parsedResult').style.display='block';
  appendTUI({t:now(),l:'ok',m:'✓ Config: '+p.address+' ('+p.method+'/'+p.type+')'});
  // detect provider and update quick ranges
  activeProvider=detectProvider(input);
  renderQuickRanges(activeProvider);
}

function updateProxyChip(addr,method,type){
  document.getElementById('proxyChipTxt').textContent=addr+' · '+method+'/'+type;
  document.getElementById('proxyChip').style.display='inline-flex';
  document.getElementById('clearProxyBtn').style.display='';
  document.getElementById('configSummary').innerHTML='⬡ '+addr+' ('+method+'/'+type+')';
}

async function clearSavedProxy(){
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:''})});
  document.getElementById('proxyChip').style.display='none';
  document.getElementById('clearProxyBtn').style.display='none';
  document.getElementById('parsedResult').style.display='none';
  document.getElementById('configSummary').textContent='No config — import a proxy link first';
  appendTUI({t:now(),l:'warn',m:'Proxy config removed'});
}

function maskUUID(u){return!u||u.length<8?u:u.slice(0,8)+'••••••••';}

// ══ COPY / EXPORT ══
function exportResults(f){window.location.href='/api/results/export?format='+f;}
function copyIP(ip){
  navigator.clipboard.writeText(ip).then(()=>appendTUI({t:now(),l:'ok',m:'⎘ '+ip})).catch(()=>{
    const el=document.createElement('textarea');el.value=ip;document.body.appendChild(el);el.select();document.execCommand('copy');document.body.removeChild(el);
    appendTUI({t:now(),l:'ok',m:'⎘ '+ip});
  });
}
function copyWithIP(newIP){
  // server-side build-link — درست‌تر از regex، vmess هم پشتیبانی می‌کنه
  fetch('/api/config/build-link',{
    method:'POST',
    headers:{'Content-Type':'application/json'},
    body:JSON.stringify({ip:newIP})
  }).then(r=>r.json()).then(d=>{
    if(d.link){
      navigator.clipboard.writeText(d.link)
        .then(()=>appendTUI({t:now(),l:'ok',m:'⬡ Link with '+newIP+' copied'}));
    } else {
      // fallback: فقط IP رو copy کن
      appendTUI({t:now(),l:'warn',m:'⚠ build-link: '+(d.error||'no link')});
      copyIP(newIP);
    }
  }).catch(()=>{
    // اگه server نبود، regex fallback
    const rawLink=document.getElementById('linkInput').value.trim();
    if(!rawLink){copyIP(newIP);return;}
    try{
      const updated=rawLink.replace(/(@)([^:@\/?#\[\]]+)(:\d+)/,'$1'+newIP+'$3');
      navigator.clipboard.writeText(updated).then(()=>appendTUI({t:now(),l:'ok',m:'⬡ Link (regex) with '+newIP+' copied'}));
    }catch(e){copyIP(newIP);}
  });
}

// ══ SETTINGS LOAD ══
function loadSavedSettings(){
  fetch('/api/config/load').then(r=>r.json()).then(d=>{
    if(d.hasProxy){
      try{
        const pc=JSON.parse(d.proxyConfig);
        const addr=pc.proxy?.address||'';
        if(addr) updateProxyChip(addr,pc.proxy?.method||'tls',pc.proxy?.type||'ws');
      }catch(e){}
      // rawUrl رو در linkInput بریز تا copy-with-IP کار کنه
      if(d.rawUrl) document.getElementById('linkInput').value=d.rawUrl;
    }
    if(d.scanConfig){
      try{
        const sc=JSON.parse(d.scanConfig);
        const s=sc.scan||{},f=sc.fragment||{},x=sc.xray||{};
        const sv=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        const sc2=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.checked=!!v;};
        const ss=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        if(s.threads!=null){sv('cfgThreads',s.threads);sv('qThreads',s.threads);}
        if(s.timeout!=null){sv('cfgTimeout',s.timeout);sv('qTimeout',s.timeout);}
        if(s.maxLatency!=null){sv('cfgMaxLat',s.maxLatency);sv('qMaxLat',s.maxLatency);}
        if(s.retries!=null) sv('cfgRetries',s.retries);
        if(s.maxIPs!=null) sv('cfgMaxIPs',s.maxIPs);
        if(s.sampleSize!=null){sv('cfgSampleSize',s.sampleSize);sv('sampleSize',s.sampleSize);}
        if(s.testUrl) sv('cfgTestURL',s.testUrl);
        if(s.shuffle!=null) sc2('cfgShuffle',s.shuffle);
        if(s.stabilityRounds!=null){sv('cfgRounds',s.stabilityRounds);sv('qRounds',s.stabilityRounds);}
        if(s.stabilityInterval!=null) sv('cfgInterval',s.stabilityInterval);
        if(s.packetLossCount!=null) sv('cfgPLCount',s.packetLossCount);
        if(s.maxPacketLossPct!=null) sv('cfgMaxPL',s.maxPacketLossPct);
        if(s.minDownloadMbps!=null) sv('cfgMinDL',s.minDownloadMbps);
        if(s.minUploadMbps!=null) sv('cfgMinUL',s.minUploadMbps);
        if(s.jitterTest!=null) sc2('cfgJitter',s.jitterTest);
        if(f.mode){ss('cfgFragMode',f.mode);onFragModeChange(f.mode);}
        if(f.packets) sv('cfgFragPkts',f.packets);
        if(f.manual?.length) sv('cfgFragLen',f.manual.length);
        if(f.manual?.interval) sv('cfgFragInt',f.manual.interval);
        if(x.logLevel) ss('cfgXrayLog',x.logLevel);
        if(x.mux?.concurrency!=null) sv('cfgMuxConc',x.mux.concurrency);
        if(x.mux?.enabled!=null) sc2('cfgMuxEnabled',x.mux.enabled);
        const p3=sc.phase3||{};
        if(p3.enabled!=null){sc2('cfgP3Enabled',p3.enabled);document.getElementById('p3Settings').style.display=p3.enabled?'':'none';}
        if(p3.downloadUrl) sv('cfgDLURL',p3.downloadUrl);
        if(p3.uploadUrl) sv('cfgULURL',p3.uploadUrl);
        if(p3.testUpload!=null) sc2('cfgP3Upload',p3.testUpload);
        if(p3.minDlMbps!=null) sv('cfgMinDL',p3.minDlMbps);
        if(p3.minUlMbps!=null) sv('cfgMinUL',p3.minUlMbps);
        updateConfigSummary(s,f);
      }catch(e){console.warn('load err',e);}
    }
    fetch('/api/tui/stream').then(r=>r.json()).then(data=>{
      if(data.lines) data.lines.forEach(l=>{try{appendTUI(JSON.parse(l));}catch(e){}});
    }).catch(()=>{});
  });
  // Load history badge
  const hist=loadHistory();
  updateHistoryBadge(hist.length);
}

// ══ QUICK RANGES ══
const CF_RANGES=[
  {label:"CF 104.16/20",cidr:"104.16.0.0/20"},
  {label:"CF 104.17/20",cidr:"104.17.0.0/20"},
  {label:"CF 104.18/20",cidr:"104.18.0.0/20"},
  {label:"CF 104.19/20",cidr:"104.19.0.0/20"},
  {label:"CF 162.158/15",cidr:"162.158.0.0/15"},
  {label:"CF 172.64/13",cidr:"172.64.0.0/13"},
  {label:"CF 198.41/24",cidr:"198.41.128.0/24"},
  {label:"CF 141.101/17",cidr:"141.101.64.0/17"},
];
const FASTLY_RANGES=[
  {label:"Fastly 151.101/18",cidr:"151.101.0.0/18"},
  {label:"Fastly 199.232/24",cidr:"199.232.0.0/24"},
  {label:"Fastly 23.235/24",cidr:"23.235.32.0/24"},
  {label:"Fastly 104.156/22",cidr:"104.156.80.0/22"},
  {label:"Fastly 185.31/22",cidr:"185.31.16.0/22"},
  {label:"Fastly 157.52/22",cidr:"157.52.64.0/22"},
  {label:"Fastly 167.82/22",cidr:"167.82.0.0/22"},
  {label:"Fastly 130.211/16",cidr:"130.211.0.0/16"},
];

let activeProvider='cf'; // 'cf' or 'fastly'

function detectProvider(rawUrl){
  if(!rawUrl) return 'cf';
  const u=rawUrl.toLowerCase();
  const cfHosts=['cloudflare','cdn-cgi','fastly.net','cloudfront'];
  // fastly indicators
  if(u.includes('fastly')||u.includes('global.ssl.fastly')||u.includes('starz-net')||u.includes('fastlylb')) return 'fastly';
  return 'cf';
}

function renderQuickRanges(provider){
  const ranges=provider==='fastly'?FASTLY_RANGES:CF_RANGES;
  const label=provider==='fastly'?'FASTLY RANGES':'CLOUDFLARE RANGES';
  const el=document.getElementById('quickRanges');
  const lbl=document.getElementById('rangeLabel');
  if(lbl) lbl.textContent=label;
  if(!el) return;
  el.innerHTML='';
  ranges.forEach(r=>{
    const btn=document.createElement('button');
    btn.textContent=r.label;
    btn.dataset.cidr=r.cidr;
    btn.style.cssText='font-family:var(--font-mono);font-size:9px;padding:3px 8px;background:var(--bg3);border:1px solid var(--bd2);border-radius:var(--rad-xs);color:var(--tx2);cursor:pointer;margin:0';
    btn.addEventListener('mouseover',function(){this.style.borderColor='var(--c)';this.style.color='var(--c)'});
    btn.addEventListener('mouseout',function(){this.style.borderColor='var(--bd2)';this.style.color='var(--tx2)'});
    btn.addEventListener('click',function(){addRange(this.dataset.cidr)});
    el.appendChild(btn);
  });
}

function addRange(cidr){
  const ta=document.getElementById('ipInput');
  const current=ta.value.trim();
  ta.value=current?(current+'\n'+cidr):cidr;
}

// ══ PHASE 2 PROGRESS FIX ══
function onPhase2Progress(r){
  const done=r.done||0,total=r.total||1;
  const pct=r.pct||Math.round(done/total*100);
  document.getElementById('progBar').style.width=pct+'%';
  document.getElementById('progPct').textContent=pct+'%';
  document.getElementById('progTxt').textContent=done+' / '+total+' (P2)';
  document.getElementById('tbProgress').textContent='P2: '+done+'/'+total+' · '+pct+'%';
  if(r.eta) setStatValue('stETA',r.eta,'var(--y)');
  if(r.rate>0) document.getElementById('progRate').textContent=(r.rate||0).toFixed(1)+' IP/s';
  if(r.passed!=null) setStatValue('stPass',r.passed,'var(--g)');

  const grade=r.grade||scoreToGrade(r.score||0);
  const gc=gradeColor(grade);
  const passed=r.passed?'p2':'fail';
  const lat=r.latency?Math.round(r.latency)+'ms':'';
  const dlStr=r.dl&&r.dl!=='—'?' ↓'+r.dl:'';
  const rowTxt='['+grade+'] '+r.ip+' · '+lat+dlStr+(r.failReason?' · '+r.failReason:'');
  addFeedRow(rowTxt, passed);

  // live phase2 result در جدول
  if(!p2Results) p2Results=[];
  const existing=p2Results.findIndex(x=>x.IP===r.ip);
  const entry={
    IP:r.ip, Passed:r.passed, AvgLatencyMs:r.latency||0,
    JitterMs:r.jitter||0, PacketLossPct:r.loss||0,
    DownloadMbps:parseFloat(r.dl)||0, UploadMbps:parseFloat(r.ul)||0,
    StabilityScore:r.score||0, FailReason:r.failReason||''
  };
  if(existing>=0) p2Results[existing]=entry;
  else p2Results.push(entry);
}

// ══ TEMPLATES ══
let templates=[];
async function loadTemplates(){
  const res=await fetch('/api/templates');
  const d=await res.json();
  templates=d.templates||[];
  document.getElementById('nbTemplates').textContent=templates.length||'';
  renderTemplates();
}
function renderTemplates(){
  const el=document.getElementById('templateList');
  if(!el) return;
  el.innerHTML='';
  if(!templates.length){
    el.innerHTML='<div style="color:var(--dim);font-size:12px;text-align:center;padding:24px">هنوز کانفیگی ذخیره نشده</div>';
    return;
  }
  templates.forEach(t=>{
    const div=document.createElement('div');
    div.className='card';
    div.style.cssText='padding:12px 14px;display:flex;align-items:center;justify-content:space-between;gap:10px';
    div.innerHTML='<div style="min-width:0"><div style="font-weight:600;font-size:13px;color:var(--tx)">'+t.name+'</div>'+
      '<div style="font-family:var(--font-mono);font-size:9px;color:var(--dim);margin-top:2px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">'+(t.rawUrl||'').substring(0,60)+'…</div></div>'+
      '<div style="display:flex;gap:6px;flex-shrink:0">'+
        '<button class="btn btn-sm" data-id="'+t.id+'" data-action="use">▶ Use</button>'+
        '<button class="btn btn-sm" style="color:var(--r);border-color:var(--r)" data-id="'+t.id+'" data-action="del">✕</button>'+
      '</div>';
    div.querySelector('[data-action="use"]').onclick=function(){useTemplate(this.dataset.id)};
    div.querySelector('[data-action="del"]').onclick=function(){deleteTemplate(this.dataset.id)};
    el.appendChild(div);
  });
}
async function saveTemplate(){
  const name=document.getElementById('tmplName').value.trim();
  const rawUrl=document.getElementById('tmplURL').value.trim();
  if(!name||!rawUrl) return;
  await fetch('/api/templates/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,rawUrl})});
  document.getElementById('tmplName').value='';
  document.getElementById('tmplURL').value='';
  loadTemplates();
}
function useTemplate(id){
  const t=templates.find(x=>x.id===id);
  if(!t) return;
  document.getElementById('linkInput').value=t.rawUrl;
  parseLink();
  nav('import');
}
async function deleteTemplate(id){
  await fetch('/api/templates/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id})});
  loadTemplates();
}

// ══ HEALTH MONITOR ══
async function loadHealth(){
  const res=await fetch('/api/health');
  const d=await res.json();
  renderHealth(d.entries||[]);
}

// ── Sparkline SVG generator ──
function makeSparkline(history, times, width, height){
  if(!history||history.length<2) return '<svg width="'+width+'" height="'+height+'"><text x="50%" y="60%" text-anchor="middle" fill="var(--dim)" font-size="8">no data</text></svg>';
  const max=Math.max(...history.filter(v=>v>0),1);
  const pts=history.map((v,i)=>{
    const x=Math.round(i/(history.length-1)*(width-4))+2;
    const y=v===0?height-2:Math.round((1-v/max)*(height-6))+3;
    return {x,y,v,t:times?times[i]:0};
  });

  let path='';
  let dots='';
  let prev=null;
  pts.forEach((p,i)=>{
    if(p.v===0){
      // fail = red dot
      dots+='<circle cx="'+p.x+'" cy="'+(height-4)+'" r="2" fill="var(--r)" opacity="0.8"/>';
      prev=null;
    } else {
      if(prev!==null){
        const col=p.v>300?'var(--r)':p.v>150?'var(--y)':'var(--g)';
        path+='<line x1="'+prev.x+'" y1="'+prev.y+'" x2="'+p.x+'" y2="'+p.y+'" stroke="'+col+'" stroke-width="1.5" opacity="0.9"/>';
      }
      const title=p.t?new Date(p.t).toLocaleTimeString()+' '+p.v+'ms':p.v+'ms';
      dots+='<circle cx="'+p.x+'" cy="'+p.y+'" r="3" fill="transparent" stroke="none"><title>'+title+'</title></circle>';
      prev=p;
    }
  });

  return '<svg width="'+width+'" height="'+height+'" style="cursor:default">'+path+dots+'</svg>';
}

let healthCache=[];
function renderHealth(entries){
  healthCache=entries;
  renderHealthList();
}

function renderHealthList(){
  const el=document.getElementById('healthList');
  if(!el) return;
  const entries=healthCache;
  el.innerHTML='';
  if(!entries.length){
    el.innerHTML='<div style="color:var(--dim);font-size:12px;text-align:center;padding:32px">'+
      'هیچ IP ای در مانیتور نیست.<br>'+
      '<span style="color:var(--tx2)">از صفحه Results، IP ها رو انتخاب کن و Add to Monitor بزن</span>'+
    '</div>';
    return;
  }
  const showGraph=document.getElementById('monitorShowGraph')?.checked!==false;
  const nb=document.getElementById('nbMonitor');
  if(nb){nb.style.display='';nb.textContent=entries.filter(e=>e.status==='alive'||e.status==='recovered').length;}

  entries.forEach(e=>{
    const sc=e.status||'unknown';
    const col=sc==='alive'?'var(--g)':sc==='recovered'?'var(--y)':sc==='dead'?'var(--r)':'var(--dim)';
    const icon=sc==='alive'?'●':sc==='recovered'?'◑':sc==='dead'?'○':'?';
    const lastChk=e.lastCheck?new Date(e.lastCheck).toLocaleTimeString():'checking...';
    const lat=e.latencyMs?Math.round(e.latencyMs)+'ms':'—';
    const uptime=e.uptimePct!=null&&e.totalChecks>0?e.uptimePct.toFixed(0)+'%':'—';

    // GeoIP badge
    const geo=e.geoInfo;
    const geoBadge=geo?
      '<span style="font-size:9px;padding:1px 5px;background:var(--bg3);border-radius:3px;color:var(--dim);font-family:var(--font-mono)" title="'+(geo.isp||'')+'">'+( geo.countryCode||'?')+' '+(geo.city||'')+'</span>':
      '<button onclick="fetchGeoIP(\''+e.ip+'\')" style="font-size:9px;background:none;border:1px solid var(--dim);border-radius:3px;color:var(--dim);cursor:pointer;padding:1px 5px;font-family:var(--font-mono)">GeoIP</button>';

    // Sparkline
    const spark=showGraph&&e.latencyHistory&&e.latencyHistory.length>1?
      makeSparkline(e.latencyHistory,e.checkTimes,120,30):'';

    const div=document.createElement('div');
    div.className='card';
    div.dataset.ip=e.ip;
    div.style.cssText='padding:10px 14px;display:flex;align-items:center;gap:12px';
    div.innerHTML=
      '<span class="health-icon" style="color:'+col+';font-size:18px;flex-shrink:0">'+icon+'</span>'+
      '<div style="flex:1;min-width:0">'+
        '<div style="display:flex;align-items:center;gap:6px;flex-wrap:wrap">'+
          '<span style="font-family:var(--font-mono);font-size:12px;font-weight:700;color:var(--c)">'+e.ip+'</span>'+
          '<span class="health-badge" style="font-size:10px;padding:1px 6px;background:'+col+'20;color:'+col+';border-radius:3px;font-family:var(--font-mono)">'+sc.toUpperCase()+'</span>'+
          geoBadge+
        '</div>'+
        '<div style="font-family:var(--font-mono);font-size:10px;color:var(--dim);margin-top:3px">'+
          'lat: <span class="health-lat" style="color:var(--y)">'+lat+'</span>'+
          ' · uptime: <span class="health-up" style="color:var(--g)">'+uptime+'</span>'+
          ' · <span class="health-chk">'+lastChk+'</span>'+
          ' · '+e.totalChecks+' checks'+
        '</div>'+
      '</div>'+
      (spark?'<div class="health-graph" style="flex-shrink:0">'+spark+'</div>':'')+
      '<button class="copy-btn" style="color:var(--r)" data-ip="'+e.ip+'" title="Remove">✕</button>';
    div.querySelector('.copy-btn').onclick=function(){removeFromMonitor(this.dataset.ip)};
    el.appendChild(div);
  });
}

async function fetchGeoIP(ip){
  const res=await fetch('/api/geoip?ip='+encodeURIComponent(ip));
  const geo=await res.json();
  if(geo.countryCode){
    // cache توی healthCache
    const e=healthCache.find(x=>x.ip===ip);
    if(e) e.geoInfo=geo;
    renderHealthList();
  }
}

async function checkAllNow(){
  appendTUI({t:now(),l:'info',m:'⚡ Triggering health checks...'});
  healthCache.forEach(e=>{
    const el=document.querySelector('[data-ip="'+e.ip+'"]');
    if(el){const badge=el.querySelector('.health-badge');if(badge){badge.textContent='CHECKING...';badge.style.color='var(--y)';}}
  });
  await fetch('/api/health/check-now',{method:'POST'});
  appendTUI({t:now(),l:'ok',m:'✓ Health checks triggered — منتظر نتیجه باش'});
}

// live update
function handleHealthUpdate(payload){
  const ip=payload.ip;
  if(!ip){loadHealth();return;}
  // آپدیت cache
  const cached=healthCache.find(e=>e.ip===ip);
  if(cached){
    Object.assign(cached,payload);
    // re-render همون entry
    const el=document.querySelector('[data-ip="'+ip+'"]');
    if(!el){renderHealthList();return;}
    const sc=payload.status||'unknown';
    const col=sc==='alive'?'var(--g)':sc==='recovered'?'var(--y)':sc==='dead'?'var(--r)':'var(--dim)';
    el.querySelector('.health-icon').style.color=col;
    el.querySelector('.health-icon').textContent=sc==='alive'?'●':sc==='recovered'?'◑':sc==='dead'?'○':'?';
    el.querySelector('.health-badge').textContent=sc.toUpperCase();
    el.querySelector('.health-badge').style.color=col;
    el.querySelector('.health-badge').style.background=col+'20';
    const latEl=el.querySelector('.health-lat');
    if(latEl) latEl.textContent=payload.latencyMs?Math.round(payload.latencyMs)+'ms':'—';
    const upEl=el.querySelector('.health-up');
    if(upEl) upEl.textContent=payload.uptimePct?payload.uptimePct.toFixed(0)+'%':'—';
    el.querySelector('.health-chk').textContent=new Date().toLocaleTimeString();
    // آپدیت sparkline
    if(payload.latencyHistory){
      const graphEl=el.querySelector('.health-graph');
      if(graphEl) graphEl.innerHTML=makeSparkline(payload.latencyHistory,payload.checkTimes,120,30);
    }
  } else {
    loadHealth();
  }
  if(payload.error) appendTUI({t:now(),l:'warn',m:'Health ['+ip+']: '+payload.error});
}
function handleHealthError(payload){
  appendTUI({t:now(),l:'err',m:'Monitor Error: '+payload.message});
}

async function addIPToMonitor(ip, score){
  await fetch('/api/health/add',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ip,baseLatencyMs:score||0})});
  appendTUI({t:now(),l:'ok',m:'♡ '+ip+' به monitor اضافه شد'});
}
async function addToMonitor(){
  const ip=document.getElementById('monitorIPInput').value.trim();
  if(!ip) return;
  await addIPToMonitor(ip,0);
  document.getElementById('monitorIPInput').value='';
  loadHealth();
}
async function removeFromMonitor(ip){
  await fetch('/api/health/remove',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ip})});
  healthCache=healthCache.filter(e=>e.ip!==ip);
  renderHealthList();
}

// ── Multi-Import ──
function toggleMultiMode(on){
  document.getElementById('linkInputLabel').textContent=on?'یه لینک در هر خط paste کن:':'vless:// or vmess:// or trojan://';
  document.getElementById('linkInput').placeholder=on?'vless://...\nvmess://...\ntrojan://...':'vless://uuid@domain:443?...';
  document.getElementById('linkInput').rows=on?8:3;
  document.getElementById('btnMultiParse').style.display=on?'':'none';
}
async function parseMultiLinks(){
  const input=document.getElementById('linkInput').value.trim();
  if(!input) return;
  const res=await fetch('/api/config/multi-parse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})});
  const d=await res.json();
  const cont=document.getElementById('multiResults');
  const list=document.getElementById('multiResultsList');
  cont.style.display='';
  list.innerHTML='';
  if(d.errors&&d.errors.length){
    d.errors.forEach(e=>{
      const div=document.createElement('div');
      div.style.cssText='color:var(--r);font-size:10px;font-family:var(--font-mono);padding:4px 8px';
      div.textContent='✗ '+e;
      list.appendChild(div);
    });
  }
  (d.results||[]).forEach((r,i)=>{
    const div=document.createElement('div');
    div.className='card';
    div.style.cssText='padding:10px 14px;display:flex;align-items:center;gap:10px;margin-bottom:6px';
    div.innerHTML=
      '<span style="font-size:10px;padding:1px 6px;background:var(--c)20;color:var(--c);border-radius:3px;font-family:var(--font-mono)">'+r.protocol.toUpperCase()+'</span>'+
      '<div style="flex:1;min-width:0">'+
        '<div style="font-family:var(--font-mono);font-size:12px;font-weight:700;color:var(--c)">'+(r.remark||'Config '+(i+1))+'</div>'+
        '<div style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">'+r.address+':'+r.port+'</div>'+
      '</div>'+
      '<button class="btn btn-sm btn-primary-real" onclick="useMultiConfig('+i+')">▶ Use</button>'+
      '<button class="btn btn-sm" onclick="saveMultiAsTemplate('+i+')">+ Template</button>';
    list.appendChild(div);
  });
  window._multiResults=d.results||[];
  appendTUI({t:now(),l:'ok',m:'⬡ '+d.count+' کانفیگ parse شد — '+((d.errors||[]).length)+' خطا'});
}
async function useMultiConfig(idx){
  const r=window._multiResults[idx];
  if(!r) return;
  document.getElementById('linkInput').value=r.rawUrl;
  document.getElementById('multiImportMode').checked=false;
  toggleMultiMode(false);
  document.getElementById('multiResults').style.display='none';
  parseLink();
}
async function saveMultiAsTemplate(idx){
  const r=window._multiResults[idx];
  if(!r) return;
  const name=(r.remark||'Config')+' — '+r.address;
  await fetch('/api/templates/save',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({name,rawUrl:r.rawUrl})});
  appendTUI({t:now(),l:'ok',m:'✓ Template saved: '+name});
}

// ══ SUBNET STATS ══
async function loadSubnets(){
  const res=await fetch('/api/subnets');
  const d=await res.json();
  renderSubnets(d.subnets||[]);
}
function renderSubnets(subnets){
  const el=document.getElementById('subnetList');
  if(!el) return;
  el.innerHTML='';
  if(!subnets.length){
    el.innerHTML='<div style="color:var(--dim);font-size:12px;text-align:center;padding:32px">بعد از اسکن نتایج اینجا میان</div>';
    return;
  }
  subnets.slice(0,30).forEach((s,i)=>{
    const pct=s.passRate||0;
    const col=pct>=50?'var(--g)':pct>=20?'var(--y)':'var(--r)';
    const div=document.createElement('div');
    div.className='card';
    div.style.cssText='padding:10px 14px';
    div.innerHTML='<div style="display:flex;align-items:center;gap:10px">'+
      '<span style="font-family:var(--font-mono);font-size:11px;color:var(--dim);min-width:24px">'+(i+1)+'</span>'+
      '<span style="font-family:var(--font-mono);font-size:12px;font-weight:700;color:var(--c);flex:1">'+s.subnet+'</span>'+
      '<span style="font-size:11px;color:'+col+';font-family:var(--font-mono)">'+pct.toFixed(1)+'%</span>'+
      '<span style="font-size:11px;color:var(--tx2);font-family:var(--font-mono)">'+s.passed+'/'+s.total+'</span>'+
      (s.avgLatMs>0?'<span style="font-size:11px;color:var(--y);font-family:var(--font-mono)">'+Math.round(s.avgLatMs)+'ms</span>':'')+
      '<button class="copy-btn" data-subnet="'+s.subnet+'" style="font-size:9px;padding:2px 7px;background:var(--cd);border:1px solid var(--c);border-radius:3px;color:var(--c)">+ Use</button>'+
    '</div>'+
    '<div style="margin-top:6px;background:var(--bg3);border-radius:2px;height:3px">'+
      '<div style="width:'+Math.min(pct,100)+'%;height:100%;background:'+col+';border-radius:2px"></div>'+
    '</div>';
    div.querySelector('.copy-btn').onclick=function(){addRange(this.dataset.subnet)};
    el.appendChild(div);
  });
}

// hook health updates from WS
function handleHealthUpdate(payload){
  // آپدیت یه entry خاص بدون reload کامل
  const ip=payload.ip;
  if(!ip){loadHealth();return;}
  // پیدا کن container رو
  const el=document.getElementById('healthList');
  if(!el) return;
  // اگه container خالیه یا entry وجود نداره، reload کامل بزن
  const existing=el.querySelector('[data-ip="'+ip+'"]');
  if(!existing){loadHealth();return;}
  // آپدیت status و latency
  const sc=payload.status||'unknown';
  const col=sc==='alive'?'var(--g)':sc==='recovered'?'var(--y)':sc==='dead'?'var(--r)':'var(--dim)';
  const icon=sc==='alive'?'●':sc==='recovered'?'◑':sc==='dead'?'○':'?';
  const iconEl=existing.querySelector('.health-icon');
  const statusBadge=existing.querySelector('.health-badge');
  const latEl=existing.querySelector('.health-lat');
  const upEl=existing.querySelector('.health-up');
  const chkEl=existing.querySelector('.health-chk');
  if(iconEl){iconEl.textContent=icon;iconEl.style.color=col;}
  if(statusBadge){statusBadge.textContent=sc.toUpperCase();statusBadge.style.color=col;statusBadge.style.background=col+'20';}
  if(latEl) latEl.textContent=payload.latencyMs?Math.round(payload.latencyMs)+'ms':'—';
  if(upEl) upEl.textContent=payload.uptimePct?payload.uptimePct.toFixed(0)+'%':'—';
  if(chkEl) chkEl.textContent=new Date().toLocaleTimeString();
  if(payload.error&&payload.error!==''){
    appendTUI({t:now(),l:'warn',m:'Health ['+ip+']: '+payload.error});
  }
}

// hook health_error
function handleHealthError(payload){
  appendTUI({t:now(),l:'err',m:'Monitor Error: '+payload.message});
}

// hook phase2_done برای subnet stats
function handlePhase2Done(payload){
  if(payload.subnets) renderSubnets(payload.subnets);
  if(payload.results&&!viewingSession){
    p2Results=payload.results;
    updatePassedChips();renderP2();
    document.getElementById('resSummary').textContent=(p2Results||[]).filter(r=>r.Passed).length+' passed out of '+(p2Results||[]).length+' tested';
    document.getElementById('passedBadge').textContent=(p2Results||[]).filter(r=>r.Passed).length;
    document.getElementById('nbResults').textContent=(p2Results||[]).filter(r=>r.Passed).length;
  }
  addFeedRow('✓ Scan complete — '+((payload.results||[]).filter(r=>r.Passed).length)+' passed','ok');
}


// ══ PHASE 3 ══
function toggleP3Settings(on){
  document.getElementById('p3Settings').style.display=on?'':'none';
}

async function runPhase3(){
  const passed=(p2Results||[]).filter(r=>r.Passed);
  if(!passed.length){appendTUI({t:now(),l:'warn',m:'No passed IPs for Phase 3'});return;}
  const dlUrl='https://speed.cloudflare.com/__down?bytes=5000000';
  const ulUrl='https://speed.cloudflare.com/__up';
  const testUpload=document.getElementById('cfgP3Upload').checked;
  const ips=passed.map(r=>r.IP);
  const btn=document.getElementById('btnP3');
  if(btn){btn.disabled=true;btn.textContent='🚀 Running...';}
  const res=await fetch('/api/phase3/run',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({ips,downloadUrl:dlUrl,uploadUrl:ulUrl,testUpload})});
  const d=await res.json();
  if(!d.ok) appendTUI({t:now(),l:'err',m:'Phase 3 error: '+d.error});
  else appendTUI({t:now(),l:'ok',m:'🚀 Phase 3 (Speed Test) شروع شد — '+ips.length+' IP'});
  if(btn){btn.disabled=false;btn.textContent='🚀 Speed Test (Phase 3)';}
}


// ══ UI MODE ══
function setUIMode(mode){
  const body=document.body;
  const btnFull=document.getElementById('uiModeFull');
  const btnCompact=document.getElementById('uiModeCompact');
  if(mode==='compact'){
    body.classList.add('compact-mode');
    btnCompact.classList.add('active');
    btnFull.classList.remove('active');
  } else {
    body.classList.remove('compact-mode');
    btnFull.classList.add('active');
    btnCompact.classList.remove('active');
  }
  localStorage.setItem('uiMode', mode);
}
function initUIMode(){
  const saved=localStorage.getItem('uiMode')||'full';
  setUIMode(saved);
}
// ══ MONITOR SETTINGS ══
async function saveMonitorSettings(){
  const enabled=document.getElementById('monitorEnabled').checked;
  const intervalMins=parseInt(document.getElementById('monitorInterval').value)||3;
  const trafficDetect=document.getElementById('monitorTrafficDetect').checked;
  await fetch('/api/health/settings',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({enabled,intervalMins,trafficDetect})});
  appendTUI({t:now(),l:'ok',m:'✓ Monitor settings: interval='+intervalMins+'min enabled='+enabled});
}

async function loadMonitorSettings(){
  try{
    const res=await fetch('/api/health/settings');
    const d=await res.json();
    const en=document.getElementById('monitorEnabled');
    const iv=document.getElementById('monitorInterval');
    const td=document.getElementById('monitorTrafficDetect');
    if(en&&d.enabled!=null) en.checked=d.enabled;
    if(iv&&d.intervalMins) iv.value=d.intervalMins;
    if(td&&d.trafficDetect!=null) td.checked=d.trafficDetect;
  }catch(e){}
}

// stubs for removed chart functions
function initCharts(){}
function pushChartData(){}

// ══ SUBSCRIPTION FETCH ══
async function fetchSubscription(){
  const url=document.getElementById('subUrlInput').value.trim();
  if(!url){return;}
  const btn=document.getElementById('btnFetchSub');
  const status=document.getElementById('subStatus');
  const cont=document.getElementById('subResults');
  btn.disabled=true;btn.textContent='...';
  status.style.display='';status.textContent='در حال دریافت...';
  cont.style.display='none';
  try{
    const res=await fetch('/api/subscription/fetch',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url})});
    const d=await res.json();
    if(!d.ok){status.textContent='✗ خطا: '+(d.error||'unknown');btn.disabled=false;btn.textContent='↓ Fetch';return;}
    status.textContent='✓ '+d.count+' کانفیگ دریافت شد';
    cont.style.display='flex';
    cont.innerHTML='';
    (d.results||[]).forEach((r,i)=>{
      const div=document.createElement('div');
      div.className='card';
      div.style.cssText='padding:9px 12px;display:flex;align-items:center;gap:8px';
      div.innerHTML=
        '<span style="font-size:9px;padding:1px 5px;background:var(--c)20;color:var(--c);border-radius:3px;font-family:var(--font-mono);flex-shrink:0">'+r.protocol.toUpperCase()+'</span>'+
        '<div style="flex:1;min-width:0">'+
          '<div style="font-family:var(--font-mono);font-size:11px;font-weight:700;color:var(--c);white-space:nowrap;overflow:hidden;text-overflow:ellipsis">'+(r.remark||'Config '+(i+1))+'</div>'+
          '<div style="font-family:var(--font-mono);font-size:9px;color:var(--dim)">'+r.address+':'+r.port+'</div>'+
        '</div>'+
        '<button class="btn btn-sm btn-primary-real" style="font-size:10px;padding:2px 8px" onclick="useSubConfig('+i+')">▶ Use</button>'+
        '<button class="btn btn-sm" style="font-size:10px;padding:2px 8px" onclick="saveSubAsTemplate('+i+')">+ Save</button>';
      cont.appendChild(div);
    });
    window._subResults=d.results||[];
    appendTUI({t:now(),l:'ok',m:'🔗 Sub: '+d.count+' کانفیگ از '+url});
  }catch(e){status.textContent='✗ '+e.message;}
  btn.disabled=false;btn.textContent='↓ Fetch';
}
function useSubConfig(idx){
  const r=window._subResults[idx];if(!r)return;
  document.getElementById('linkInput').value=r.rawUrl;
  document.getElementById('multiImportMode').checked=false;
  toggleMultiMode(false);
  parseLink();
  nav('import');
}
async function saveSubAsTemplate(idx){
  const r=window._subResults[idx];if(!r)return;
  const name=(r.remark||'Config')+' — '+r.address;
  await fetch('/api/templates/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,rawUrl:r.rawUrl})});
  appendTUI({t:now(),l:'ok',m:'✓ Template saved: '+name});
  loadTemplates();
}

// ══ FILTER & SORT P2 TABLE ══
let _p2FilteredResults=[];
function applyP2Filter(){
  if(!p2Results||!p2Results.length){renderP2();return;}
  const ipFilter=(document.getElementById('filterIP')||{}).value||'';
  const statusFilter=(document.getElementById('filterStatus')||{}).value||'all';
  const sortBy=(document.getElementById('sortP2')||{}).value||'score';
  let filtered=p2Results.filter(r=>{
    if(ipFilter&&!r.IP.includes(ipFilter))return false;
    if(statusFilter==='pass'&&!r.Passed)return false;
    if(statusFilter==='fail'&&r.Passed)return false;
    return true;
  });
  filtered=[...filtered].sort((a,b)=>{
    if(sortBy==='latency')return(a.AvgLatencyMs||0)-(b.AvgLatencyMs||0);
    if(sortBy==='dl')return(b.DownloadMbps||0)-(a.DownloadMbps||0);
    if(sortBy==='loss')return(a.PacketLossPct||0)-(b.PacketLossPct||0);
    if(sortBy==='jitter')return(a.JitterMs||0)-(b.JitterMs||0);
    return(b.StabilityScore||0)-(a.StabilityScore||0);
  });
  _p2FilteredResults=filtered;
  renderP2WithData(filtered);
}
function renderP2WithData(results){
  const tbody=document.getElementById('p2Tbody');
  document.getElementById('p2CountBadge').textContent=(results||[]).length+' IPs';
  if(!results||!results.length){
    tbody.innerHTML='<tr><td colspan="10" style="text-align:center;color:var(--dim);padding:32px;font-family:var(--font-mono)">No results</td></tr>';
    return;
  }
  tbody.innerHTML=results.map((r,i)=>{
    const sc=r.StabilityScore||0;
    const scc=sc>=75?'var(--g)':sc>=50?'var(--y)':'var(--r)';
    const lc=r.AvgLatencyMs<=500?'var(--g)':r.AvgLatencyMs<=1500?'var(--y)':'var(--r)';
    const badge=r.Passed?'<span class="badge bg">PASS</span>':'<span class="badge br" title="'+(r.FailReason||'')+'">FAIL</span>';
    let dl=0,ul=0;
    if(typeof r.DownloadMbps==='number')dl=r.DownloadMbps;
    else if(typeof r.DownloadMbps==='string')dl=parseFloat(r.DownloadMbps)||0;
    if(typeof r.UploadMbps==='number')ul=r.UploadMbps;
    else if(typeof r.UploadMbps==='string')ul=parseFloat(r.UploadMbps)||0;
    const dlTxt=dl>0?dl.toFixed(1)+' M':'—';const ulTxt=ul>0?ul.toFixed(1)+' M':'—';
    const dlc=dl<=0?'var(--dim)':dl>=5?'var(--g)':dl>=1?'var(--y)':'var(--r)';
    const pl=r.PacketLossPct||0;const plc=pl<=5?'var(--g)':pl<=20?'var(--y)':'var(--r)';
    const jt=r.JitterMs||0;const jc=jt<=20?'var(--g)':jt<=80?'var(--y)':'var(--r)';
    return '<tr class="'+(r.Passed?'pass-row':'fail-row')+'">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700;font-size:12px;font-family:var(--font-mono)">'+r.IP+'</td>'+
      '<td style="color:'+scc+';font-weight:700;font-size:14px;font-family:var(--font-mono)">'+sc.toFixed(0)+'</td>'+
      '<td style="color:'+lc+';font-family:var(--font-mono)">'+Math.round(r.AvgLatencyMs||0)+'ms</td>'+
      '<td style="color:'+jc+';font-family:var(--font-mono)">'+(jt>0?jt.toFixed(0)+'ms':'—')+'</td>'+
      '<td style="color:'+plc+';font-family:var(--font-mono)">'+pl.toFixed(0)+'%</td>'+
      '<td style="color:'+dlc+';font-family:var(--font-mono)">'+dlTxt+'</td>'+
      '<td style="color:var(--tx2);font-family:var(--font-mono)">'+ulTxt+'</td>'+
      '<td>'+badge+'</td>'+
      '<td><div style="display:flex;gap:3px">'+
        '<button class="copy-btn" data-ip="'+r.IP+'" data-action="copyip" title="Copy IP">⎘ IP</button>'+
        '<button class="copy-btn" data-ip="'+r.IP+'" data-action="copyvless2" title="Copy link with this IP" style="color:var(--c)">⬡ Link</button>'+
        '<button class="copy-btn" data-ip="'+r.IP+'" data-action="addmonitor" title="Add to Health Monitor" style="color:var(--g)">♡</button>'+
      '</div></td></tr>';
  }).join('');
  // event delegation
  if(tbody)tbody.onclick=function(e){
    const btn=e.target.closest('[data-action]');if(!btn)return;
    const ip=btn.dataset.ip;const action=btn.dataset.action;
    if(action==='copyip'){copyIP(ip);}
    else if(action==='copyvless2'){copyWithIP(ip);}
    else if(action==='addmonitor'){
      addIPToMonitor(ip,0);
      btn.textContent='✓';btn.style.color='var(--g)';
      setTimeout(()=>{btn.textContent='♡';btn.style.color='var(--g)';},2000);
    }
  };
}

// ══ SYSTEM INFO ══
async function loadSysInfo(){
  try{
    const res=await fetch('/api/sysinfo');
    const d=await res.json();
    const si=(id,v)=>{const el=document.getElementById(id);if(el)el.textContent=v||'—';};
    si('siUptime',d.uptime);
    si('siThreads',d.threads||'—');
    si('siPersistPath',d.persistPath);
    // JS runtime stats
    if(performance&&performance.memory){
      si('siMemMB',Math.round(performance.memory.usedJSHeapSize/1024/1024)+' MB');
      si('siMemSys','total: '+Math.round(performance.memory.totalJSHeapSize/1024/1024)+' MB');
    }
    si('siGoroutines','N/A');
    si('siGC','N/A');
    si('siVersion','piyazche');
    si('siGoVer','Go (embedded)');
    si('siOS',navigator.platform||'—');
  }catch(e){}
}

// ══ HISTORY — persist to server on scan done ══
function syncHistoryToServer(){
  const history=loadHistory();
  if(!history.length)return;
  fetch('/api/sessions/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({sessions:history})}).catch(()=>{});
}

// ══ INIT ══
initUIMode();
connectWS();
fetch('/api/status').then(r=>r.json()).then(d=>setStatus(d.status||'idle',d.phase||''));
loadSavedSettings();
renderQuickRanges('cf');
loadTemplates();
loadMonitorSettings();
</script>
</body>
</html>
`
