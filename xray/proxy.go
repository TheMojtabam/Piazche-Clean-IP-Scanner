package xray

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"
)

// TestResult contains the result of a connectivity test
type TestResult struct {
	IP         string
	Success    bool
	Latency    time.Duration
	Error      error
	StatusCode int
	BytesRead  int64
}

// makeSOCKSClient یه http.Client با SOCKS5 proxy می‌سازه
func makeSOCKSClient(socksPort int, host string, timeout time.Duration, keepAlive bool) (*http.Client, error) {
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			return dialer.Dial(network, addr)
		},
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false, ServerName: host},
		DisableKeepAlives:   !keepAlive,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

// TestConnectivity tests connectivity through the SOCKS proxy
func TestConnectivity(socksPort int, testURL string, timeout time.Duration) *TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return TestConnectivityWithContext(ctx, socksPort, testURL, timeout)
}

// TestConnectivityWithContext tests connectivity with context support for cancellation
func TestConnectivityWithContext(ctx context.Context, socksPort int, testURL string, timeout time.Duration) *TestResult {
	result := &TestResult{}

	parsedURL, err := url.Parse(testURL)
	if err != nil {
		result.Error = fmt.Errorf("invalid test URL: %w", err)
		return result
	}

	client, err := makeSOCKSClient(socksPort, parsedURL.Hostname(), timeout, false)
	if err != nil {
		result.Error = err
		return result
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			result.Error = ctx.Err()
		} else {
			result.Error = fmt.Errorf("request failed: %w", err)
		}
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if err != nil {
		if ctx.Err() != nil {
			result.Error = ctx.Err()
		} else {
			result.Error = fmt.Errorf("failed to read response: %w", err)
		}
		return result
	}

	result.Latency = time.Since(start)
	result.StatusCode = resp.StatusCode
	result.BytesRead = int64(len(body))
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
	return result
}

// TestSpeed performs a speed test by downloading a file
func TestSpeed(socksPort int, testURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return TestDownloadSpeed(ctx, socksPort, testURL, timeout)
}

// TestDownloadSpeed measures download speed through the SOCKS proxy
func TestDownloadSpeed(ctx context.Context, socksPort int, downloadURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %w", err)
	}
	client, err := makeSOCKSClient(socksPort, parsedURL.Hostname(), timeout, false)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil && ctx.Err() == nil {
		return 0, err
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 || n == 0 {
		return 0, fmt.Errorf("no data received")
	}
	return float64(n) / elapsed, nil
}

// TestUploadSpeed measures upload speed through the SOCKS proxy
func TestUploadSpeed(ctx context.Context, socksPort int, uploadURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	parsedURL, err := url.Parse(uploadURL)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %w", err)
	}
	client, err := makeSOCKSClient(socksPort, parsedURL.Hostname(), timeout, false)
	if err != nil {
		return 0, err
	}

	// 5MB upload
	const uploadSize = 5 * 1024 * 1024
	pr, pw := io.Pipe()
	go func() {
		buf := make([]byte, 32*1024)
		var written int64
		for written < uploadSize {
			n := int64(len(buf))
			if written+n > uploadSize {
				n = uploadSize - written
			}
			pw.Write(buf[:n])
			written += n
		}
		pw.Close()
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
	if err != nil {
		pr.Close()
		return 0, err
	}
	req.ContentLength = uploadSize

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil && ctx.Err() == nil {
		return 0, err
	}
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, fmt.Errorf("upload too fast / no data")
	}
	return uploadSize / elapsed, nil
}

// PacketLossResult نتیجه تست packet loss
type PacketLossResult struct {
	Sent     int
	Received int
	Lost     int
	LossPct  float64
	AvgRTT   time.Duration
	MinRTT   time.Duration
	MaxRTT   time.Duration
}

// TestPacketLossAdvanced - packet loss رو با connection pool بهینه اندازه‌گیری می‌کنه
// بهتر از TestPacketLoss: concurrent pings، connection reuse، آمار کامل‌تر
func TestPacketLossAdvanced(ctx context.Context, socksPort int, testURL string, count int, pingTimeout time.Duration) (*PacketLossResult, error) {
	if count <= 0 {
		count = 5
	}

	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// یه client با keepalive برای reuse connection — پینگ واقعی‌تر
	client, err := makeSOCKSClient(socksPort, parsedURL.Hostname(), pingTimeout*time.Duration(count)+5*time.Second, true)
	if err != nil {
		return nil, err
	}

	// concurrent pings با سمافور (max 3 همزمان)
	concurrency := 3
	if count < concurrency {
		concurrency = count
	}

	type pingResult struct {
		rtt  time.Duration
		ok   bool
	}

	results := make([]pingResult, count)
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	_ = mu

	// warm up connection
	warmReq, _ := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
	client.Do(warmReq)

	var successCount int64

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			// بقیه رو lost حساب کن
			for j := i; j < count; j++ {
				results[j] = pingResult{ok: false}
			}
			goto compute
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			defer cancel()

			req, e := http.NewRequestWithContext(pingCtx, "HEAD", testURL, nil)
			if e != nil {
				results[idx] = pingResult{ok: false}
				return
			}

			start := time.Now()
			resp, e := client.Do(req)
			rtt := time.Since(start)

			if e != nil || resp == nil {
				results[idx] = pingResult{ok: false}
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			ok := resp.StatusCode >= 100 && resp.StatusCode < 600
			results[idx] = pingResult{rtt: rtt, ok: ok}
			if ok {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)

		// کمی بین pingها صبر کن تا burst نباشه
		if i < count-1 {
			select {
			case <-ctx.Done():
			case <-time.After(150 * time.Millisecond):
			}
		}
	}

	wg.Wait()

compute:
	res := &PacketLossResult{Sent: count}
	var totalRTT time.Duration
	first := true

	for _, r := range results {
		if r.ok {
			res.Received++
			totalRTT += r.rtt
			if first {
				res.MinRTT = r.rtt
				res.MaxRTT = r.rtt
				first = false
			} else {
				if r.rtt < res.MinRTT {
					res.MinRTT = r.rtt
				}
				if r.rtt > res.MaxRTT {
					res.MaxRTT = r.rtt
				}
			}
		}
	}
	res.Lost = res.Sent - res.Received
	if res.Received > 0 {
		res.AvgRTT = totalRTT / time.Duration(res.Received)
	}
	res.LossPct = float64(res.Lost) / float64(res.Sent) * 100
	return res, nil
}

// TestPacketLoss - backward compat wrapper
func TestPacketLoss(ctx context.Context, socksPort int, testURL string, count int, pingTimeout time.Duration) (lossPercent float64, err error) {
	res, err := TestPacketLossAdvanced(ctx, socksPort, testURL, count, pingTimeout)
	if err != nil {
		return 100, err
	}
	return res.LossPct, nil
}

// zeroReader is an io.Reader that returns zeros
type zeroReader struct{}

func (z zeroReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
