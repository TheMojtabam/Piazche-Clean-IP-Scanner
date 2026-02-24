package shodan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"piyazche/utils"
)

// ---------- Shodan API types ----------

type SearchResult struct {
	Total   int      `json:"total"`
	Matches []Banner `json:"matches"`
}

type Banner struct {
	IP        string   `json:"ip_str"`
	Port      int      `json:"port"`
	Org       string   `json:"org"`
	ISP       string   `json:"isp"`
	ASN       string   `json:"asn"`
	Hostnames []string `json:"hostnames"`
	Domains   []string `json:"domains"`
	Timestamp string   `json:"timestamp"`
	Location  Location `json:"location"`
}

type Location struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// ---------- Harvester config ----------

// HarvestConfig تنظیمات جمع‌آوری IP از Shodan
type HarvestConfig struct {
	// APIKey کلید Shodan (لازم)
	APIKey string `json:"apiKey"`

	// Query کوئری سرچ - اگه خالی باشه از DefaultQuery استفاده می‌شه
	Query string `json:"query"`

	// UseDefaultQuery از کوئری پیش‌فرض (non-CF ranges with CF header) استفاده کن
	UseDefaultQuery bool `json:"useDefaultQuery"`

	// Pages تعداد صفحات نتیجه (هر صفحه 100 IP، default:1)
	Pages int `json:"pages"`

	// ExcludeCFRanges اگه true باشه رنج‌های اصلی CF از نتایج حذف می‌شن
	ExcludeCFRanges bool `json:"excludeCFRanges"`

	// MinConfidence حداقل امتیاز اطمینان برای IP (0-100)
	MinConfidence int `json:"minConfidence"`
}

// DefaultHarvestConfig تنظیمات پیش‌فرض
func DefaultHarvestConfig() HarvestConfig {
	return HarvestConfig{
		APIKey:          "",
		UseDefaultQuery: true,
		Pages:           1,
		ExcludeCFRanges: true,
		MinConfidence:   0,
	}
}

// ---------- Query templates ----------

// DefaultShodanQuery کوئری که IP های non-CF رو با CF header پیدا می‌کنه
// این IP ها CDN-as-a-service هستن - پشت CF هستن ولی رنج اصلی CF نیستن
const DefaultShodanQuery = `ssl:"Cloudflare Inc ECC CA" port:443 -net:173.245.48.0/20 -net:103.21.244.0/22 -net:103.22.200.0/22 -net:103.31.4.0/22 -net:141.101.64.0/18 -net:108.162.192.0/18 -net:190.93.240.0/20 -net:188.114.96.0/20 -net:197.234.240.0/22 -net:198.41.128.0/17 -net:162.158.0.0/15 -net:104.16.0.0/13 -net:104.24.0.0/14 -net:172.64.0.0/13 -net:131.0.72.0/22`

// AlternativeQuery کوئری جایگزین با http.headers
const AlternativeQuery = `http.headers:"CF-RAY" port:443 -org:"Cloudflare Inc."`

// ---------- Harvester ----------

// Harvester جمع‌آوری IP از Shodan API
type Harvester struct {
	cfg    HarvestConfig
	client *http.Client
}

// NewHarvester یه Harvester جدید می‌سازه
func NewHarvester(cfg HarvestConfig) *Harvester {
	return &Harvester{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HarvestResult نتیجه جمع‌آوری
type HarvestResult struct {
	IPs        []string
	Banners    []Banner
	TotalFound int
	Pages      int
	Query      string
}

// Harvest IP ها رو از Shodan جمع می‌کنه
func (h *Harvester) Harvest(ctx context.Context) (*HarvestResult, error) {
	if h.cfg.APIKey == "" {
		return nil, fmt.Errorf("shodan API key is required")
	}

	query := h.cfg.Query
	if query == "" || h.cfg.UseDefaultQuery {
		query = DefaultShodanQuery
	}

	fmt.Printf("\n%s%s▸ Shodan Harvest%s\n", utils.Bold, utils.Cyan, utils.Reset)
	fmt.Printf("  %sQuery:%s %s%.80s...%s\n",
		utils.Gray, utils.Reset, utils.Dim, query, utils.Reset)

	pages := h.cfg.Pages
	if pages <= 0 {
		pages = 1
	}

	result := &HarvestResult{
		Query: query,
		Pages: pages,
	}

	seenIPs := make(map[string]bool)

	for page := 1; page <= pages; page++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		fmt.Printf("  %sFetching page %d/%d...%s", utils.Dim, page, pages, utils.Reset)

		banners, total, err := h.searchPage(ctx, query, page)
		if err != nil {
			fmt.Printf(" %s✗ %v%s\n", utils.Red, err, utils.Reset)
			return result, fmt.Errorf("shodan search failed (page %d): %w", page, err)
		}

		result.TotalFound = total
		newCount := 0
		for _, b := range banners {
			if b.IP == "" {
				continue
			}
			if h.cfg.ExcludeCFRanges && isCFRange(b.IP) {
				continue
			}
			if !seenIPs[b.IP] {
				seenIPs[b.IP] = true
				result.IPs = append(result.IPs, b.IP)
				result.Banners = append(result.Banners, b)
				newCount++
			}
		}

		fmt.Printf("\r  %s✓ Page %d/%d%s  %s+%d IPs%s  (total in DB: %s%d%s)\n",
			utils.Green, page, pages, utils.Reset,
			utils.Yellow, newCount, utils.Reset,
			utils.Cyan, total, utils.Reset)

		// rate limit: shodan اجازه می‌ده 1 req/sec روی free plan
		if page < pages {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(1100 * time.Millisecond):
			}
		}
	}

	fmt.Printf("\n  %s✓ Harvested %d unique IPs from Shodan%s\n\n",
		utils.Green, len(result.IPs), utils.Reset)

	return result, nil
}

// searchPage یه صفحه از نتایج Shodan رو می‌گیره
func (h *Harvester) searchPage(ctx context.Context, query string, page int) ([]Banner, int, error) {
	apiURL := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s&page=%d&minify=true",
		url.QueryEscape(h.cfg.APIKey),
		url.QueryEscape(query),
		page)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != 200 {
		// parse error message از Shodan
		var apiErr struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Error != "" {
			return nil, 0, fmt.Errorf("shodan API error: %s", apiErr.Error)
		}
		return nil, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Matches, result.Total, nil
}

// CheckCredits چک می‌کنه چند تا query credit داری
func (h *Harvester) CheckCredits(ctx context.Context) (int, error) {
	apiURL := fmt.Sprintf("https://api.shodan.io/api-info?key=%s", url.QueryEscape(h.cfg.APIKey))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var info struct {
		QueryCredits int    `json:"query_credits"`
		ScanCredits  int    `json:"scan_credits"`
		Plan         string `json:"plan"`
		Error        string `json:"error"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &info); err != nil {
		return 0, err
	}
	if info.Error != "" {
		return 0, fmt.Errorf("shodan: %s", info.Error)
	}

	return info.QueryCredits, nil
}

// ---------- CF range check ----------

// cfRanges رنج‌های اصلی Cloudflare
var cfRanges = []string{
	"173.245.48.", "103.21.244.", "103.22.200.", "103.31.4.",
	"141.101.64.", "141.101.65.", "141.101.66.", "141.101.67.",
	"108.162.192.", "190.93.240.", "188.114.96.", "197.234.240.",
	"198.41.128.", "162.158.", "104.16.", "104.17.", "104.18.",
	"104.19.", "104.20.", "104.21.", "104.22.", "104.23.", "104.24.",
	"104.25.", "104.26.", "104.27.", "104.28.", "104.29.", "104.30.",
	"104.31.", "172.64.", "172.65.", "172.66.", "172.67.", "172.68.",
	"172.69.", "172.70.", "131.0.72.",
}

func isCFRange(ip string) bool {
	for _, prefix := range cfRanges {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}
	return false
}
