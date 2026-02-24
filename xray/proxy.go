package xray

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
// Uses a timed writer to measure actual bytes-sent-per-second, not total RTT.
func TestUploadSpeed(ctx context.Context, socksPort int, uploadURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	parsedURL, err := url.Parse(uploadURL)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %w", err)
	}
	client, err := makeSOCKSClient(socksPort, parsedURL.Hostname(), timeout, false)
	if err != nil {
		return 0, err
	}

	// 8MB non-compressible payload (random-ish pattern)
	const uploadSize = 8 * 1024 * 1024
	buf := make([]byte, uploadSize)
	for i := range buf {
		buf[i] = byte((i*7 + 13) & 0xFF)
	}

	// Pipe: goroutine writes to pw, HTTP reads from pr
	pr, pw := io.Pipe()
	writeStart := time.Now() // start timer before goroutine (conservative)
	var bytesWritten int64

	go func() {
		n, werr := pw.Write(buf)
		bytesWritten = int64(n)
		pw.CloseWithError(werr)
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
	if err != nil {
		pr.Close()
		return 0, err
	}
	req.ContentLength = uploadSize
	req.Header.Set("Content-Type", "application/octet-stream")
	// Cloudflare __up needs this header
	req.Header.Set("Content-Length", fmt.Sprintf("%d", uploadSize))

	resp, err := client.Do(req)
	elapsed := time.Since(writeStart).Seconds()

	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	if err != nil && ctx.Err() == nil {
		return 0, fmt.Errorf("upload request failed: %w", err)
	}
	if elapsed <= 0.1 || bytesWritten == 0 {
		return 0, fmt.Errorf("upload elapsed too short: %.2fs", elapsed)
	}
	return float64(bytesWritten) / elapsed, nil
}

// repeatReader یه reader که یه slice داده رو یکبار میده
type repeatReader struct {
	data      []byte
	pos       int
	remaining int
}

func (r *repeatReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	if n > r.remaining {
		n = r.remaining
	}
	r.pos += n
	if r.pos >= len(r.data) {
		r.pos = 0
	}
	r.remaining -= n
	return n, nil
}

// TestPacketLoss - sequential pings through SOCKS proxy
func TestPacketLoss(ctx context.Context, socksPort int, testURL string, count int, pingTimeout time.Duration) (lossPercent float64, err error) {
	if count <= 0 {
		count = 5
	}

	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return 100, fmt.Errorf("invalid URL: %w", err)
	}

	lost := 0
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			lost += count - i
			break
		default:
		}

		// هر ping یه client جدید با DisableKeepAlives — مثل ping واقعی
		client, e := makeSOCKSClient(socksPort, parsedURL.Hostname(), pingTimeout, false)
		if e != nil {
			lost++
			continue
		}

		func() {
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			defer cancel()

			req, e := http.NewRequestWithContext(pingCtx, "HEAD", testURL, nil)
			if e != nil {
				lost++
				return
			}
			resp, e := client.Do(req)
			if e != nil || resp == nil {
				lost++
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode < 100 || resp.StatusCode >= 600 {
				lost++
			}
		}()

		if i < count-1 {
			select {
			case <-ctx.Done():
				lost += count - i - 1
				goto done
			case <-time.After(200 * time.Millisecond):
			}
		}
	}
done:
	return float64(lost) / float64(count) * 100, nil
}

// IsPortOpen checks if a TCP port is accepting connections
func IsPortOpen(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// zeroReader is an io.Reader that returns zeros
type zeroReader struct{}

func (z zeroReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
