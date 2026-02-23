package scanner

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
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
	JitterMs       float64 // standard deviation of latency
	PacketLossPct  float64 // averaged across all stability rounds
	DownloadMbps   float64
	UploadMbps     float64
	StabilityScore float64 // 0-100, higher=better
	Passed         bool
	FailReason     string
}

// RunPhase2 takes the successful IPs from phase-1 and runs deep tests
func RunPhase2(ctx context.Context, cfg *config.Config, phase1Results []Result) []Phase2Result {
	rounds := cfg.Scan.StabilityRounds
	if rounds <= 0 {
		rounds = 1
	}
	interval := time.Duration(cfg.Scan.StabilityInterval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	// Only keep successful phase-1 IPs
	var candidates []Result
	for _, r := range phase1Results {
		if r.Success {
			candidates = append(candidates, r)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	fmt.Printf("\n%s%s▸ Phase 2: Deep Stability Test%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("  %sIPs:%s %d  %sRounds:%s %d  %sInterval:%s %s\n",
		utils.Gray, utils.Reset, len(candidates),
		utils.Gray, utils.Reset, rounds,
		utils.Gray, utils.Reset, interval)

	// کمی صبر کن تا xray های فاز ۱ کاملاً آزاد بشن
	fmt.Printf("  %sWaiting for resources to free up...%s\n\n", utils.Dim, utils.Reset)
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(3 * time.Second):
	}

	// فاز ۲ رو sequential اجرا کن تا port exhaustion نشه
	// (فاز ۱ با thread بالا خیلی پورت مصرف کرده، باید صبر کرد)
	for i, candidate := range candidates {
		select {
		case <-ctx.Done():
			break
		default:
		}

		p2 := testIPPhase2(ctx, cfg, candidate.IP, rounds, interval)
		applyFilters(cfg, &p2)

		statusIcon := utils.Green + "✓" + utils.Reset
		statusStr := ""
		if !p2.Passed {
			statusIcon = utils.Red + "✗" + utils.Reset
			statusStr = fmt.Sprintf(" %s(%s)%s", utils.Red, p2.FailReason, utils.Reset)
		}

		jitterStr := ""
		if cfg.Scan.JitterTest {
			jitterColor := utils.Green
			if p2.JitterMs > 50 {
				jitterColor = utils.Red
			} else if p2.JitterMs > 20 {
				jitterColor = utils.Yellow
			}
			jitterStr = fmt.Sprintf(" %sJ:%s%s%.0fms%s", utils.Gray, utils.Reset, jitterColor, p2.JitterMs, utils.Reset)
		}

		plColor := utils.Green
		if p2.PacketLossPct > 20 {
			plColor = utils.Red
		} else if p2.PacketLossPct > 5 {
			plColor = utils.Yellow
		}

		speedStr := ""
		if cfg.Scan.SpeedTest {
			speedStr = fmt.Sprintf(" %s↓%s%.1fM %s↑%s%.1fM",
				utils.Blue, utils.Reset, p2.DownloadMbps,
				utils.Blue, utils.Reset, p2.UploadMbps)
		}

		fmt.Printf("[%d/%d] %s %s%s%s %s─%s %s%.0fms%s %sPL:%s%s%.0f%%%s%s%s%s\n",
			i+1, len(candidates),
			statusIcon,
			utils.Cyan, candidate.IP, utils.Reset,
			utils.Gray, utils.Reset,
			utils.Yellow, p2.AvgLatencyMs, utils.Reset,
			utils.Gray, utils.Reset, plColor, p2.PacketLossPct, utils.Reset,
			jitterStr, speedStr, statusStr)

		final = append(final, p2)
	}

	// sort by stability score then latency
	sort.Slice(final, func(i, j int) bool {
		if final[i].StabilityScore != final[j].StabilityScore {
			return final[i].StabilityScore > final[j].StabilityScore
		}
		return final[i].AvgLatencyMs < final[j].AvgLatencyMs
	})

	return final
}

func testIPPhase2(ctx context.Context, cfg *config.Config, ip string, rounds int, interval time.Duration) Phase2Result {
	p2 := Phase2Result{IP: ip}

	port := utils.AcquirePort()
	defer utils.ReleasePort(port)

	xrayConfig, err := config.GenerateXrayConfig(cfg, ip, port)
	if err != nil {
		p2.FailReason = "config error"
		return p2
	}

	manager := xray.NewManagerWithDebug(false)
	if err := manager.Start(xrayConfig, port); err != nil {
		p2.FailReason = "xray start failed"
		return p2
	}
	defer manager.Stop()

	readyCtx, readyCancel := context.WithTimeout(ctx, 3*time.Second)
	if err := manager.WaitForReadyWithContext(readyCtx, 3*time.Second); err != nil {
		readyCancel()
		p2.FailReason = "xray not ready"
		return p2
	}
	readyCancel()

	timeout := time.Duration(cfg.Scan.Timeout) * time.Second

	var latencies []int64
	var lossTotal float64
	roundsDone := 0

	plCount := cfg.Scan.PacketLossCount
	if plCount <= 0 {
		plCount = 5
	}

	for round := 0; round < rounds; round++ {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		// Latency sample
		connResult := xray.TestConnectivityWithContext(ctx, port, cfg.Scan.TestURL, timeout)
		if connResult.Success {
			latencies = append(latencies, connResult.Latency.Milliseconds())
		}

		// Packet loss for this round
		pingTimeout := 3 * time.Second
		plTotal := pingTimeout*time.Duration(plCount) + 2*time.Second
		plCtx, plCancel := context.WithTimeout(ctx, plTotal)
		loss, plErr := xray.TestPacketLoss(plCtx, port, cfg.Scan.TestURL, plCount, pingTimeout)
		plCancel()
		if plErr == nil {
			lossTotal += loss
		} else {
			// اگه خطا داشت یعنی همه پینگ‌ها fail شدن = 100% loss
			lossTotal += 100
		}
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
	// Calculate stats
	if len(latencies) > 0 {
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

		// Jitter = standard deviation
		if cfg.Scan.JitterTest && len(latencies) > 1 {
			var variance float64
			for _, l := range latencies {
				diff := float64(l) - p2.AvgLatencyMs
				variance += diff * diff
			}
			variance /= float64(len(latencies))
			p2.JitterMs = math.Sqrt(variance)
		}
	}

	if roundsDone > 0 {
		p2.PacketLossPct = lossTotal / float64(roundsDone)
	}

	// Speed test (once, after stability rounds)
	if cfg.Scan.SpeedTest {
		dlURL := cfg.Scan.DownloadURL
		if dlURL == "" {
			dlURL = "https://speed.cloudflare.com/__down?bytes=1000000"
		}
		ulURL := cfg.Scan.UploadURL
		if ulURL == "" {
			ulURL = "https://speed.cloudflare.com/__up"
		}

		dlCtx, dlCancel := context.WithTimeout(ctx, timeout)
		dlBps, dlErr := xray.TestDownloadSpeed(dlCtx, port, dlURL, timeout)
		dlCancel()
		if dlErr == nil && dlBps > 0 {
			p2.DownloadMbps = dlBps / 1024 / 1024 * 8
		}

		ulCtx, ulCancel := context.WithTimeout(ctx, timeout)
		ulBps, ulErr := xray.TestUploadSpeed(ulCtx, port, ulURL, timeout)
		ulCancel()
		if ulErr == nil && ulBps > 0 {
			p2.UploadMbps = ulBps / 1024 / 1024 * 8
		}
	}

	// Stability Score: 0-100
	// - 50 pts from packet loss (0% = 50, 100% = 0)
	// - 30 pts from latency (<=100ms=30, >=3000ms=0)
	// - 20 pts from jitter (<=10ms=20, >=200ms=0)
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

// applyFilters checks phase-2 result against configured thresholds
func applyFilters(cfg *config.Config, p2 *Phase2Result) {
	if !p2.Passed {
		return
	}

	// MaxPacketLossPct: -1=disabled, >=0=active
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

	if cfg.Scan.MinUploadMbps > 0 && p2.UploadMbps < cfg.Scan.MinUploadMbps {
		p2.Passed = false
		p2.FailReason = fmt.Sprintf("upload %.1fMbps < min %.1fMbps", p2.UploadMbps, cfg.Scan.MinUploadMbps)
		return
	}
}

// PrintPhase2Results prints the final phase-2 table
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

	// Build header dynamically
	fmt.Printf("%s┌──────────────────────┬────────┬──────────┬───────────", utils.Gray)
	if hasJitter {
		fmt.Printf("┬──────────")
	}
	if hasSpeed {
		fmt.Printf("┬──────────────┬──────────────")
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
		fmt.Printf("%s│%s %12s %s│%s %12s ",
			utils.Gray, utils.Reset, utils.Bold+"Download"+utils.Reset,
			utils.Gray, utils.Reset, utils.Bold+"Upload"+utils.Reset)
	}
	fmt.Printf("%s│%s\n", utils.Gray, utils.Reset)

	fmt.Printf("%s├──────────────────────┼────────┼──────────┼───────────", utils.Gray)
	if hasJitter {
		fmt.Printf("┼──────────")
	}
	if hasSpeed {
		fmt.Printf("┼──────────────┼──────────────")
	}
	fmt.Printf("┤%s\n", utils.Reset)

	for i := 0; i < n; i++ {
		r := passed[i]

		scoreColor := utils.Green
		if r.StabilityScore < 50 {
			scoreColor = utils.Red
		} else if r.StabilityScore < 75 {
			scoreColor = utils.Yellow
		}

		latColor := utils.Green
		if r.AvgLatencyMs > 2000 {
			latColor = utils.Red
		} else if r.AvgLatencyMs > 1000 {
			latColor = utils.Yellow
		}

		plColor := utils.Green
		if r.PacketLossPct > 20 {
			plColor = utils.Red
		} else if r.PacketLossPct > 5 {
			plColor = utils.Yellow
		}

		rank := fmt.Sprintf("%d.", i+1)
		fmt.Printf("%s│%s %s%-2s%-17s %s│%s %s%5.1f%s %s│%s %s%6.0fms%s %s│%s %s%7.1f%%%s  ",
			utils.Gray, utils.Reset, utils.Dim, rank, utils.Cyan+r.IP+utils.Reset, utils.Gray, utils.Reset,
			scoreColor, r.StabilityScore, utils.Reset, utils.Gray, utils.Reset,
			latColor, r.AvgLatencyMs, utils.Reset, utils.Gray, utils.Reset,
			plColor, r.PacketLossPct, utils.Reset)

		if hasJitter {
			jColor := utils.Green
			if r.JitterMs > 50 {
				jColor = utils.Red
			} else if r.JitterMs > 20 {
				jColor = utils.Yellow
			}
			fmt.Printf("%s│%s %s%6.0fms%s ", utils.Gray, utils.Reset, jColor, r.JitterMs, utils.Reset)
		}

		if hasSpeed {
			dlColor := utils.Green
			if r.DownloadMbps < 1 {
				dlColor = utils.Red
			} else if r.DownloadMbps < 5 {
				dlColor = utils.Yellow
			}
			ulColor := utils.Green
			if r.UploadMbps < 0.5 {
				ulColor = utils.Red
			} else if r.UploadMbps < 2 {
				ulColor = utils.Yellow
			}
			fmt.Printf("%s│%s %s%9.2f Mbps%s %s│%s %s%9.2f Mbps%s ",
				utils.Gray, utils.Reset,
				dlColor, r.DownloadMbps, utils.Reset,
				utils.Gray, utils.Reset,
				ulColor, r.UploadMbps, utils.Reset)
		}

		fmt.Printf("%s│%s\n", utils.Gray, utils.Reset)
	}

	fmt.Printf("%s└──────────────────────┴────────┴──────────┴───────────", utils.Gray)
	if hasJitter {
		fmt.Printf("┴──────────")
	}
	if hasSpeed {
		fmt.Printf("┴──────────────┴──────────────")
	}
	fmt.Printf("┘%s\n\n", utils.Reset)
}

// SavePhase2ToCSV saves phase-2 results to CSV
func SavePhase2ToCSV(results []Phase2Result, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"IP", "Score", "Avg Latency (ms)", "Min Lat (ms)", "Max Lat (ms)", "Jitter (ms)", "Packet Loss (%)", "Download (Mbps)", "Upload (Mbps)", "Passed", "Fail Reason"}
	writer.Write(header)

	for _, r := range results {
		passed := "yes"
		if !r.Passed {
			passed = "no"
		}
		writer.Write([]string{
			r.IP,
			fmt.Sprintf("%.1f", r.StabilityScore),
			fmt.Sprintf("%.0f", r.AvgLatencyMs),
			fmt.Sprintf("%d", r.MinLatencyMs),
			fmt.Sprintf("%d", r.MaxLatencyMs),
			fmt.Sprintf("%.1f", r.JitterMs),
			fmt.Sprintf("%.1f", r.PacketLossPct),
			fmt.Sprintf("%.2f", r.DownloadMbps),
			fmt.Sprintf("%.2f", r.UploadMbps),
			passed,
			r.FailReason,
		})
	}
	return nil
}
