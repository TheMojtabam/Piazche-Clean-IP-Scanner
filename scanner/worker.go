package scanner

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"piyazche/config"
	"piyazche/utils"
	"piyazche/xray"
)

// Worker represents a scanner worker
type Worker struct {
	id        int
	cfg       *config.Config
	results   *ResultCollector
	wg        *sync.WaitGroup
	jobs      <-chan string
	quit      <-chan struct{}
	ctx       context.Context
	processed *atomic.Int64
	logger    chan<- string
	debug     bool
	debugOnce *sync.Once
}

// NewWorker creates a new scanner worker
func NewWorker(id int, cfg *config.Config, results *ResultCollector,
	wg *sync.WaitGroup, jobs <-chan string, quit <-chan struct{}, ctx context.Context, processed *atomic.Int64, logger chan<- string, debug bool, debugOnce *sync.Once) *Worker {
	return &Worker{
		id:        id,
		cfg:       cfg,
		results:   results,
		wg:        wg,
		jobs:      jobs,
		quit:      quit,
		ctx:       ctx,
		processed: processed,
		logger:    logger,
		debug:     debug,
		debugOnce: debugOnce,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	go w.run()
}

func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.quit:
			return
		case ip, ok := <-w.jobs:
			if !ok {
				return
			}
			select {
			case <-w.ctx.Done():
				return
			default:
			}
			w.processIP(ip)
		}
	}
}

func (w *Worker) processIP(ip string) {
	defer w.processed.Add(1)

	select {
	case <-w.ctx.Done():
		return
	default:
	}

	result := Result{
		IP: ip,
	}

	// Give failed connections a few chances
	maxRetries := w.cfg.Scan.Retries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		testResult := w.testIP(ip)
		if testResult.Success {
			result.Success = true
			result.Latency = testResult.Latency
			result.StatusCode = testResult.StatusCode
			break
		}
		lastErr = testResult.Error

		// Small pause before retry so we don't flood the server
		if attempt < maxRetries-1 {
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}

	if !result.Success && lastErr != nil {
		result.Error = lastErr.Error()
	}

	// If connectivity test passed and speed test is enabled, measure speed & packet loss
	if result.Success {
		port := utils.AcquirePort()
		defer utils.ReleasePort(port)

		xrayConfig, err := config.GenerateXrayConfig(w.cfg, result.IP, port)
		if err == nil {
			manager := xray.NewManagerWithDebug(false)
			if startErr := manager.Start(xrayConfig, port); startErr == nil {
				readyTimeout := 2 * time.Second
				if readyErr := manager.WaitForReadyWithContext(w.ctx, readyTimeout); readyErr == nil {
					timeout := time.Duration(w.cfg.Scan.Timeout) * time.Second

					// Packet loss test
					plCount := w.cfg.Scan.PacketLossCount
					if plCount <= 0 {
						plCount = 5
					}
					plCtx, plCancel := context.WithTimeout(w.ctx, timeout)
					loss, plErr := xray.TestPacketLoss(plCtx, port, w.cfg.Scan.TestURL, plCount, timeout)
					plCancel()
					if plErr == nil {
						result.PacketLossPct = loss
					}

					// Speed test (download + upload)
					if w.cfg.Scan.SpeedTest {
						dlURL := w.cfg.Scan.DownloadURL
						if dlURL == "" {
							dlURL = "https://speed.cloudflare.com/__down?bytes=1000000"
						}
						ulURL := w.cfg.Scan.UploadURL
						if ulURL == "" {
							ulURL = "https://speed.cloudflare.com/__up"
						}

						dlCtx, dlCancel := context.WithTimeout(w.ctx, timeout)
						dlBps, dlErr := xray.TestDownloadSpeed(dlCtx, port, dlURL, timeout)
						dlCancel()
						if dlErr == nil {
							result.DownloadMbps = dlBps / 1024 / 1024 * 8
						}

						ulCtx, ulCancel := context.WithTimeout(w.ctx, timeout)
						ulBps, ulErr := xray.TestUploadSpeed(ulCtx, port, ulURL, timeout)
						ulCancel()
						if ulErr == nil {
							result.UploadMbps = ulBps / 1024 / 1024 * 8
						}
					}
				}
				manager.Stop()
			}
		}
	}

	w.results.Add(result)

	if w.logger != nil {
		var logMsg string
		if result.Success {
			speedInfo := ""
			if result.DownloadMbps > 0 || result.UploadMbps > 0 {
				speedInfo = fmt.Sprintf(" %s↓%s%s%.1fMbps%s %s↑%s%s%.1fMbps%s",
					utils.Blue, utils.Reset, utils.Green, result.DownloadMbps, utils.Reset,
					utils.Blue, utils.Reset, utils.Yellow, result.UploadMbps, utils.Reset)
			}
			plColor := utils.Green
			if result.PacketLossPct > 20 {
				plColor = utils.Red
			} else if result.PacketLossPct > 5 {
				plColor = utils.Yellow
			}
			plInfo := fmt.Sprintf(" %sPL:%s%s%.0f%%%s", utils.Gray, utils.Reset, plColor, result.PacketLossPct, utils.Reset)
			logMsg = fmt.Sprintf("%s✓%s %s%s%s %s─%s %s%dms%s%s%s",
				utils.Green, utils.Reset, utils.Cyan, ip, utils.Reset, utils.Gray, utils.Reset, utils.Yellow, result.Latency.Milliseconds(), utils.Reset, plInfo, speedInfo)
		} else {
			errMsg := result.Error
			if len(errMsg) > 50 {
				errMsg = errMsg[:50] + "..."
			}
			logMsg = fmt.Sprintf("%s✗%s %s%s%s %s─ %s%s", utils.Red, utils.Reset, utils.Gray, ip, utils.Reset, utils.Gray, errMsg, utils.Reset)
		}
		select {
		case w.logger <- logMsg:
		default:
		}
	}
}

func (w *Worker) testIP(ip string) *xray.TestResult {
	select {
	case <-w.ctx.Done():
		return &xray.TestResult{IP: ip, Error: w.ctx.Err()}
	default:
	}

	// Grab a free port for this test's SOCKS proxy
	port := utils.AcquirePort()
	defer utils.ReleasePort(port)

	xrayConfig, err := config.GenerateXrayConfig(w.cfg, ip, port)
	if err != nil {
		return &xray.TestResult{
			IP:    ip,
			Error: fmt.Errorf("failed to generate xray config: %w", err),
		}
	}

	// Debug output only for the first IP to avoid log spam
	showDebug := false
	if w.debug {
		w.debugOnce.Do(func() {
			showDebug = true
			w.printFragmentDebugInfo(ip, port)
		})
	}

	manager := xray.NewManagerWithDebug(showDebug)
	if err := manager.Start(xrayConfig, port); err != nil {
		return &xray.TestResult{
			IP:    ip,
			Error: fmt.Errorf("failed to start xray: %w", err),
		}
	}
	defer manager.Stop()

	select {
	case <-w.ctx.Done():
		return &xray.TestResult{IP: ip, Error: w.ctx.Err()}
	default:
	}

	// Wait for xray to spin up and open the port
	readyTimeout := 2 * time.Second
	if err := manager.WaitForReadyWithContext(w.ctx, readyTimeout); err != nil {
		return &xray.TestResult{
			IP:    ip,
			Error: fmt.Errorf("xray not ready: %w", err),
		}
	}

	// Run a test request through the proxy and time it
	timeout := time.Duration(w.cfg.Scan.Timeout) * time.Second
	testResult := xray.TestConnectivityWithContext(w.ctx, port, w.cfg.Scan.TestURL, timeout)
	testResult.IP = ip

	// Too slow? Mark as failed even if it connected
	if testResult.Success && w.cfg.Scan.MaxLatency > 0 {
		if testResult.Latency.Milliseconds() > int64(w.cfg.Scan.MaxLatency) {
			testResult.Success = false
			testResult.Error = fmt.Errorf("latency %dms exceeds max %dms",
				testResult.Latency.Milliseconds(), w.cfg.Scan.MaxLatency)
		}
	}

	return testResult
}

// printFragmentDebugInfo prints colorized fragment debug information
func (w *Worker) printFragmentDebugInfo(ip string, port int) {
	fmt.Printf("\n%s%sDebug Info - First IP Test%s\n", utils.Bold, utils.Magenta, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%s%s\n", utils.Gray, "Target IP:", utils.Reset, utils.Cyan, ip, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Local Port:", utils.Reset, utils.Yellow, port, utils.Reset)

	// Fragment settings
	fmt.Printf("\n  %s%s▸ Fragment Configuration%s\n", utils.Bold, utils.Yellow, utils.Reset)

	modeColor := utils.Green
	if w.cfg.Fragment.Mode == "off" {
		modeColor = utils.Red
	} else if w.cfg.Fragment.Mode == "auto" {
		modeColor = utils.Blue
	}
	fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Mode:", utils.Reset, modeColor, w.cfg.Fragment.Mode, utils.Reset)
	fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Packets:", utils.Reset, utils.Magenta, w.cfg.Fragment.Packets, utils.Reset)

	if w.cfg.Fragment.Mode == "manual" || w.cfg.Fragment.Mode == "" {
		fmt.Printf("    %s%-16s%s %s%s%s %sbytes%s\n", utils.Gray, "Length:", utils.Reset,
			utils.Green, w.cfg.Fragment.Manual.Length, utils.Reset, utils.Dim, utils.Reset)
		fmt.Printf("    %s%-16s%s %s%s%s %sms%s\n", utils.Gray, "Interval:", utils.Reset,
			utils.Green, w.cfg.Fragment.Manual.Interval, utils.Reset, utils.Dim, utils.Reset)
	}

	// Connection info
	fmt.Printf("\n  %s%s▸ Connection Details%s\n", utils.Bold, utils.Yellow, utils.Reset)

	// Show SNI based on method
	if w.cfg.Proxy.Method == "tls" && w.cfg.Proxy.TLS != nil {
		fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "SNI:", utils.Reset, utils.Cyan, w.cfg.Proxy.TLS.SNI, utils.Reset)
		fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Fingerprint:", utils.Reset, utils.Magenta, w.cfg.Proxy.TLS.Fingerprint, utils.Reset)
	} else if w.cfg.Proxy.Method == "reality" && w.cfg.Proxy.Reality != nil {
		fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Server Name:", utils.Reset, utils.Cyan, w.cfg.Proxy.Reality.ServerName, utils.Reset)
		fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Fingerprint:", utils.Reset, utils.Magenta, w.cfg.Proxy.Reality.Fingerprint, utils.Reset)
	}
	fmt.Printf("    %s%-16s%s %s%s%s\n", utils.Gray, "Network:", utils.Reset, utils.Green, w.cfg.Proxy.Type, utils.Reset)
	fmt.Printf("    %s%-16s%s %s%d%s\n", utils.Gray, "Remote Port:", utils.Reset, utils.Yellow, w.cfg.Proxy.Port, utils.Reset)

	fmt.Println()
}
