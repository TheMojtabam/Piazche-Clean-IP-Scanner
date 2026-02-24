package scanner

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"piyazche/config"
	"piyazche/utils"
	"piyazche/xray"
)

// Phase2Result holds the deep-test results for a single IP
type Phase2Result struct {
	IP             string
	AvgLatencyMs   float64
	MinLatencyMs   int64
	MaxLatencyMs   int64
	JitterMs       float64
	PacketLossPct  float64
	DownloadMbps   float64
	UploadMbps     float64
	StabilityScore float64
	Passed         bool
	FailReason     string
}

// RunPhase2 takes the successful IPs from phase-1 and runs deep tests
func RunPhase2(ctx context.Context, cfg *config.Config, phase1Results []Result) []Phase2Result {
	return RunPhase2WithCallback(ctx, cfg, phase1Results, nil)
}

// RunPhase2WithCallback مثل RunPhase2 ولی هر بار که یه IP تموم شد callback میزنه
// معماری جدید: هر IP یه xray instance جداگانه داره ولی log level رو none میذاره
// تا terminal پر از Error نشه. Upload test حذف شد چون دقیق نبود.
func RunPhase2WithCallback(ctx context.Context, cfg *config.Config, phase1Results []Result, onDone func(Phase2Result)) []Phase2Result {
	rounds := cfg.Scan.StabilityRounds
	if rounds <= 0 {
		rounds = 1
	}
	interval := time.Duration(cfg.Scan.StabilityInterval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	var candidates []Result
	for _, r := range phase1Results {
		if r.Success {
			candidates = append(candidates, r)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// concurrency محدود: هر worker یه xray instance داره
	// بیشتر از 4 باعث port exhaustion میشه
	concurrency := 4
	if len(candidates) < concurrency {
		concurrency = len(candidates)
	}

	fmt.Printf("\n%s%s▸ Phase 2: Deep Stability Test%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("  %sIPs:%s %d  %sRounds:%s %d  %sInterval:%s %s  %sWorkers:%s %d\n",
		utils.Gray, utils.Reset, len(candidates),
		utils.Gray, utils.Reset, rounds,
		utils.Gray, utils.Reset, interval,
		utils.Gray, utils.Reset, concurrency)

	total := len(candidates)
	final := make([]Phase2Result, total)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var doneCount int64

	for i, candidate := range candidates {
		select {
		case <-ctx.Done():
			goto waitAll
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			p2 := testIPPhase2(ctx, cfg, ip, rounds, interval)
			applyFilters(cfg, &p2)

			done := int(atomic.AddInt64(&doneCount, 1))
			_ = done
			_ = done

			statusIcon := fmt.Sprintf("%s✓%s", utils.Green, utils.Reset)
			statusStr := ""
			plColor := utils.Green
			if !p2.Passed {
				statusIcon = fmt.Sprintf("%s✗%s", utils.Red, utils.Reset)
				statusStr = fmt.Sprintf(" %s(%s)%s", utils.Red, p2.FailReason, utils.Reset)
				plColor = utils.Red
			}
			if p2.PacketLossPct > 30 {
				plColor = utils.Red
			} else if p2.PacketLossPct > 10 {
				plColor = utils.Yellow
			}

			jitterStr := ""
			if cfg.Scan.JitterTest {
				jitterStr = fmt.Sprintf(" %sJ:%s%s%.0fms%s", utils.Gray, utils.Reset, utils.Magenta, p2.JitterMs, utils.Reset)
			}
			speedStr := ""
			if cfg.Scan.SpeedTest && p2.DownloadMbps > 0 {
				dlColor := utils.Green
				if p2.DownloadMbps < 1 {
					dlColor = utils.Red
				} else if p2.DownloadMbps < 5 {
					dlColor = utils.Yellow
				}
				speedStr = fmt.Sprintf(" %s↓%s%s%.1fM%s", utils.Blue, utils.Reset, dlColor, p2.DownloadMbps, utils.Reset)
			}

			mu.Lock()
			fmt.Printf("[%d/%d] %s %s%s%s %s─%s %s%.0fms%s %sPL:%s%s%.0f%%%s%s%s%s\n",
				done, total,
				statusIcon,
				utils.Cyan, ip, utils.Reset,
				utils.Gray, utils.Reset,
				utils.Yellow, p2.AvgLatencyMs, utils.Reset,
				utils.Gray, utils.Reset, plColor, p2.PacketLossPct, utils.Reset,
				jitterStr, speedStr, statusStr)
			final[idx] = p2
			mu.Unlock()

			if onDone != nil {
				onDone(p2)
			}
		}(i, candidate.IP)
	}

waitAll:
	wg.Wait()

	var nonEmpty []Phase2Result
	for _, r := range final {
		if r.IP != "" {
			nonEmpty = append(nonEmpty, r)
		}
	}

	sort.Slice(nonEmpty, func(i, j int) bool {
		if nonEmpty[i].StabilityScore != nonEmpty[j].StabilityScore {
			return nonEmpty[i].StabilityScore > nonEmpty[j].StabilityScore
		}
		return nonEmpty[i].AvgLatencyMs < nonEmpty[j].AvgLatencyMs
	})

	return nonEmpty
}

func testIPPhase2(ctx context.Context, cfg *config.Config, ip string, rounds int, interval time.Duration) Phase2Result {
	p2 := Phase2Result{IP: ip}

	port := utils.AcquirePort()
	defer utils.ReleasePort(port)

	// log level رو none بذار تا terminal پر از Error نشه
	p2Cfg := *cfg
	p2Cfg.Xray.LogLevel = "none"

	xrayConfig, err := config.GenerateXrayConfig(&p2Cfg, ip, port)
	if err != nil {
		p2.FailReason = "config error"
		return p2
	}

	manager := xray.NewManagerWithDebug(false)

	var startErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			utils.ReleasePort(port)
			time.Sleep(200 * time.Millisecond)
			port = utils.AcquirePort()
			xrayConfig, err = config.GenerateXrayConfig(&p2Cfg, ip, port)
			if err != nil {
				p2.FailReason = "config error"
				return p2
			}
		}
		startErr = manager.Start(xrayConfig, port)
		if startErr == nil {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if startErr != nil {
		p2.FailReason = "xray start failed"
		return p2
	}
	defer manager.Stop()

	readyCtx, readyCancel := context.WithTimeout(ctx, 6*time.Second)
	if err := manager.WaitForReadyWithContext(readyCtx, 6*time.Second); err != nil {
		readyCancel()
		p2.FailReason = "xray not ready"
		return p2
	}
	readyCancel()

	connTimeout := time.Duration(cfg.Scan.Timeout) * time.Second
	if connTimeout < 10*time.Second {
		connTimeout = 10 * time.Second
	}

	plCount := cfg.Scan.PacketLossCount
	if plCount <= 0 {
		plCount = 5
	}

	var latencies []int64
	var lossTotal float64
	roundsDone := 0

	for round := 0; round < rounds; round++ {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		// ۱. Latency — یه HTTP request ساده
		connResult := xray.TestConnectivityWithContext(ctx, port, cfg.Scan.TestURL, connTimeout)
		if connResult.Success {
			latencies = append(latencies, connResult.Latency.Milliseconds())
		}

		// ۲. Packet Loss — sequential HEAD requests (نه concurrent، نه keepalive)
		lost := 0
		pingTimeout := 4 * time.Second
		for i := 0; i < plCount; i++ {
			select {
			case <-ctx.Done():
				lost += plCount - i
				goto plDone
			default:
			}

			pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
			ok := doSimplePing(pingCtx, port, cfg.Scan.TestURL)
			pingCancel()
			if !ok {
				lost++
			}

			// کمی بین پینگها صبر کن
			if i < plCount-1 {
				select {
				case <-ctx.Done():
				case <-time.After(300 * time.Millisecond):
				}
			}
		}
	plDone:
		lossTotal += float64(lost) / float64(plCount) * 100
		roundsDone++

		if round < rounds-1 {
			select {
			case <-ctx.Done():
				goto done
			case <-time.After(interval):
			}
		}
	}

done:
	if len(latencies) == 0 {
		p2.FailReason = "no successful latency samples"
		return p2
	}

	// آمار latency
	var sum int64
	p2.MinLatencyMs = latencies[0]
	p2.MaxLatencyMs = latencies[0]
	for _, l := range latencies {
		sum += l
		if l < p2.MinLatencyMs {
			p2.MinLatencyMs = l
		}
		if l > p2.MaxLatencyMs {
			p2.MaxLatencyMs = l
		}
	}
	p2.AvgLatencyMs = float64(sum) / float64(len(latencies))

	// Jitter
	if cfg.Scan.JitterTest && len(latencies) > 1 {
		var variance float64
		for _, l := range latencies {
			diff := float64(l) - p2.AvgLatencyMs
			variance += diff * diff
		}
		variance /= float64(len(latencies))
		p2.JitterMs = math.Sqrt(variance)
	}

	// Packet Loss میانگین
	if roundsDone > 0 {
		p2.PacketLossPct = lossTotal / float64(roundsDone)
	}

	// Speed Test — فقط download (upload حذف شد چون دقیق نبود)
	if cfg.Scan.SpeedTest {
		dlURL := cfg.Scan.DownloadURL
		if dlURL == "" {
			dlURL = "https://speed.cloudflare.com/__down?bytes=5000000"
		}
		dlCtx, dlCancel := context.WithTimeout(ctx, 25*time.Second)
		dlBps, dlErr := xray.TestDownloadSpeed(dlCtx, port, dlURL, 25*time.Second)
		dlCancel()
		if dlErr == nil && dlBps > 0 {
			p2.DownloadMbps = dlBps / 1024 / 1024 * 8
		}
	}

	// Stability Score: 0-100
	plScore := math.Max(0, 50*(1-p2.PacketLossPct/100))
	latScore := math.Max(0, math.Min(30, 30*(1-(p2.AvgLatencyMs-100)/2900)))
	jitterScore := 20.0
	if cfg.Scan.JitterTest && p2.JitterMs > 10 {
		jitterScore = math.Max(0, 20*(1-(p2.JitterMs-10)/190))
	}
	p2.StabilityScore = plScore + latScore + jitterScore
	p2.Passed = true

	return p2
}

// doSimplePing یه HEAD request ساده بدون keepalive میزنه
func doSimplePing(ctx context.Context, socksPort int, testURL string) bool {
	// از TestConnectivityWithContext استفاده میکنیم که ساده‌ترین روشه
	result := xray.TestConnectivityWithContext(ctx, socksPort, testURL, 4*time.Second)
	return result.Success
}


func applyFilters(cfg *config.Config, p2 *Phase2Result) {
	if !p2.Passed {
		return
	}

	if cfg.Scan.MaxPacketLossPct >= 0 && p2.PacketLossPct > cfg.Scan.MaxPacketLossPct {
		p2.Passed = false
		p2.FailReason = fmt.Sprintf("packet loss %.0f%% > max %.0f%%", p2.PacketLossPct, cfg.Scan.MaxPacketLossPct)
		return
	}

	if cfg.Scan.MinDownloadMbps > 0 && p2.DownloadMbps < cfg.Scan.MinDownloadMbps {
		p2.Passed = false
		p2.FailReason = fmt.Sprintf("download %.1fMbps < min %.1fMbps", p2.DownloadMbps, cfg.Scan.MinDownloadMbps)
		return
	}
}

func PrintPhase2Results(results []Phase2Result, n int, hasSpeed bool, hasJitter bool) {
	var passed []Phase2Result
	for _, r := range results {
		if r.Passed {
			passed = append(passed, r)
		}
	}

	fmt.Printf("\n%s%s▸ Phase 2 Results%s  (%s%d passed%s / %s%d total%s)\n",
		utils.Bold, utils.Cyan, utils.Reset,
		utils.Green, len(passed), utils.Reset,
		utils.White, len(results), utils.Reset)

	if len(passed) == 0 {
		fmt.Printf("%sNo IPs passed phase-2 filters.%s\n", utils.Yellow, utils.Reset)
		return
	}

	if n > len(passed) {
		n = len(passed)
	}

	fmt.Printf("%s┌──────────────────────┬────────┬──────────┬───────────", utils.Gray)
	if hasJitter {
		fmt.Printf("┬──────────")
	}
	if hasSpeed {
		fmt.Printf("┬──────────────")
	}
	fmt.Printf("┬──────────┐%s\n", utils.Reset)

	fmt.Printf("%s│%s %-20s %s│%s %6s %s│%s %8s %s│%s %9s ",
		utils.Gray, utils.Reset, utils.Bold+"IP"+utils.Reset, utils.Gray, utils.Reset,
		utils.Bold+"Score"+utils.Reset, utils.Gray, utils.Reset,
		utils.Bold+"Avg Lat"+utils.Reset, utils.Gray, utils.Reset,
		utils.Bold+"Pkt Loss"+utils.Reset)
	if hasJitter {
		fmt.Printf("%s│%s %8s ", utils.Gray, utils.Reset, utils.Bold+"Jitter"+utils.Reset)
	}
	if hasSpeed {
		fmt.Printf("%s│%s %12s ", utils.Gray, utils.Reset, utils.Bold+"Download"+utils.Reset)
	}
	fmt.Printf("%s│%s\n", utils.Gray, utils.Reset)

	for i, r := range passed {
		if i >= n {
			break
		}

		scoreColor := utils.Green
		if r.StabilityScore < 40 {
			scoreColor = utils.Red
		} else if r.StabilityScore < 65 {
			scoreColor = utils.Yellow
		}

		plColor := utils.Green
		if r.PacketLossPct > 30 {
			plColor = utils.Red
		} else if r.PacketLossPct > 10 {
			plColor = utils.Yellow
		}

		fmt.Printf("%s│%s %-20s %s│%s %s%6.0f%s %s│%s %s%6.0fms%s %s│%s %s%7.0f%%%s ",
			utils.Gray, utils.Reset, r.IP,
			utils.Gray, utils.Reset,
			scoreColor, r.StabilityScore, utils.Reset,
			utils.Gray, utils.Reset,
			utils.Cyan, r.AvgLatencyMs, utils.Reset,
			utils.Gray, utils.Reset,
			plColor, r.PacketLossPct, utils.Reset)

		if hasJitter {
			fmt.Printf("%s│%s %s%6.0fms%s ", utils.Gray, utils.Reset, utils.Magenta, r.JitterMs, utils.Reset)
		}
		if hasSpeed {
			dlColor := utils.Dim
			dlStr := "    —"
			if r.DownloadMbps > 0 {
				dlStr = fmt.Sprintf("%5.1fMbps", r.DownloadMbps)
				dlColor = utils.Green
				if r.DownloadMbps < 1 {
					dlColor = utils.Red
				} else if r.DownloadMbps < 5 {
					dlColor = utils.Yellow
				}
			}
			fmt.Printf("%s│%s %s%12s%s ", utils.Gray, utils.Reset, dlColor, dlStr, utils.Reset)
		}
		fmt.Printf("%s│%s\n", utils.Gray, utils.Reset)
	}

	fmt.Printf("%s└──────────────────────┴────────┴──────────┴───────────", utils.Gray)
	if hasJitter {
		fmt.Printf("┴──────────")
	}
	if hasSpeed {
		fmt.Printf("┴──────────────")
	}
	fmt.Printf("┘%s\n", utils.Reset)
}

func SavePhase2Results(results []Phase2Result, format string, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if format == "json" {
		jf, err := os.Create(path)
		if err != nil {
			return err
		}
		defer jf.Close()
		enc := json.NewEncoder(jf)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// default: CSV
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	hasSpeed := false
	for _, r := range results {
		if r.DownloadMbps > 0 {
			hasSpeed = true
			break
		}
	}
	header := []string{"ip", "avg_latency_ms", "min_latency_ms", "max_latency_ms", "jitter_ms", "packet_loss_pct", "stability_score", "passed", "fail_reason"}
	if hasSpeed {
		header = append(header, "download_mbps")
	}
	w.Write(header)

	for _, r := range results {
		row := []string{
			r.IP,
			fmt.Sprintf("%.1f", r.AvgLatencyMs),
			fmt.Sprintf("%d", r.MinLatencyMs),
			fmt.Sprintf("%d", r.MaxLatencyMs),
			fmt.Sprintf("%.1f", r.JitterMs),
			fmt.Sprintf("%.1f", r.PacketLossPct),
			fmt.Sprintf("%.1f", r.StabilityScore),
			fmt.Sprintf("%t", r.Passed),
			r.FailReason,
		}
		if hasSpeed {
			row = append(row, fmt.Sprintf("%.2f", r.DownloadMbps))
		}
		w.Write(row)
	}
	w.Flush()
	return w.Error()
}

// GeneratePhase2OutputPath generates a timestamped output file path for phase2 results
func GeneratePhase2OutputPath(format string) string {
	timestamp := time.Now().Format("2006-01-02_150405")
	ext := format
	if ext == "" {
		ext = "csv"
	}
	filename := fmt.Sprintf("%s_phase2.%s", timestamp, ext)
	return filepath.Join("results", filename)
}
