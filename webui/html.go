package webui

const indexHTMLContent = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Piyazche Scanner</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=Vazirmatn:wght@300;400;500;700&display=swap" rel="stylesheet">
<style>
:root{
  --bg:#050608;--bg2:#0a0c10;--bg3:#0f1218;--bg4:#14181f;
  --bd:#181c26;--bd2:#1e2330;--bd3:#252d3d;
  --tx:#c8d0e4;--tx2:#8891aa;--dim:#3d4560;
  --g:#2dd4a0;--gd:#081f18;--g2:#0f9e72;
  --c:#38b6f8;--cd:#061828;--c2:#0ea5e9;
  --y:#f5c842;--yd:#201a04;
  --r:#f06060;--rd:#220c0c;
  --p:#a78bfa;--pd:#130f28;
  --o:#fb923c;
  --rad:10px;--rad-sm:6px;--rad-xs:4px;
}
*{margin:0;padding:0;box-sizing:border-box}
html{height:100%}
body{font-family:'Vazirmatn',Tahoma,sans-serif;background:var(--bg);color:var(--tx);height:100%;font-size:14px;line-height:1.6;overflow:hidden}
.app{display:grid;grid-template-columns:220px 1fr;grid-template-rows:52px 1fr;height:100vh}

/* ══ TOPBAR ══ */
.topbar{grid-column:1/-1;background:rgba(10,12,16,.98);border-bottom:1px solid var(--bd);display:flex;align-items:center;padding:0 18px;gap:12px;position:relative;z-index:100}
.logo{font-family:'IBM Plex Mono',monospace;font-size:17px;font-weight:600;letter-spacing:-0.5px;user-select:none;display:flex;align-items:baseline;gap:4px}
.logo-pi{color:var(--tx2)}.logo-az{color:var(--c);font-weight:700}
.logo-che{color:var(--tx2)}.logo-ver{color:var(--dim);font-size:10px;margin-right:2px}
.status-pill{display:flex;align-items:center;gap:6px;padding:4px 12px;border-radius:20px;font-size:11px;background:var(--bg3);border:1px solid var(--bd2);font-family:'IBM Plex Mono',monospace}
.dot{width:6px;height:6px;border-radius:50%;flex-shrink:0;transition:all .3s}
.dot-idle{background:var(--dim)}
.dot-scan{background:var(--g);box-shadow:0 0 8px var(--g);animation:pulse 1.2s ease-in-out infinite}
.dot-warn{background:var(--y)}
.dot-done{background:var(--c)}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:.3;transform:scale(.7)}}
.tb-right{margin-right:auto;display:flex;align-items:center;gap:10px}
.proxy-chip{display:inline-flex;align-items:center;gap:5px;padding:3px 10px;border-radius:5px;font-size:11px;font-family:'IBM Plex Mono',monospace;background:rgba(45,212,160,.1);border:1px solid rgba(45,212,160,.3);color:var(--g);cursor:pointer}
.proxy-chip:hover{background:rgba(45,212,160,.15)}

/* ══ SIDEBAR ══ */
.sidebar{background:var(--bg2);border-left:1px solid var(--bd);display:flex;flex-direction:column;overflow-y:auto;padding-bottom:16px}
.nav-group{padding:14px 14px 4px;font-size:9px;letter-spacing:2.5px;text-transform:uppercase;color:var(--dim);font-family:'IBM Plex Mono',monospace}
.nav-item{display:flex;align-items:center;gap:9px;padding:8px 16px;cursor:pointer;transition:all .12s;color:var(--tx2);font-size:12.5px;border:none;background:none;width:100%;text-align:right;border-right:2px solid transparent;font-family:'Vazirmatn',sans-serif}
.nav-item:hover{background:var(--bg3);color:var(--tx)}
.nav-item.active{background:linear-gradient(90deg,rgba(56,182,248,.08),transparent);color:var(--c);border-right-color:var(--c)}
.nav-icon{font-size:13px;min-width:16px;text-align:center;opacity:.8}
.nav-badge{margin-right:auto;background:var(--bd2);color:var(--dim);font-size:9px;padding:1px 6px;border-radius:8px;font-family:'IBM Plex Mono',monospace}
.nav-badge.live{background:rgba(45,212,160,.15);color:var(--g)}

/* ══ MAIN ══ */
.main{overflow-y:auto;overflow-x:hidden}
.main::-webkit-scrollbar{width:4px}
.main::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
.page{display:none;padding:20px 22px 30px;min-height:100%}
.page.active{display:block}

/* ══ PAGE HEADER ══ */
.phd{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:16px;gap:12px}
.phd-l h2{font-size:18px;font-weight:700;letter-spacing:-.3px}
.phd-l p{font-size:11px;color:var(--dim);margin-top:2px}
.phd-r{display:flex;gap:7px;align-items:center;flex-shrink:0}

/* ══ STATS ══ */
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:8px;margin-bottom:14px}
.stat-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad-sm);padding:11px 14px;position:relative;overflow:hidden;transition:border-color .2s}
.stat-card::before{content:'';position:absolute;inset:0;background:linear-gradient(135deg,transparent 60%,rgba(255,255,255,.01));pointer-events:none}
.stat-v{font-size:22px;font-weight:700;font-family:'IBM Plex Mono',monospace;line-height:1.2;transition:color .3s}
.stat-l{font-size:9px;color:var(--dim);margin-top:3px;letter-spacing:1.5px;text-transform:uppercase}

/* ══ PROGRESS ══ */
.prog-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:14px}
.prog-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 16px;font-size:11px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:'IBM Plex Mono',monospace}
.prog-hd-l{display:flex;align-items:center;gap:7px}
.prog-bd{padding:14px 16px}
.prog-meta{display:flex;justify-content:space-between;font-size:11px;color:var(--dim);margin-bottom:5px}
.prog-pct{font-family:'IBM Plex Mono',monospace;color:var(--c);font-weight:600;font-size:13px}
.prog-wrap{background:var(--bg);border-radius:3px;height:5px;overflow:hidden;margin-bottom:10px}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:3px;transition:width .5s cubic-bezier(.4,0,.2,1);width:0%}
.prog-bar.p2{background:linear-gradient(90deg,var(--p),var(--c))}

/* ══ LIVE FEED ══ */
.live-feed{background:var(--bg);border:1px solid var(--bd);border-radius:var(--rad-sm);overflow:hidden}
.live-feed-hd{padding:6px 12px;border-bottom:1px solid var(--bd);display:flex;align-items:center;gap:8px;font-size:10px;color:var(--dim);font-family:'IBM Plex Mono',monospace}
.live-feed-body{height:120px;overflow-y:auto;padding:8px 12px;display:flex;flex-direction:column-reverse;gap:2px}
.live-feed-body::-webkit-scrollbar{width:3px}
.live-feed-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.live-row{display:flex;align-items:center;gap:8px;font-family:'IBM Plex Mono',monospace;font-size:11px;padding:1px 0;animation:fadeIn .2s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(-3px)}to{opacity:1;transform:none}}
.live-row-ok{color:var(--g)}
.live-row-fail{color:var(--dim)}
.live-row-scan{color:var(--tx2)}
.live-row-p2{color:var(--p)}
.live-ip{color:var(--c);min-width:120px;font-weight:500}
.live-lat{color:var(--y);min-width:60px}
.live-tag{font-size:9px;padding:1px 5px;border-radius:3px;flex-shrink:0}
.tag-ok{background:rgba(45,212,160,.15);color:var(--g)}.tag-fail{background:rgba(240,96,96,.1);color:var(--r)}.tag-p2{background:rgba(167,139,250,.15);color:var(--p)}
.spin{width:12px;height:12px;border:1.5px solid var(--bd3);border-top-color:var(--c);border-radius:50%;animation:spin .6s linear infinite;flex-shrink:0}
@keyframes spin{to{transform:rotate(360deg)}}

/* ══ CARDS ══ */
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:12px}
.card-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 14px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:'IBM Plex Mono',monospace;letter-spacing:.5px}
.card-bd{padding:14px}

/* ══ TABLE ══ */
.tbl-wrap{overflow-x:auto}
.tbl{width:100%;border-collapse:collapse;font-size:11.5px;font-family:'IBM Plex Mono',monospace}
.tbl th{padding:7px 10px;text-align:right;color:var(--dim);font-weight:500;border-bottom:1px solid var(--bd);background:var(--bg3);font-size:10px;letter-spacing:.5px;white-space:nowrap;position:sticky;top:0}
.tbl td{padding:6px 10px;border-bottom:1px solid var(--bd);vertical-align:middle;white-space:nowrap}
.tbl tbody tr:hover td{background:rgba(255,255,255,.015)}
.tbl tr.pass-row td:first-child{border-right:2px solid var(--g)}
.tbl tr.fail-row td:first-child{border-right:2px solid transparent}
.tbl tr.p1-row td:first-child{border-right:2px solid var(--c)}

/* ══ BADGES ══ */
.badge{display:inline-flex;align-items:center;padding:2px 7px;border-radius:8px;font-size:9px;font-family:'IBM Plex Mono',monospace;font-weight:600;letter-spacing:.5px}
.bg{background:var(--gd);color:var(--g);border:1px solid rgba(45,212,160,.2)}
.bc{background:var(--cd);color:var(--c);border:1px solid rgba(56,182,248,.2)}
.by{background:var(--yd);color:var(--y);border:1px solid rgba(245,200,66,.2)}
.br{background:var(--rd);color:var(--r);border:1px solid rgba(240,96,96,.2)}
.bp{background:var(--pd);color:var(--p);border:1px solid rgba(167,139,250,.2)}

/* ══ BUTTONS ══ */
.btn{display:inline-flex;align-items:center;gap:6px;padding:7px 15px;border-radius:var(--rad-sm);border:1px solid var(--bd2);background:var(--bg3);color:var(--tx);cursor:pointer;font-size:12px;font-family:'Vazirmatn',sans-serif;transition:all .12s;white-space:nowrap;font-weight:500}
.btn:hover{background:var(--bd2);border-color:var(--bd3)}.btn:active{transform:scale(.97)}.btn:disabled{opacity:.35;cursor:not-allowed;pointer-events:none}
.btn-primary{background:linear-gradient(135deg,var(--cd),rgba(56,182,248,.1));border-color:rgba(56,182,248,.4);color:var(--c)}.btn-primary:hover{background:var(--c2);color:#000;border-color:var(--c2)}
.btn-success{background:var(--gd);border-color:rgba(45,212,160,.3);color:var(--g)}.btn-success:hover{background:var(--g2);color:#000;border-color:var(--g)}
.btn-danger{background:var(--rd);border-color:rgba(240,96,96,.3);color:var(--r)}.btn-danger:hover{background:var(--r);color:#fff}
.btn-warn{background:var(--yd);border-color:rgba(245,200,66,.3);color:var(--y)}.btn-warn:hover{background:var(--y);color:#000}
.btn-sm{padding:5px 11px;font-size:11px}.btn-xs{padding:2px 8px;font-size:10px}

/* ══ FORMS ══ */
textarea,input[type=text],input[type=number],input[type=password],select{background:var(--bg3);border:1px solid var(--bd2);color:var(--tx);border-radius:var(--rad-sm);padding:7px 11px;font-size:12px;font-family:'IBM Plex Mono',monospace;width:100%;outline:none;direction:ltr;transition:border-color .15s}
textarea:focus,input:focus,select:focus{border-color:rgba(56,182,248,.5);box-shadow:0 0 0 2px rgba(56,182,248,.07)}
label{display:block;font-size:11px;color:var(--dim);margin-bottom:4px;text-align:right;font-family:'Vazirmatn',sans-serif}
.f-row{margin-bottom:11px}.f-grid{display:grid;grid-template-columns:1fr 1fr;gap:11px}.f-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:11px}
.f-sep{height:1px;background:var(--bd);margin:14px 0}
.chk-row{display:flex;align-items:center;gap:7px;cursor:pointer;font-size:12px;color:var(--tx2);font-family:'Vazirmatn',sans-serif}
.chk-row input{width:auto;cursor:pointer;accent-color:var(--c)}
.parsed-box{background:var(--bg3);border:1px solid rgba(45,212,160,.2);border-radius:var(--rad-sm);padding:12px;font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--g);direction:ltr;line-height:1.9}
.parsed-box .k{color:var(--dim)}.parsed-box .v{color:var(--c)}
.cfg-chip{display:flex;flex-wrap:wrap;gap:6px;padding:8px 10px;background:var(--bg3);border:1px solid var(--bd2);border-radius:var(--rad-sm);font-size:10px;font-family:'IBM Plex Mono',monospace}
.cfg-item{color:var(--dim)}.cfg-item b{color:var(--tx2)}

/* ══ CHIPS ══ */
.ip-chips{display:flex;flex-wrap:wrap;gap:5px;padding:4px 0}
.ip-chip{background:var(--cd);border:1px solid rgba(56,182,248,.25);border-radius:4px;padding:3px 9px;font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--c);cursor:pointer;display:flex;align-items:center;gap:5px;transition:all .12s}
.ip-chip:hover{background:var(--c);color:#000;border-color:var(--c)}
.ip-chip .lat{font-size:9px;opacity:.6}

/* ══ HIST ══ */
.hist-item{background:var(--bg3);border:1px solid var(--bd);border-radius:var(--rad-sm);padding:10px 14px;display:flex;align-items:center;gap:12px;margin-bottom:7px;cursor:pointer;transition:all .12s;font-family:'IBM Plex Mono',monospace;font-size:11px}
.hist-item:hover{border-color:var(--bd3);background:var(--bg4)}

/* ══ TUI ══ */
.tui-wrap{background:#020304;border:1px solid var(--bd2);border-radius:var(--rad-sm);overflow:hidden;font-family:'IBM Plex Mono',monospace;font-size:11.5px}
.tui-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:6px 12px;display:flex;align-items:center;gap:7px;font-size:10px;color:var(--dim)}
.tui-dots{display:flex;gap:4px}
.tui-dot{width:9px;height:9px;border-radius:50%}
.tui-body{padding:10px 12px;height:400px;overflow-y:auto;line-height:1.85}
.tui-body::-webkit-scrollbar{width:3px}
.tui-body::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
.tui-line{display:flex;gap:8px}
.tui-t{color:#1e2535;flex-shrink:0;user-select:none}
.tui-ok{color:#2dd4a0}.tui-err{color:#f06060}.tui-info{color:#38b6f8}.tui-warn{color:#f5c842}.tui-scan{color:#5a6680}.tui-p2{color:#a78bfa}
.cursor{display:inline-block;width:7px;height:12px;background:var(--c);animation:blink 1s step-end infinite;vertical-align:middle;margin-right:2px}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}

/* ══ ALERTS ══ */
.alert{border-radius:var(--rad-sm);padding:9px 12px;font-size:12px;margin-bottom:11px;border-right:3px solid;display:flex;align-items:flex-start;gap:8px}
.alert-info{background:var(--cd);border-color:var(--c);color:var(--c)}.alert-warn{background:var(--yd);border-color:var(--y);color:var(--y)}.alert-err{background:var(--rd);border-color:var(--r);color:var(--r)}

/* ══ RESULT TABS ══ */
.tab-bar{display:flex;gap:2px;border-bottom:1px solid var(--bd);margin-bottom:14px}
.tab{padding:7px 16px;font-size:12px;font-family:'Vazirmatn',sans-serif;background:none;border:none;color:var(--tx2);cursor:pointer;border-bottom:2px solid transparent;transition:all .12s;margin-bottom:-1px}
.tab:hover{color:var(--tx)}.tab.active{color:var(--c);border-bottom-color:var(--c);background:rgba(56,182,248,.05)}

/* ══ COPY BTN ══ */
.copy-btn{background:none;border:none;cursor:pointer;color:var(--dim);padding:2px 5px;border-radius:3px;font-size:10px;transition:all .12s}
.copy-btn:hover{color:var(--c);background:var(--cd)}

/* ══ TOOLTIP ══ */
[data-tip]{position:relative}
[data-tip]:hover::after{content:attr(data-tip);position:absolute;bottom:calc(100% + 4px);right:50%;transform:translateX(50%);background:#1a1e28;color:var(--tx);font-size:10px;padding:3px 7px;border-radius:4px;white-space:nowrap;font-family:'IBM Plex Mono',monospace;z-index:100;border:1px solid var(--bd2)}

/* ══ RESPONSIVE ══ */
@media(max-width:768px){.app{grid-template-columns:1fr}.sidebar{display:none}.stats-row{grid-template-columns:repeat(2,1fr)}.f-grid,.f-grid-3{grid-template-columns:1fr}}
::-webkit-scrollbar{width:4px;height:4px}
::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
</style>
</head>
<body>
<div class="app">

<!-- TOPBAR -->
<div class="topbar">
  <div class="logo"><span class="logo-pi">piy</span><span class="logo-az">az</span><span class="logo-che">che</span><span class="logo-ver">v4</span></div>
  <div class="status-pill">
    <div class="dot dot-idle" id="sDot"></div>
    <span id="sTxt" style="font-family:'IBM Plex Mono',monospace;font-size:11px">idle</span>
    <span id="sPhase" style="color:var(--dim);font-size:10px"></span>
  </div>
  <div id="proxyChip" style="display:none" class="proxy-chip" onclick="nav('import')" title="کانفیگ وارد شده — کلیک برای تغییر">
    <span>🔒</span><span id="proxyChipTxt"></span>
  </div>
  <div class="tb-right">
    <span id="tbProgress" style="font-family:'IBM Plex Mono',monospace;font-size:11px;color:var(--dim)"></span>
  </div>
</div>

<!-- SIDEBAR -->
<div class="sidebar">
  <div class="nav-group">اسکنر</div>
  <button class="nav-item active" data-page="scan" onclick="nav('scan',this)">
    <span class="nav-icon">⚡</span>اسکن
    <span class="nav-badge live" id="nbScan" style="display:none">LIVE</span>
  </button>
  <button class="nav-item" data-page="results" onclick="nav('results',this)">
    <span class="nav-icon">📊</span>نتایج
    <span class="nav-badge" id="nbResults">0</span>
  </button>
  <button class="nav-item" data-page="history" onclick="nav('history',this)">
    <span class="nav-icon">🕐</span>تاریخچه
  </button>
  <div class="nav-group">پیکربندی</div>
  <button class="nav-item" data-page="config" onclick="nav('config',this)">
    <span class="nav-icon">⚙️</span>تنظیمات
  </button>
  <button class="nav-item" data-page="import" onclick="nav('import',this)">
    <span class="nav-icon">🔗</span>وارد کردن لینک
  </button>
  <div class="nav-group">ابزار</div>
  <button class="nav-item" data-page="shodan" onclick="nav('shodan',this)">
    <span class="nav-icon">🔍</span>Shodan
  </button>
  <button class="nav-item" data-page="tui" onclick="nav('tui',this)">
    <span class="nav-icon">⌨️</span>لاگ
  </button>
</div>

<div class="main">

<!-- ══ SCAN ══ -->
<div id="page-scan" class="page active">
  <div class="phd">
    <div class="phd-l"><h2>اسکن جدید</h2><p>رنج IP رو بده و شروع کن</p></div>
    <div class="phd-r">
      <button class="btn btn-primary" id="btnStart" onclick="startScan()">▶ شروع</button>
      <button class="btn btn-warn" id="btnPause" onclick="pauseScan()" style="display:none">⏸ توقف</button>
      <button class="btn btn-danger" id="btnStop" onclick="stopScan()" style="display:none">■ متوقف</button>
    </div>
  </div>

  <!-- Stats -->
  <div class="stats-row">
    <div class="stat-card"><div class="stat-v" id="stTotal" style="color:var(--dim)">—</div><div class="stat-l">کل IP</div></div>
    <div class="stat-card"><div class="stat-v" id="stDone" style="color:var(--c)">0</div><div class="stat-l">بررسی شده</div></div>
    <div class="stat-card"><div class="stat-v" id="stOk" style="color:var(--g)">0</div><div class="stat-l">موفق ✓</div></div>
    <div class="stat-card"><div class="stat-v" id="stFail" style="color:var(--r)">0</div><div class="stat-l">ناموفق ✗</div></div>
    <div class="stat-card"><div class="stat-v" id="stETA" style="color:var(--y)">—</div><div class="stat-l">زمان باقی</div></div>
  </div>

  <!-- Progress -->
  <div class="prog-card">
    <div class="prog-hd">
      <div class="prog-hd-l"><span id="phaseLabel" style="color:var(--c)">Phase 1</span><span style="color:var(--dim)">— اسکن اولیه</span></div>
      <span id="pctLabel" class="prog-pct">0%</span>
    </div>
    <div class="prog-bd">
      <div class="prog-meta"><span id="progDetail" style="color:var(--dim)">در انتظار شروع...</span><span id="rateLabel" style="color:var(--dim)"></span></div>
      <div class="prog-wrap"><div class="prog-bar" id="progBar"></div></div>
      <!-- Live Feed -->
      <div class="live-feed">
        <div class="live-feed-hd">
          <div class="spin" id="feedSpin" style="display:none"></div>
          <span id="feedPhase" style="color:var(--dim)">feed</span>
          <span id="feedCount" style="margin-right:auto;color:var(--dim)"></span>
        </div>
        <div class="live-feed-body" id="liveFeed">
          <div class="live-row live-row-scan" style="color:var(--dim)">⊙ منتظر شروع اسکن...</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Input row -->
  <div class="f-grid">
    <div class="card">
      <div class="card-hd"><div>🌐 رنج IP</div><button class="btn btn-xs" onclick="previewIPs()">پیش‌نمایش</button></div>
      <div class="card-bd">
        <div class="f-row">
          <label>هر خط: IP یا CIDR — خالی = از ipv4.txt</label>
          <textarea id="ipInput" rows="6" placeholder="104.16.0.0/12&#10;185.42.0.0/16&#10;45.12.33.91"></textarea>
        </div>
        <div class="f-grid">
          <div class="f-row"><label>حداکثر IP (0=همه)</label><input type="number" id="maxIPs" value="0" min="0"></div>
          <div class="f-row"><label>IP از هر subnet</label><input type="number" id="sampleSize" value="1" min="1"></div>
        </div>
        <div id="ipPreview" style="display:none;margin-top:6px;font-size:11px;color:var(--dim);font-family:'IBM Plex Mono',monospace"></div>
      </div>
    </div>
    <div class="card">
      <div class="card-hd"><div>⚡ تنظیمات سریع</div></div>
      <div class="card-bd">
        <div class="f-grid">
          <div class="f-row"><label>Threads</label><input type="number" id="qThreads" value="200" min="1"></div>
          <div class="f-row"><label>Timeout (ثانیه)</label><input type="number" id="qTimeout" value="8" min="1"></div>
          <div class="f-row"><label>Max Latency (ms)</label><input type="number" id="qMaxLat" value="3500"></div>
          <div class="f-row"><label>Stability Rounds</label><input type="number" id="qRounds" value="3" min="0"></div>
        </div>
        <div class="cfg-chip" id="configSummary">
          <span class="cfg-item">پیش‌فرض — لینک وارد نشده</span>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- ══ RESULTS ══ -->
<div id="page-results" class="page">
  <div class="phd">
    <div class="phd-l"><h2>نتایج</h2><p id="resSummary">نتیجه‌ای نیست</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="exportResults('csv')">↓ CSV</button>
      <button class="btn btn-sm" onclick="exportResults('json')">↓ JSON</button>
      <button class="btn btn-sm" onclick="copyAllPassed()">📋 کپی همه</button>
    </div>
  </div>

  <!-- IP chips -->
  <div class="card" style="margin-bottom:12px">
    <div class="card-hd"><div>✓ IP های موفق</div><span id="passedBadge" class="badge bg">0</span></div>
    <div class="card-bd"><div class="ip-chips" id="ipChips"><span style="color:var(--dim);font-size:12px">نتیجه‌ای نیست</span></div></div>
  </div>

  <!-- Tabs -->
  <div class="tab-bar">
    <button class="tab active" onclick="switchTab('p2',this)">🔬 Phase 2 (عمیق)</button>
    <button class="tab" onclick="switchTab('p1',this)">⚡ Phase 1 (اولیه)</button>
  </div>

  <!-- Phase 2 table -->
  <div id="tab-p2" class="card">
    <div class="card-hd"><div>نتایج Phase 2 — تست عمق و پایداری</div><span id="p2CountBadge" style="color:var(--dim);font-size:10px"></span></div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP</th><th>Score</th><th>Latency</th><th>Jitter</th><th>Pkt Loss</th><th>Download</th><th>وضعیت</th><th style="text-align:center">عملیات</th>
        </tr></thead>
        <tbody id="p2Tbody"><tr><td colspan="9" style="text-align:center;color:var(--dim);padding:28px">نتیجه Phase 2 نیست</td></tr></tbody>
      </table>
    </div>
  </div>

  <!-- Phase 1 table -->
  <div id="tab-p1" class="card" style="display:none">
    <div class="card-hd"><div>نتایج Phase 1 — اسکن اولیه</div><span id="p1CountBadge" style="color:var(--dim);font-size:10px"></span></div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP</th><th>Latency</th><th>وضعیت</th><th style="text-align:center">عملیات</th>
        </tr></thead>
        <tbody id="p1Tbody"><tr><td colspan="5" style="text-align:center;color:var(--dim);padding:28px">نتیجه Phase 1 نیست</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<!-- ══ HISTORY ══ -->
<div id="page-history" class="page">
  <div class="phd">
    <div class="phd-l"><h2>تاریخچه</h2><p>اسکن‌های قبلی</p></div>
  </div>
  <div id="histList"><p style="color:var(--dim)">هنوز اسکنی انجام نشده</p></div>
</div>

<!-- ══ CONFIG ══ -->
<div id="page-config" class="page">
  <div class="phd">
    <div class="phd-l"><h2>تنظیمات اسکنر</h2><p>تنظیمات ذخیره و در اسکن بعدی اعمال میشه</p></div>
    <div class="phd-r">
      <button class="btn btn-success" onclick="saveConfig()">💾 ذخیره</button>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>⚡ Phase 1 — اسکن اولیه</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Threads</label><input type="number" id="cfgThreads" value="200" min="1"></div>
        <div class="f-row"><label>Timeout (ثانیه)</label><input type="number" id="cfgTimeout" value="8" min="1"></div>
        <div class="f-row"><label>Max Latency (ms)</label><input type="number" id="cfgMaxLat" value="3500"></div>
        <div class="f-row"><label>Retries</label><input type="number" id="cfgRetries" value="2" min="0"></div>
        <div class="f-row"><label>Max IPs (0=همه)</label><input type="number" id="cfgMaxIPs" value="0" min="0"></div>
        <div class="f-row"><label>Sample per Subnet</label><input type="number" id="cfgSampleSize" value="1" min="1"></div>
      </div>
      <div class="f-row"><label>Test URL</label><input type="text" id="cfgTestURL" value="https://www.gstatic.com/generate_204" placeholder="https://..."></div>
      <label class="chk-row"><input type="checkbox" id="cfgShuffle" checked> Shuffle IPs</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>🔬 Phase 2 — تست عمق و پایداری</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Rounds (0=غیرفعال)</label><input type="number" id="cfgRounds" value="3" min="0"></div>
        <div class="f-row"><label>Interval (ثانیه)</label><input type="number" id="cfgInterval" value="5" min="1"></div>
        <div class="f-row"><label>Ping Count (packet loss)</label><input type="number" id="cfgPLCount" value="5" min="1"></div>
        <div class="f-row"><label>Max Packet Loss (%، -1=غیرفعال)</label><input type="number" id="cfgMaxPL" value="-1" min="-1" max="100"></div>
        <div class="f-row"><label>Min Download (Mbps، 0=غیرفعال)</label><input type="number" id="cfgMinDL" value="0" min="0"></div>
        <div class="f-row"><label>Min Upload (Mbps، 0=غیرفعال)</label><input type="number" id="cfgMinUL" value="0" min="0"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap;margin-bottom:10px">
        <label class="chk-row"><input type="checkbox" id="cfgJitter"> اندازه‌گیری Jitter</label>
        <label class="chk-row"><input type="checkbox" id="cfgSpeed"> Speed Test (کند)</label>
      </div>
      <div class="f-grid" id="speedURLs" style="display:none">
        <div class="f-row"><label>Download URL</label><input type="text" id="cfgDLURL" value="https://speed.cloudflare.com/__down?bytes=5000000"></div>
        <div class="f-row"><label>Upload URL (غیرفعال)</label><input type="text" id="cfgULURL" value="https://speed.cloudflare.com/__up" disabled style="opacity:.4"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>🔀 Fragment</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Mode</label><select id="cfgFragMode"><option value="manual">manual</option><option value="auto">auto</option><option value="off">off</option></select></div>
        <div class="f-row"><label>Packets</label><input type="text" id="cfgFragPkts" value="tlshello"></div>
        <div></div>
        <div class="f-row"><label>Length</label><input type="text" id="cfgFragLen" value="10-20"></div>
        <div class="f-row"><label>Interval (ms)</label><input type="text" id="cfgFragInt" value="10-20"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>🎛️ Xray</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Log Level</label><select id="cfgXrayLog"><option value="none">none</option><option value="error">error</option><option value="warning">warning</option><option value="info">info</option><option value="debug">debug</option></select></div>
        <div class="f-row"><label>Mux Concurrency (-1=off)</label><input type="number" id="cfgMuxConc" value="-1"></div>
        <div style="display:flex;align-items:flex-end;padding-bottom:11px"><label class="chk-row"><input type="checkbox" id="cfgMuxEnabled"> Mux</label></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>🔍 Shodan</div></div>
    <div class="card-bd">
      <div class="f-grid">
        <div class="f-row"><label>API Key</label><input type="password" id="cfgShodanKey" placeholder="xxxxx"></div>
        <div class="f-row"><label>Mode</label><select id="cfgShodanMode"><option value="off">off</option><option value="harvest">harvest</option><option value="scan">scan</option><option value="both">both</option></select></div>
        <div class="f-row"><label>Pages (هر page=100 IP)</label><input type="number" id="cfgShodanPages" value="1" min="1"></div>
        <div class="f-row"><label>Save IPs to file</label><input type="text" id="cfgShodanSave" placeholder="harvested_ips.txt"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap">
        <label class="chk-row"><input type="checkbox" id="cfgShodanUseDefault" checked> کوئری پیش‌فرض</label>
        <label class="chk-row"><input type="checkbox" id="cfgShodanExcludeCF" checked> حذف رنج CF</label>
        <label class="chk-row"><input type="checkbox" id="cfgShodanAppend"> Append to file</label>
      </div>
    </div>
  </div>
</div>

<!-- ══ IMPORT ══ -->
<div id="page-import" class="page">
  <div class="phd">
    <div class="phd-l"><h2>وارد کردن کانفیگ</h2><p>لینک vless/vmess/trojan بده</p></div>
    <div class="phd-r" id="clearProxyBtn" style="display:none">
      <button class="btn btn-danger btn-sm" onclick="clearSavedProxy()">🗑 حذف کانفیگ</button>
    </div>
  </div>
  <div class="card">
    <div class="card-hd"><div>🔗 لینک پروکسی</div></div>
    <div class="card-bd">
      <div class="f-row"><label>لینک vless:// یا vmess:// یا trojan://</label><textarea id="linkInput" rows="3" placeholder="vless://uuid@domain:443?..."></textarea></div>
      <button class="btn btn-primary" onclick="parseLink()">🔍 Parse و ذخیره</button>
    </div>
  </div>
  <div id="parsedResult" style="display:none" class="card">
    <div class="card-hd"><div>✓ کانفیگ parse شد</div></div>
    <div class="card-bd">
      <div class="parsed-box" id="parsedBox"></div>
    </div>
  </div>
</div>

<!-- ══ SHODAN ══ -->
<div id="page-shodan" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Shodan Harvest</h2><p>جمع‌آوری IP از Shodan</p></div>
    <div class="phd-r">
      <button class="btn btn-primary" id="btnShodan" onclick="startShodan()">🔍 شروع</button>
    </div>
  </div>
  <div id="shodanAlert" class="alert alert-info" style="display:none"></div>
  <div class="card">
    <div class="card-hd"><div>⚙️ تنظیمات Shodan</div></div>
    <div class="card-bd">
      <div class="f-grid">
        <div class="f-row"><label>API Key</label><input type="password" id="shodanKey" placeholder="xxxxx"></div>
        <div class="f-row"><label>Pages (هر page=100 IP, 1 credit)</label><input type="number" id="shodanPages" value="1" min="1"></div>
      </div>
      <div class="f-row"><label>Query سفارشی (خالی = پیش‌فرض non-CF)</label><textarea id="shodanQuery" rows="2"></textarea></div>
      <div style="display:flex;gap:20px;flex-wrap:wrap">
        <label class="chk-row"><input type="checkbox" id="shodanExcludeCF" checked> حذف رنج CF</label>
        <label class="chk-row"><input type="checkbox" id="shodanAutoScan"> بعد از harvest اسکن کن</label>
      </div>
    </div>
  </div>
  <div id="shodanTicker" style="display:none;align-items:center;gap:10px;padding:10px;background:var(--bg3);border-radius:var(--rad-sm);margin-bottom:12px;font-family:'IBM Plex Mono',monospace;font-size:12px">
    <div class="spin"></div><span id="shodanTickerText">در حال جمع‌آوری...</span>
  </div>
  <div class="card" id="shodanResults" style="display:none">
    <div class="card-hd"><div>📋 IP های جمع‌آوری شده</div><span id="shodanCount" style="color:var(--g);font-family:'IBM Plex Mono',monospace">0</span></div>
    <div class="card-bd">
      <div class="ip-chips" id="shodanChips"></div>
      <div style="margin-top:10px;display:flex;gap:7px">
        <button class="btn btn-sm" onclick="copyAllShodan()">📋 کپی همه</button>
        <button class="btn btn-primary btn-sm" onclick="scanShodanIPs()">⚡ اسکن این IP ها</button>
      </div>
    </div>
  </div>
</div>

<!-- ══ TUI LOG ══ -->
<div id="page-tui" class="page">
  <div class="phd">
    <div class="phd-l"><h2>لاگ زنده</h2><p>همه رویدادهای اسکنر</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="clearTUI()">🗑 پاک</button>
      <button class="btn btn-sm" id="btnAS" onclick="toggleAS()">⬇ Auto</button>
    </div>
  </div>
  <div class="tui-wrap">
    <div class="tui-hd">
      <div class="tui-dots">
        <div class="tui-dot" style="background:#f06060"></div>
        <div class="tui-dot" style="background:#f5c842"></div>
        <div class="tui-dot" style="background:#2dd4a0"></div>
      </div>
      <span style="margin-right:8px;font-size:11px">piyazche — scanner log</span>
      <span id="tuiStatus" style="margin-right:auto;color:var(--dim);font-size:10px">idle</span>
    </div>
    <div class="tui-body" id="tuiBody">
      <div class="tui-line"><span class="tui-t">--:--:--</span><span class="tui-info">piyazche آماده<span class="cursor"></span></span></div>
    </div>
  </div>
</div>

</div><!-- /main -->
</div><!-- /app -->

<script>
// ══ STATE ══
let ws=null, p1Results=[], p2Results=[], shodanIPs=[], tuiAS=true;
let feedRows=[], maxFeedRows=80;
let currentTab='p2';

// ══ NAV ══
function nav(page,btn){
  document.querySelectorAll('.page').forEach(p=>p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(b=>b.classList.remove('active'));
  document.getElementById('page-'+page).classList.add('active');
  if(btn) btn.classList.add('active');
  else { const b=document.querySelector('[data-page="'+page+'"]'); if(b) b.classList.add('active'); }
  if(page==='results') refreshResults();
  if(page==='history') refreshHistory();
}

// ══ TABS ══
function switchTab(tab, btn){
  currentTab=tab;
  document.querySelectorAll('.tab').forEach(t=>t.classList.remove('active'));
  if(btn) btn.classList.add('active');
  document.getElementById('tab-p2').style.display=tab==='p2'?'':'none';
  document.getElementById('tab-p1').style.display=tab==='p1'?'':'none';
}

// ══ WS ══
function connectWS(){
  const proto=location.protocol==='https:'?'wss':'ws';
  ws=new WebSocket(proto+'://'+location.host+'/ws');
  ws.onmessage=e=>{try{handleWS(JSON.parse(e.data));}catch(err){}};
  ws.onclose=()=>setTimeout(connectWS,2000);
}

function handleWS(msg){
  const {type,payload}=msg;
  switch(type){
    case 'status': setStatus(payload.status,payload.phase); break;
    case 'progress': onProgress(payload); break;
    case 'live_ip': addFeedRow(payload.ip,'scan'); break;
    case 'ip_result':
      if(payload.success) addFeedRow(payload.ip+' · '+payload.latency+'ms','ok');
      // else addFeedRow(payload.ip,'fail'); // too noisy
      break;
    case 'tui': appendTUI(payload); break;
    case 'phase2_start':
      document.getElementById('phaseLabel').textContent='Phase 2';
      document.getElementById('progBar').classList.add('p2');
      document.getElementById('feedPhase').textContent='🔬 Phase 2';
      setStatus('scanning','phase2');
      break;
    case 'phase2_progress':{
      const r=payload;
      const pct=r.total>0?Math.round(r.done/r.total*100):0;
      document.getElementById('progBar').style.width=pct+'%';
      document.getElementById('pctLabel').textContent=pct+'%';
      const dl=r.dl&&r.dl!=='—'?' ↓'+r.dl:'';
      const details='Phase 2: '+r.done+'/'+r.total;
      document.getElementById('progDetail').textContent=details;
      const rowTxt=r.ip+' · '+Math.round(r.latency)+'ms · loss:'+r.loss.toFixed(0)+'%'+dl;
      addFeedRow(rowTxt, r.passed?'p2':'fail');
      break;
    }
    case 'phase2_done':
      p2Results=payload.results||[];
      renderP2();
      updatePassedChips();
      break;
    case 'scan_done':
      setStatus('done','');
      showBtns(false);
      document.getElementById('nbScan').style.display='none';
      clearFeedSpin();
      refreshResults();
      refreshHistory();
      break;
    case 'error': appendTUI({t:now(),l:'err',m:payload.message}); break;
    case 'shodan_status':
      document.getElementById('shodanTicker').style.display='flex';
      break;
    case 'shodan_done':
      shodanIPs=payload.ips||[];
      renderShodanResults(shodanIPs);
      document.getElementById('shodanTicker').style.display='none';
      document.getElementById('btnShodan').disabled=false;
      appendTUI({t:now(),l:'ok',m:'Shodan: '+shodanIPs.length+' IP'});
      break;
    case 'shodan_error':
      showShodanAlert(payload.message,'err');
      document.getElementById('shodanTicker').style.display='none';
      document.getElementById('btnShodan').disabled=false;
      break;
  }
}

function now(){return new Date().toLocaleTimeString('fa-IR');}

// ══ FEED ══
function addFeedRow(txt, type){
  const feed=document.getElementById('liveFeed');
  const row=document.createElement('div');
  const cls={ok:'live-row-ok',fail:'live-row-fail',scan:'live-row-scan',p2:'live-row-p2'}[type]||'live-row-scan';
  const icon={ok:'✓',fail:'✗',scan:'›',p2:'◈'}[type]||'›';
  const iconColor={ok:'var(--g)',fail:'var(--r)',scan:'var(--dim)',p2:'var(--p)'}[type];
  row.className='live-row '+cls;
  row.innerHTML='<span style="color:'+iconColor+';flex-shrink:0">'+icon+'</span><span>'+escH(txt)+'</span>';
  // prepend (newest on top due to column-reverse)
  feed.insertBefore(row, feed.firstChild);
  feedRows.push(row);
  if(feedRows.length>maxFeedRows){
    const old=feedRows.shift();
    if(old.parentNode) old.parentNode.removeChild(old);
  }
  // update count
  if(type==='ok'){
    const cur=parseInt(document.getElementById('feedCount').textContent)||0;
    document.getElementById('feedCount').textContent=(cur+1)+' موفق';
  }
}

function clearFeedSpin(){
  document.getElementById('feedSpin').style.display='none';
}

// ══ TUI ══
function appendTUI(entry){
  const body=document.getElementById('tuiBody');
  const line=document.createElement('div');
  line.className='tui-line';
  const cls={ok:'tui-ok',err:'tui-err',info:'tui-info',warn:'tui-warn',scan:'tui-scan',phase2:'tui-p2'}[entry.l]||'tui-info';
  line.innerHTML='<span class="tui-t">'+(entry.t||'')+'</span><span class="'+cls+'">'+escH(entry.m)+'</span>';
  body.appendChild(line);
  while(body.children.length>600) body.removeChild(body.firstChild);
  if(tuiAS) body.scrollTop=body.scrollHeight;
  document.getElementById('tuiStatus').textContent=entry.m.slice(0,50);
}
function escH(s){return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');}
function clearTUI(){document.getElementById('tuiBody').innerHTML='<div class="tui-line"><span class="tui-t">--:--:--</span><span class="tui-info">پاک شد<span class="cursor"></span></span></div>';}
function toggleAS(){tuiAS=!tuiAS;document.getElementById('btnAS').textContent='⬇ Auto '+(tuiAS?'ON':'OFF');}

// ══ STATUS ══
function setStatus(st,phase){
  const dot=document.getElementById('sDot');
  const txt=document.getElementById('sTxt');
  const ph=document.getElementById('sPhase');
  const labels={idle:'idle',scanning:'scanning',paused:'paused',done:'done'};
  txt.textContent=labels[st]||st;
  dot.className='dot';
  const cls={scanning:'dot-scan',paused:'dot-warn',done:'dot-done',idle:'dot-idle'}[st]||'dot-idle';
  dot.classList.add(cls);
  ph.textContent=phase?'· '+phase:'';
  // show/hide scan badge
  const nb=document.getElementById('nbScan');
  if(st==='scanning'){nb.style.display='';nb.textContent='LIVE';}
  else if(st==='done'){nb.style.display='';nb.textContent='DONE';}
}

// ══ PROGRESS ══
function onProgress(p){
  document.getElementById('stTotal').textContent=p.Total||'—';
  document.getElementById('stDone').textContent=p.Done||0;
  document.getElementById('stOk').textContent=p.Succeeded||0;
  document.getElementById('stFail').textContent=p.Failed||0;
  document.getElementById('stETA').textContent=p.ETA||'—';
  const pct=p.Total>0?Math.round(p.Done/p.Total*100):0;
  document.getElementById('progBar').style.width=pct+'%';
  document.getElementById('pctLabel').textContent=pct+'%';
  const rate=(p.Rate||0).toFixed(1);
  document.getElementById('progDetail').textContent=p.Done+'/'+(p.Total||'?');
  document.getElementById('rateLabel').textContent=rate+' ip/s';
  document.getElementById('tbProgress').textContent=p.Done+'/'+(p.Total||'?')+' ('+rate+' ip/s)';
  document.getElementById('feedSpin').style.display='';
  document.getElementById('feedPhase').textContent='⚡ Phase 1';
}

// ══ SCAN ══
async function startScan(){
  const ipInput=document.getElementById('ipInput').value.trim();
  const maxIPs=parseInt(document.getElementById('maxIPs').value)||0;
  const quickSettings=JSON.stringify({
    threads:parseInt(document.getElementById('qThreads').value)||200,
    timeout:parseInt(document.getElementById('qTimeout').value)||8,
    maxLatency:parseInt(document.getElementById('qMaxLat').value)||3500,
    stabilityRounds:parseInt(document.getElementById('qRounds').value)||3,
    sampleSize:parseInt(document.getElementById('sampleSize').value)||1,
  });
  const btn=document.getElementById('btnStart');
  btn.disabled=true;
  p1Results=[]; p2Results=[];
  feedRows=[];
  document.getElementById('liveFeed').innerHTML='<div class="live-row live-row-scan"><span style="color:var(--dim)">›</span><span>اسکن شروع شد...</span></div>';
  document.getElementById('feedCount').textContent='';
  document.getElementById('progBar').classList.remove('p2');
  const res=await fetch('/api/scan/start',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({quickSettings,ipRanges:ipInput,maxIPs})});
  const data=await res.json();
  btn.disabled=false;
  if(!data.ok){appendTUI({t:now(),l:'err',m:'خطا: '+data.error});return;}
  setStatus('scanning','phase1');
  showBtns(true);
  appendTUI({t:now(),l:'ok',m:'▶ اسکن شروع شد'});
}

async function stopScan(){
  await fetch('/api/scan/stop',{method:'POST'});
  setStatus('idle','');
  showBtns(false);
  clearFeedSpin();
}

async function pauseScan(){
  const res=await fetch('/api/scan/pause',{method:'POST'});
  const d=await res.json();
  const btn=document.getElementById('btnPause');
  if(d.message==='paused'){btn.textContent='▶ ادامه';setStatus('paused','');}
  else{btn.textContent='⏸ توقف';setStatus('scanning','');}
}

function showBtns(r){
  document.getElementById('btnStart').style.display=r?'none':'inline-flex';
  document.getElementById('btnPause').style.display=r?'inline-flex':'none';
  document.getElementById('btnStop').style.display=r?'inline-flex':'none';
}

// ══ IP PREVIEW ══
async function previewIPs(){
  const input=document.getElementById('ipInput').value.trim();
  const maxIPs=parseInt(document.getElementById('maxIPs').value)||0;
  if(!input){document.getElementById('ipPreview').style.display='none';return;}
  const res=await fetch('/api/ips/expand',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ipRanges:input,maxIPs})});
  const data=await res.json();
  const el=document.getElementById('ipPreview');
  el.style.display='block';
  el.textContent='مجموع: '+data.count+' IP'+(data.preview?' — مثال: '+data.preview.join(', '):'');
}

// ══ RESULTS ══
function refreshResults(){
  fetch('/api/results').then(r=>r.json()).then(data=>{
    p1Results=data.phase1||[];
    p2Results=data.phase2||[];
    renderP1();
    renderP2();
    updatePassedChips();
  });
}

function updatePassedChips(){
  const passed=(p2Results||[]).filter(r=>r.Passed);
  document.getElementById('resSummary').textContent=passed.length+' IP موفق از '+(p2Results||[]).length+' تست شده';
  document.getElementById('passedBadge').textContent=passed.length;
  document.getElementById('nbResults').textContent=passed.length;
  const chips=document.getElementById('ipChips');
  chips.innerHTML=passed.length
    ?passed.map(r=>'<div class="ip-chip" onclick="copyIP(\''+r.IP+'\')" title="کلیک=کپی IP">'+r.IP+'<span class="lat">'+Math.round(r.AvgLatencyMs)+'ms</span></div>').join('')
    :'<span style="color:var(--dim);font-size:12px">نتیجه‌ای نیست</span>';
}

function renderP2(){
  const tbody=document.getElementById('p2Tbody');
  document.getElementById('p2CountBadge').textContent=(p2Results||[]).length+' IP';
  if(!p2Results||!p2Results.length){
    tbody.innerHTML='<tr><td colspan="9" style="text-align:center;color:var(--dim);padding:28px">نتیجه Phase 2 نیست</td></tr>';
    return;
  }
  tbody.innerHTML=p2Results.map((r,i)=>{
    const sc=r.StabilityScore||0;
    const scc=sc>=75?'var(--g)':sc>=50?'var(--y)':'var(--r)';
    const lc=r.AvgLatencyMs<=500?'var(--g)':r.AvgLatencyMs<=1500?'var(--y)':'var(--r)';
    const badge=r.Passed?'<span class="badge bg">PASS</span>':'<span class="badge br" title="'+(r.FailReason||'')+'">FAIL</span>';
    const dl=r.DownloadMbps||0;
    const dlTxt=dl>0?dl.toFixed(1)+' M':'—';
    const dlc=dl<=0?'var(--dim)':dl>=5?'var(--g)':dl>=1?'var(--y)':'var(--r)';
    const plc=(r.PacketLossPct||0)<=5?'var(--g)':(r.PacketLossPct||0)<=20?'var(--y)':'var(--r)';
    const rowCls=r.Passed?'pass-row':'fail-row';
    return '<tr class="'+rowCls+'">'+
      '<td style="color:var(--dim)">'+(i+1)+'.</td>'+
      '<td style="color:var(--c);font-weight:500">'+r.IP+'</td>'+
      '<td style="color:'+scc+';font-weight:600">'+sc.toFixed(0)+'</td>'+
      '<td style="color:'+lc+'">'+Math.round(r.AvgLatencyMs||0)+'ms</td>'+
      '<td style="color:var(--dim)">'+(r.JitterMs>0?r.JitterMs.toFixed(0)+'ms':'—')+'</td>'+
      '<td style="color:'+plc+'">'+(r.PacketLossPct||0).toFixed(0)+'%</td>'+
      '<td style="color:'+dlc+'">'+dlTxt+'</td>'+
      '<td>'+badge+'</td>'+
      '<td style="display:flex;gap:3px;justify-content:center">'+
        '<button class="copy-btn" onclick="copyIP(\''+r.IP+'\')" title="کپی IP">📋</button>'+
        '<button class="copy-btn" onclick="copyWithIP(\''+r.IP+'\')" title="کپی لینک">🔗</button>'+
      '</td></tr>';
  }).join('');
}

function renderP1(){
  const tbody=document.getElementById('p1Tbody');
  document.getElementById('p1CountBadge').textContent=(p1Results||[]).length+' IP اسکن شده';
  const succ=(p1Results||[]).filter(r=>r.success||r.Success);
  if(!p1Results||!p1Results.length){
    tbody.innerHTML='<tr><td colspan="5" style="text-align:center;color:var(--dim);padding:28px">نتیجه Phase 1 نیست</td></tr>';
    return;
  }
  // فقط موفق‌ها رو نشون بده (fail خیلی زیاده)
  tbody.innerHTML=succ.map((r,i)=>{
    const ip=r.ip||r.IP||'';
    const lat=r.latency_ms||r.LatencyMs||0;
    const lc=lat<=500?'var(--g)':lat<=1500?'var(--y)':'var(--r)';
    return '<tr class="p1-row">'+
      '<td style="color:var(--dim)">'+(i+1)+'.</td>'+
      '<td style="color:var(--c);font-weight:500">'+ip+'</td>'+
      '<td style="color:'+lc+'">'+Math.round(lat)+'ms</td>'+
      '<td><span class="badge bg">OK</span></td>'+
      '<td><button class="copy-btn" onclick="copyIP(\''+ip+'\')" title="کپی IP">📋</button></td>'+
    '</tr>';
  }).join('') + (p1Results.length>succ.length?'<tr><td colspan="5" style="text-align:center;color:var(--dim);padding:10px;font-size:10px">'+(p1Results.length-succ.length)+' IP ناموفق پنهان شدند</td></tr>':'');
}

function copyAllPassed(){
  const passed=(p2Results||[]).filter(r=>r.Passed).map(r=>r.IP);
  if(!passed.length) return;
  navigator.clipboard.writeText(passed.join('\n')).then(()=>appendTUI({t:now(),l:'ok',m:'📋 '+passed.length+' IP کپی شد'}));
}

// ══ HISTORY ══
function refreshHistory(){
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const el=document.getElementById('histList');
    if(!sessions||!sessions.length){el.innerHTML='<p style="color:var(--dim)">هنوز اسکنی انجام نشده</p>';return;}
    el.innerHTML=sessions.map(s=>
      '<div class="hist-item" onclick="showSession(\''+s.id+'\')">'+
        '<span style="color:var(--c)">'+new Date(s.startedAt).toLocaleString('fa-IR')+'</span>'+
        '<span style="color:var(--dim)">'+s.duration+'</span>'+
        '<span style="color:var(--dim)">'+s.totalIPs+' IP</span>'+
        '<span class="badge bg">'+s.passed+' passed</span>'+
      '</div>'
    ).join('');
  });
}
function showSession(id){
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const s=sessions.find(x=>x.id===id);if(!s)return;
    p2Results=s.results||[];p1Results=[];
    renderP2();renderP1();updatePassedChips();nav('results');
  });
}

// ══ CONFIG ══
document.getElementById('cfgSpeed').addEventListener('change',function(){
  document.getElementById('speedURLs').style.display=this.checked?'':'none';
});

function saveConfig(){
  const scanCfg={
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
  // sync quick
  document.getElementById('qThreads').value=scanCfg.scan.threads;
  document.getElementById('qTimeout').value=scanCfg.scan.timeout;
  document.getElementById('qMaxLat').value=scanCfg.scan.maxLatency;
  document.getElementById('qRounds').value=scanCfg.scan.stabilityRounds;
  document.getElementById('sampleSize').value=scanCfg.scan.sampleSize;
  fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({scanConfig:JSON.stringify(scanCfg)})}).then(()=>{
    appendTUI({t:now(),l:'ok',m:'✓ تنظیمات ذخیره شد'});
    updateConfigSummary(scanCfg.scan);
    nav('scan');
  });
}

function updateConfigSummary(s){
  const el=document.getElementById('configSummary');
  el.innerHTML=[
    '<span class="cfg-item"><b>threads:</b> '+s.threads+'</span>',
    '<span class="cfg-item"><b>timeout:</b> '+s.timeout+'s</span>',
    '<span class="cfg-item"><b>maxLat:</b> '+s.maxLatency+'ms</span>',
    '<span class="cfg-item"><b>rounds:</b> '+s.stabilityRounds+'</span>',
    s.speedTest?'<span class="cfg-item" style="color:var(--g)"><b>speed:</b> ON</span>':'',
  ].filter(Boolean).join('');
}

// ══ IMPORT ══
async function parseLink(){
  const input=document.getElementById('linkInput').value.trim();
  if(!input) return;
  const res=await fetch('/api/config/parse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})});
  const data=await res.json();
  if(!data.ok){appendTUI({t:now(),l:'err',m:'خطا: '+data.error});return;}
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:data.config})});
  const p=data.parsed;
  updateProxyChip(p.address,p.method,p.type);
  document.getElementById('parsedBox').innerHTML=
    '<span class="k">uuid: </span><span class="v">'+maskUUID(p.uuid)+'</span><br>'+
    '<span class="k">address: </span><span class="v">'+p.address+'</span><br>'+
    '<span class="k">port: </span><span class="v">'+p.port+'</span><br>'+
    '<span class="k">type: </span><span class="v">'+p.type+'</span><br>'+
    '<span class="k">method: </span><span class="v">'+p.method+'</span>'+
    (p.sni?'<br><span class="k">sni: </span><span class="v">'+p.sni+'</span>':'')+
    (p.path?'<br><span class="k">path: </span><span class="v">'+p.path+'</span>':'')+
    (p.fp?'<br><span class="k">fp: </span><span class="v">'+p.fp+'</span>':'');
  document.getElementById('parsedResult').style.display='block';
  appendTUI({t:now(),l:'ok',m:'✓ کانفیگ: '+p.address+' ('+p.method+'/'+p.type+')'});
}

function updateProxyChip(addr,method,type){
  const chip=document.getElementById('proxyChip');
  document.getElementById('proxyChipTxt').textContent=addr+' · '+method+'/'+type;
  chip.style.display='inline-flex';
  document.getElementById('clearProxyBtn').style.display='';
  document.getElementById('configSummary').innerHTML='<span class="cfg-item" style="color:var(--g)">🔒 '+addr+'</span>';
}

async function clearSavedProxy(){
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:''})});
  document.getElementById('proxyChip').style.display='none';
  document.getElementById('clearProxyBtn').style.display='none';
  document.getElementById('parsedResult').style.display='none';
  document.getElementById('configSummary').innerHTML='<span class="cfg-item">پیش‌فرض — لینک وارد نشده</span>';
  appendTUI({t:now(),l:'warn',m:'کانفیگ proxy حذف شد'});
}

function maskUUID(u){return!u||u.length<8?u:u.slice(0,8)+'••••••••';}

// ══ SHODAN ══
async function startShodan(){
  const key=document.getElementById('shodanKey').value.trim();
  if(!key){showShodanAlert('API Key الزامی است','err');return;}
  document.getElementById('btnShodan').disabled=true;
  document.getElementById('shodanTicker').style.display='flex';
  document.getElementById('shodanResults').style.display='none';
  document.getElementById('shodanAlert').style.display='none';
  const res=await fetch('/api/shodan/harvest',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({apiKey:key,query:document.getElementById('shodanQuery').value.trim(),pages:parseInt(document.getElementById('shodanPages').value)||1,excludeCF:document.getElementById('shodanExcludeCF').checked,autoScan:document.getElementById('shodanAutoScan').checked})});
  const data=await res.json();
  if(!data.ok){document.getElementById('shodanTicker').style.display='none';document.getElementById('btnShodan').disabled=false;showShodanAlert(data.error,'err');}
}
function renderShodanResults(ips){
  document.getElementById('shodanCount').textContent=ips.length;
  document.getElementById('shodanResults').style.display='block';
  const chips=document.getElementById('shodanChips');
  chips.innerHTML=ips.slice(0,200).map(ip=>'<div class="ip-chip" onclick="copyIP(\''+ip+'\')">'+ip+'</div>').join('');
  if(ips.length>200) chips.innerHTML+='<span style="color:var(--dim);font-size:11px"> +'+( ips.length-200)+' IP دیگر</span>';
}
function copyAllShodan(){navigator.clipboard.writeText(shodanIPs.join('\n'));appendTUI({t:now(),l:'ok',m:'کپی شد: '+shodanIPs.length+' IP'});}
function scanShodanIPs(){if(!shodanIPs.length)return;document.getElementById('ipInput').value=shodanIPs.join('\n');nav('scan');}
function showShodanAlert(msg,type){const el=document.getElementById('shodanAlert');el.className='alert alert-'+(type==='err'?'err':type==='warn'?'warn':'info');el.textContent=msg;el.style.display='block';}

// ══ COPY / EXPORT ══
function exportResults(f){window.location.href='/api/results/export?format='+f;}
function copyIP(ip){
  navigator.clipboard.writeText(ip).then(()=>appendTUI({t:now(),l:'ok',m:'📋 '+ip})).catch(()=>{
    const el=document.createElement('textarea');el.value=ip;document.body.appendChild(el);el.select();document.execCommand('copy');document.body.removeChild(el);
    appendTUI({t:now(),l:'ok',m:'📋 '+ip});
  });
}
function copyWithIP(newIP){
  const rawLink=document.getElementById('linkInput').value.trim();
  if(!rawLink){copyIP(newIP);return;}
  try{
    const updated=rawLink.replace(/(@)([^:@\/?#\[\]]+)(:\d+)/,'$1'+newIP+'$3');
    navigator.clipboard.writeText(updated).then(()=>appendTUI({t:now(),l:'ok',m:'🔗 لینک با '+newIP+' کپی شد'}));
  }catch(e){copyIP(newIP);}
}

// ══ SETTINGS LOAD ══
function loadSavedSettings(){
  fetch('/api/config/load').then(r=>r.json()).then(d=>{
    if(d.hasProxy){
      try{
        const pc=JSON.parse(d.proxyConfig);
        const addr=pc.proxy?.address||'';
        if(addr) updateProxyChip(addr,pc.proxy?.method||'tls',pc.proxy?.type||'ws');
      }catch(e){}
    }
    if(d.scanConfig){
      try{
        const sc=JSON.parse(d.scanConfig);
        const s=sc.scan||{},f=sc.fragment||{},x=sc.xray||{},sh=sc.shodan||{};
        const sv=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        const sc2=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.checked=!!v;};
        const ss=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        if(s.threads) sv('cfgThreads',s.threads);
        if(s.timeout) sv('cfgTimeout',s.timeout);
        if(s.maxLatency) sv('cfgMaxLat',s.maxLatency);
        if(s.retries!=null) sv('cfgRetries',s.retries);
        if(s.maxIPs!=null) sv('cfgMaxIPs',s.maxIPs);
        if(s.sampleSize) sv('cfgSampleSize',s.sampleSize);
        if(s.testUrl) sv('cfgTestURL',s.testUrl);
        if(s.shuffle!=null) sc2('cfgShuffle',s.shuffle);
        if(s.stabilityRounds!=null) sv('cfgRounds',s.stabilityRounds);
        if(s.stabilityInterval) sv('cfgInterval',s.stabilityInterval);
        if(s.packetLossCount) sv('cfgPLCount',s.packetLossCount);
        if(s.maxPacketLossPct!=null) sv('cfgMaxPL',s.maxPacketLossPct);
        if(s.minDownloadMbps!=null) sv('cfgMinDL',s.minDownloadMbps);
        if(s.minUploadMbps!=null) sv('cfgMinUL',s.minUploadMbps);
        if(s.speedTest!=null){sc2('cfgSpeed',s.speedTest);document.getElementById('speedURLs').style.display=s.speedTest?'':'none';}
        if(s.jitterTest!=null) sc2('cfgJitter',s.jitterTest);
        if(s.downloadUrl) sv('cfgDLURL',s.downloadUrl);
        if(s.uploadUrl) sv('cfgULURL',s.uploadUrl);
        if(f.mode) ss('cfgFragMode',f.mode);
        if(f.packets) sv('cfgFragPkts',f.packets);
        if(f.manual?.length) sv('cfgFragLen',f.manual.length);
        if(f.manual?.interval) sv('cfgFragInt',f.manual.interval);
        if(x.logLevel) ss('cfgXrayLog',x.logLevel);
        if(x.mux?.concurrency!=null) sv('cfgMuxConc',x.mux.concurrency);
        if(x.mux?.enabled!=null) sc2('cfgMuxEnabled',x.mux.enabled);
        if(sh.mode) ss('cfgShodanMode',sh.mode);
        if(sh.apiKey) sv('cfgShodanKey',sh.apiKey);
        if(sh.pages) sv('cfgShodanPages',sh.pages);
        if(sh.saveHarvestedIPs) sv('cfgShodanSave',sh.saveHarvestedIPs);
        if(sh.useDefaultQuery!=null) sc2('cfgShodanUseDefault',sh.useDefaultQuery);
        if(sh.excludeCFRanges!=null) sc2('cfgShodanExcludeCF',sh.excludeCFRanges);
        if(sh.appendToExisting!=null) sc2('cfgShodanAppend',sh.appendToExisting);
        // sync quick
        if(s.threads) sv('qThreads',s.threads);
        if(s.timeout) sv('qTimeout',s.timeout);
        if(s.maxLatency) sv('qMaxLat',s.maxLatency);
        if(s.stabilityRounds!=null) sv('qRounds',s.stabilityRounds);
        if(s.sampleSize) sv('sampleSize',s.sampleSize);
        updateConfigSummary(s);
      }catch(e){console.warn('load settings err',e);}
    }
    // TUI history
    fetch('/api/tui/stream').then(r=>r.json()).then(data=>{
      if(data.lines) data.lines.forEach(l=>{try{appendTUI(JSON.parse(l));}catch(e){}});
    }).catch(()=>{});
  });
}

// ══ RESULTS API needs phase1 ══
// Override refreshResults to fetch phase1 too
const _origRefresh=refreshResults;

// ══ INIT ══
connectWS();
fetch('/api/status').then(r=>r.json()).then(d=>setStatus(d.status||'idle',d.phase||''));
loadSavedSettings();
</script>
</body>
</html>
`
