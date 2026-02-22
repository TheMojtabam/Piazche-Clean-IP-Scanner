package scanner

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"piyazche/utils"
)

// Result represents the scan result for a single IP
type Result struct {
	IP            string        `json:"ip"`
	Success       bool          `json:"success"`
	Latency       time.Duration `json:"latency"`
	LatencyMs     int64         `json:"latency_ms"`
	StatusCode    int           `json:"status_code,omitempty"`
	Error         string        `json:"error,omitempty"`
	TestedAt      time.Time     `json:"tested_at"`
	DownloadMbps  float64       `json:"download_mbps,omitempty"`
	UploadMbps    float64       `json:"upload_mbps,omitempty"`
	PacketLossPct float64       `json:"packet_loss_pct,omitempty"`
}

// ResultCollector collects and manages scan results
type ResultCollector struct {
	results []Result
	mu      sync.Mutex
}

// NewResultCollector creates a new result collector
func NewResultCollector() *ResultCollector {
	return &ResultCollector{
		results: make([]Result, 0),
	}
}

// Add adds a result to the collection
func (rc *ResultCollector) Add(result Result) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	result.LatencyMs = result.Latency.Milliseconds()
	result.TestedAt = time.Now()
	rc.results = append(rc.results, result)
}

// GetResults returns a copy of all results
func (rc *ResultCollector) GetResults() []Result {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	results := make([]Result, len(rc.results))
	copy(results, rc.results)
	return results
}

// GetSuccessful returns only successful results
func (rc *ResultCollector) GetSuccessful() []Result {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var successful []Result
	for _, r := range rc.results {
		if r.Success {
			successful = append(successful, r)
		}
	}
	return successful
}

// GetSortedByLatency returns results sorted by latency (ascending)
func (rc *ResultCollector) GetSortedByLatency() []Result {
	results := rc.GetSuccessful()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Latency < results[j].Latency
	})
	return results
}

// Count returns the total number of results
func (rc *ResultCollector) Count() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return len(rc.results)
}

// SuccessCount returns the number of successful results
func (rc *ResultCollector) SuccessCount() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	count := 0
	for _, r := range rc.results {
		if r.Success {
			count++
		}
	}
	return count
}

// SaveToCSV saves results to a CSV file
func (rc *ResultCollector) SaveToCSV(path string) error {
	results := rc.GetSortedByLatency()

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

	header := []string{"IP", "Latency (ms)", "Download (Mbps)", "Upload (Mbps)", "Packet Loss (%)", "Status", "Tested At"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, r := range results {
		status := "failed"
		if r.StatusCode >= 200 && r.StatusCode < 300 {
			status = "success"
		}

		row := []string{
			r.IP,
			fmt.Sprintf("%d", r.LatencyMs),
			fmt.Sprintf("%.2f", r.DownloadMbps),
			fmt.Sprintf("%.2f", r.UploadMbps),
			fmt.Sprintf("%.1f", r.PacketLossPct),
			status,
			r.TestedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// SaveToJSON saves results to a JSON file
func (rc *ResultCollector) SaveToJSON(path string) error {
	results := rc.GetSortedByLatency()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// GenerateOutputPath generates a timestamped output file path
func GenerateOutputPath(format string) string {
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("%s_results.%s", timestamp, format)
	return filepath.Join("results", filename)
}

// PrintTopResults prints the top N results to stdout with colors
func (rc *ResultCollector) PrintTopResults(n int) {
	results := rc.GetSortedByLatency()
	if len(results) == 0 {
		fmt.Printf("%sNo successful results found.%s\n", utils.Yellow, utils.Reset)
		return
	}

	if n > len(results) {
		n = len(results)
	}

	// Detect if speed test was performed
	hasSpeed := false
	for _, r := range results[:n] {
		if r.DownloadMbps > 0 || r.UploadMbps > 0 {
			hasSpeed = true
			break
		}
	}

	if hasSpeed {
		fmt.Printf("\n%s%sTop %d IPs by Latency%s\n", utils.Bold, utils.Cyan, n, utils.Reset)
		fmt.Printf("%s┌──────────────────────┬──────────────┬──────────────────┬──────────────────┬──────────────┐%s\n", utils.Gray, utils.Reset)
		fmt.Printf("%s│%s %-20s %s│%s %12s %s│%s %16s %s│%s %16s %s│%s %12s %s│%s\n",
			utils.Gray, utils.Reset, utils.Bold+"IP"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Latency"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Download"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Upload"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Pkt Loss"+utils.Reset, utils.Gray, utils.Reset)
		fmt.Printf("%s├──────────────────────┼──────────────┼──────────────────┼──────────────────┼──────────────┤%s\n", utils.Gray, utils.Reset)

		for i := 0; i < n; i++ {
			r := results[i]
			latencyColor := utils.Green
			if r.LatencyMs > 2000 {
				latencyColor = utils.Red
			} else if r.LatencyMs > 1000 {
				latencyColor = utils.Yellow
			}

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

			plColor := utils.Green
			if r.PacketLossPct > 20 {
				plColor = utils.Red
			} else if r.PacketLossPct > 5 {
				plColor = utils.Yellow
			}

			rank := fmt.Sprintf("%d.", i+1)
			fmt.Printf("%s│%s %s%-2s%-17s %s│%s %s%8dms%s   %s│%s %s%12.2f Mbps%s %s│%s %s%12.2f Mbps%s %s│%s %s%9.1f%%%s    %s│%s\n",
				utils.Gray, utils.Reset, utils.Dim, rank, utils.Cyan+r.IP+utils.Reset, utils.Gray, utils.Reset,
				latencyColor, r.LatencyMs, utils.Reset, utils.Gray, utils.Reset,
				dlColor, r.DownloadMbps, utils.Reset, utils.Gray, utils.Reset,
				ulColor, r.UploadMbps, utils.Reset, utils.Gray, utils.Reset,
				plColor, r.PacketLossPct, utils.Reset, utils.Gray, utils.Reset)
		}
		fmt.Printf("%s└──────────────────────┴──────────────┴──────────────────┴──────────────────┴──────────────┘%s\n\n", utils.Gray, utils.Reset)
	} else {
		fmt.Printf("\n%s%sTop %d IPs by Latency%s\n", utils.Bold, utils.Cyan, n, utils.Reset)
		fmt.Printf("%s┌──────────────────────┬──────────────┬──────────────┐%s\n", utils.Gray, utils.Reset)
		fmt.Printf("%s│%s %-20s %s│%s %12s %s│%s %12s %s│%s\n",
			utils.Gray, utils.Reset, utils.Bold+"IP"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Latency"+utils.Reset, utils.Gray, utils.Reset,
			utils.Bold+"Pkt Loss"+utils.Reset, utils.Gray, utils.Reset)
		fmt.Printf("%s├──────────────────────┼──────────────┼──────────────┤%s\n", utils.Gray, utils.Reset)

		for i := 0; i < n; i++ {
			r := results[i]
			latencyColor := utils.Green
			if r.LatencyMs > 2000 {
				latencyColor = utils.Red
			} else if r.LatencyMs > 1000 {
				latencyColor = utils.Yellow
			}

			plColor := utils.Green
			if r.PacketLossPct > 20 {
				plColor = utils.Red
			} else if r.PacketLossPct > 5 {
				plColor = utils.Yellow
			}

			rank := fmt.Sprintf("%d.", i+1)
			fmt.Printf("%s│%s %s%-2s%-17s %s│%s %s%8dms%s   %s│%s %s%9.1f%%%s    %s│%s\n",
				utils.Gray, utils.Reset, utils.Dim, rank, utils.Cyan+r.IP+utils.Reset, utils.Gray, utils.Reset,
				latencyColor, r.LatencyMs, utils.Reset, utils.Gray, utils.Reset,
				plColor, r.PacketLossPct, utils.Reset, utils.Gray, utils.Reset)
		}
		fmt.Printf("%s└──────────────────────┴──────────────┴──────────────┘%s\n\n", utils.Gray, utils.Reset)
	}
}
