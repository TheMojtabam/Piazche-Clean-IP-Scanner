package webui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"piyazche/config"
)

// ParseProxyURL تبدیل vless:// vmess:// trojan:// یا JSON به config.Config
func ParseProxyURL(input string) (*config.Config, error) {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "vless://") {
		return parseVless(input)
	}
	if strings.HasPrefix(input, "vmess://") {
		return parseVmess(input)
	}
	if strings.HasPrefix(input, "trojan://") {
		return parseTrojan(input)
	}
	if strings.HasPrefix(input, "{") {
		return parseJSONConfig(input)
	}

	return nil, fmt.Errorf("فرمت شناخته‌شده نیست — vless:// vmess:// trojan:// یا JSON بفرست")
}

// --- VLESS parser ---
// vless://uuid@host:port?type=ws&security=tls&sni=...&path=...&host=...&fp=...&pbk=...&sid=...#remark
func parseVless(raw string) (*config.Config, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("vless URL parse error: %w", err)
	}

	uuid := u.User.Username()
	if uuid == "" {
		return nil, fmt.Errorf("vless: UUID خالی است")
	}

	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		portStr = "443"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("vless: port نامعتبر: %s", portStr)
	}

	q := u.Query()
	transportType := q.Get("type")
	if transportType == "" {
		transportType = "tcp"
	}
	if transportType == "h2" {
		transportType = "xhttp"
	}
	security := q.Get("security")
	if security == "" {
		security = "tls"
	}

	sni := firstNonEmpty(q.Get("sni"), q.Get("serverName"), host)
	fp := firstNonEmpty(q.Get("fp"), "chrome")
	alpnStr := q.Get("alpn")
	alpn := parseALPN(alpnStr)

	cfg := config.DefaultConfig()
	cfg.Proxy.UUID = uuid
	cfg.Proxy.Address = host
	cfg.Proxy.Port = port
	cfg.Proxy.Type = transportType
	cfg.Proxy.Method = security

	if security == "reality" {
		cfg.Proxy.Reality = &config.RealityConfig{
			PublicKey:   q.Get("pbk"),
			ShortId:     firstNonEmpty(q.Get("sid"), q.Get("shortId")),
			SpiderX:     q.Get("spx"),
			Fingerprint: fp,
			ServerName:  sni,
		}
		cfg.Proxy.TLS = nil
	} else if security == "none" || security == "" && q.Get("security") == "none" {
		// No TLS — plaintext connection
		cfg.Proxy.Method = "none"
		cfg.Proxy.TLS = nil
	} else {
		cfg.Proxy.Method = "tls"
		allowInsecure := q.Get("allowInsecure") == "1" || q.Get("allowInsecure") == "true"
		cfg.Proxy.TLS = &config.TlsConfig{
			SNI:           sni,
			Fingerprint:   fp,
			ALPN:          alpn,
			AllowInsecure: allowInsecure,
		}
	}

	switch transportType {
	case "ws":
		wsPath := firstNonEmpty(q.Get("path"), "/")
		wsHost := firstNonEmpty(q.Get("host"), sni)
		cfg.Proxy.WS = &config.WsConfig{Host: wsHost, Path: wsPath}

	case "grpc":
		cfg.Proxy.Grpc = &config.GrpcConfig{
			ServiceName: q.Get("serviceName"),
			Authority:   q.Get("authority"),
			MultiMode:   q.Get("mode") == "multi",
		}

	case "xhttp", "h2":
		cfg.Proxy.Xhttp = &config.XhttpConfig{
			Host: firstNonEmpty(q.Get("host"), sni),
			Path: firstNonEmpty(q.Get("path"), "/"),
			Mode: firstNonEmpty(q.Get("mode"), "auto"),
		}

	case "httpupgrade":
		cfg.Proxy.WS = &config.WsConfig{
			Host: firstNonEmpty(q.Get("host"), sni),
			Path: firstNonEmpty(q.Get("path"), "/"),
		}
	}

	return cfg, nil
}

// --- VMESS parser ---
func parseVmess(raw string) (*config.Config, error) {
	b64 := strings.TrimPrefix(raw, "vmess://")
	b64 = strings.TrimRight(b64, "=")
	switch len(b64) % 4 {
	case 2:
		b64 += "=="
	case 3:
		b64 += "="
	}

	var data []byte
	var err error
	data, err = base64.StdEncoding.DecodeString(b64)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(b64)
		if err != nil {
			data, err = base64.RawStdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("vmess: base64 decode error: %w", err)
			}
		}
	}

	var v struct {
		ID      string      `json:"id"`
		Add     string      `json:"add"`
		Port    interface{} `json:"port"`
		Net     string      `json:"net"`
		Type    string      `json:"type"`
		Host    string      `json:"host"`
		Path    string      `json:"path"`
		TLS     string      `json:"tls"`
		SNI     string      `json:"sni"`
		Fp      string      `json:"fp"`
		Alpn    string      `json:"alpn"`
		Aid     interface{} `json:"aid"`
		// grpc
		ServiceName string `json:"serviceName"`
		// xhttp
		Mode string `json:"mode"`
		// allow insecure
		AllowInsecure string `json:"allowInsecure"`
		SkipCertVerify bool  `json:"skip-cert-verify"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("vmess: JSON parse error: %w", err)
	}

	if v.ID == "" {
		return nil, fmt.Errorf("vmess: UUID خالی است")
	}

	port := 443
	switch p := v.Port.(type) {
	case float64:
		port = int(p)
	case string:
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	transportType := firstNonEmpty(v.Net, "tcp")
	if transportType == "h2" {
		transportType = "xhttp"
	}

	sni := firstNonEmpty(v.SNI, v.Host, v.Add)
	fp := firstNonEmpty(v.Fp, "chrome")
	alpn := parseALPN(v.Alpn)

	cfg := config.DefaultConfig()
	cfg.Proxy.UUID = v.ID
	cfg.Proxy.Address = v.Add
	cfg.Proxy.Port = port
	cfg.Proxy.Type = transportType
	cfg.Proxy.Method = "tls"

	allowInsecure := v.AllowInsecure == "1" || v.AllowInsecure == "true" || v.SkipCertVerify
	cfg.Proxy.TLS = &config.TlsConfig{
		SNI:           sni,
		Fingerprint:   fp,
		ALPN:          alpn,
		AllowInsecure: allowInsecure,
	}

	switch transportType {
	case "ws":
		cfg.Proxy.WS = &config.WsConfig{
			Host: firstNonEmpty(v.Host, sni),
			Path: firstNonEmpty(v.Path, "/"),
		}
	case "grpc":
		cfg.Proxy.Grpc = &config.GrpcConfig{
			ServiceName: firstNonEmpty(v.ServiceName, v.Path),
			MultiMode:   v.Type == "multi",
		}
	case "xhttp":
		cfg.Proxy.Xhttp = &config.XhttpConfig{
			Host: firstNonEmpty(v.Host, sni),
			Path: firstNonEmpty(v.Path, "/"),
			Mode: firstNonEmpty(v.Mode, "auto"),
		}
	case "httpupgrade":
		cfg.Proxy.WS = &config.WsConfig{
			Host: firstNonEmpty(v.Host, sni),
			Path: firstNonEmpty(v.Path, "/"),
		}
	}

	return cfg, nil
}

// --- Trojan parser ---
// trojan://password@host:port?type=tcp&security=tls&sni=...#remark
func parseTrojan(raw string) (*config.Config, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("trojan URL parse error: %w", err)
	}

	password := u.User.Username()
	host := u.Hostname()
	portStr := firstNonEmpty(u.Port(), "443")
	port, _ := strconv.Atoi(portStr)
	if port <= 0 {
		port = 443
	}

	q := u.Query()
	transportType := firstNonEmpty(q.Get("type"), "tcp")
	if transportType == "h2" {
		transportType = "xhttp"
	}
	sni := firstNonEmpty(q.Get("sni"), q.Get("serverName"), host)
	fp := firstNonEmpty(q.Get("fp"), "chrome")
	alpn := parseALPN(q.Get("alpn"))

	cfg := config.DefaultConfig()
	cfg.Proxy.UUID = password // trojan uses password as uuid field
	cfg.Proxy.Address = host
	cfg.Proxy.Port = port
	cfg.Proxy.Type = transportType
	cfg.Proxy.Method = "tls"
	cfg.Proxy.TLS = &config.TlsConfig{
		SNI:         sni,
		Fingerprint: fp,
		ALPN:        alpn,
	}

	if transportType == "ws" {
		cfg.Proxy.WS = &config.WsConfig{
			Host: firstNonEmpty(q.Get("host"), sni),
			Path: firstNonEmpty(q.Get("path"), "/"),
		}
	}
	if transportType == "grpc" {
		cfg.Proxy.Grpc = &config.GrpcConfig{
			ServiceName: q.Get("serviceName"),
		}
	}

	return cfg, nil
}

func parseJSONConfig(raw string) (*config.Config, error) {
	cfg := config.DefaultConfig()
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}
	return cfg, nil
}

// --- helpers ---

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseALPN(s string) []string {
	if s == "" {
		return []string{"http/1.1", "h2"}
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"http/1.1", "h2"}
	}
	return out
}
