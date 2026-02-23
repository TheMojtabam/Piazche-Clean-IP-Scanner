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
  --g:#3dd68c;--gd:#0d2b1c;--g2:#2ab574;
  --c:#38bdf8;--cd:#0a2030;--c2:#0ea5e9;
  --y:#fbbf24;--yd:#271d06;
  --r:#f87171;--rd:#2a0f0f;
  --p:#a78bfa;--pd:#1a1040;
  --o:#fb923c;--od:#271508;
  --radius:10px;--radius-sm:6px;
  --shadow:0 4px 24px rgba(0,0,0,.4);
}
*{margin:0;padding:0;box-sizing:border-box}
html{scroll-behavior:smooth}
body{
  font-family:'Vazirmatn',Tahoma,sans-serif;
  background:var(--bg);
  color:var(--tx);
  min-height:100vh;
  font-size:14px;
  line-height:1.6;
}
/* ─── Layout ─── */
.layout{display:grid;grid-template-columns:240px 1fr;grid-template-rows:56px 1fr;min-height:100vh}
.topbar{
  grid-column:1/-1;
  background:rgba(13,15,18,.92);
  border-bottom:1px solid var(--bd);
  backdrop-filter:blur(12px);
  display:flex;align-items:center;padding:0 20px;gap:14px;
  position:sticky;top:0;z-index:100;
}
.sidebar{
  background:var(--bg2);
  border-left:1px solid var(--bd);
  padding:12px 0 20px;
  display:flex;flex-direction:column;
  gap:2px;
  overflow-y:auto;
}
.main{overflow-y:auto;padding:20px 24px}
/* ─── Logo ─── */
.logo{
  font-family:'IBM Plex Mono',monospace;
  font-size:18px;font-weight:600;
  color:var(--c);
  letter-spacing:-1px;
  user-select:none;
}
.logo b{color:var(--g)}
.logo span{color:var(--dim);font-size:11px;margin-right:6px;font-weight:400}
/* ─── Status badge ─── */
.status-badge{
  display:flex;align-items:center;gap:7px;
  padding:5px 14px;border-radius:20px;font-size:12px;
  background:var(--bg3);border:1px solid var(--bd2);
  font-family:'IBM Plex Mono',monospace;
}
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0}
.dot-idle{background:var(--dim)}
.dot-live{background:var(--g);box-shadow:0 0 8px var(--g);animation:pulse 1.4s infinite}
.dot-warn{background:var(--y);box-shadow:0 0 8px var(--y)}
.dot-done{background:var(--c);box-shadow:0 0 8px var(--c)}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:.5;transform:scale(.85)}}
.topbar-right{margin-right:auto;display:flex;align-items:center;gap:10px}
.topbar-speed{font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--dim)}
/* ─── Nav ─── */
.nav-group{padding:6px 12px 3px;font-size:10px;letter-spacing:2px;text-transform:uppercase;color:var(--dim);margin-top:8px}
.nav-item{
  display:flex;align-items:center;gap:10px;
  padding:9px 14px;cursor:pointer;
  transition:all .15s;color:var(--tx2);
  font-size:13px;border:none;background:none;
  width:100%;text-align:right;
  border-right:2px solid transparent;
}
.nav-item:hover{background:var(--bg3);color:var(--tx)}
.nav-item.active{background:var(--cd);color:var(--c);border-right-color:var(--c)}
.nav-icon{font-size:14px;min-width:18px;text-align:center;flex-shrink:0}
/* ─── Pages ─── */
.page{display:none}
.page.active{display:block}
/* ─── Page header ─── */
.page-hd{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:18px;gap:12px}
.page-hd-left h2{font-size:19px;font-weight:600;color:var(--tx)}
.page-hd-left p{font-size:12px;color:var(--dim);margin-top:2px}
.page-hd-actions{display:flex;gap:8px;align-items:center;flex-shrink:0}
/* ─── Cards ─── */
.card{
  background:var(--bg2);
  border:1px solid var(--bd);
  border-radius:var(--radius);
  overflow:hidden;
  margin-bottom:14px;
}
.card-hd{
  background:var(--bg3);
  border-bottom:1px solid var(--bd);
  padding:9px 16px;
  font-size:11px;color:var(--dim);
  display:flex;align-items:center;justify-content:space-between;
  font-family:'IBM Plex Mono',monospace;
  letter-spacing:.5px;
}
.card-hd-left{display:flex;align-items:center;gap:7px}
.card-bd{padding:16px}
/* ─── Stats ─── */
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:10px;margin-bottom:14px}
.stat{
  background:var(--bg2);border:1px solid var(--bd);
  border-radius:var(--radius-sm);padding:12px 14px;
  text-align:center;
  transition:border-color .2s;
}
.stat:hover{border-color:var(--bd3)}
.stat-v{font-size:24px;font-weight:700;font-family:'IBM Plex Mono',monospace;line-height:1.2}
.stat-l{font-size:10px;color:var(--dim);margin-top:4px;letter-spacing:.5px;text-transform:uppercase}
/* ─── Progress ─── */
.prog-wrap{background:var(--bg3);border-radius:4px;height:6px;overflow:hidden;margin:8px 0}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:4px;transition:width .4s ease}
.prog-bar.phase2{background:linear-gradient(90deg,var(--p),var(--c))}
/* ─── Live IP ticker ─── */
.live-ticker{
  background:var(--bg3);
  border:1px solid var(--bd2);
  border-radius:var(--radius-sm);
  padding:10px 14px;
  font-family:'IBM Plex Mono',monospace;
  font-size:12px;
  min-height:42px;
  display:flex;align-items:center;gap:10px;
  margin-top:10px;
  overflow:hidden;
}
.spinner{
  width:16px;height:16px;
  border:2px solid var(--bd3);
  border-top-color:var(--c);
  border-radius:50%;
  animation:spin .7s linear infinite;
  flex-shrink:0;
}
@keyframes spin{to{transform:rotate(360deg)}}
.live-ip-text{color:var(--c);font-weight:500;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.live-ip-phase{color:var(--dim);font-size:10px;margin-right:auto}
.ticker-idle{color:var(--dim)}
/* ─── Table ─── */
.tbl{width:100%;border-collapse:collapse;font-size:12px;font-family:'IBM Plex Mono',monospace}
.tbl th{
  padding:8px 10px;text-align:right;
  color:var(--dim);font-weight:500;
  border-bottom:1px solid var(--bd);
  background:var(--bg3);
  white-space:nowrap;
  font-size:11px;letter-spacing:.5px;
}
.tbl td{padding:7px 10px;border-bottom:1px solid var(--bd);vertical-align:middle}
.tbl tr:last-child td{border-bottom:none}
.tbl tr:hover td{background:rgba(255,255,255,.02)}
/* ─── Badges ─── */
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:10px;font-size:10px;font-family:'IBM Plex Mono',monospace;font-weight:500;letter-spacing:.3px}
.bg{background:var(--gd);color:var(--g)}
.by{background:var(--yd);color:var(--y)}
.br{background:var(--rd);color:var(--r)}
.bc{background:var(--cd);color:var(--c)}
.bp{background:var(--pd);color:var(--p)}
/* ─── Buttons ─── */
.btn{
  display:inline-flex;align-items:center;gap:6px;
  padding:8px 16px;border-radius:var(--radius-sm);
  border:1px solid var(--bd2);background:var(--bg3);
  color:var(--tx);cursor:pointer;font-size:13px;
  font-family:inherit;transition:all .15s;
  white-space:nowrap;
}
.btn:hover{background:var(--bd2);border-color:var(--bd3)}
.btn:active{transform:scale(.97)}
.btn-primary{background:var(--cd);border-color:var(--c2);color:var(--c)}
.btn-primary:hover{background:var(--c2);color:#000}
.btn-danger{background:var(--rd);border-color:var(--r);color:var(--r)}
.btn-danger:hover{background:var(--r);color:#fff}
.btn-warn{background:var(--yd);border-color:var(--y);color:var(--y)}
.btn-warn:hover{background:var(--y);color:#000}
.btn-sm{padding:5px 10px;font-size:11px}
.btn-xs{padding:3px 8px;font-size:10px}
.btn:disabled{opacity:.4;cursor:not-allowed;pointer-events:none}
/* ─── Forms ─── */
textarea,input[type=text],input[type=number],input[type=password],input[type=email],select{
  background:var(--bg3);
  border:1px solid var(--bd2);
  color:var(--tx);
  border-radius:var(--radius-sm);
  padding:8px 12px;
  font-size:13px;
  font-family:'IBM Plex Mono',monospace;
  width:100%;outline:none;
  direction:ltr;
  transition:border-color .15s;
}
textarea:focus,input:focus,select:focus{border-color:var(--c)}
select option{background:var(--bg3)}
label{display:block;font-size:12px;color:var(--dim);margin-bottom:5px;text-align:right;font-family:'Vazirmatn',sans-serif}
.form-row{margin-bottom:12px}
.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.form-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.form-sep{height:1px;background:var(--bd);margin:16px 0}
.check-row{display:flex;align-items:center;gap:8px;cursor:pointer;font-size:13px;color:var(--tx2)}
.check-row input[type=checkbox]{width:auto;cursor:pointer;accent-color:var(--c)}
/* ─── Parsed box ─── */
.parsed-box{
  background:var(--bg3);border:1px solid var(--gd);
  border-radius:var(--radius-sm);
  padding:14px;font-family:'IBM Plex Mono',monospace;
  font-size:12px;color:var(--g);direction:ltr;line-height:1.8;
}
.parsed-box .k{color:var(--dim)}
.parsed-box .v{color:var(--c)}
/* ─── Log ─── */
.log-box{
  background:#050607;
  border-radius:var(--radius-sm);
  padding:12px;font-family:'IBM Plex Mono',monospace;
  font-size:12px;line-height:1.75;
  max-height:340px;overflow-y:auto;
  direction:ltr;
}
.log-ok{color:var(--g)}
.log-err{color:var(--r)}
.log-info{color:var(--c)}
.log-warn{color:var(--y)}
.log-dim{color:var(--dim)}
/* ─── Copy btn ─── */
.copy-btn{background:none;border:none;cursor:pointer;color:var(--dim);padding:2px 6px;border-radius:4px;font-size:11px;transition:all .15s}
.copy-btn:hover{color:var(--c);background:var(--cd)}
/* ─── IP chips ─── */
.ip-list{display:flex;flex-wrap:wrap;gap:6px;padding:4px 0}
.ip-chip{
  background:var(--cd);border:1px solid var(--c);
  border-radius:5px;padding:4px 10px;
  font-family:'IBM Plex Mono',monospace;font-size:12px;
  color:var(--c);cursor:pointer;
  display:flex;align-items:center;gap:6px;
  transition:all .15s;
}
.ip-chip:hover{background:var(--c);color:#000}
/* ─── Config summary ─── */
.cfg-summary{
  background:var(--bg3);border:1px solid var(--bd2);
  border-radius:var(--radius-sm);
  padding:8px 12px;font-size:12px;color:var(--dim);
  font-family:'IBM Plex Mono',monospace;
}
/* ─── Alert box ─── */
.alert{border-radius:var(--radius-sm);padding:10px 14px;font-size:12px;margin-bottom:12px;border-left:3px solid}
.alert-info{background:var(--cd);border-color:var(--c);color:var(--c)}
.alert-warn{background:var(--yd);border-color:var(--y);color:var(--y)}
.alert-err{background:var(--rd);border-color:var(--r);color:var(--r)}
/* ─── History ─── */
.hist-item{
  background:var(--bg3);border:1px solid var(--bd);
  border-radius:var(--radius-sm);
  padding:12px 16px;display:flex;align-items:center;gap:14px;
  margin-bottom:8px;cursor:pointer;
  transition:all .15s;
}
.hist-item:hover{border-color:var(--bd3);background:var(--bg4)}
/* ─── Section title ─── */
.sec-title{
  font-size:10px;letter-spacing:3px;text-transform:uppercase;
  color:var(--dim);margin:18px 0 10px;
  display:flex;align-items:center;gap:10px;
}
.sec-title::after{content:'';flex:1;height:1px;background:var(--bd)}
/* ─── Scrollbar ─── */
::-webkit-scrollbar{width:5px;height:5px}
::-webkit-scrollbar-track{background:transparent}
::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:3px}
/* ─── Responsive ─── */
@media(max-width:768px){
  .layout{grid-template-columns:1fr}
  .sidebar{display:none}
  .stats-row{grid-template-columns:repeat(2,1fr)}
  .form-grid,.form-grid-3{grid-template-columns:1fr}
}
/* ─── Transitions ─── */
.fade-in{animation:fadeIn .25s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
</style>
</head>
<body>
<div class="layout">

<!-- TOPBAR -->
<div class="topbar">
  <div class="logo">piy<b>az</b>che <span>scanner</span></div>
  <div class="status-badge" id="statusBadge">
    <div class="dot dot-idle" id="statusDot"></div>
    <span id="statusText" style="font-family:'IBM Plex Mono',monospace">آماده</span>
  </div>
  <div class="topbar-right">
    <span class="topbar-speed" id="progressText"></span>
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
  <button class="nav-item" data-page="shodan" onclick="nav('shodan',this)"><span class="nav-icon">🔍</span>Shodan Harvest</button>
  <button class="nav-item" data-page="log" onclick="nav('log',this)"><span class="nav-icon">📝</span>لاگ <span id="logBadge" style="display:none;background:var(--r);color:#fff;border-radius:10px;padding:0 5px;font-size:10px;margin-right:auto">!</span></button>
</div>

<!-- MAIN -->
<div class="main">

<!-- ───── SCAN PAGE ───── -->
<div id="page-scan" class="page active fade-in">
  <div class="page-hd">
    <div class="page-hd-left"><h2>اسکن جدید</h2><p>رنج IP بده، تنظیمات رو بزن، شروع کن</p></div>
    <div class="page-hd-actions">
      <button class="btn btn-primary" id="btnStart" onclick="startScan()">▶ شروع اسکن</button>
      <button class="btn btn-warn" id="btnPause" onclick="pauseScan()" style="display:none">⏸ توقف</button>
      <button class="btn btn-danger" id="btnStop" onclick="stopScan()" style="display:none">■ بستن</button>
    </div>
  </div>

  <!-- Stats row -->
  <div class="stats-row">
    <div class="stat"><div class="stat-v" id="stTotal" style="color:var(--tx)">—</div><div class="stat-l">کل IP</div></div>
    <div class="stat"><div class="stat-v" id="stDone" style="color:var(--c)">0</div><div class="stat-l">بررسی شده</div></div>
    <div class="stat"><div class="stat-v" id="stOk" style="color:var(--g)">0</div><div class="stat-l">موفق</div></div>
    <div class="stat"><div class="stat-v" id="stFail" style="color:var(--r)">0</div><div class="stat-l">ناموفق</div></div>
    <div class="stat"><div class="stat-v" id="stETA" style="color:var(--y)">—</div><div class="stat-l">زمان باقی</div></div>
  </div>

  <!-- Progress + live IP -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">📶 پیشرفت اسکن</div><span id="phaseLabel" style="color:var(--dim)">Phase 1</span></div>
    <div class="card-bd">
      <div style="display:flex;justify-content:space-between;font-size:12px;color:var(--dim);margin-bottom:4px">
        <span id="progDetail">در انتظار شروع...</span>
        <span id="pctLabel" style="font-family:'IBM Plex Mono',monospace;color:var(--c)">0%</span>
      </div>
      <div class="prog-wrap"><div class="prog-bar" id="progressBar" style="width:0%"></div></div>
      <!-- Live IP ticker -->
      <div class="live-ticker" id="liveTicker">
        <div class="ticker-idle">⊙ آماده — اسکن را شروع کن</div>
      </div>
    </div>
  </div>

  <!-- IP input + quick settings -->
  <div class="form-grid">
    <div>
      <div class="card">
        <div class="card-hd"><div class="card-hd-left">🌐 رنج IP</div></div>
        <div class="card-bd">
          <div class="form-row">
            <label>هر خط: IP یا CIDR — خالی = از ipv4.txt</label>
            <textarea id="ipInput" rows="7" placeholder="104.16.0.0/12&#10;185.42.0.0/16&#10;45.12.33.91&#10;..."></textarea>
          </div>
          <div class="form-grid">
            <div class="form-row">
              <label>حداکثر IP (0 = همه)</label>
              <input type="number" id="maxIPs" value="0" min="0">
            </div>
            <div class="form-row">
              <label>IPs در هر سابنت</label>
              <input type="number" id="sampleSize" value="1" min="1" max="255">
            </div>
          </div>
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
            <div class="form-row"><label>Max Latency (ms)</label><input type="number" id="qMaxLat" value="3500" min="100"></div>
            <div class="form-row"><label>Stability Rounds</label><input type="number" id="qRounds" value="3" min="0"></div>
          </div>
          <div style="margin-top:4px">
            <div class="cfg-summary" id="configSummary">پیش‌فرض — لینک وارد نشده</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- ───── RESULTS PAGE ───── -->
<div id="page-results" class="page fade-in">
  <div class="page-hd">
    <div class="page-hd-left"><h2>نتایج</h2><p id="resultsSummary">هنوز اسکنی نشده</p></div>
    <div class="page-hd-actions">
      <button class="btn btn-sm" onclick="exportResults('txt')">📥 IP‌ها (.txt)</button>
      <button class="btn btn-sm" onclick="exportResults('json')">📥 JSON</button>
    </div>
  </div>

  <!-- IP chips -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">✅ IP های موفق — کلیک برای کپی</div><span id="passedCount" style="color:var(--g);font-family:'IBM Plex Mono',monospace">0</span></div>
    <div class="card-bd">
      <div class="ip-list" id="ipChips"><span style="color:var(--dim);font-size:13px">نتیجه‌ای نیست</span></div>
    </div>
  </div>

  <!-- Phase2 table -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">📊 Phase 2 — جزئیات تست عمقی</div></div>
    <div class="card-bd" style="padding:0;overflow-x:auto">
      <table class="tbl" id="resultsTable">
        <thead>
          <tr>
            <th>#</th><th>IP</th><th>Score</th><th>Avg Lat</th>
            <th>Jitter</th><th>PktLoss</th><th>Download</th><th>Upload</th><th>وضعیت</th><th></th>
          </tr>
        </thead>
        <tbody id="resultsTbody">
          <tr><td colspan="10" style="text-align:center;color:var(--dim);padding:28px">هنوز نتیجه‌ای نیست</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</div>

<!-- ───── HISTORY PAGE ───── -->
<div id="page-history" class="page fade-in">
  <div class="page-hd"><div class="page-hd-left"><h2>تاریخچه</h2><p>اسکن‌های قبلی این session</p></div></div>
  <div id="historyList"><p style="color:var(--dim)">هنوز اسکنی انجام نشده</p></div>
</div>

<!-- ───── CONFIG PAGE ───── -->
<div id="page-config" class="page fade-in">
  <div class="page-hd">
    <div class="page-hd-left"><h2>تنظیمات اسکنر</h2><p>همه گزینه‌های scan, phase2, fragment, xray, shodan</p></div>
    <button class="btn btn-primary" onclick="saveConfig()">💾 ذخیره</button>
  </div>

  <!-- Scan settings -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">⚡ تنظیمات اسکن (Phase 1)</div></div>
    <div class="card-bd">
      <div class="form-grid-3">
        <div class="form-row"><label>Threads (worker ها)</label><input type="number" id="cfgThreads" value="200" min="1"></div>
        <div class="form-row"><label>Timeout (ثانیه)</label><input type="number" id="cfgTimeout" value="8" min="1"></div>
        <div class="form-row"><label>Max Latency (ms)</label><input type="number" id="cfgMaxLat" value="3500" min="100"></div>
        <div class="form-row"><label>Retries</label><input type="number" id="cfgRetries" value="2" min="0"></div>
        <div class="form-row"><label>Max IPs (0 = همه)</label><input type="number" id="cfgMaxIPs" value="0" min="0"></div>
        <div class="form-row"><label>Sample Size (IP از هر subnet)</label><input type="number" id="cfgSampleSize" value="1" min="1"></div>
      </div>
      <div class="form-row">
        <label>Test URL</label>
        <input type="text" id="cfgTestURL" value="https://www.gstatic.com/generate_204">
      </div>
      <div style="display:flex;gap:16px;flex-wrap:wrap;margin-top:6px">
        <label class="check-row"><input type="checkbox" id="cfgShuffle" checked> Shuffle IPs</label>
      </div>
    </div>
  </div>

  <!-- Phase 2 -->
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
        <div class="form-row">
          <label>Download URL</label>
          <input type="text" id="cfgDLURL" value="https://speed.cloudflare.com/__down?bytes=1000000">
        </div>
        <div class="form-row">
          <label>Upload URL</label>
          <input type="text" id="cfgULURL" value="https://speed.cloudflare.com/__up">
        </div>
      </div>
    </div>
  </div>

  <!-- Fragment -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔧 Fragment</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Mode</label>
          <select id="cfgFragMode">
            <option value="manual">manual</option>
            <option value="auto">auto</option>
            <option value="off">off</option>
          </select>
        </div>
        <div class="form-row"><label>Packets</label><input type="text" id="cfgFragPkts" value="tlshello"></div>
        <div class="form-row"><label>Manual Length</label><input type="text" id="cfgFragLen" value="10-20"></div>
        <div class="form-row"><label>Manual Interval</label><input type="text" id="cfgFragInt" value="10-20"></div>
      </div>
    </div>
  </div>

  <!-- Xray -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🚀 Xray</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Log Level</label>
          <select id="cfgXrayLog">
            <option value="none">none</option>
            <option value="error">error</option>
            <option value="warning">warning</option>
            <option value="info">info</option>
            <option value="debug">debug</option>
          </select>
        </div>
        <div class="form-row"><label>Mux Concurrency (-1=غیرفعال)</label><input type="number" id="cfgMuxConc" value="-1"></div>
      </div>
      <label class="check-row" style="margin-top:8px"><input type="checkbox" id="cfgMuxEnabled"> Mux Enabled</label>
    </div>
  </div>

  <!-- Shodan config -->
  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔍 Shodan (در config)</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>Mode</label>
          <select id="cfgShodanMode">
            <option value="off">off</option>
            <option value="harvest">harvest</option>
            <option value="scan">scan</option>
            <option value="both">both</option>
          </select>
        </div>
        <div class="form-row"><label>Pages</label><input type="number" id="cfgShodanPages" value="1" min="1"></div>
        <div class="form-row"><label>API Key</label><input type="password" id="cfgShodanKey" placeholder="shodan api key"></div>
        <div class="form-row"><label>Save Harvested IPs (مسیر فایل)</label><input type="text" id="cfgShodanSave" value="results/shodan_ips.txt"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap;margin-top:8px">
        <label class="check-row"><input type="checkbox" id="cfgShodanUseDefault" checked> کوئری پیش‌فرض</label>
        <label class="check-row"><input type="checkbox" id="cfgShodanExcludeCF" checked> حذف رنج اصلی CF</label>
        <label class="check-row"><input type="checkbox" id="cfgShodanAppend"> Append به فایل موجود</label>
      </div>
    </div>
  </div>
</div>

<!-- ───── IMPORT PAGE ───── -->
<div id="page-import" class="page fade-in">
  <div class="page-hd"><div class="page-hd-left"><h2>وارد کردن کانفیگ</h2><p>vless:// vmess:// یا JSON بده — همه فیلدها کشف میشه</p></div></div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔗 لینک یا JSON</div></div>
    <div class="card-bd">
      <div class="form-row">
        <label>vless:// یا vmess:// یا JSON config کامل</label>
        <textarea id="linkInput" rows="5" placeholder="vless://uuid@host:443?type=ws&security=tls&sni=example.com&path=/ws&host=example.com#remark&#10;&#10;یا&#10;&#10;vmess://base64...&#10;&#10;یا&#10;&#10;{ &quot;proxy&quot;: { &quot;uuid&quot;: &quot;...&quot; } }"></textarea>
      </div>
      <button class="btn btn-primary" onclick="parseLink()">🔄 تبدیل و نمایش</button>
    </div>
  </div>

  <div id="parsedResult" style="display:none" class="card">
    <div class="card-hd"><div class="card-hd-left">✅ کانفیگ تشخیص داده شد</div></div>
    <div class="card-bd">
      <div class="parsed-box" id="parsedBox"></div>
      <div style="margin-top:14px;display:flex;gap:8px">
        <button class="btn btn-primary" onclick="applyParsed()">✓ استفاده از این کانفیگ</button>
        <button class="btn" onclick="document.getElementById('parsedResult').style.display='none'">رد کردن</button>
      </div>
    </div>
  </div>
</div>

<!-- ───── SHODAN PAGE ───── -->
<div id="page-shodan" class="page fade-in">
  <div class="page-hd">
    <div class="page-hd-left"><h2>Shodan Harvest</h2><p>IP های non-CF با certificate کلودفلر — هر صفحه ۱۰۰ IP</p></div>
    <button class="btn btn-primary" id="btnShodan" onclick="startShodan()">▶ شروع Harvest</button>
  </div>

  <div id="shodanAlert" style="display:none"></div>

  <div class="card">
    <div class="card-hd"><div class="card-hd-left">🔑 تنظیمات</div></div>
    <div class="card-bd">
      <div class="form-grid">
        <div class="form-row"><label>API Key (اجباری)</label><input type="password" id="shodanKey" placeholder="your-shodan-api-key"></div>
        <div class="form-row"><label>تعداد صفحات (هر صفحه ۱ کردیت)</label><input type="number" id="shodanPages" value="1" min="1" max="20"></div>
      </div>
      <div class="form-row">
        <label>Query سفارشی — خالی = کوئری پیش‌فرض Cloudflare</label>
        <textarea id="shodanQuery" rows="2" placeholder='خالی = ssl:"Cloudflare Inc ECC CA" port:443 -net:173.245.48.0/20 ...'></textarea>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap">
        <label class="check-row"><input type="checkbox" id="shodanExcludeCF" checked> حذف رنج‌های اصلی CF</label>
        <label class="check-row"><input type="checkbox" id="shodanAutoScan"> پس از harvest اسکن کن</label>
      </div>
    </div>
  </div>

  <!-- Shodan status ticker -->
  <div class="live-ticker" id="shodanTicker" style="display:none;margin-bottom:14px">
    <div class="spinner"></div>
    <div class="live-ip-text" id="shodanTickerText">در حال جمع‌آوری...</div>
  </div>

  <div class="card" id="shodanResults" style="display:none">
    <div class="card-hd">
      <div class="card-hd-left">📋 IP های جمع‌آوری شده</div>
      <span id="shodanCount" style="color:var(--g);font-family:'IBM Plex Mono',monospace">0</span>
    </div>
    <div class="card-bd">
      <div class="ip-list" id="shodanIpChips"></div>
      <div style="margin-top:12px;display:flex;gap:8px">
        <button class="btn btn-sm" onclick="copyAllShodan()">📋 کپی همه</button>
        <button class="btn btn-primary btn-sm" onclick="scanShodanIPs()">⚡ اسکن این IP ها</button>
      </div>
    </div>
  </div>
</div>

<!-- ───── LOG PAGE ───── -->
<div id="page-log" class="page fade-in">
  <div class="page-hd">
    <div class="page-hd-left"><h2>لاگ</h2><p>رویدادهای WebSocket و سیستم</p></div>
    <button class="btn btn-sm" onclick="clearLog()">🗑 پاک کردن</button>
  </div>
  <div class="card">
    <div class="card-bd" style="padding:0">
      <div class="log-box" id="logBox"><span class="log-dim">لاگی نیست...</span></div>
    </div>
  </div>
</div>

</div><!-- /main -->
</div><!-- /layout -->

<script>
// ═══ State ═══
let ws = null;
let currentConfig = null;
let parsedConfig = null;
let scanStatus = 'idle';
let p2Results = [];
let shodanIPs = [];
let logErrors = 0;

// ═══ Navigation ═══
function nav(page, btn) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(b => b.classList.remove('active'));
  const el = document.getElementById('page-' + page);
  if (el) { el.classList.add('active'); }
  if (btn) btn.classList.add('active');
  else {
    const b = document.querySelector('[data-page="'+page+'"]');
    if (b) b.classList.add('active');
  }
  if (page === 'results') refreshResults();
  if (page === 'history') refreshHistory();
  if (page === 'log') { logErrors = 0; const lb = document.getElementById('logBadge'); if(lb) lb.style.display='none'; }
}

// ═══ WebSocket ═══
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');
  ws.onmessage = e => {
    try { handleWS(JSON.parse(e.data)); } catch(err) {}
  };
  ws.onclose = () => setTimeout(connectWS, 2000);
  ws.onerror = () => {};
}

function handleWS(msg) {
  const { type, payload } = msg;

  switch(type) {
    case 'status':
      updateStatus(payload.status, payload.phase);
      break;
    case 'progress':
      updateProgress(payload);
      break;
    case 'live_ip':
      updateLiveTicker(payload.ip, 'phase1');
      break;
    case 'phase2_start':
      document.getElementById('phaseLabel').textContent = 'Phase 2 — ' + payload.count + ' IP';
      document.getElementById('progressBar').classList.add('phase2');
      updateStatus('scanning', 'phase2');
      addLog('Phase 2 شروع شد: ' + payload.count + ' IP', 'info');
      break;
    case 'phase2_done':
      p2Results = payload.results || [];
      refreshResults();
      clearLiveTicker();
      break;
    case 'scan_done':
      updateStatus('done', '');
      addLog('✓ اسکن تموم شد: ' + payload.passed + ' IP موفق — ' + payload.duration, 'ok');
      showBtns(false);
      refreshResults();
      refreshHistory();
      clearLiveTicker();
      break;
    case 'error':
      addLog('✗ خطا: ' + payload.message, 'err');
      break;
    case 'shodan_status':
      document.getElementById('shodanTicker').style.display = 'flex';
      document.getElementById('shodanTickerText').textContent = 'در حال جمع‌آوری از Shodan...';
      break;
    case 'shodan_done':
      shodanIPs = payload.ips || [];
      renderShodanResults(shodanIPs, payload.total);
      document.getElementById('shodanTicker').style.display = 'none';
      document.getElementById('btnShodan').disabled = false;
      addLog('✓ Shodan: ' + shodanIPs.length + ' IP از ' + payload.total + ' نتیجه', 'ok');
      break;
    case 'shodan_error':
      showShodanAlert(payload.message, 'err');
      document.getElementById('shodanTicker').style.display = 'none';
      document.getElementById('btnShodan').disabled = false;
      break;
  }
}

// ═══ Live Ticker ═══
function updateLiveTicker(ip, phase) {
  const ticker = document.getElementById('liveTicker');
  const phaseLabel = phase === 'phase2' ? '🔬 Phase 2' : '⚡ Phase 1';
  ticker.innerHTML =
    '<div class="spinner"></div>' +
    '<div class="live-ip-phase">' + phaseLabel + '</div>' +
    '<div class="live-ip-text">' + ip + '</div>';
}

function clearLiveTicker() {
  document.getElementById('liveTicker').innerHTML =
    '<div class="ticker-idle">⊙ اسکن تموم شد</div>';
}

// ═══ Scan Control ═══
async function startScan() {
  const ipInput = document.getElementById('ipInput').value.trim();
  const maxIPs = parseInt(document.getElementById('maxIPs').value) || 0;
  const configJSON = buildConfigJSON();

  const btn = document.getElementById('btnStart');
  btn.disabled = true;

  const res = await fetch('/api/scan/start', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({ config: configJSON, ipRanges: ipInput, maxIPs })
  });
  const data = await res.json();
  btn.disabled = false;

  if (!data.ok) {
    addLog('✗ خطا: ' + data.error, 'err');
    return;
  }

  p2Results = [];
  document.getElementById('progressBar').classList.remove('phase2');
  updateStatus('scanning', 'phase1');
  showBtns(true);
  addLog('▶ اسکن شروع شد', 'ok');
}

async function stopScan() {
  await fetch('/api/scan/stop', { method: 'POST' });
  updateStatus('idle', '');
  showBtns(false);
  clearLiveTicker();
}

async function pauseScan() {
  const res = await fetch('/api/scan/pause', { method: 'POST' });
  const data = await res.json();
  const btn = document.getElementById('btnPause');
  if (data.message === 'paused') {
    btn.textContent = '▶ ادامه';
    updateStatus('paused', '');
  } else {
    btn.textContent = '⏸ توقف';
    updateStatus('scanning', '');
  }
}

function showBtns(running) {
  document.getElementById('btnStart').style.display = running ? 'none' : 'inline-flex';
  document.getElementById('btnPause').style.display = running ? 'inline-flex' : 'none';
  document.getElementById('btnStop').style.display = running ? 'inline-flex' : 'none';
}

// ═══ Progress & Status ═══
function updateStatus(status, phase) {
  scanStatus = status;
  const dot = document.getElementById('statusDot');
  const txt = document.getElementById('statusText');

  const labels = { idle: 'آماده', scanning: 'اسکن', paused: 'متوقف', done: 'تموم شد' };
  txt.textContent = labels[status] || status;

  dot.className = 'dot';
  if (status === 'scanning') { dot.classList.add('dot-live'); }
  else if (status === 'paused') { dot.classList.add('dot-warn'); }
  else if (status === 'done') { dot.classList.add('dot-done'); }
  else { dot.classList.add('dot-idle'); }

  if (phase === 'phase2') {
    document.getElementById('phaseLabel').textContent = 'Phase 2';
  } else if (phase === 'phase1') {
    document.getElementById('phaseLabel').textContent = 'Phase 1';
  }
}

function updateProgress(p) {
  document.getElementById('stTotal').textContent = p.Total || '—';
  document.getElementById('stDone').textContent = p.Done || 0;
  document.getElementById('stOk').textContent = p.Succeeded || 0;
  document.getElementById('stFail').textContent = p.Failed || 0;
  document.getElementById('stETA').textContent = p.ETA || '—';

  const pct = p.Total > 0 ? Math.round(p.Done / p.Total * 100) : 0;
  document.getElementById('progressBar').style.width = pct + '%';
  document.getElementById('pctLabel').textContent = pct + '%';

  const rate = (p.Rate || 0).toFixed(1);
  document.getElementById('progDetail').textContent =
    p.Done + ' / ' + (p.Total || '?') + '  ·  ' + rate + ' IP/s';
  document.getElementById('progressText').textContent =
    p.Done + '/' + (p.Total||'?') + ' (' + rate + ' ip/s)';

  // update live ticker if IP came through progress (fallback)
  if (p.CurrentIP) updateLiveTicker(p.CurrentIP, scanStatus === 'phase2' ? 'phase2' : 'phase1');
}

// ═══ Results ═══
function refreshResults() {
  fetch('/api/results').then(r => r.json()).then(data => {
    p2Results = data.phase2 || [];
    renderResults(p2Results);
  });
}

function renderResults(results) {
  const passed = (results || []).filter(r => r.Passed);
  document.getElementById('resultsSummary').textContent =
    passed.length + ' IP موفق از ' + (results || []).length + ' تست شده';
  document.getElementById('passedCount').textContent = passed.length;

  const chips = document.getElementById('ipChips');
  if (!passed.length) {
    chips.innerHTML = '<span style="color:var(--dim);font-size:13px">نتیجه‌ای نیست</span>';
  } else {
    chips.innerHTML = passed.map(r =>
      '<div class="ip-chip" onclick="copyIP(\''+r.IP+'\')" title="کلیک برای کپی">' +
      r.IP +
      '<span style="opacity:.5;font-size:10px">' + Math.round(r.AvgLatencyMs) + 'ms</span>' +
      '</div>'
    ).join('');
  }

  const tbody = document.getElementById('resultsTbody');
  if (!results || !results.length) {
    tbody.innerHTML = '<tr><td colspan="10" style="text-align:center;color:var(--dim);padding:28px">نتیجه‌ای نیست</td></tr>';
    return;
  }

  tbody.innerHTML = results.map((r, i) => {
    const sc = r.StabilityScore || 0;
    const sc2 = (typeof sc === 'number') ? sc : 0;
    const scoreColor = sc2 >= 75 ? 'var(--g)' : sc2 >= 50 ? 'var(--y)' : 'var(--r)';
    const latColor = r.AvgLatencyMs <= 500 ? 'var(--g)' : r.AvgLatencyMs <= 1500 ? 'var(--y)' : 'var(--r)';
    const badge = r.Passed
      ? '<span class="badge bg">PASS</span>'
      : '<span class="badge br" title="' + (r.FailReason||'') + '">FAIL</span>';
    return '<tr>' +
      '<td style="color:var(--dim)">' + (i+1) + '.</td>' +
      '<td style="color:var(--c)">' + r.IP + '</td>' +
      '<td style="color:' + scoreColor + '">' + sc2.toFixed(1) + '</td>' +
      '<td style="color:' + latColor + '">' + Math.round(r.AvgLatencyMs) + 'ms</td>' +
      '<td style="color:var(--dim)">' + (r.JitterMs > 0 ? r.JitterMs.toFixed(0) + 'ms' : '—') + '</td>' +
      '<td style="color:' + (r.PacketLossPct <= 5 ? 'var(--g)' : 'var(--r)') + '">' + (r.PacketLossPct||0).toFixed(0) + '%</td>' +
      '<td style="color:var(--c2)">' + (r.DownloadMbps > 0 ? r.DownloadMbps.toFixed(1) + ' M' : '—') + '</td>' +
      '<td style="color:var(--c2)">' + (r.UploadMbps > 0 ? r.UploadMbps.toFixed(1) + ' M' : '—') + '</td>' +
      '<td>' + badge + '</td>' +
      '<td><button class="copy-btn" onclick="copyIP(\'' + r.IP + '\')">📋</button></td>' +
    '</tr>';
  }).join('');
}

// ═══ History ═══
function refreshHistory() {
  fetch('/api/sessions').then(r => r.json()).then(sessions => {
    const el = document.getElementById('historyList');
    if (!sessions || !sessions.length) {
      el.innerHTML = '<p style="color:var(--dim)">هنوز اسکنی انجام نشده</p>';
      return;
    }
    el.innerHTML = sessions.map(s =>
      '<div class="hist-item" onclick="showSession(\'' + s.id + '\')">' +
      '<span style="font-family:monospace;color:var(--c);font-size:13px">' + new Date(s.startedAt).toLocaleString('fa-IR') + '</span>' +
      '<span style="color:var(--dim);font-size:12px">' + s.duration + '</span>' +
      '<span style="color:var(--dim);font-size:12px">' + s.totalIPs + ' IP</span>' +
      '<span class="badge bg">' + s.passed + ' passed</span>' +
      '</div>'
    ).join('');
  });
}

function showSession(id) {
  fetch('/api/sessions').then(r => r.json()).then(sessions => {
    const s = sessions.find(x => x.id === id);
    if (!s) return;
    p2Results = s.results || [];
    renderResults(p2Results);
    nav('results');
  });
}

// ═══ Config ═══
function buildConfigJSON() {
  if (currentConfig) return currentConfig;

  const threads = parseInt(document.getElementById('qThreads').value) || 200;
  const timeout = parseInt(document.getElementById('qTimeout').value) || 8;
  const maxLat = parseInt(document.getElementById('qMaxLat').value) || 3500;
  const rounds = parseInt(document.getElementById('qRounds').value) || 3;
  const sampleSize = parseInt(document.getElementById('sampleSize').value) || 1;

  const cfg = {
    scan: {
      threads,
      timeout,
      maxLatency: maxLat,
      stabilityRounds: rounds,
      stabilityInterval: 5,
      sampleSize,
      shuffle: true,
    }
  };
  return JSON.stringify(cfg);
}

function saveConfig() {
  const cfg = {
    scan: {
      threads: parseInt(document.getElementById('cfgThreads').value) || 200,
      timeout: parseInt(document.getElementById('cfgTimeout').value) || 8,
      maxLatency: parseInt(document.getElementById('cfgMaxLat').value) || 3500,
      retries: parseInt(document.getElementById('cfgRetries').value) || 2,
      maxIPs: parseInt(document.getElementById('cfgMaxIPs').value) || 0,
      sampleSize: parseInt(document.getElementById('cfgSampleSize').value) || 1,
      testUrl: document.getElementById('cfgTestURL').value,
      shuffle: document.getElementById('cfgShuffle').checked,
      stabilityRounds: parseInt(document.getElementById('cfgRounds').value) || 3,
      stabilityInterval: parseInt(document.getElementById('cfgInterval').value) || 5,
      packetLossCount: parseInt(document.getElementById('cfgPLCount').value) || 5,
      maxPacketLossPct: parseFloat(document.getElementById('cfgMaxPL').value),
      minDownloadMbps: parseFloat(document.getElementById('cfgMinDL').value) || 0,
      minUploadMbps: parseFloat(document.getElementById('cfgMinUL').value) || 0,
      speedTest: document.getElementById('cfgSpeed').checked,
      jitterTest: document.getElementById('cfgJitter').checked,
      downloadUrl: document.getElementById('cfgDLURL').value,
      uploadUrl: document.getElementById('cfgULURL').value,
    },
    fragment: {
      mode: document.getElementById('cfgFragMode').value,
      packets: document.getElementById('cfgFragPkts').value,
      manual: {
        length: document.getElementById('cfgFragLen').value,
        interval: document.getElementById('cfgFragInt').value,
      }
    },
    xray: {
      logLevel: document.getElementById('cfgXrayLog').value,
      mux: {
        enabled: document.getElementById('cfgMuxEnabled').checked,
        concurrency: parseInt(document.getElementById('cfgMuxConc').value) || -1,
      }
    },
    shodan: {
      mode: document.getElementById('cfgShodanMode').value,
      apiKey: document.getElementById('cfgShodanKey').value,
      pages: parseInt(document.getElementById('cfgShodanPages').value) || 1,
      useDefaultQuery: document.getElementById('cfgShodanUseDefault').checked,
      excludeCFRanges: document.getElementById('cfgShodanExcludeCF').checked,
      saveHarvestedIPs: document.getElementById('cfgShodanSave').value,
      appendToExisting: document.getElementById('cfgShodanAppend').checked,
    }
  };

  currentConfig = JSON.stringify(cfg);
  addLog('✓ تنظیمات ذخیره شد', 'ok');
  document.getElementById('configSummary').textContent =
    'threads:' + cfg.scan.threads + ' · rounds:' + cfg.scan.stabilityRounds + ' · frag:' + cfg.fragment.mode;
  nav('scan');
}

// ═══ Import Link ═══
async function parseLink() {
  const input = document.getElementById('linkInput').value.trim();
  if (!input) return;

  const res = await fetch('/api/config/parse', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({ input })
  });
  const data = await res.json();

  if (!data.ok) {
    addLog('✗ خطا: ' + data.error, 'err');
    return;
  }

  parsedConfig = data.config;
  const p = data.parsed;
  document.getElementById('parsedBox').innerHTML =
    '<span class="k">uuid: </span><span class="v">' + maskUUID(p.uuid) + '</span><br>' +
    '<span class="k">address: </span><span class="v">' + p.address + '</span><br>' +
    '<span class="k">port: </span><span class="v">' + p.port + '</span><br>' +
    '<span class="k">type: </span><span class="v">' + p.type + '</span><br>' +
    '<span class="k">method: </span><span class="v">' + p.method + '</span><br>' +
    (p.sni ? '<span class="k">sni: </span><span class="v">' + p.sni + '</span><br>' : '') +
    (p.path ? '<span class="k">path: </span><span class="v">' + p.path + '</span><br>' : '') +
    (p.fp ? '<span class="k">fingerprint: </span><span class="v">' + p.fp + '</span>' : '');

  document.getElementById('parsedResult').style.display = 'block';
  addLog('✓ کانفیگ parse شد: ' + p.address + ':' + p.port + ' (' + p.method + '/' + p.type + ')', 'ok');
}

function applyParsed() {
  if (!parsedConfig) return;
  currentConfig = parsedConfig;
  document.getElementById('configSummary').textContent = '✓ کانفیگ از لینک';
  document.getElementById('parsedResult').style.display = 'none';
  addLog('✓ کانفیگ اعمال شد', 'ok');
  nav('scan');
}

function maskUUID(uuid) {
  if (!uuid || uuid.length < 8) return uuid;
  return uuid.slice(0, 8) + '••••••••';
}

// ═══ Shodan ═══
async function startShodan() {
  const key = document.getElementById('shodanKey').value.trim();
  if (!key) {
    showShodanAlert('API Key الزامی است', 'err');
    return;
  }

  document.getElementById('btnShodan').disabled = true;
  document.getElementById('shodanTicker').style.display = 'flex';
  document.getElementById('shodanTickerText').textContent = 'در حال ارتباط با Shodan...';
  document.getElementById('shodanResults').style.display = 'none';
  document.getElementById('shodanAlert').style.display = 'none';

  const res = await fetch('/api/shodan/harvest', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      apiKey: key,
      query: document.getElementById('shodanQuery').value.trim(),
      pages: parseInt(document.getElementById('shodanPages').value) || 1,
      excludeCF: document.getElementById('shodanExcludeCF').checked,
      autoScan: document.getElementById('shodanAutoScan').checked,
    })
  });
  const data = await res.json();

  if (!data.ok) {
    document.getElementById('shodanTicker').style.display = 'none';
    document.getElementById('btnShodan').disabled = false;
    showShodanAlert(data.error, 'err');
  } else {
    addLog('Shodan harvest شروع شد...', 'info');
  }
}

function renderShodanResults(ips, total) {
  shodanIPs = ips || [];
  document.getElementById('shodanCount').textContent = shodanIPs.length;
  document.getElementById('shodanResults').style.display = 'block';

  const chips = document.getElementById('shodanIpChips');
  chips.innerHTML = shodanIPs.slice(0, 200).map(ip =>
    '<div class="ip-chip" onclick="copyIP(\'' + ip + '\')">' + ip + '</div>'
  ).join('');

  if (shodanIPs.length > 200) {
    chips.innerHTML += '<span style="color:var(--dim);font-size:12px"> ... و ' + (shodanIPs.length - 200) + ' IP دیگر</span>';
  }
}

function copyAllShodan() {
  const text = shodanIPs.join('\n');
  navigator.clipboard.writeText(text).then(() => addLog('✓ همه IP ها کپی شد (' + shodanIPs.length + ')', 'ok'));
}

function scanShodanIPs() {
  if (!shodanIPs.length) return;
  document.getElementById('ipInput').value = shodanIPs.join('\n');
  nav('scan');
  addLog('IP های Shodan به صفحه اسکن منتقل شد', 'info');
}

function showShodanAlert(msg, type) {
  const el = document.getElementById('shodanAlert');
  const cls = type === 'err' ? 'alert-err' : type === 'warn' ? 'alert-warn' : 'alert-info';
  el.className = 'alert ' + cls;
  el.textContent = msg;
  el.style.display = 'block';
}

// ═══ Export ═══
function exportResults(format) {
  window.location.href = '/api/results/export?format=' + format;
}

// ═══ Copy ═══
function copyIP(ip) {
  navigator.clipboard.writeText(ip).then(() => {
    addLog('📋 کپی: ' + ip, 'ok');
  }).catch(() => {
    const el = document.createElement('textarea');
    el.value = ip;
    document.body.appendChild(el);
    el.select();
    document.execCommand('copy');
    document.body.removeChild(el);
    addLog('📋 کپی: ' + ip, 'ok');
  });
}

// ═══ Log ═══
let logLines = 0;
function addLog(msg, type = 'dim') {
  const box = document.getElementById('logBox');
  if (box.children.length === 1 && box.firstChild.textContent === 'لاگی نیست...') {
    box.innerHTML = '';
  }
  const classes = { ok: 'log-ok', err: 'log-err', info: 'log-info', warn: 'log-warn', dim: 'log-dim' };
  const el = document.createElement('div');
  el.className = classes[type] || 'log-dim';
  const time = new Date().toLocaleTimeString('fa-IR');
  el.textContent = time + '  ' + msg;
  box.appendChild(el);
  box.scrollTop = box.scrollHeight;
  while (box.children.length > 300) box.removeChild(box.firstChild);

  if (type === 'err') {
    logErrors++;
    const lb = document.getElementById('logBadge');
    const activePage = document.querySelector('.nav-item.active');
    if (activePage && activePage.getAttribute('data-page') !== 'log') {
      if(lb) { lb.style.display='inline'; lb.textContent = logErrors; }
    }
  }
}

function clearLog() {
  document.getElementById('logBox').innerHTML = '<span class="log-dim">لاگی نیست...</span>';
  logErrors = 0;
  const lb = document.getElementById('logBadge');
  if(lb) lb.style.display='none';
}

// ═══ Init ═══
connectWS();
fetch('/api/status').then(r => r.json()).then(d => {
  updateStatus(d.status || 'idle', d.phase || '');
});
</script>
</body>
</html>
`
