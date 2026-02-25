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

// ─── BUILD — معکوس parse ──────────────────────────────────────────────────────

// BuildProxyURL یه config رو با IP جدید به لینک برمیگردونه
func BuildProxyURL(cfg *config.Config, newIP string, origRaw string) (string, error) {
	origRaw = strings.TrimSpace(origRaw)
	proto, remark := detectProtoAndRemark(origRaw)
	switch proto {
	case "vless":
		return buildVless(cfg, newIP, remark)
	case "vmess":
		return buildVmess(cfg, newIP, remark)
	case "trojan":
		return buildTrojan(cfg, newIP, remark)
	}
	return "", fmt.Errorf("پروتکل شناخته‌شده نیست")
}

func detectProtoAndRemark(raw string) (proto, remark string) {
	if strings.HasPrefix(raw, "vless://") {
		proto = "vless"
	} else if strings.HasPrefix(raw, "vmess://") {
		proto = "vmess"
	} else if strings.HasPrefix(raw, "trojan://") {
		proto = "trojan"
	} else {
		return "", ""
	}
	if proto == "vmess" {
		if idx := strings.LastIndex(raw, "#"); idx != -1 {
			r, _ := url.PathUnescape(raw[idx+1:])
			remark = r
		}
	} else {
		if u, err := url.Parse(raw); err == nil {
			r, _ := url.PathUnescape(u.Fragment)
			remark = r
		}
	}
	return
}

func buildVless(cfg *config.Config, newIP, remark string) (string, error) {
	p := cfg.Proxy
	q := url.Values{}
	q.Set("type", p.Type)

	if p.Method == "reality" && p.Reality != nil {
		q.Set("security", "reality")
		q.Set("pbk", p.Reality.PublicKey)
		if p.Reality.ShortId != "" { q.Set("sid", p.Reality.ShortId) }
		if p.Reality.Fingerprint != "" { q.Set("fp", p.Reality.Fingerprint) }
		if p.Reality.ServerName != "" { q.Set("sni", p.Reality.ServerName) }
		if p.Reality.SpiderX != "" { q.Set("spx", p.Reality.SpiderX) }
	} else if p.Method == "none" {
		q.Set("security", "none")
	} else if p.TLS != nil {
		q.Set("security", "tls")
		q.Set("sni", p.TLS.SNI)
		if p.TLS.Fingerprint != "" { q.Set("fp", p.TLS.Fingerprint) }
		if len(p.TLS.ALPN) > 0 { q.Set("alpn", strings.Join(p.TLS.ALPN, ",")) }
		if p.TLS.AllowInsecure { q.Set("allowInsecure", "1") }
	}

	switch p.Type {
	case "ws", "httpupgrade":
		if p.WS != nil { q.Set("path", p.WS.Path); q.Set("host", p.WS.Host) }
	case "grpc":
		if p.Grpc != nil {
			q.Set("serviceName", p.Grpc.ServiceName)
			if p.Grpc.MultiMode { q.Set("mode", "multi") }
		}
	case "xhttp", "h2":
		if p.Xhttp != nil {
			q.Set("path", p.Xhttp.Path)
			q.Set("host", p.Xhttp.Host)
			if p.Xhttp.Mode != "" && p.Xhttp.Mode != "auto" { q.Set("mode", p.Xhttp.Mode) }
		}
	}

	link := fmt.Sprintf("vless://%s@%s:%d?%s", p.UUID, newIP, p.Port, q.Encode())
	if remark != "" { link += "#" + url.PathEscape(remark) }
	return link, nil
}

func buildVmess(cfg *config.Config, newIP, remark string) (string, error) {
	p := cfg.Proxy
	sni, fp, alpn := "", "chrome", ""
	allowInsecure := false
	if p.TLS != nil {
		sni = p.TLS.SNI
		if p.TLS.Fingerprint != "" { fp = p.TLS.Fingerprint }
		if len(p.TLS.ALPN) > 0 { alpn = strings.Join(p.TLS.ALPN, ",") }
		allowInsecure = p.TLS.AllowInsecure
	}

	net_ := p.Type
	if net_ == "xhttp" { net_ = "h2" }
	host_, path_, svcName := "", "/", ""
	switch p.Type {
	case "ws", "httpupgrade":
		if p.WS != nil { host_ = p.WS.Host; path_ = p.WS.Path }
	case "grpc":
		if p.Grpc != nil { svcName = p.Grpc.ServiceName }
	case "xhttp", "h2":
		if p.Xhttp != nil { host_ = p.Xhttp.Host; path_ = p.Xhttp.Path }
	}

	v := map[string]interface{}{
		"v": "2", "ps": remark, "add": newIP, "port": p.Port, "id": p.UUID, "aid": 0,
		"net": net_, "type": "none", "host": host_, "path": path_,
		"tls": map[bool]string{true: "tls", false: ""}[p.Method == "tls"],
		"sni": sni, "fp": fp, "alpn": alpn,
	}
	if svcName != "" { v["serviceName"] = svcName }
	if allowInsecure { v["allowInsecure"] = "1" }

	b, err := json.Marshal(v)
	if err != nil { return "", err }
	return "vmess://" + base64.StdEncoding.EncodeToString(b), nil
}

func buildTrojan(cfg *config.Config, newIP, remark string) (string, error) {
	p := cfg.Proxy
	q := url.Values{}
	q.Set("type", p.Type)
	q.Set("security", "tls")
	if p.TLS != nil {
		q.Set("sni", p.TLS.SNI)
		if p.TLS.Fingerprint != "" { q.Set("fp", p.TLS.Fingerprint) }
		if len(p.TLS.ALPN) > 0 { q.Set("alpn", strings.Join(p.TLS.ALPN, ",")) }
	}
	switch p.Type {
	case "ws":
		if p.WS != nil { q.Set("path", p.WS.Path); q.Set("host", p.WS.Host) }
	case "grpc":
		if p.Grpc != nil { q.Set("serviceName", p.Grpc.ServiceName) }
	}
	link := fmt.Sprintf("trojan://%s@%s:%d?%s", p.UUID, newIP, p.Port, q.Encode())
	if remark != "" { link += "#" + url.PathEscape(remark) }
	return link, nil
}

// ─── MULTI IMPORT ────────────────────────────────────────────────────────────

type ParsedProxy struct {
	RawURL     string `json:"rawUrl"`
	ConfigJSON string `json:"configJson"`
	Remark     string `json:"remark"`
	Protocol   string `json:"protocol"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
}

func ParseMultiProxy(input string) ([]ParsedProxy, []string) {
	var results []ParsedProxy
	var errs []string
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		cfg, err := ParseProxyURL(line)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%.40s… → %v", line, err))
			continue
		}
		cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
		proto, remark := detectProtoAndRemark(line)
		results = append(results, ParsedProxy{
			RawURL: line, ConfigJSON: string(cfgJSON),
			Remark: remark, Protocol: proto,
			Address: cfg.Proxy.Address, Port: cfg.Proxy.Port,
		})
	}
	return results, errs
}

// ─── EXPORT ───────────────────────────────────────────────────────────────────

// BuildClashProxies خروجی Clash YAML برای لیست IP ها
func BuildClashProxies(cfg *config.Config, ips []string, origRaw string) string {
	var sb strings.Builder
	sb.WriteString("proxies:\n")
	for i, ip := range ips {
		link, err := BuildProxyURL(cfg, ip, origRaw)
		if err != nil { continue }
		name := fmt.Sprintf("piyazche-%d", i+1)
		sb.WriteString(buildClashProxy(cfg, ip, name, link))
	}
	sb.WriteString("\nproxy-groups:\n")
	sb.WriteString("  - name: Piyazche\n    type: url-test\n    url: http://www.gstatic.com/generate_204\n    interval: 300\n    proxies:\n")
	for i := range ips {
		sb.WriteString(fmt.Sprintf("      - piyazche-%d\n", i+1))
	}
	return sb.String()
}

func buildClashProxy(cfg *config.Config, ip, name, _ string) string {
	p := cfg.Proxy
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  - name: %s\n", name))

	// protocol
	if strings.HasPrefix(cfg.Proxy.UUID, "") {
		switch {
		case p.Method == "reality":
			sb.WriteString("    type: vless\n")
		default:
			sb.WriteString("    type: vless\n")
		}
	}

	sb.WriteString(fmt.Sprintf("    server: %s\n", ip))
	sb.WriteString(fmt.Sprintf("    port: %d\n", p.Port))
	sb.WriteString(fmt.Sprintf("    uuid: %s\n", p.UUID))

	if p.Method == "reality" && p.Reality != nil {
		sb.WriteString("    tls: true\n    servername: " + p.Reality.ServerName + "\n")
		sb.WriteString("    reality-opts:\n")
		sb.WriteString("      public-key: " + p.Reality.PublicKey + "\n")
		if p.Reality.ShortId != "" {
			sb.WriteString("      short-id: " + p.Reality.ShortId + "\n")
		}
		if p.Reality.Fingerprint != "" {
			sb.WriteString("    client-fingerprint: " + p.Reality.Fingerprint + "\n")
		}
	} else if p.TLS != nil {
		sb.WriteString("    tls: true\n")
		sb.WriteString("    servername: " + p.TLS.SNI + "\n")
		if p.TLS.Fingerprint != "" {
			sb.WriteString("    client-fingerprint: " + p.TLS.Fingerprint + "\n")
		}
		if len(p.TLS.ALPN) > 0 {
			sb.WriteString("    alpn: [" + strings.Join(p.TLS.ALPN, ", ") + "]\n")
		}
	}

	switch p.Type {
	case "ws":
		sb.WriteString("    network: ws\n")
		if p.WS != nil {
			sb.WriteString("    ws-opts:\n      path: " + p.WS.Path + "\n      headers:\n        Host: " + p.WS.Host + "\n")
		}
	case "grpc":
		sb.WriteString("    network: grpc\n")
		if p.Grpc != nil {
			sb.WriteString("    grpc-opts:\n      grpc-service-name: " + p.Grpc.ServiceName + "\n")
		}
	case "h2", "xhttp":
		sb.WriteString("    network: h2\n")
		if p.Xhttp != nil {
			sb.WriteString("    h2-opts:\n      path: " + p.Xhttp.Path + "\n")
		}
	}
	return sb.String()
}

// BuildSingboxOutbounds خروجی Sing-box JSON
func BuildSingboxOutbounds(cfg *config.Config, ips []string) string {
	p := cfg.Proxy
	var outbounds []map[string]interface{}
	var tags []string

	for i, ip := range ips {
		tag := fmt.Sprintf("piyazche-%d", i+1)
		tags = append(tags, tag)

		ob := map[string]interface{}{
			"tag":        tag,
			"type":       "vless",
			"server":     ip,
			"server_port": p.Port,
			"uuid":       p.UUID,
		}

		// TLS
		tlsObj := map[string]interface{}{"enabled": true}
		if p.Method == "reality" && p.Reality != nil {
			tlsObj["server_name"] = p.Reality.ServerName
			tlsObj["utls"] = map[string]string{"enabled": "true", "fingerprint": firstNonEmpty(p.Reality.Fingerprint, "chrome")}
			tlsObj["reality"] = map[string]interface{}{
				"enabled":    true,
				"public_key": p.Reality.PublicKey,
				"short_id":   p.Reality.ShortId,
			}
		} else if p.TLS != nil {
			tlsObj["server_name"] = p.TLS.SNI
			if p.TLS.Fingerprint != "" {
				tlsObj["utls"] = map[string]string{"enabled": "true", "fingerprint": p.TLS.Fingerprint}
			}
			if p.TLS.AllowInsecure { tlsObj["insecure"] = true }
			if len(p.TLS.ALPN) > 0 { tlsObj["alpn"] = p.TLS.ALPN }
		}
		ob["tls"] = tlsObj

		// Transport
		switch p.Type {
		case "ws":
			t := map[string]interface{}{"type": "ws"}
			if p.WS != nil {
				t["path"] = p.WS.Path
				t["headers"] = map[string]string{"Host": p.WS.Host}
			}
			ob["transport"] = t
		case "grpc":
			t := map[string]interface{}{"type": "grpc"}
			if p.Grpc != nil { t["service_name"] = p.Grpc.ServiceName }
			ob["transport"] = t
		case "xhttp", "h2":
			t := map[string]interface{}{"type": "http"}
			if p.Xhttp != nil { t["path"] = p.Xhttp.Path }
			ob["transport"] = t
		}

		outbounds = append(outbounds, ob)
	}

	// selector
	selector := map[string]interface{}{
		"tag":      "proxy",
		"type":     "urltest",
		"outbounds": tags,
		"url":      "http://www.gstatic.com/generate_204",
		"interval": "5m",
	}
	outbounds = append([]map[string]interface{}{selector}, outbounds...)

	result := map[string]interface{}{"outbounds": outbounds}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b)
}
