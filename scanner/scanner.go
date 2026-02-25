package scanner

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"piyazche/config"
	"piyazche/utils"

	"github.com/schollz/progressbar/v3"
)

// Scanner is the main scanner orchestrator
type Scanner struct {
	cfg        *config.Config
	ips        []string
	results    *ResultCollector
	quit       chan struct{}
	quitOnce   sync.Once
	pauseCh    chan struct{} // close to pause, recreate to resume
	pauseMu    sync.Mutex
	paused     bool
	ctx        context.Context
	cancel     context.CancelFunc
	startTime  time.Time
	debug      bool
	OnIPStart  func(ip string)
}

// NewScanner creates a new scanner
func NewScanner(cfg *config.Config) *Scanner {
	return NewScannerWithDebug(cfg, false)
}

// NewScannerWithDebug creates a new scanner with debug option
func NewScannerWithDebug(cfg *config.Config, debug bool) *Scanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scanner{
		cfg:     cfg,
		results: NewResultCollector(),
		quit:    make(chan struct{}),
		pauseCh: make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
		debug:   debug,
	}
}

// LoadIPs loads IPs from file or CIDR string
func (s *Scanner) LoadIPs(source string, maxIPs int, shuffle bool) error {
	var ips []string
	var err error

	if _, statErr := os.Stat(source); statErr == nil {
		ips, err = utils.ParseIPsFromFile(source, s.cfg.Scan.SampleSize, shuffle)
	} else {
		ips, err = utils.ParseCIDRList(source, s.cfg.Scan.SampleSize)
		if shuffle {
			utils.ShuffleIPs(ips)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to load IPs: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IPs found in source")
	}

	if maxIPs > 0 && maxIPs < len(ips) {
		ips = ips[:maxIPs]
	}

	s.ips = ips
	return nil
}

// Run starts the scanning process
func (s *Scanner) Run() error {
	if len(s.ips) == 0 {
		return fmt.Errorf("no IPs loaded")
	}

	s.startTime = time.Now()
	threads := s.cfg.Scan.Threads
	if threads <= 0 {
		threads = 16
	}

	fmt.Printf("%s%sStarting Scan%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("   %sIPs:%s %d  %sWorkers:%s %d\n\n", utils.Gray, utils.Reset, len(s.ips), utils.Gray, utils.Reset, threads)

	jobs := make(chan string, threads*2)
	logger := make(chan string, threads*4)

	bar := progressbar.NewOptions(len(s.ips),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("IPs"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("[cyan]Scanning[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionUseANSICodes(true),
	)

	var processed atomic.Int64

	logDone := make(chan struct{})
	go func() {
		defer close(logDone)
		for msg := range logger {
			fmt.Printf("\r\033[K%s\n", msg)
			bar.RenderBlank()
		}
	}()

	var debugOnce sync.Once

	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		worker := NewWorker(i, s.cfg, s.results, &wg, jobs, s.quit, s.ctx, &processed, logger, s.debug, &debugOnce)
		worker.Start()
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		lastProcessed := int64(0)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				current := processed.Load()
				if current != lastProcessed {
					bar.Set(int(current))
					lastProcessed = current
				}
			}
		}
	}()

	go func() {
		for _, ip := range s.ips {
			// بررسی pause قبل از هر IP
			for {
				pauseCh := s.pauseChannel()
				select {
				case <-s.quit:
					close(jobs)
					return
				case <-s.ctx.Done():
					close(jobs)
					return
				case <-pauseCh:
					// paused — منتظر resume بمون
					for s.IsPaused() {
						select {
						case <-s.quit:
							close(jobs)
							return
						case <-time.After(200 * time.Millisecond):
						}
					}
					continue
				default:
				}
				break
			}

			select {
			case <-s.quit:
				close(jobs)
				return
			case jobs <- ip:
				if s.OnIPStart != nil {
					s.OnIPStart(ip)
				}
			}
		}
		close(jobs)
	}()

	wg.Wait()
	close(done)
	close(logger)
	<-logDone
	bar.Finish()

	s.printSummary()

	return nil
}

// Stop gracefully stops the scanner (safe to call multiple times)
func (s *Scanner) Stop() {
	s.cancel()
	s.quitOnce.Do(func() {
		close(s.quit)
	})
}

// Pause موقتاً اسکن رو متوقف می‌کنه
func (s *Scanner) Pause() {
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	if !s.paused {
		s.paused = true
		// بستن pauseCh باعث میشه workerها block بشن
		close(s.pauseCh)
	}
}

// Resume از pause برمی‌گرده
func (s *Scanner) Resume() {
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	if s.paused {
		s.paused = false
		s.pauseCh = make(chan struct{})
	}
}

// IsPaused وضعیت pause رو برمیگردونه
func (s *Scanner) IsPaused() bool {
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	return s.paused
}

// pauseChannel کانال pause فعلی رو برمیگردونه (برای select در worker)
func (s *Scanner) pauseChannel() chan struct{} {
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	return s.pauseCh
}

// GetResults returns the result collector
func (s *Scanner) GetResults() *ResultCollector {
	return s.results
}

func (s *Scanner) printSummary() {
	duration := time.Since(s.startTime)
	total := s.results.Count()
	successful := s.results.SuccessCount()
	successRate := float64(successful) / float64(total) * 100

	fmt.Printf("\n%s%sScan Complete%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("  %s%-18s%s %s%v%s\n", utils.Gray, "Duration:", utils.Reset, utils.White, duration.Round(time.Second), utils.Reset)
	fmt.Printf("  %s%-18s%s %s%d%s\n", utils.Gray, "Total IPs tested:", utils.Reset, utils.White, total, utils.Reset)

	rateColor := utils.Green
	if successRate < 20 {
		rateColor = utils.Red
	} else if successRate < 50 {
		rateColor = utils.Yellow
	}
	fmt.Printf("  %s%-18s%s %s%d%s (%s%.1f%%%s)\n",
		utils.Gray, "Successful:", utils.Reset,
		utils.Green, successful, utils.Reset,
		rateColor, successRate, utils.Reset)

	if successful > 0 {
		sorted := s.results.GetSortedByLatency()
		latencyColor := utils.Green
		if sorted[0].LatencyMs > 2000 {
			latencyColor = utils.Red
		} else if sorted[0].LatencyMs > 1000 {
			latencyColor = utils.Yellow
		}
		fmt.Printf("  %s%-18s%s %s%dms%s %s(%s)%s\n",
			utils.Gray, "Best latency:", utils.Reset,
			latencyColor, sorted[0].LatencyMs, utils.Reset,
			utils.Dim, sorted[0].IP, utils.Reset)
	}
}

// SaveResults saves results to file
func (s *Scanner) SaveResults(format string, path string) error {
	if path == "" {
		path = GenerateOutputPath(format)
	}

	switch format {
	case "csv":
		return s.results.SaveToCSV(path)
	case "json":
		return s.results.SaveToJSON(path)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// LoadIPsFromList IP ها رو مستقیم از یه slice لود می‌کنه (برای Shodan integration)
func (s *Scanner) LoadIPsFromList(ips []string, maxIPs int, shuffle bool) {
	if shuffle {
		utils.ShuffleIPs(ips)
	}
	if maxIPs > 0 && maxIPs < len(ips) {
		ips = ips[:maxIPs]
	}
	s.ips = ips
}

// IPCount تعداد IP های لود شده رو برمیگردونه
func (s *Scanner) IPCount() int {
	return len(s.ips)
}
