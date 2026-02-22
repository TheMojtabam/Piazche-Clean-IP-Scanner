# Piyazche

A CDN IP scanner that tests connectivity through xray-core proxy or raw ICMP/TCP.

## What it does

Scans a list of IPs (usually Cloudflare ranges) and finds the ones with lowest latency for your specific proxy configuration. Useful when you need to find clean IPs that work well with your setup.

## Install

### Download pre-built binaries

Download the latest release from the [Releases page](https://github.com/redhatx7/piyazche-scanner/releases).

Available platforms:
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64 / Apple Silicon)

### Build from source

```bash
go build -o piyazche .
```

## Usage

```bash
# Basic scan with xray proxy
./piyazche -c config.json -s ipv4.txt -t 16

# ICMP ping scan (no proxy, just find reachable IPs)
sudo ./piyazche --scan-mode icmp -s ipv4.txt -t 64

# Check single IP connection
./piyazche -c config.json --check

# Auto-optimize fragment settings
./piyazche -c config.json --fragment-mode auto --test-ip 104.27.68.140
```

## How the xray scan timing works

```
                    Per-IP Test Flow
                    ================

    ┌─────────────────────────────────────────────────────────┐
    │                      For each IP                        │
    └─────────────────────────────────────────────────────────┘
                              │
                              ▼
    ┌─────────────────────────────────────────────────────────┐
    │  1. Generate xray config with target IP + fragment      │
    │  2. Start xray-core instance on random local port       │
    │  3. Wait for xray to be ready (up to 10s)               │
    └─────────────────────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │         Retry Loop (default: 3)        │
         │                                        │
         │  ┌──────────────────────────────────┐  │
         │  │  HTTP GET through SOCKS5 proxy   │  │
         │  │  Target: testUrl (gstatic 204)   │  │
         │  │  Timeout: scan.timeout seconds   │  │
         │  └──────────────────────────────────┘  │
         │                 │                      │
         │     ┌───────────┴───────────┐          │
         │     ▼                       ▼          │
         │  Success                  Fail         │
         │     │                       │          │
         │     │              Retry if attempts   │
         │     │                 remaining        │
         │     │                       │          │
         └─────┼───────────────────────┼──────────┘
               │                       │
               ▼                       ▼
    ┌─────────────────┐     ┌─────────────────────┐
    │ Check latency   │     │ Mark as failed      │
    │ vs maxLatency   │     │ Move to next IP     │
    └─────────────────┘     └─────────────────────┘
               │
       ┌───────┴───────┐
       ▼               ▼
    Within           Exceeds
    limit            limit
       │               │
       ▼               ▼
    Record          Discard
    result          result
```

**Timing parameters:**
- `timeout`: How long to wait for HTTP response (per attempt)
- `maxLatency`: Discard results slower than this (in ms)
- `retries`: Number of attempts before giving up on an IP

Example: With `timeout=4`, `maxLatency=2000`, `retries=3`:
- Each attempt waits up to 4 seconds
- If all 3 attempts fail, IP is marked failed
- If any attempt succeeds but takes >2000ms, result is discarded

## How fragment auto-optimization works

```
                Fragment Optimizer Flow
                =======================

    ┌─────────────────────────────────────────────────────────┐
    │  Test 3 zones in sequence: tlshello, 1-3, 1-5          │
    │  Each zone = different packet fragmentation strategy    │
    └─────────────────────────────────────────────────────────┘
                              │
                              ▼
    ┌─────────────────────────────────────────────────────────┐
    │              For each zone (e.g. tlshello)              │
    │                                                         │
    │  Start with full range:                                 │
    │    size = [10-60]  interval = [10-32]                   │
    └─────────────────────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │      Test Loop (max 20 attempts)       │
         │                                        │
         │  1. Test current size+interval range   │
         │  2. Measure success/failure + latency  │
         └────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           Success                          Failure
              │                               │
              ▼                               ▼
    ┌─────────────────┐           ┌─────────────────────┐
    │ Save if best    │           │ Shift to unexplored │
    │ latency so far  │           │ region (4 strategies│
    │                 │           │ rotate each fail):  │
    │ Narrow range:   │           │                     │
    │ [10-60]->[17-48]│           │ 0: shift lower      │
    │ (shrink 38%     │           │ 1: shift upper      │
    │  from edges)    │           │ 2: small+slow combo │
    │                 │           │ 3: widen range      │
    └─────────────────┘           └─────────────────────┘
              │                               │
              └───────────────┬───────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │           Exit conditions:             │
         │                                        │
         │  - Hit success threshold (50% of max)  │
         │  - 3 consecutive failures + have best  │
         │  - Max attempts reached                │
         └────────────────────────────────────────┘
                              │
                              ▼
    ┌─────────────────────────────────────────────────────────┐
    │  Pick best result across all 3 zones (lowest latency)  │
    │  Apply: packets=zone, length=size, interval=interval   │
    └─────────────────────────────────────────────────────────┘
```

**How narrowing works (step by step):**

Each success shrinks the range by ~38% (19% from each edge), converging toward the sweet spot:

```
Fragment SIZE range narrowing
─────────────────────────────

Step 0 (initial):
    10                                                    60
    ├──────────────────────────────────────────────────────┤
    [==========================================================]

Step 1 (success → narrow):
         17                                          48
          ├──────────────────────────────────────────┤
          [============================================]
    cut ──┘▲                                        ▲└── cut
           └── 19% from left                19% from right

Step 2 (success → narrow):
              22                                39
               ├────────────────────────────────┤
               [================================]

Step 3 (success → narrow):
                  26                       35
                   ├───────────────────────┤
                   [=======================]

Step 4 (converged):
                      28              33
                       ├──────────────┤
                       [==============]  ← sweet spot found


Fragment INTERVAL range narrowing
─────────────────────────────────

Step 0:  [10 ─────────────────────────────────────── 32]
Step 1:  [10 ─────────────────────────────────────── 32]  (same range)
              ↓ narrow on success
Step 1:      [14 ─────────────────────────────── 28]
              ↓ narrow on success
Step 2:          [17 ─────────────────────── 25]
              ↓ narrow on success  
Step 3:              [19 ───────────── 23]
              ↓ converged
Step 4:                  [20 ───── 22]  ← optimal interval
```

**The math:**
- Golden ratio constant: `φ = 0.618`
- Narrow ratio: `(1 - φ) / 2 ≈ 0.19` (19% cut from each side)
- Each success: `newMin = min + range×0.19`, `newMax = max - range×0.19`

**Shift strategies on failure:**
- Rotates through 4 approaches to explore different regions
- Avoids getting stuck in a non-working area

```
Shift strategies visualization (on failure)
───────────────────────────────────────────

Current range that failed:
         [==========]
          20     40

Strategy 0 - Shift Lower (explore smaller values):
    [==========]
     10     30
    ←── moved left

Strategy 1 - Shift Upper (explore larger values):
              [==========]
               30     50
               moved right ──→

Strategy 2 - Small + Slow (conservative combo):
       [======]
        15   30
    smaller range, lower values

Strategy 3 - Widen (expand search area):
    [==================]
     10            50
    ←── wider range ──→
```

## Config parameters

### proxy

| Field | Description |
|-------|-------------|
| `uuid` | Your vless UUID |
| `address` | Server address (domain or IP) |
| `port` | Server port (usually 443) |
| `method` | Security method: `tls` or `reality` |
| `type` | Transport type: `ws`, `xhttp`, `grpc`, `httpupgrade`, `tcp` |

### proxy.tls (when method=tls)

| Field | Description |
|-------|-------------|
| `sni` | Server Name Indication |
| `alpn` | ALPN protocols, e.g. `["http/1.1", "h2"]` |
| `fingerprint` | TLS fingerprint: `chrome`, `firefox`, `safari`, etc. |
| `allowInsecure` | Skip certificate verification (not recommended) |

### proxy.reality (when method=reality)

| Field | Description |
|-------|-------------|
| `publicKey` | Reality public key from server |
| `shortId` | Reality short ID |
| `spiderX` | Spider path (usually `/`) |
| `fingerprint` | uTLS fingerprint |
| `serverName` | SNI for reality (usually a real website like `www.domain.com`) |

### proxy.ws (when type=ws)

| Field | Description |
|-------|-------------|
| `host` | WebSocket Host header |
| `path` | WebSocket path |

### proxy.xhttp (when type=xhttp)

| Field | Description |
|-------|-------------|
| `host` | HTTP host |
| `path` | HTTP path |
| `mode` | Mode: `auto`, `stream`, etc. |

### proxy.grpc (when type=grpc)

| Field | Description |
|-------|-------------|
| `serviceName` | gRPC service name |
| `authority` | gRPC authority (optional) |
| `multiMode` | Enable multi-mode |

### fragment

| Field | Description |
|-------|-------------|
| `enabled` | Enable TLS fragmentation |
| `mode` | `manual` (use settings below) or `auto` (discover optimal) |
| `packets` | Which packets to fragment: `tlshello`, `1-3`, `1-5` |
| `manual.length` | Fragment size range, e.g. `"10-20"` |
| `manual.interval` | Delay between fragments in ms, e.g. `"10-20"` |

### scan

| Field | Description |
|-------|-------------|
| `threads` | Number of concurrent workers |
| `timeout` | HTTP request timeout in seconds |
| `testUrl` | URL to test (default: gstatic 204) |
| `maxLatency` | Max acceptable latency in ms |
| `retries` | Retry count per IP |
| `sampleSize` | IPs to sample per subnet |

### xray.mux

| Field | Description |
|-------|-------------|
| `enabled` | Enable mux multiplexing |
| `concurrency` | Number of concurrent streams |

## Sample configs

### WebSocket + TLS

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server.com",
    "port": 443,
    "method": "tls",
    "type": "ws",
    "tls": {
      "sni": "your-sni.com",
      "alpn": ["http/1.1"],
      "fingerprint": "chrome"
    },
    "ws": {
      "host": "your-host.com",
      "path": "/ws-path"
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "manual",
    "packets": "tlshello",
    "manual": {
      "length": "10-20",
      "interval": "10-20"
    }
  },
  "scan": {
    "threads": 8,
    "timeout": 4,
    "testUrl": "https://www.gstatic.com/generate_204",
    "maxLatency": 3000,
    "retries": 2
  }
}
```

### WebSocket + Reality

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server-ip",
    "port": 443,
    "method": "reality",
    "type": "ws",
    "reality": {
      "publicKey": "your-public-key",
      "shortId": "your-short-id",
      "spiderX": "/",
      "fingerprint": "chrome",
      "serverName": "www.domain.com"
    },
    "ws": {
      "host": "",
      "path": "/"
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "auto"
  },
  "scan": {
    "threads": 1,
    "timeout": 10,
    "maxLatency": 3000,
    "retries": 2
  }
}
```

### XHTTP + TLS

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server.com",
    "port": 443,
    "method": "tls",
    "type": "xhttp",
    "tls": {
      "sni": "your-sni.com",
      "alpn": ["h2", "http/1.1"],
      "fingerprint": "chrome"
    },
    "xhttp": {
      "host": "your-host.com",
      "path": "/xhttp-path",
      "mode": "auto"
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "manual",
    "packets": "tlshello",
    "manual": {
      "length": "15-30",
      "interval": "10-20"
    }
  },
  "scan": {
    "threads": 8,
    "timeout": 4,
    "maxLatency": 2500,
    "retries": 3
  }
}
```

### XHTTP + Reality

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server-ip",
    "port": 443,
    "method": "reality",
    "type": "xhttp",
    "reality": {
      "publicKey": "your-public-key",
      "shortId": "your-short-id",
      "spiderX": "/",
      "fingerprint": "chrome",
      "serverName": "www.domain.com"
    },
    "xhttp": {
      "host": "",
      "path": "/",
      "mode": "auto"
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "auto"
  },
  "scan": {
    "threads": 1,
    "timeout": 10,
    "maxLatency": 3000,
    "retries": 2
  }
}
```

### gRPC + TLS

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server.com",
    "port": 443,
    "method": "tls",
    "type": "grpc",
    "tls": {
      "sni": "your-sni.com",
      "alpn": ["h2"],
      "fingerprint": "chrome"
    },
    "grpc": {
      "serviceName": "your-grpc-service",
      "multiMode": false
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "manual",
    "packets": "tlshello",
    "manual": {
      "length": "10-20",
      "interval": "10-20"
    }
  },
  "scan": {
    "threads": 8,
    "timeout": 5,
    "maxLatency": 3000,
    "retries": 2
  }
}
```

### gRPC + Reality

```json
{
  "proxy": {
    "uuid": "your-uuid",
    "address": "your-server-ip",
    "port": 443,
    "method": "reality",
    "type": "grpc",
    "reality": {
      "publicKey": "your-public-key",
      "shortId": "your-short-id",
      "spiderX": "/",
      "fingerprint": "chrome",
      "serverName": "www.domain.com"
    },
    "grpc": {
      "serviceName": "grpc-service",
      "multiMode": false
    }
  },
  "fragment": {
    "enabled": true,
    "mode": "auto"
  },
  "scan": {
    "threads": 1,
    "timeout": 10,
    "maxLatency": 3000,
    "retries": 2
  }
}
```

## CLI flags

```
-c, --config         Config file path (default: config.json)
-s, --subnets        IP list file or CIDR (default: ipv4.txt)
-t, --threads        Worker count (overrides config)
-o, --output         Output format: csv, json
    --max-ips        Limit number of IPs to scan
    --shuffle        Randomize IP order (default: true)
    --top            Show top N results (default: 10)
    --debug          Print xray config for first IP
    --fragment-mode  Fragment mode: manual, auto, off
    --test-ip        IP for fragment optimization
    --check          Test single connection (required for reality)
    --mux            Enable mux: true, false
    --scan-mode      Scan mode: xray (default), icmp
```

## Notes

- Reality mode requires `--check` flag or `--fragment-mode auto` (scanner mode not supported)
- ICMP scan mode needs root for real ICMP, falls back to TCP connect without root
- Results are saved to `results/` directory as CSV or JSON
- Higher thread count = faster scan but more resource usage
# P
# P
# P
# P
# P
