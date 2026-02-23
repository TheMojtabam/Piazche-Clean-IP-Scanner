package webui

const indexHTMLContent = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Piyazche Scanner</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=Vazirmatn:wght@300;400;500;600;700&display=swap" rel="stylesheet">
<style>
:root{
  --bg:#070809;--bg2:#0d0f12;--bg3:#12151a;--bg4:#181c22;
  --bd:#1a1e28;--bd2:#232836;--bd3:#2a3040;
  --tx:#d4d8e8;--tx2:#9ba3bc;--dim:#4a5168;
  --g:#3dd68c;--gd:#0d2b1c;
  --c:#38bdf8;--cd:#0a2030;--c2:#0ea5e9;
  --y:#fbbf24;--yd:#271d06;
  --r:#f87171;--rd:#2a0f0f;
  --p:#a78bfa;--pd:#1a1040;
  --o:#fb923c;
  --radius:10px;--radius-sm:6px;
}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Vazirmatn',Tahoma,sans-serif;background:var(--bg);color:var(--tx);min-height:100vh;font-size:14px;line-height:1.6}
.layout{display:grid;grid-template-columns:240px 1fr;grid-template-rows:56px 1fr;min-height:100vh}
.topbar{grid-column:1/-1;background:rgba(13,15,18,.95);border-bottom:1px solid var(--bd);backdrop-filter:blur(12px);display:flex;align-items:center;padding:0 20px;gap:14px;position:sticky;top:0;z-index:100}
.sidebar{background:var(--bg2);border-left:1px solid var(--bd);padding:12px 0 20px;display:flex;flex-direction:column;gap:2px;overflow-y:auto}
.main{overflow-y:auto;padding:20px 24px}
.logo{font-family:'IBM Plex Mono',monospace;font-size:18px;font-weight:600;color:var(--c);letter-spacing:-1px;user-select:none}
.logo b{color:var(--g)}
.logo small{color:var(--dim);font-size:11px;margin-right:6px;font-weight:400}
.status-badge{display:flex;align-items:center;gap:7px;padding:5px 14px;border-radius:20px;font-size:12px;background:var(--bg3);border:1px solid var(--bd2);font-family:'IBM Plex Mono',monospace}
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0;transition:background .3s}
.dot-idle{background:var(--dim)}
.dot-live{background:var(--g);box-shadow:0 0 8px var(--g);animation:pulse 1.4s infinite}
.dot-warn{background:var(--y)}
.dot-done{background:var(--c)}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:.4;transform:scale(.8)}}
.topbar-right{margin-right:auto;display:flex;align-items:center;gap:10px}
.nav-group{padding:6px 12px 3px;font-size:10px;letter-spacing:2px;text-transform:uppercase;color:var(--dim);margin-top:8px}
.nav-item{display:flex;align-items:center;gap:10px;padding:9px 14px;cursor:pointer;transition:all .15s;color:var(--tx2);font-size:13px;border:none;background:none;width:100%;text-align:right;border-right:2px solid transparent}
.nav-item:hover{background:var(--bg3);color:var(--tx)}
.nav-item.active{background:var(--cd);color:var(--c);border-right-color:var(--c)}
.nav-icon{font-size:14px;min-width:18px;text-align:center;flex-shrink:0}
.page{display:none}.page.active{display:block}
.page-hd{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:18px;gap:12px}
.page-hd-left h2{font-size:19px;font-weight:600}.page-hd-left p{font-size:12px;color:var(--dim);margin-top:2px}
.page-hd-actions{display:flex;gap:8px;align-items:center;flex-shrink:0}
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--radius);overflow:hidden;margin-bottom:14px}
.card-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:9px 16px;font-size:11px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:'IBM Plex Mono',monospace;letter-spacing:.5px}
.card-hd-left{display:flex;align-items:center;gap:7px}
.card-bd{padding:16px}
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:10px;margin-bottom:14px}
.stat{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--radius-sm);padding:12px 14px;text-align:center}
.stat-v{font-size:24px;font-weight:700;font-family:'IBM Plex Mono',monospace;line-height:1.2}
.stat-l{font-size:10px;color:var(--dim);margin-top:4px;letter-spacing:.5px;text-transform:uppercase}
.prog-wrap{background:var(--bg3);border-radius:4px;height:6px;overflow:hidden;margin:8px 0}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:4px;transition:width .4s ease}
.prog-bar.ph2{background:linear-gradient(90deg,var(--p),var(--c))}
/* Live ticker */
.live-ticker{background:var(--bg3);border:1px solid var(--bd2);border-radius:var(--radius-sm);padding:10px 14px;font-family:'IBM Plex Mono',monospace;font-size:12px;min-height:42px;display:flex;align-items:center;gap:10px;margin-top:10px;overflow:hidden}
.spinner{width:16px;height:16px;border:2px solid var(--bd3);border-top-color:var(--c);border-radius:50%;animation:spin .7s linear infinite;flex-shrink:0}
@keyframes spin{to{transform:rotate(360deg)}}
.live-ip-text{color:var(--c);font-weight:500;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.live-phase{color:var(--dim);font-size:10px;background:var(--bd2);padding:2px 6px;border-radius:4px;flex-shrink:0}
/* TUI Terminal */
.tui-wrap{background:#020304;border:1px solid var(--bd2);border-radius:var(--radius-sm);overflow:hidden;font-family:'IBM Plex Mono',monospace;font-size:12px}
.tui-header{background:var(--bg3);border-bottom:1px solid var(--bd);padding:6px 14px;display:flex;align-items:center;gap:8px;font-size:11px;color:var(--dim)}
.tui-dots{display:flex;gap:5px}
.tui-dot{width:10px;height:10px;border-radius:50%}
.tui-body{padding:12px 14px;max-height:380px;overflow-y:auto;line-height:1.8}
.tui-body::-webkit-scrollbar{width:4px}
.tui-body::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
.tui-line{display:flex;gap:10px;align-items:flex-start}
.tui-time{color:#2d3550;flex-shrink:0;user-select:none}
.tui-ok{color:#3dd68c}
.tui-err{color:#f87171}
.tui-info{color:#38bdf8}
.tui-warn{color:#fbbf24}
.tui-scan{color:#6b7a9e}
.tui-phase2{color:#a78bfa}
.tui-cursor{display:inline-block;width:8px;height:14px;background:var(--c);animation:blink 1s step-end infinite;vertical-align:middle;margin-right:2px}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}
/* Table */
.tbl{width:100%;border-collapse:collapse;font-size:12px;font-family:'IBM Plex Mono',monospace}
.tbl th{padding:8px 10px;text-align:right;color:var(--dim);font-weight:500;border-bottom:1px solid var(--bd);background:var(--bg3);font-size:11px;letter-spacing:.5px;white-space:nowrap}
.tbl td{padding:7px 10px;border-bottom:1px solid var(--bd);vertical-align:middle}
.tbl tr:last-child td{border-bottom:none}
.tbl tr:hover td{background:rgba(255,255,255,.02)}
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:10px;font-size:10px;font-family:'IBM Plex Mono',monospace;font-weight:500}
.bg{background:var(--gd);color:var(--g)}.by{background:var(--yd);color:var(--y)}.br{background:var(--rd);color:var(--r)}.bc{background:var(--cd);color:var(--c)}.bp{background:var(--pd);color:var(--p)}
.btn{display:inline-flex;align-items:center;gap:6px;padding:8px 16px;border-radius:var(--radius-sm);border:1px solid var(--bd2);background:var(--bg3);color:var(--tx);cursor:pointer;font-size:13px;font-family:inherit;transition:all .15s;white-space:nowrap}
.btn:hover{background:var(--bd2)}.btn:active{transform:scale(.97)}.btn:disabled{opacity:.4;cursor:not-allowed;pointer-events:none}
.btn-primary{background:var(--cd);border-color:var(--c2);color:var(--c)}.btn-primary:hover{background:var(--c2);color:#000}
.btn-danger{background:var(--rd);border-color:var(--r);color:var(--r)}.btn-danger:hover{background:var(--r);color:#fff}
.btn-warn{background:var(--yd);border-color:var(--y);color:var(--y)}.btn-warn:hover{background:var(--y);color:#000}
.btn-sm{padding:5px 10px;font-size:11px}.btn-xs{padding:3px 8px;font-size:10px}
textarea,input[type=text],input[type=number],input[type=password],select{background:var(--bg3);border:1px solid var(--bd2);color:var(--tx);border-radius:var(--radius-sm);padding:8px 12px;font-size:13px;font-family:'IBM Plex Mono',monospace;width:100%;outline:none;direction:ltr;transition:border-color .15s}
textarea:focus,input:focus,select:focus{border-color:var(--c)}
label{display:block;font-size:12px;color:var(--dim);margin-bottom:5px;text-align:right;font-family:'Vazirmatn',sans-serif}
.form-row{margin-bottom:12px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}.form-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.form-sep{height:1px;background:var(--bd);margin:16px 0}
.check-row{display:flex;align-items:center;gap:8px;cursor:pointer;font-size:13px;color:var(--tx2)}
.check-row input{width:auto;cursor:pointer;accent-color:var(--c)}
.parsed-box{background:var(--bg3);border:1px solid var(--gd);border-radius:var(--radius-sm);padding:14px;font-family:'IBM Plex Mono',monospace;font-size:12px;color:var(--g);direction:ltr;line-height:1.8}
.parsed-box .k{color:var(--dim)}.parsed-box .v{color:var(--c)}
.cfg-summary{background:var(--bg3);border:1px solid var(--bd2);border-radius:var(--radius-sm);padding:8px 12px;font-size:12px;color:var(--dim);font-family:'IBM Plex Mono',monospace}
.proxy-badge{display:inline-flex;align-items:center;gap:6px;padding:4px 10px;border-radius:5px;font-size:11px;font-family:'IBM Plex Mono',monospace;background:var(--gd);border:1px solid var(--g);color:var(--g)}
.alert{border-radius:var(--radius-sm);padding:10px 14px;font-size:12px;margin-bottom:12px;border-left:3px solid}
.alert-info{background:var(--cd);border-color:var(--c);color:var(--c)}.alert-warn{background:var(--yd);border-color:var(--y);color:var(--y)}.alert-err{background:var(--rd);border-color:var(--r);color:var(--r)}
.ip-list{display:flex;flex-wrap:wrap;gap:6px;padding:4px 0}
.ip-chip{background:var(--cd);border:1px solid var(--c);border-radius:5px;padding:4px 10px;font-family:'IBM Plex Mono',monospace;font-size:12px;color:var(--c);cursor:pointer;display:flex;align-items:center;gap:6px;transition:all .15s}
.ip-chip:hover{background:var(--c);color:#000}
.hist-item{background:var(--bg3);border:1px solid var(--bd);border-radius:var(--radius-sm);padding:12px 16px;display:flex;align-items:center;gap:14px;margin-bottom:8px;cursor:pointer;transition:all .15s}
.hist-item:hover{border-color:var(--bd3)}
.copy-btn{background:none;border:none;cursor:pointer;color:var(--dim);padding:2px 6px;border-radius:4px;font-size:11px;transition:all .15s}
.copy-btn:hover{color:var(--c);background:var(--cd)}
::-webkit-scrollbar{width:5px;height:5px}
::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:3px}
@media(max-width:768px){.layout{grid-template-columns:1fr}.sidebar{display:none}.stats-row{grid-template-columns:repeat(2,1fr)}.form-grid,.form-grid-3{grid-template-columns:1fr}}
</style>
</head>
<body>
<div class="layout">

<!-- TOPBAR -->
<div class="topbar">
  <div class="logo">piy<b>az</b>che <small>scanner</small></div>
  <div class="status-badge" id="statusBadge">
    <div class="dot dot-idle" id="statusDot"></div>
    <span id="statusText" style="font-family:'IBM Plex Mono',monospace">آماده</span>
  </div>
  <div id="proxyBadge" style="display:none" class="proxy-badge">✓ <span id="proxyBadgeText">کانفیگ لود شده</span></div>
  <div class="topbar-right">
    <span style="font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--dim)" id="progressText"></span>
  </div>
</div>

<!-- SIDEBAR -->
<div class="sidebar">
  <div class="nav-group">اسکن</div>
  <button class="nav-item active" data-page="scan" onclick="nav('scan',this)"><span class="nav-icon">⚡</span>اسکن جدید</button>
  <button class="nav-item" data-page="results" onclick="nav('results',this)"><span class="nav-icon">📊</span>نتایج</button>
  <button class="nav-item" data-page="history" onclick="nav('history',this)"><span class="nav-icon">🕐</span>تاریخچه</button>
  <div class="nav-group">پیکربندی</div>
  <button class="nav-item" data-page="config" onclick="nav('config',this)"><span class="nav-icon">⚙️</span>تنظیمات اسکنر</button>
  <button class="nav-item" data-page="import" onclick="nav('import',this)"><span class="nav-icon">🔗</span>وارد کردن کانفیگ</button>
  <div class="nav-group">ابزارها</div>
  <button class="nav-item" data-page="shodan" onclick="nav('shodan',this)"><span class="nav-icon">🔍</span>Shodan</button>
  <button class="nav-item" data-page="tui" onclick="nav('tui',this)"><span class="nav-icon">⌨️</span>ترمینال</button>
</div>

<div class="main">

<!-- ══ SCAN ══ -->
<div id="page-scan" class="page active">
  <div class="page-hd">
    <div class="page-hd-left"><h2>اسکن جدید</h2><p>رنج IP بده و شروع کن</p></div>
    <div class="page-hd-actions">
      <button class="btn btn-primary" id="btnStart" onclick="startScan()">▶ شروع</button>
      <button class="btn btn-warn" id="btnPause" onclick="pauseScan()" style="display:none">⏸ توقف</button>
      <button class="btn btn-danger" id="btnStop" onclick="stopScan()" style="display:none">■ متوقف</button>
    </div>
  </div>

  <div class="stats-row">
    <div class="stat"><div class="stat-v" id="stTotal" style="color:var(--tx)">—</div><div class="stat-l">کل IP</div></div>
    <div class="stat"><div class="stat-v" id="stDone" style="color:var(--c)">0</div><div class="stat-l">بررسی شده</div></div>
    <div class="stat"><div class="stat-v" id="stOk" style="color:var(--g)">0</div><div class="stat-l">موفق</div></div>
    <div class="stat"><div class="stat-v" id="stFail" style="color:var(--r)">0</div><div class="stat-l">ناموفق</div></div>
    <div class="stat"><div class="stat-v" id="stETA" style="color:var(--y)">—</div><div class="stat-l">زمان باقی</div></div>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">📶 پیشرفت</div><span id="phaseLabel" style="color:var(--dim);font-size:11px">Phase 1</span></div>
    <div class="card-bd">
      <div style="display:flex;justify-content:space-between;font-size:12px;color:var(--dim);margin-bottom:4px">
        <span id="progDetail">در انتظار شروع...</span>
        <span id="pctLabel" style="font-family:'IBM Plex Mono',monospace;color:var(--c);font-weight:600">0%</span>
      </div>
      <div class="prog-wrap"><div class="prog-bar" id="progressBar" style="width:0%"></div></div>
      <div class="live-ticker" id="liveTicker">
        <div style="color:var(--dim)">⊙ آماده — اسکن را شروع کن</div>
      </div>
    </div>
  </div>

  <div class="form-grid">
    <div>
      <div class="card">
        <div class="card-hd"><div class="card-hd-left">🌐 رنج IP</div><button class="btn-xs btn" onclick="previewIPs()" style="padding:2px 8px;font-size:10px">پیش‌نمایش</button></div>
        <div class="card-bd">
          <div class="form-row">
            <label>هر خط: IP یا CIDR — خالی = از ipv4.txt</label>
            <textarea id="ipInput" rows="7" placeholder="104.16.0.0/12&#10;185.42.0.0/16&#10;45.12.33.91"></textarea>
          </div>
          <div class="form-grid">
            <div class="form-row"><label>حداکثر IP (0=همه)</label><input type="number" id="maxIPs" value="0" min="0"></div>
            <div class="form-row"><label>IP از هر سابنت</label><input type="number" id="sampleSize" value="1" min="1"></div>
          </div>
          <div id="ipPreview" style="display:none;margin-top:8px;font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--dim)"></div>
        </div>
      </div>
    </div>
    <div>
      <div class="card">
        <div class="card-hd"><div class="card-hd-left">⚡ تنظیمات سریع</div></div>
        <div class="card-bd">
          <div class="form-grid">
            <div class="form-row"><label>Threads</label><input type="number" id="qThreads" value="200" min="1"></div>
            <div class="form-row"><label>Timeout (ثانیه)</label><input type="number" id="qTimeout" value="8" min="1"></div>
            <div class="form-row"><label>Max Latency (ms)</label><input type="number" id="qMaxLat" value="3500"></div>
            <div class="form-row"><label>Stability Rounds</label><input type="number" id="qRounds" value="3" min="0"></div>
          </div>
          <div class="cfg-summary" id="configSummary" style="margin-top:8px">پیش‌فرض — لینک وارد نشده</div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- ══ RESULTS ══ -->
<div id="page-results" class="page">
  <div class="page-hd">
    <div class="page-hd-left"><h2>نتایج</h2><p id="resultsSummary">هنوز اسکنی نشده</p></div>
    <div class="page-hd-actions">
      <button class="btn btn-sm" onclick="exportResults('txt')">📥 .txt</button>
      <button class="btn btn-sm" onclick="exportResults('json')">📥 JSON</button>
    </div>
  </div>
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">✅ IP های موفق — کلیک برای کپی</div><span id="passedCount" style="color:var(--g);font-family:'IBM Plex Mono',monospace">0</span></div>
    <div class="card-bd"><div class="ip-list" id="ipChips"><span style="color:var(--dim)">نتیجه‌ای نیست</span></div></div>
  </div>
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">📊 Phase 2 جزئیات</div></div>
    <div class="card-bd" style="padding:0;overflow-x:auto">
      <table class="tbl">
        <thead><tr><th>#</th><th>IP</th><th>Score</th><th>Avg Lat</th><th>Jitter</th><th>PktLoss</th><th>Download</th><th>Upload</th><th>وضعیت</th><th></th></tr></thead>
        <tbody id="resultsTbody"><tr><td colspan="10" style="text-align:center;color:var(--dim);padding:28px">نتیجه‌ای نیست</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<!-- ══ HISTORY ══ -->
<div id="page-history" class="page">
  <div class="page-hd"><div class="page-hd-left"><h2>تاریخچه</h2></div></div>
  <div id="historyList"><p style="color:var(--dim)">هنوز اسکنی انجام نشده</p></div>
</div>

<!-- ══ CONFIG ══ -->
<div id="page-config" class="page">
  <div class="page-hd">
    <div class="page-hd-left"><h2>تنظیمات اسکنر</h2><p>scan · phase2 · fragment · xray · shodan</p></div>
    <button class="btn btn-primary" onclick="saveConfig()">💾 ذخیره و اعمال</button>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">⚡ Phase 1 — اسکن اولیه</div></div>
    <div class="card-bd">
      <div class="form-grid-3">
        <div class="form-row"><label>Threads</label><input type="number" id="cfgThreads" value="200" min="1"></div>
        <div class="form-row"><label>Timeout (ثانیه)</label><input type="number" id="cfgTimeout" value="8" min="1"></div>
        <div class="form-row"><label>Max Latency (ms)</label><input type="number" id="cfgMaxLat" value="3500"></div>
        <div class="form-row"><label>Retries</label><input type="number" id="cfgRetries" value="2" min="0"></div>
        <div class="form-row"><label>Max IPs (0=همه)</label><input type="number" id="cfgMaxIPs" value="0" min="0"></div>
        <div class="form-row"><label>Sample از هر subnet</label><input type="number" id="cfgSampleSize" value="1" min="1"></div>
      </div>
      <div class="form-row"><label>Test URL</label><input type="text" id="cfgTestURL" value="https://www.gstatic.com/generate_204"></div>
      <label class="check-row" style="margin-top:8px"><input type="checkbox" id="cfgShuffle" checked> Shuffle IPs</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔬 Phase 2 — تست عمقی</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Stability Rounds (0=غیرفعال)</label><input type="number" id="cfgRounds" value="3" min="0"></div>
        <div class="form-row"><label>Interval بین rounds (ثانیه)</label><input type="number" id="cfgInterval" value="5" min="1"></div>
        <div class="form-row"><label>Packet Loss Count (ping تعداد)</label><input type="number" id="cfgPLCount" value="5" min="1"></div>
        <div class="form-row"><label>Max Packet Loss % (-1=غیرفعال)</label><input type="number" id="cfgMaxPL" value="-1" step="0.1"></div>
        <div class="form-row"><label>Min Download Mbps (0=غیرفعال)</label><input type="number" id="cfgMinDL" value="0" step="0.1"></div>
        <div class="form-row"><label>Min Upload Mbps (0=غیرفعال)</label><input type="number" id="cfgMinUL" value="0" step="0.1"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap;margin-top:6px">
        <label class="check-row"><input type="checkbox" id="cfgSpeed"> Speed Test</label>
        <label class="check-row"><input type="checkbox" id="cfgJitter" checked> Jitter Test</label>
      </div>
      <div class="form-sep"></div>
      <div class="form-grid">
        <div class="form-row"><label>Download URL</label><input type="text" id="cfgDLURL" value="https://speed.cloudflare.com/__down?bytes=1000000"></div>
        <div class="form-row"><label>Upload URL</label><input type="text" id="cfgULURL" value="https://speed.cloudflare.com/__up"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔧 Fragment</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Mode</label><select id="cfgFragMode"><option value="manual">manual</option><option value="auto">auto</option><option value="off">off</option></select></div>
        <div class="form-row"><label>Packets</label><input type="text" id="cfgFragPkts" value="tlshello"></div>
        <div class="form-row"><label>Manual Length</label><input type="text" id="cfgFragLen" value="10-20"></div>
        <div class="form-row"><label>Manual Interval</label><input type="text" id="cfgFragInt" value="10-20"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🚀 Xray</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Log Level</label><select id="cfgXrayLog"><option value="none">none</option><option value="error">error</option><option value="warning">warning</option><option value="info">info</option><option value="debug">debug</option></select></div>
        <div class="form-row"><label>Mux Concurrency (-1=off)</label><input type="number" id="cfgMuxConc" value="-1"></div>
      </div>
      <label class="check-row" style="margin-top:8px"><input type="checkbox" id="cfgMuxEnabled"> Mux Enabled</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔍 Shodan</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Mode</label><select id="cfgShodanMode"><option value="off">off</option><option value="harvest">harvest</option><option value="scan">scan</option><option value="both">both</option></select></div>
        <div class="form-row"><label>Pages</label><input type="number" id="cfgShodanPages" value="1" min="1"></div>
        <div class="form-row"><label>API Key</label><input type="password" id="cfgShodanKey"></div>
        <div class="form-row"><label>Save Path</label><input type="text" id="cfgShodanSave" value="results/shodan_ips.txt"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap;margin-top:8px">
        <label class="check-row"><input type="checkbox" id="cfgShodanUseDefault" checked> کوئری پیش‌فرض</label>
        <label class="check-row"><input type="checkbox" id="cfgShodanExcludeCF" checked> حذف رنج CF</label>
        <label class="check-row"><input type="checkbox" id="cfgShodanAppend"> Append</label>
      </div>
    </div>
  </div>
</div>

<!-- ══ IMPORT ══ -->
<div id="page-import" class="page">
  <div class="page-hd"><div class="page-hd-left"><h2>وارد کردن کانفیگ</h2><p>vless · vmess · trojan · JSON — همه فیلدها کشف میشه و server-side ذخیره میشه</p></div></div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔗 لینک یا JSON</div></div>
    <div class="card-bd">
      <div class="form-row">
        <label>vless:// یا vmess:// یا trojan:// یا JSON config</label>
        <textarea id="linkInput" rows="5" placeholder="vless://uuid@host:443?type=ws&security=tls&sni=example.com&path=/ws#name"></textarea>
      </div>
      <button class="btn btn-primary" onclick="parseLink()">🔄 تبدیل و ذخیره</button>
    </div>
  </div>

  <div id="parsedResult" style="display:none" class="card">
    <div class="card-hd"><div class="card-hd-left">✅ کانفیگ تشخیص داده شد — در سرور ذخیره شد</div></div>
    <div class="card-bd">
      <div class="parsed-box" id="parsedBox"></div>
      <div style="margin-top:14px;display:flex;gap:8px">
        <button class="btn btn-primary" onclick="nav('scan')">⚡ برو به اسکن</button>
        <button class="btn btn-danger btn-sm" onclick="clearSavedProxy()">🗑 حذف کانفیگ ذخیره شده</button>
      </div>
    </div>
  </div>
</div>

<!-- ══ SHODAN ══ -->
<div id="page-shodan" class="page">
  <div class="page-hd">
    <div class="page-hd-left"><h2>Shodan Harvest</h2><p>IP های non-CF با certificate کلودفلر</p></div>
    <button class="btn btn-primary" id="btnShodan" onclick="startShodan()">▶ شروع</button>
  </div>
  <div id="shodanAlert" style="display:none"></div>
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔑 تنظیمات</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>API Key</label><input type="password" id="shodanKey"></div>
        <div class="form-row"><label>صفحات (هر صفحه ۱ کردیت)</label><input type="number" id="shodanPages" value="1" min="1" max="20"></div>
      </div>
      <div class="form-row"><label>Query سفارشی (خالی = پیش‌فرض)</label><textarea id="shodanQuery" rows="2"></textarea></div>
      <div style="display:flex;gap:20px;flex-wrap:wrap">
        <label class="check-row"><input type="checkbox" id="shodanExcludeCF" checked> حذف رنج CF</label>
        <label class="check-row"><input type="checkbox" id="shodanAutoScan"> بعد از harvest اسکن کن</label>
      </div>
    </div>
  </div>
  <div class="live-ticker" id="shodanTicker" style="display:none;margin-bottom:14px">
    <div class="spinner"></div>
    <div class="live-ip-text" id="shodanTickerText">در حال جمع‌آوری...</div>
  </div>
  <div class="card" id="shodanResults" style="display:none">
    <div class="card-hd"><div class="card-hd-left">📋 IP های جمع‌آوری شده</div><span id="shodanCount" style="color:var(--g);font-family:'IBM Plex Mono',monospace">0</span></div>
    <div class="card-bd">
      <div class="ip-list" id="shodanIpChips"></div>
      <div style="margin-top:12px;display:flex;gap:8px">
        <button class="btn btn-sm" onclick="copyAllShodan()">📋 کپی همه</button>
        <button class="btn btn-primary btn-sm" onclick="scanShodanIPs()">⚡ اسکن این IP ها</button>
      </div>
    </div>
  </div>
</div>

<!-- ══ TUI TERMINAL ══ -->
<div id="page-tui" class="page">
  <div class="page-hd">
    <div class="page-hd-left"><h2>ترمینال</h2><p>لاگ زنده اسکنر — همه رویدادها</p></div>
    <div class="page-hd-actions">
      <button class="btn btn-sm" onclick="clearTUI()">🗑 پاک</button>
      <button class="btn btn-sm" onclick="toggleAutoScroll()" id="btnAutoScroll">⬇ Auto-scroll: ON</button>
    </div>
  </div>
  <div class="tui-wrap">
    <div class="tui-header">
      <div class="tui-dots">
        <div class="tui-dot" style="background:#f87171"></div>
        <div class="tui-dot" style="background:#fbbf24"></div>
        <div class="tui-dot" style="background:#3dd68c"></div>
      </div>
      <span style="margin-right:8px">piyazche — scanner log</span>
      <span id="tuiStatus" style="margin-right:auto;color:var(--dim)">idle</span>
    </div>
    <div class="tui-body" id="tuiBody">
      <div class="tui-line"><span class="tui-time">--:--:--</span><span class="tui-info">piyazche scanner آماده‌ست — کانفیگ بده و اسکن رو شروع کن</span></div>
      <div class="tui-line"><span class="tui-time">--:--:--</span><span class="tui-info">WebSocket متصل شد<span class="tui-cursor"></span></span></div>
    </div>
  </div>
</div>

</div></div>

<script>
// ══ State ══
let ws = null;
let p2Results = [];
let p2LiveResults = [];
let shodanIPs = [];
let tuiAutoScroll = true;
let savedProxyAddr = '';

// ══ Nav ══
function nav(page, btn) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(b => b.classList.remove('active'));
  document.getElementById('page-' + page).classList.add('active');
  if (btn) btn.classList.add('active');
  else { const b = document.querySelector('[data-page="'+page+'"]'); if(b) b.classList.add('active'); }
  if (page === 'results') refreshResults();
  if (page === 'history') refreshHistory();
}

// ══ WebSocket ══
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');
  ws.onmessage = e => { try { handleWS(JSON.parse(e.data)); } catch(err){} };
  ws.onclose = () => setTimeout(connectWS, 2000);
}

function handleWS(msg) {
  const { type, payload } = msg;
  switch(type) {
    case 'status': updateStatus(payload.status, payload.phase); break;
    case 'progress': updateProgress(payload); break;
    case 'live_ip': updateLiveTicker(payload.ip, 'phase1'); break;
    case 'tui': appendTUI(payload); break;
    case 'phase2_start':
      document.getElementById('phaseLabel').textContent = 'Phase 2 — ' + payload.count + ' IP';
      document.getElementById('progressBar').classList.add('ph2');
      updateStatus('scanning','phase2');
      p2LiveResults = [];
      break;
    case 'phase2_progress': {
      const r = payload;
      // update live ticker
      const dlTxt = r.dl!=='—' ? '  ↓'+r.dl : '';
      const ulTxt = r.ul!=='—' ? '  ↑'+r.ul : '';
      updateLiveTicker(r.ip+'  '+Math.round(r.latency)+'ms  loss:'+r.loss.toFixed(0)+'%'+dlTxt+ulTxt, 'phase2');
      // update progress bar
      const pct2 = r.total>0 ? Math.round(r.done/r.total*100) : 0;
      document.getElementById('progressBar').style.width = pct2+'%';
      document.getElementById('pctLabel').textContent = pct2+'%';
      document.getElementById('progDetail').textContent = 'Phase 2: '+r.done+'/'+r.total+' · score:'+r.score.toFixed(0);
      break;
    }
    case 'phase2_done':
      p2Results = payload.results || [];
      refreshResults();
      clearLiveTicker('Phase 2 تموم شد');
      break;
    case 'scan_done':
      updateStatus('done','');
      showBtns(false);
      refreshResults();
      refreshHistory();
      clearLiveTicker('✓ اسکن تموم شد — ' + payload.passed + ' IP موفق');
      break;
    case 'error': appendTUI({t: now(), l:'err', m: payload.message}); break;
    case 'shodan_status':
      document.getElementById('shodanTicker').style.display='flex';
      break;
    case 'shodan_done':
      shodanIPs = payload.ips || [];
      renderShodanResults(shodanIPs, payload.total);
      document.getElementById('shodanTicker').style.display='none';
      document.getElementById('btnShodan').disabled=false;
      appendTUI({t:now(),l:'ok',m:'Shodan: '+shodanIPs.length+' IP جمع‌آوری شد'});
      break;
    case 'shodan_error':
      showShodanAlert(payload.message,'err');
      document.getElementById('shodanTicker').style.display='none';
      document.getElementById('btnShodan').disabled=false;
      break;
  }
}

function now() { return new Date().toLocaleTimeString('fa-IR'); }

// ══ TUI ══
function appendTUI(entry) {
  const body = document.getElementById('tuiBody');
  const line = document.createElement('div');
  line.className = 'tui-line';
  const cls = {ok:'tui-ok',err:'tui-err',info:'tui-info',warn:'tui-warn',scan:'tui-scan',phase2:'tui-phase2'}[entry.l] || 'tui-info';
  line.innerHTML = '<span class="tui-time">' + (entry.t||'') + '</span><span class="' + cls + '">' + escHtml(entry.m) + '</span>';
  body.appendChild(line);
  while(body.children.length > 600) body.removeChild(body.firstChild);
  if(tuiAutoScroll) body.scrollTop = body.scrollHeight;
  document.getElementById('tuiStatus').textContent = entry.m.slice(0,40);
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function clearTUI() {
  document.getElementById('tuiBody').innerHTML =
    '<div class="tui-line"><span class="tui-time">--:--:--</span><span class="tui-info">پاک شد<span class="tui-cursor"></span></span></div>';
}

function toggleAutoScroll() {
  tuiAutoScroll = !tuiAutoScroll;
  document.getElementById('btnAutoScroll').textContent = '⬇ Auto-scroll: ' + (tuiAutoScroll ? 'ON' : 'OFF');
}

// ══ Live Ticker ══
function updateLiveTicker(ip, phase) {
  document.getElementById('liveTicker').innerHTML =
    '<div class="spinner"></div>' +
    '<div class="live-phase">' + (phase==='phase2'?'🔬 Phase 2':'⚡ Phase 1') + '</div>' +
    '<div class="live-ip-text">' + ip + '</div>';
}
function clearLiveTicker(msg) {
  document.getElementById('liveTicker').innerHTML =
    '<div style="color:var(--dim)">⊙ ' + (msg||'آماده') + '</div>';
}

// ══ IP Preview ══
async function previewIPs() {
  const input = document.getElementById('ipInput').value.trim();
  const maxIPs = parseInt(document.getElementById('maxIPs').value)||0;
  if(!input){ document.getElementById('ipPreview').style.display='none'; return; }
  const res = await fetch('/api/ips/expand', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({ipRanges: input, maxIPs})
  });
  const data = await res.json();
  const el = document.getElementById('ipPreview');
  el.style.display = 'block';
  el.textContent = 'مجموع: ' + data.count + ' IP' +
    (data.preview ? '  —  مثال: ' + data.preview.join(', ') : '');
}

// ══ Scan ══
async function startScan() {
  const ipInput = document.getElementById('ipInput').value.trim();
  const maxIPs = parseInt(document.getElementById('maxIPs').value)||0;

  // Build scan config from quick settings (proxy config is server-side)
  const scanCfg = JSON.stringify({
    scan: {
      threads: parseInt(document.getElementById('qThreads').value)||200,
      timeout: parseInt(document.getElementById('qTimeout').value)||8,
      maxLatency: parseInt(document.getElementById('qMaxLat').value)||3500,
      stabilityRounds: parseInt(document.getElementById('qRounds').value)||3,
      sampleSize: parseInt(document.getElementById('sampleSize').value)||1,
      stabilityInterval: 5,
      shuffle: true,
    }
  });

  const btn = document.getElementById('btnStart');
  btn.disabled = true;
  const res = await fetch('/api/scan/start', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({config: scanCfg, ipRanges: ipInput, maxIPs})
  });
  const data = await res.json();
  btn.disabled = false;
  if(!data.ok){ appendTUI({t:now(),l:'err',m:'خطا: '+data.error}); return; }

  p2Results = [];
  document.getElementById('progressBar').classList.remove('ph2');
  updateStatus('scanning','phase1');
  showBtns(true);
  appendTUI({t:now(),l:'ok',m:'▶ اسکن شروع شد'});
}

async function stopScan() {
  await fetch('/api/scan/stop',{method:'POST'});
  updateStatus('idle','');
  showBtns(false);
  clearLiveTicker('متوقف شد');
}

async function pauseScan() {
  const res = await fetch('/api/scan/pause',{method:'POST'});
  const data = await res.json();
  const btn = document.getElementById('btnPause');
  if(data.message==='paused'){ btn.textContent='▶ ادامه'; updateStatus('paused',''); }
  else { btn.textContent='⏸ توقف'; updateStatus('scanning',''); }
}

function showBtns(r) {
  document.getElementById('btnStart').style.display = r?'none':'inline-flex';
  document.getElementById('btnPause').style.display = r?'inline-flex':'none';
  document.getElementById('btnStop').style.display = r?'inline-flex':'none';
}

// ══ Status / Progress ══
function updateStatus(status, phase) {
  const dot = document.getElementById('statusDot');
  const txt = document.getElementById('statusText');
  const labels = {idle:'آماده',scanning:'اسکن',paused:'متوقف',done:'تموم شد'};
  txt.textContent = labels[status]||status;
  dot.className = 'dot';
  const cls = {scanning:'dot-live',paused:'dot-warn',done:'dot-done',idle:'dot-idle'}[status]||'dot-idle';
  dot.classList.add(cls);
  if(phase==='phase2') document.getElementById('phaseLabel').textContent='Phase 2';
  else if(phase==='phase1') document.getElementById('phaseLabel').textContent='Phase 1';
}

function updateProgress(p) {
  document.getElementById('stTotal').textContent = p.Total||'—';
  document.getElementById('stDone').textContent = p.Done||0;
  document.getElementById('stOk').textContent = p.Succeeded||0;
  document.getElementById('stFail').textContent = p.Failed||0;
  document.getElementById('stETA').textContent = p.ETA||'—';
  const pct = p.Total>0?Math.round(p.Done/p.Total*100):0;
  document.getElementById('progressBar').style.width = pct+'%';
  document.getElementById('pctLabel').textContent = pct+'%';
  const rate = (p.Rate||0).toFixed(1);
  document.getElementById('progDetail').textContent = p.Done+'/'+(p.Total||'?')+'  ·  '+rate+' IP/s';
  document.getElementById('progressText').textContent = p.Done+'/'+(p.Total||'?')+' ('+rate+' ip/s)';
  if(p.CurrentIP) updateLiveTicker(p.CurrentIP, 'phase1');
}

// ══ Results ══
function refreshResults() {
  fetch('/api/results').then(r=>r.json()).then(data=>{
    p2Results = data.phase2||[];
    renderResults(p2Results);
  });
}

function renderResults(results) {
  const passed = (results||[]).filter(r=>r.Passed);
  document.getElementById('resultsSummary').textContent = passed.length+' IP موفق از '+(results||[]).length+' تست شده';
  document.getElementById('passedCount').textContent = passed.length;
  const chips = document.getElementById('ipChips');
  chips.innerHTML = passed.length
    ? passed.map(r=>'<div class="ip-chip" onclick="copyIP(\''+r.IP+'\')" title="کلیک=کپی">'+r.IP+'<span style="opacity:.5;font-size:10px">'+Math.round(r.AvgLatencyMs)+'ms</span></div>').join('')
    : '<span style="color:var(--dim)">نتیجه‌ای نیست</span>';
  const tbody = document.getElementById('resultsTbody');
  if(!results||!results.length){ tbody.innerHTML='<tr><td colspan="10" style="text-align:center;color:var(--dim);padding:28px">نتیجه‌ای نیست</td></tr>'; return; }
  tbody.innerHTML = results.map((r,i)=>{
    const sc=r.StabilityScore||0;
    const scc=sc>=75?'var(--g)':sc>=50?'var(--y)':'var(--r)';
    const lc=r.AvgLatencyMs<=500?'var(--g)':r.AvgLatencyMs<=1500?'var(--y)':'var(--r)';
    const badge=r.Passed?'<span class="badge bg">PASS</span>':'<span class="badge br" title="'+(r.FailReason||'')+'">FAIL</span>';
    const dl=r.DownloadMbps||0;
    const ul=r.UploadMbps||0;
    const dlc=dl<=0?'var(--dim)':dl>=5?'var(--g)':dl>=1?'var(--y)':'var(--r)';
    const ulc=ul<=0?'var(--dim)':ul>=2?'var(--g)':ul>=0.5?'var(--y)':'var(--r)';
    const dlTxt=dl>0?dl.toFixed(1)+' M':'—';
    const ulTxt=ul>0?ul.toFixed(1)+' M':'—';
    return '<tr><td style="color:var(--dim)">'+(i+1)+'.</td>'+
      '<td style="color:var(--c)">'+r.IP+'</td>'+
      '<td style="color:'+scc+'">'+sc.toFixed(1)+'</td>'+
      '<td style="color:'+lc+'">'+Math.round(r.AvgLatencyMs)+'ms</td>'+
      '<td style="color:var(--dim)">'+(r.JitterMs>0?r.JitterMs.toFixed(0)+'ms':'—')+'</td>'+
      '<td style="color:'+(r.PacketLossPct<=5?'var(--g)':'var(--r)')+'">'+(r.PacketLossPct||0).toFixed(0)+'%</td>'+
      '<td style="color:'+dlc+'">'+dlTxt+'</td>'+
      '<td style="color:'+ulc+'">'+ulTxt+'</td>'+
      '<td>'+badge+'</td>'+
      '<td style="display:flex;gap:4px">'+
        '<button class="copy-btn" onclick="copyIP(\''+r.IP+'\')" title="کپی IP">📋</button>'+
        '<button class="copy-btn" onclick="copyWithIP(\''+r.IP+'\')" title="کپی لینک با این IP">🔗</button>'+
      '</td></tr>';
  }).join('');
}

// ══ History ══
function refreshHistory() {
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const el = document.getElementById('historyList');
    if(!sessions||!sessions.length){ el.innerHTML='<p style="color:var(--dim)">هنوز اسکنی انجام نشده</p>'; return; }
    el.innerHTML = sessions.map(s=>'<div class="hist-item" onclick="showSession(\''+s.id+'\')"><span style="font-family:monospace;color:var(--c);font-size:13px">'+new Date(s.startedAt).toLocaleString('fa-IR')+'</span><span style="color:var(--dim);font-size:12px">'+s.duration+'</span><span style="color:var(--dim);font-size:12px">'+s.totalIPs+' IP</span><span class="badge bg">'+s.passed+' passed</span></div>').join('');
  });
}
function showSession(id) {
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const s=sessions.find(x=>x.id===id); if(!s)return;
    p2Results=s.results||[]; renderResults(p2Results); nav('results');
  });
}

// ══ Config ══
function saveConfig() {
  const scanCfg = {
    scan:{
      threads:parseInt(document.getElementById('cfgThreads').value)||200,
      timeout:parseInt(document.getElementById('cfgTimeout').value)||8,
      maxLatency:parseInt(document.getElementById('cfgMaxLat').value)||3500,
      retries:parseInt(document.getElementById('cfgRetries').value)||2,
      maxIPs:parseInt(document.getElementById('cfgMaxIPs').value)||0,
      sampleSize:parseInt(document.getElementById('cfgSampleSize').value)||1,
      testUrl:document.getElementById('cfgTestURL').value,
      shuffle:document.getElementById('cfgShuffle').checked,
      stabilityRounds:parseInt(document.getElementById('cfgRounds').value)||3,
      stabilityInterval:parseInt(document.getElementById('cfgInterval').value)||5,
      packetLossCount:parseInt(document.getElementById('cfgPLCount').value)||5,
      maxPacketLossPct:parseFloat(document.getElementById('cfgMaxPL').value),
      minDownloadMbps:parseFloat(document.getElementById('cfgMinDL').value)||0,
      minUploadMbps:parseFloat(document.getElementById('cfgMinUL').value)||0,
      speedTest:document.getElementById('cfgSpeed').checked,
      jitterTest:document.getElementById('cfgJitter').checked,
      downloadUrl:document.getElementById('cfgDLURL').value,
      uploadUrl:document.getElementById('cfgULURL').value,
    },
    fragment:{
      mode:document.getElementById('cfgFragMode').value,
      packets:document.getElementById('cfgFragPkts').value,
      manual:{length:document.getElementById('cfgFragLen').value,interval:document.getElementById('cfgFragInt').value}
    },
    xray:{logLevel:document.getElementById('cfgXrayLog').value,mux:{enabled:document.getElementById('cfgMuxEnabled').checked,concurrency:parseInt(document.getElementById('cfgMuxConc').value)||-1}},
    shodan:{mode:document.getElementById('cfgShodanMode').value,apiKey:document.getElementById('cfgShodanKey').value,pages:parseInt(document.getElementById('cfgShodanPages').value)||1,useDefaultQuery:document.getElementById('cfgShodanUseDefault').checked,excludeCFRanges:document.getElementById('cfgShodanExcludeCF').checked,saveHarvestedIPs:document.getElementById('cfgShodanSave').value,appendToExisting:document.getElementById('cfgShodanAppend').checked}
  };

  // sync quick settings too
  document.getElementById('qThreads').value = scanCfg.scan.threads;
  document.getElementById('qTimeout').value = scanCfg.scan.timeout;
  document.getElementById('qMaxLat').value = scanCfg.scan.maxLatency;
  document.getElementById('qRounds').value = scanCfg.scan.stabilityRounds;
  document.getElementById('sampleSize').value = scanCfg.scan.sampleSize;

  // Save to server
  fetch('/api/config/save', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({scanConfig: JSON.stringify(scanCfg)})
  }).then(()=>{
    appendTUI({t:now(),l:'ok',m:'✓ تنظیمات ذخیره شد'});
    document.getElementById('configSummary').textContent =
      'threads:'+scanCfg.scan.threads+' · timeout:'+scanCfg.scan.timeout+'s · rounds:'+scanCfg.scan.stabilityRounds+' · frag:'+scanCfg.fragment.mode;
    nav('scan');
  });
}

// ══ Import Link ══
async function parseLink() {
  const input = document.getElementById('linkInput').value.trim();
  if(!input) return;
  const res = await fetch('/api/config/parse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})});
  const data = await res.json();
  if(!data.ok){ appendTUI({t:now(),l:'err',m:'خطا: '+data.error}); return; }

  // Save proxy config to server
  await fetch('/api/config/save',{
    method:'POST',headers:{'Content-Type':'application/json'},
    body: JSON.stringify({proxyConfig: data.config})
  });

  const p = data.parsed;
  savedProxyAddr = p.address+':'+p.port;
  updateProxyBadge(p.address, p.method, p.type);

  document.getElementById('parsedBox').innerHTML =
    '<span class="k">uuid: </span><span class="v">'+maskUUID(p.uuid)+'</span><br>'+
    '<span class="k">address: </span><span class="v">'+p.address+'</span><br>'+
    '<span class="k">port: </span><span class="v">'+p.port+'</span><br>'+
    '<span class="k">type: </span><span class="v">'+p.type+'</span><br>'+
    '<span class="k">method: </span><span class="v">'+p.method+'</span>'+
    (p.sni?'<br><span class="k">sni: </span><span class="v">'+p.sni+'</span>':'')+
    (p.path?'<br><span class="k">path: </span><span class="v">'+p.path+'</span>':'')+
    (p.fp?'<br><span class="k">fp: </span><span class="v">'+p.fp+'</span>':'');

  document.getElementById('parsedResult').style.display='block';
  appendTUI({t:now(),l:'ok',m:'✓ کانفیگ ذخیره شد: '+p.address+':'+p.port+' ('+p.method+'/'+p.type+')'});
}

function updateProxyBadge(addr, method, type) {
  const badge = document.getElementById('proxyBadge');
  document.getElementById('proxyBadgeText').textContent = addr+' · '+method+'/'+type;
  badge.style.display = 'inline-flex';
  document.getElementById('configSummary').textContent = '✓ '+addr+' · '+method+'/'+type;
}

async function clearSavedProxy() {
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:''})});
  document.getElementById('proxyBadge').style.display='none';
  document.getElementById('configSummary').textContent='پیش‌فرض — لینک وارد نشده';
  document.getElementById('parsedResult').style.display='none';
  appendTUI({t:now(),l:'warn',m:'کانفیگ proxy حذف شد'});
}

function maskUUID(uuid){ return !uuid||uuid.length<8?uuid:uuid.slice(0,8)+'••••••••'; }

// ══ Shodan ══
async function startShodan() {
  const key = document.getElementById('shodanKey').value.trim();
  if(!key){ showShodanAlert('API Key الزامی است','err'); return; }
  document.getElementById('btnShodan').disabled=true;
  document.getElementById('shodanTicker').style.display='flex';
  document.getElementById('shodanResults').style.display='none';
  document.getElementById('shodanAlert').style.display='none';
  const res = await fetch('/api/shodan/harvest',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({apiKey:key,query:document.getElementById('shodanQuery').value.trim(),pages:parseInt(document.getElementById('shodanPages').value)||1,excludeCF:document.getElementById('shodanExcludeCF').checked,autoScan:document.getElementById('shodanAutoScan').checked})});
  const data = await res.json();
  if(!data.ok){ document.getElementById('shodanTicker').style.display='none'; document.getElementById('btnShodan').disabled=false; showShodanAlert(data.error,'err'); }
}
function renderShodanResults(ips,total){ shodanIPs=ips||[]; document.getElementById('shodanCount').textContent=shodanIPs.length; document.getElementById('shodanResults').style.display='block'; const chips=document.getElementById('shodanIpChips'); chips.innerHTML=shodanIPs.slice(0,200).map(ip=>'<div class="ip-chip" onclick="copyIP(\''+ip+'\')">'+ip+'</div>').join(''); if(shodanIPs.length>200) chips.innerHTML+='<span style="color:var(--dim);font-size:12px"> +' + (shodanIPs.length-200)+' IP دیگر</span>'; }
function copyAllShodan(){ navigator.clipboard.writeText(shodanIPs.join('\n')); appendTUI({t:now(),l:'ok',m:'کپی شد: '+shodanIPs.length+' IP'}); }
function scanShodanIPs(){ if(!shodanIPs.length)return; document.getElementById('ipInput').value=shodanIPs.join('\n'); nav('scan'); }
function showShodanAlert(msg,type){ const el=document.getElementById('shodanAlert'); el.className='alert alert-'+(type==='err'?'err':type==='warn'?'warn':'info'); el.textContent=msg; el.style.display='block'; }

// ══ Export / Copy ══
function exportResults(f){ window.location.href='/api/results/export?format='+f; }
function copyIP(ip){ navigator.clipboard.writeText(ip).then(()=>appendTUI({t:now(),l:'ok',m:'📋 '+ip})).catch(()=>{ const el=document.createElement('textarea');el.value=ip;document.body.appendChild(el);el.select();document.execCommand('copy');document.body.removeChild(el); appendTUI({t:now(),l:'ok',m:'📋 '+ip}); }); }

// کپی لینک vless/vmess با IP جدید
function copyWithIP(newIP) {
  const linkEl = document.getElementById('linkInput');
  const rawLink = linkEl ? linkEl.value.trim() : '';
  if(!rawLink){ copyIP(newIP); return; }
  try {
    let updated = rawLink;
    // برای vless://uuid@IP:port و vmess و trojan
    updated = rawLink.replace(/(@)([^:@\/?#\[\]]+)(:\d+)/, '$1'+newIP+'$3');
    navigator.clipboard.writeText(updated).then(()=>{
      appendTUI({t:now(),l:'ok',m:'🔗 لینک با '+newIP+' کپی شد'});
    });
  } catch(e){ copyIP(newIP); }
}

// ══ Settings Persist ══
function loadSavedSettings() {
  fetch('/api/config/load').then(r=>r.json()).then(d=>{
    // Load proxy badge
    if(d.hasProxy){
      try {
        const pc = JSON.parse(d.proxyConfig);
        const addr = pc.proxy?.address||'';
        const method = pc.proxy?.method||'tls';
        const type = pc.proxy?.type||'ws';
        if(addr) updateProxyBadge(addr, method, type);
      } catch(e){}
    }
    // Load scan config settings into form fields
    if(d.scanConfig){
      try {
        const sc = JSON.parse(d.scanConfig);
        const s = sc.scan||{};
        const f = sc.fragment||{};
        const x = sc.xray||{};
        const sh = sc.shodan||{};
        // Phase 1
        if(s.threads) setVal('cfgThreads', s.threads);
        if(s.timeout) setVal('cfgTimeout', s.timeout);
        if(s.maxLatency) setVal('cfgMaxLat', s.maxLatency);
        if(s.retries!=null) setVal('cfgRetries', s.retries);
        if(s.maxIPs!=null) setVal('cfgMaxIPs', s.maxIPs);
        if(s.sampleSize) setVal('cfgSampleSize', s.sampleSize);
        if(s.testUrl) setVal('cfgTestURL', s.testUrl);
        if(s.shuffle!=null) setChk('cfgShuffle', s.shuffle);
        // Phase 2
        if(s.stabilityRounds!=null) setVal('cfgRounds', s.stabilityRounds);
        if(s.stabilityInterval) setVal('cfgInterval', s.stabilityInterval);
        if(s.packetLossCount) setVal('cfgPLCount', s.packetLossCount);
        if(s.maxPacketLossPct!=null) setVal('cfgMaxPL', s.maxPacketLossPct);
        if(s.minDownloadMbps!=null) setVal('cfgMinDL', s.minDownloadMbps);
        if(s.minUploadMbps!=null) setVal('cfgMinUL', s.minUploadMbps);
        if(s.speedTest!=null) setChk('cfgSpeed', s.speedTest);
        if(s.jitterTest!=null) setChk('cfgJitter', s.jitterTest);
        if(s.downloadUrl) setVal('cfgDLURL', s.downloadUrl);
        if(s.uploadUrl) setVal('cfgULURL', s.uploadUrl);
        // Fragment
        if(f.mode) setSelVal('cfgFragMode', f.mode);
        if(f.packets) setVal('cfgFragPkts', f.packets);
        if(f.manual?.length) setVal('cfgFragLen', f.manual.length);
        if(f.manual?.interval) setVal('cfgFragInt', f.manual.interval);
        // Xray
        if(x.logLevel) setSelVal('cfgXrayLog', x.logLevel);
        if(x.mux?.concurrency!=null) setVal('cfgMuxConc', x.mux.concurrency);
        if(x.mux?.enabled!=null) setChk('cfgMuxEnabled', x.mux.enabled);
        // Shodan
        if(sh.mode) setSelVal('cfgShodanMode', sh.mode);
        if(sh.apiKey) setVal('cfgShodanKey', sh.apiKey);
        if(sh.pages) setVal('cfgShodanPages', sh.pages);
        if(sh.saveHarvestedIPs) setVal('cfgShodanSave', sh.saveHarvestedIPs);
        if(sh.useDefaultQuery!=null) setChk('cfgShodanUseDefault', sh.useDefaultQuery);
        if(sh.excludeCFRanges!=null) setChk('cfgShodanExcludeCF', sh.excludeCFRanges);
        if(sh.appendToExisting!=null) setChk('cfgShodanAppend', sh.appendToExisting);
        // sync quick settings
        if(s.threads) setVal('qThreads', s.threads);
        if(s.timeout) setVal('qTimeout', s.timeout);
        if(s.maxLatency) setVal('qMaxLat', s.maxLatency);
        if(s.stabilityRounds!=null) setVal('qRounds', s.stabilityRounds);
        if(s.sampleSize) setVal('sampleSize', s.sampleSize);
      } catch(e){ console.warn('load settings err', e); }
    }
    // Load saved TUI history
    fetch('/api/tui/stream').then(r=>r.json()).then(data=>{
      if(data.lines&&data.lines.length){
        data.lines.forEach(line=>{ try{ appendTUI(JSON.parse(line)); }catch(e){} });
      }
    });
  });
}
function setVal(id, v){ const el=document.getElementById(id); if(el) el.value=v; }
function setChk(id, v){ const el=document.getElementById(id); if(el) el.checked=!!v; }
function setSelVal(id, v){ const el=document.getElementById(id); if(el) el.value=v; }

// ══ Init ══
connectWS();
fetch('/api/status').then(r=>r.json()).then(d=>{ updateStatus(d.status||'idle',d.phase||''); });
loadSavedSettings();
</script>
</body>
</html>
`
