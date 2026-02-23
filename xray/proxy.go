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

	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		result.Error = fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		return result
	}

	// Route HTTP traffic through the xray SOCKS proxy
	transport := &http.Transport{
		DialContext: func(dialCtx context.Context, network, addr string) (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			return dialer.Dial(network, addr)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         parsedURL.Hostname(),
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
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

	// Read the body so we measure the full request time
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
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	start := time.Now()
	resp, err := client.Get(testURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, err
	}

	duration := time.Since(start).Seconds()
	if duration > 0 {
		bytesPerSecond = float64(n) / duration
	}

	return bytesPerSecond, nil
}

// IsPortOpen checks if a port is open on the given host
func IsPortOpen(host string, port int) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// SpeedTestResult contains download and upload speeds
type SpeedTestResult struct {
	DownloadBps float64
	UploadBps   float64
	Error       error
}

// TestDownloadSpeed measures download speed through the SOCKS proxy
func TestDownloadSpeed(ctx context.Context, socksPort int, downloadURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(c context.Context, network, addr string) (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			return dialer.Dial(network, addr)
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
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

	duration := time.Since(start).Seconds()
	if duration > 0 && n > 0 {
		bytesPerSecond = float64(n) / duration
	}
	return bytesPerSecond, nil
}

// TestUploadSpeed measures upload speed through the SOCKS proxy
func TestUploadSpeed(ctx context.Context, socksPort int, uploadURL string, timeout time.Duration) (bytesPerSecond float64, err error) {
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	// 1 MB upload payload
	uploadSize := int64(1 * 1024 * 1024)
	payload := io.LimitReader(zeroReader{}, uploadSize)

	transport := &http.Transport{
		DialContext: func(c context.Context, network, addr string) (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			return dialer.Dial(network, addr)
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, payload)
	if err != nil {
		return 0, err
	}
	req.ContentLength = uploadSize
	req.Header.Set("Content-Type", "application/octet-stream")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	duration := time.Since(start).Seconds()
	if duration > 0 {
		bytesPerSecond = float64(uploadSize) / duration
	}
	return bytesPerSecond, nil
}

// zeroReader is an io.Reader that returns zeros
type zeroReader struct{}

func (z zeroReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// TestPacketLoss sends n pings through the proxy and measures packet loss
func TestPacketLoss(ctx context.Context, socksPort int, testURL string, count int, pingTimeout time.Duration) (lossPercent float64, err error) {
	if count <= 0 {
		count = 5
	}

	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return 100, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
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
			goto done
		default:
		}

		func() {
			// هر ping timeout مستقل داره
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			defer cancel()

			transport := &http.Transport{
				DialContext: func(c context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
					ServerName:         parsedURL.Hostname(),
				},
				DisableKeepAlives: true,
			}

			client := &http.Client{
				Transport: transport,
				Timeout:   pingTimeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			req, e := http.NewRequestWithContext(pingCtx, "GET", testURL, nil)
			if e != nil {
				lost++
				return
			}
			resp, e := client.Do(req)
			if e != nil {
				lost++
			} else {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
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
	lossPercent = float64(lost) / float64(count) * 100
	return lossPercent, nil
}
