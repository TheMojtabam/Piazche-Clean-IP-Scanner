package webui

const indexHTMLContent = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Piyazche Scanner</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&family=Space+Mono:wght@400;700&family=Bebas+Neue&family=Courier+Prime:wght@400;700&display=swap" rel="stylesheet">
<style>
/* ══════════════════════════════════════════════════
   THEME VARIABLES
══════════════════════════════════════════════════ */
:root{
  /* Neon Noir (default) */
  --bg:#030408;--bg2:#060810;--bg3:#0a0d18;--bg4:#0e1220;
  --bd:#0f1422;--bd2:#161c2e;--bd3:#1e263a;
  --tx:#d0d8f0;--tx2:#6070a0;--dim:#2a3050;
  --g:#00ff88;--gd:rgba(0,255,136,.06);--g2:#00cc66;
  --c:#00c8ff;--cd:rgba(0,200,255,.06);--c2:#0099cc;
  --y:#ffe000;--yd:rgba(255,224,0,.06);
  --r:#ff3366;--rd:rgba(255,51,102,.06);
  --p:#b400ff;--pd:rgba(180,0,255,.06);
  --o:#ff6600;
  --rad:8px;--rad-sm:5px;--rad-xs:3px;
  --font-head:'Space Grotesk',sans-serif;
  --font-mono:'Space Mono',monospace;
  --glow-g:0 0 20px rgba(0,255,136,.3);
  --glow-c:0 0 20px rgba(0,200,255,.3);
  --glow-r:0 0 20px rgba(255,51,102,.3);
}
[data-theme="newspaper"]{
  --bg:#0e0d0b;--bg2:#141310;--bg3:#1a1815;--bg4:#201e1a;
  --bd:#242018;--bd2:#2e2a22;--bd3:#38342c;
  --tx:#d8d0b8;--tx2:#705e40;--dim:#3a3028;
  --g:#88cc00;--gd:rgba(136,204,0,.06);--g2:#668800;
  --c:#c8a040;--cd:rgba(200,160,64,.06);--c2:#a07820;
  --y:#e8c830;--yd:rgba(232,200,48,.06);
  --r:#cc3300;--rd:rgba(204,51,0,.06);
  --p:#885500;--pd:rgba(136,85,0,.06);
  --glow-g:0 0 15px rgba(136,204,0,.2);
  --glow-c:0 0 15px rgba(200,160,64,.2);
  --glow-r:0 0 15px rgba(204,51,0,.2);
}
*{margin:0;padding:0;box-sizing:border-box}
html{height:100%}
body{font-family:var(--font-head);background:var(--bg);color:var(--tx);height:100%;font-size:14px;line-height:1.5;overflow:hidden}
.app{display:grid;grid-template-columns:200px 1fr;grid-template-rows:56px 1fr;height:100vh}

/* ══ TOPBAR ══ */
.topbar{
  grid-column:1/-1;
  background:var(--bg2);
  border-bottom:1px solid var(--bd2);
  display:flex;align-items:center;
  padding:0 18px;gap:14px;
  position:relative;z-index:100;
}
.logo{
  font-family:'Bebas Neue',sans-serif;
  font-size:26px;letter-spacing:3px;
  user-select:none;
  background:linear-gradient(90deg,var(--c),var(--g));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
  text-shadow:none;
}
[data-theme="newspaper"] .logo{
  font-family:'Courier Prime',monospace;
  font-size:20px;letter-spacing:2px;font-weight:700;
  background:none;-webkit-text-fill-color:var(--c);
}
.status-pill{
  display:flex;align-items:center;gap:6px;
  padding:4px 12px;border-radius:20px;
  font-size:11px;background:var(--bg3);
  border:1px solid var(--bd2);
  font-family:var(--font-mono);
}
.dot{width:7px;height:7px;border-radius:50%;flex-shrink:0;transition:all .3s}
.dot-idle{background:var(--dim)}
.dot-scan{background:var(--g);box-shadow:var(--glow-g);animation:pulse 1.2s ease-in-out infinite}
.dot-warn{background:var(--y)}
.dot-done{background:var(--c)}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:.3;transform:scale(.7)}}
.tb-right{margin-left:auto;display:flex;align-items:center;gap:10px}
.proxy-chip{
  display:inline-flex;align-items:center;gap:5px;
  padding:3px 10px;border-radius:5px;
  font-size:11px;font-family:var(--font-mono);
  background:var(--gd);border:1px solid rgba(0,255,136,.3);
  color:var(--g);cursor:pointer;
}
.proxy-chip:hover{background:rgba(0,255,136,.1)}
.theme-btn{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx2);padding:4px 12px;border-radius:4px;
  font-size:11px;cursor:pointer;font-family:var(--font-mono);
  transition:all .15s;
}
.theme-btn:hover{border-color:var(--bd3);color:var(--tx)}

/* ══ SIDEBAR ══ */
.sidebar{
  background:var(--bg2);border-right:1px solid var(--bd);
  display:flex;flex-direction:column;overflow-y:auto;
}
.nav-group{
  padding:14px 14px 4px;
  font-size:9px;letter-spacing:2.5px;text-transform:uppercase;
  color:var(--dim);font-family:var(--font-mono);
}
.nav-item{
  display:flex;align-items:center;gap:9px;
  padding:9px 16px;cursor:pointer;
  transition:all .12s;color:var(--tx2);
  font-size:13px;border:none;background:none;
  width:100%;text-align:left;
  border-left:2px solid transparent;
  font-family:var(--font-head);font-weight:500;
}
.nav-item:hover{background:var(--bg3);color:var(--tx)}
.nav-item.active{
  background:linear-gradient(90deg,var(--cd),transparent);
  color:var(--c);border-left-color:var(--c);
}
.nav-icon{font-size:14px;min-width:18px;text-align:center}
.nav-badge{
  margin-left:auto;background:var(--bd2);
  color:var(--dim);font-size:9px;padding:1px 6px;
  border-radius:8px;font-family:var(--font-mono);
}
.nav-badge.live{background:var(--gd);color:var(--g)}

/* ══ MAIN ══ */
.main{overflow-y:auto;overflow-x:hidden;background:var(--bg)}
.main::-webkit-scrollbar{width:4px}
.main::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}
.page{display:none;padding:22px 24px 32px;min-height:100%}
.page.active{display:block}

/* ══ PAGE HEADER ══ */
.phd{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:18px;gap:12px}
.phd-l h2{
  font-family:var(--font-head);font-size:22px;font-weight:700;
  letter-spacing:-.5px;
}
[data-theme="newspaper"] .phd-l h2{font-family:'Courier Prime',monospace;font-size:20px;text-transform:uppercase;letter-spacing:2px}
.phd-l p{font-size:11px;color:var(--dim);margin-top:3px}
.phd-r{display:flex;gap:7px;align-items:center;flex-shrink:0}

/* ══ STATS ROW ══ */
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:10px;margin-bottom:16px}
.stat-card{
  background:var(--bg2);border:1px solid var(--bd);
  border-radius:var(--rad-sm);padding:14px 16px;
  position:relative;overflow:hidden;
  transition:border-color .2s;
}
.stat-v{
  font-family:var(--font-mono);font-weight:700;
  line-height:1;transition:color .3s;
  font-size:32px; /* BIG numbers */
  letter-spacing:-1px;
}
[data-theme="newspaper"] .stat-v{font-family:'Courier Prime',monospace;font-size:28px}
.stat-l{font-size:9px;color:var(--dim);margin-top:6px;letter-spacing:2px;text-transform:uppercase}

/* ══ PROGRESS ══ */
.prog-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:14px}
.prog-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 16px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono)}
.prog-hd-l{display:flex;align-items:center;gap:7px}
.prog-bd{padding:14px 16px}
.prog-meta{display:flex;justify-content:space-between;font-size:11px;color:var(--dim);margin-bottom:6px;font-family:var(--font-mono)}
.prog-pct{color:var(--c);font-weight:700;font-size:14px}
.prog-wrap{background:var(--bg);border-radius:3px;height:6px;overflow:hidden;margin-bottom:10px}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:3px;transition:width .5s cubic-bezier(.4,0,.2,1);width:0%;box-shadow:var(--glow-c)}
.prog-bar.p2{background:linear-gradient(90deg,var(--p),var(--c))}

/* ══ LIVE FEED ══ */
.live-feed{background:var(--bg);border:1px solid var(--bd);border-radius:var(--rad-sm);overflow:hidden}
.live-feed-hd{padding:7px 12px;border-bottom:1px solid var(--bd);display:flex;align-items:center;gap:8px;font-size:10px;color:var(--dim);font-family:var(--font-mono)}
.live-feed-body{height:130px;overflow-y:auto;padding:8px 12px;display:flex;flex-direction:column-reverse;gap:2px}
.live-feed-body::-webkit-scrollbar{width:3px}
.live-feed-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.live-row{display:flex;align-items:center;gap:8px;font-family:var(--font-mono);font-size:11px;padding:1px 0;animation:fadeIn .2s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(-3px)}to{opacity:1;transform:none}}
.live-row-ok{color:var(--g)}.live-row-fail{color:var(--dim)}
.live-row-scan{color:var(--tx2)}.live-row-p2{color:var(--p)}
.live-ip{color:var(--c);min-width:128px;font-weight:700}
.live-lat{color:var(--y);min-width:62px}
.live-tag{font-size:9px;padding:1px 5px;border-radius:3px;flex-shrink:0}
.tag-ok{background:var(--gd);color:var(--g)}.tag-fail{background:var(--rd);color:var(--r)}.tag-p2{background:var(--pd);color:var(--p)}
.spin{width:12px;height:12px;border:1.5px solid var(--bd3);border-top-color:var(--c);border-radius:50%;animation:spin .6s linear infinite;flex-shrink:0}
@keyframes spin{to{transform:rotate(360deg)}}

/* ══ CARDS ══ */
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:12px}
.card-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 14px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono);letter-spacing:.5px}
.card-bd{padding:14px}

/* ══ TABLE ══ */
.tbl-wrap{overflow-x:auto}
.tbl{width:100%;border-collapse:collapse;font-size:12px;font-family:var(--font-mono)}
.tbl th{padding:8px 10px;text-align:left;color:var(--dim);font-weight:500;border-bottom:1px solid var(--bd);background:var(--bg3);font-size:9px;letter-spacing:1px;white-space:nowrap;position:sticky;top:0;text-transform:uppercase}
.tbl td{padding:7px 10px;border-bottom:1px solid var(--bd);vertical-align:middle;white-space:nowrap}
.tbl tbody tr:hover td{background:rgba(255,255,255,.015)}
.tbl tr.pass-row td:first-child{border-left:2px solid var(--g)}
.tbl tr.fail-row td:first-child{border-left:2px solid transparent}

/* ══ BADGES ══ */
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:3px;font-size:9px;font-family:var(--font-mono);font-weight:700;letter-spacing:1px;text-transform:uppercase}
.bg{background:var(--gd);color:var(--g);border:1px solid rgba(0,255,136,.2)}
.bc{background:var(--cd);color:var(--c);border:1px solid rgba(0,200,255,.2)}
.by{background:var(--yd);color:var(--y);border:1px solid rgba(255,224,0,.2)}
.br{background:var(--rd);color:var(--r);border:1px solid rgba(255,51,102,.2)}
.bp{background:var(--pd);color:var(--p);border:1px solid rgba(180,0,255,.2)}

/* ══ BUTTONS ══ */
.btn{display:inline-flex;align-items:center;gap:6px;padding:7px 16px;border-radius:var(--rad-sm);border:1px solid var(--bd2);background:var(--bg3);color:var(--tx);cursor:pointer;font-size:12px;font-family:var(--font-head);transition:all .12s;white-space:nowrap;font-weight:500}
.btn:hover{background:var(--bd2);border-color:var(--bd3)}.btn:active{transform:scale(.97)}.btn:disabled{opacity:.35;cursor:not-allowed;pointer-events:none}
.btn-primary{background:var(--cd);border-color:rgba(0,200,255,.35);color:var(--c)}.btn-primary:hover{background:var(--c2);color:#000;border-color:var(--c2)}
.btn-success{background:var(--gd);border-color:rgba(0,255,136,.3);color:var(--g)}.btn-success:hover{background:var(--g2);color:#000}
.btn-danger{background:var(--rd);border-color:rgba(255,51,102,.3);color:var(--r)}.btn-danger:hover{background:var(--r);color:#fff}
.btn-warn{background:var(--yd);border-color:rgba(255,224,0,.3);color:var(--y)}.btn-warn:hover{background:var(--y);color:#000}
.btn-sm{padding:5px 12px;font-size:11px}.btn-xs{padding:2px 8px;font-size:10px}

/* ══ FORMS ══ */
textarea,input[type=text],input[type=number],input[type=password],select{
  background:var(--bg3);border:1px solid var(--bd2);color:var(--tx);
  border-radius:var(--rad-sm);padding:7px 11px;
  font-size:12px;font-family:var(--font-mono);width:100%;
  outline:none;transition:border-color .15s;
}
textarea:focus,input:focus,select:focus{border-color:rgba(0,200,255,.5);box-shadow:0 0 0 2px rgba(0,200,255,.07)}
label{display:block;font-size:11px;color:var(--dim);margin-bottom:4px;font-family:var(--font-head);font-weight:500}
.f-row{margin-bottom:11px}.f-grid{display:grid;grid-template-columns:1fr 1fr;gap:11px}.f-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:11px}
.f-sep{height:1px;background:var(--bd);margin:14px 0}
.chk-row{display:flex;align-items:center;gap:7px;cursor:pointer;font-size:12px;color:var(--tx2);font-family:var(--font-head)}
.chk-row input{width:auto;cursor:pointer;accent-color:var(--c)}
.parsed-box{background:var(--bg3);border:1px solid rgba(0,255,136,.2);border-radius:var(--rad-sm);padding:12px;font-family:var(--font-mono);font-size:11px;color:var(--g);line-height:1.9}
.parsed-box .k{color:var(--dim)}.parsed-box .v{color:var(--c)}
.cfg-chip{display:flex;flex-wrap:wrap;gap:6px;padding:8px 10px;background:var(--bg3);border:1px solid var(--bd2);border-radius:var(--rad-sm);font-size:10px;font-family:var(--font-mono)}
.cfg-item{color:var(--dim)}.cfg-item b{color:var(--tx2)}

/* ══ IP CHIPS ══ */
.ip-chips{display:flex;flex-wrap:wrap;gap:5px;padding:4px 0}
.ip-chip{
  background:var(--cd);border:1px solid rgba(0,200,255,.2);
  border-radius:4px;padding:4px 10px;
  font-family:var(--font-mono);font-size:11px;color:var(--c);
  cursor:pointer;display:flex;align-items:center;gap:5px;transition:all .12s;
}
.ip-chip:hover{background:var(--c);color:#000;border-color:var(--c)}
.ip-chip .lat{font-size:9px;opacity:.5}

/* ══ HISTORY ══ */
.hist-item{background:var(--bg3);border:1px solid var(--bd);border-radius:var(--rad-sm);padding:12px 16px;display:flex;align-items:center;gap:12px;margin-bottom:8px;cursor:pointer;transition:all .12s;font-family:var(--font-mono);font-size:11px}
.hist-item:hover{border-color:var(--bd3);background:var(--bg4);transform:translateX(2px)}
.hist-n{font-size:22px;font-weight:700;color:var(--g);min-width:50px;font-family:var(--font-mono)}
.hist-info{flex:1}.hist-date{color:var(--dim);font-size:10px}

/* ══ TUI ══ */
.tui-wrap{background:#020304;border:1px solid var(--bd2);border-radius:var(--rad-sm);overflow:hidden;font-family:var(--font-mono);font-size:11.5px}
.tui-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:7px 12px;display:flex;align-items:center;gap:7px;font-size:10px;color:var(--dim)}
.tui-dots{display:flex;gap:4px}.tui-dot{width:9px;height:9px;border-radius:50%}
.tui-body{padding:10px 14px;height:420px;overflow-y:auto;line-height:1.9}
.tui-body::-webkit-scrollbar{width:3px}.tui-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.tui-line{display:flex;gap:8px}
.tui-t{color:var(--dim);flex-shrink:0;user-select:none;font-size:10px}
.tui-ok{color:var(--g)}.tui-err{color:var(--r)}.tui-info{color:var(--c)}.tui-warn{color:var(--y)}.tui-scan{color:var(--tx2)}.tui-p2{color:var(--p)}
.cursor{display:inline-block;width:7px;height:13px;background:var(--c);animation:blink 1s step-end infinite;vertical-align:middle}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}

/* ══ ALERTS ══ */
.alert{border-radius:var(--rad-sm);padding:9px 14px;font-size:12px;margin-bottom:11px;border-left:3px solid;display:flex;align-items:flex-start;gap:8px}
.alert-info{background:var(--cd);border-color:var(--c);color:var(--c)}
.alert-warn{background:var(--yd);border-color:var(--y);color:var(--y)}
.alert-err{background:var(--rd);border-color:var(--r);color:var(--r)}

/* ══ TABS ══ */
.tab-bar{display:flex;gap:2px;border-bottom:1px solid var(--bd);margin-bottom:14px}
.tab{padding:8px 18px;font-size:12px;font-family:var(--font-head);background:none;border:none;color:var(--tx2);cursor:pointer;border-bottom:2px solid transparent;transition:all .12s;margin-bottom:-1px;font-weight:500}
.tab:hover{color:var(--tx)}.tab.active{color:var(--c);border-bottom-color:var(--c);background:var(--cd)}

/* ══ COPY BTN ══ */
.copy-btn{background:none;border:none;cursor:pointer;color:var(--dim);padding:2px 5px;border-radius:3px;font-size:11px;transition:all .12s}
.copy-btn:hover{color:var(--c);background:var(--cd)}

/* ══ SESSION BANNER ══ */
.session-banner{background:var(--rd);color:var(--r);border:1px solid rgba(255,51,102,.3);padding:8px 14px;border-radius:var(--rad-sm);margin-bottom:12px;display:flex;justify-content:space-between;align-items:center;font-size:11px;font-family:var(--font-mono)}

/* ══ NEWSPAPER extras ══ */
[data-theme="newspaper"] .live-feed{border:1px solid var(--bd3)}
[data-theme="newspaper"] .stat-card{border-radius:0;border-width:1px}
[data-theme="newspaper"] .card{border-radius:0}
[data-theme="newspaper"] .tbl th{letter-spacing:2px;font-size:8px}

/* ══ SCROLLBAR ══ */
::-webkit-scrollbar{width:4px;height:4px}
::-webkit-scrollbar-thumb{background:var(--bd3);border-radius:2px}

/* ══ RESPONSIVE ══ */
@media(max-width:768px){.app{grid-template-columns:1fr}.sidebar{display:none}.stats-row{grid-template-columns:repeat(2,1fr)}.f-grid,.f-grid-3{grid-template-columns:1fr}}
</style>
</head>
<body>
<div class="app">

<!-- TOPBAR -->
<div class="topbar">
  <div class="logo">PIYAZCHE</div>
  <div class="status-pill">
    <div class="dot dot-idle" id="sDot"></div>
    <span id="sTxt" style="font-family:var(--font-mono);font-size:11px">idle</span>
    <span id="sPhase" style="color:var(--dim);font-size:10px"></span>
  </div>
  <div id="proxyChip" style="display:none" class="proxy-chip" onclick="nav('import')" title="Active config — click to change">
    <span>⬡</span><span id="proxyChipTxt"></span>
  </div>
  <div class="tb-right">
    <span id="tbProgress" style="font-family:var(--font-mono);font-size:11px;color:var(--dim)"></span>
    <button class="theme-btn" onclick="toggleTheme()" id="themeBtn">◈ NEWSPAPER</button>
  </div>
</div>

<!-- SIDEBAR -->
<div class="sidebar">
  <div class="nav-group">Scanner</div>
  <button class="nav-item active" data-page="scan" onclick="nav('scan',this)">
    <span class="nav-icon">⚡</span>Scan
    <span class="nav-badge live" id="nbScan" style="display:none">LIVE</span>
  </button>
  <button class="nav-item" data-page="results" onclick="nav('results',this)">
    <span class="nav-icon">◈</span>Results
    <span class="nav-badge" id="nbResults">0</span>
  </button>
  <button class="nav-item" data-page="history" onclick="nav('history',this)">
    <span class="nav-icon">◷</span>History
  </button>
  <div class="nav-group">Config</div>
  <button class="nav-item" data-page="config" onclick="nav('config',this)">
    <span class="nav-icon">⚙</span>Settings
  </button>
  <button class="nav-item" data-page="import" onclick="nav('import',this)">
    <span class="nav-icon">⬡</span>Import Link
  </button>
  <div class="nav-group">Tools</div>
  <button class="nav-item" data-page="tui" onclick="nav('tui',this)">
    <span class="nav-icon">▸</span>Live Log
  </button>
</div>

<!-- MAIN -->
<div class="main">

<!-- ══ SCAN PAGE ══ -->
<div id="page-scan" class="page active">
  <div class="phd">
    <div class="phd-l"><h2>Scan</h2><p id="configSummary" style="font-family:var(--font-mono);font-size:10px;color:var(--dim)">No config — import a proxy link first</p></div>
    <div class="phd-r">
      <button class="btn btn-success" id="btnStart" onclick="startScan()">▶ Start</button>
      <button class="btn btn-danger btn-sm" id="btnStop" onclick="stopScan()" style="display:none">■ Stop</button>
      <button class="btn btn-warn btn-sm" id="btnPause" onclick="pauseScan()" style="display:none">⏸</button>
    </div>
  </div>

  <!-- Stats -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-v" id="stTotal" style="color:var(--tx2)">—</div>
      <div class="stat-l">Total IPs</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stDone" style="color:var(--tx2)">—</div>
      <div class="stat-l">Scanned</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stPass" style="color:var(--g)">0</div>
      <div class="stat-l">Passed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stFail" style="color:var(--dim)">0</div>
      <div class="stat-l">Failed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stETA" style="color:var(--c);font-size:22px">—</div>
      <div class="stat-l">ETA</div>
    </div>
  </div>

  <!-- Progress -->
  <div class="prog-card">
    <div class="prog-hd">
      <div class="prog-hd-l">
        <div class="dot dot-idle" id="pDot"></div>
        <span id="progLabel">Ready</span>
      </div>
      <span id="progRate" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="prog-bd">
      <div class="prog-meta">
        <span id="progTxt">0 / 0</span>
        <span class="prog-pct" id="progPct">0%</span>
      </div>
      <div class="prog-wrap"><div class="prog-bar" id="progBar"></div></div>
      <!-- Quick settings -->
      <div style="display:grid;grid-template-columns:repeat(5,1fr);gap:8px;margin-top:4px">
        <div><label style="font-size:9px">Threads</label><input type="number" id="qThreads" value="200" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label style="font-size:9px">Timeout (s)</label><input type="number" id="qTimeout" value="8" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label style="font-size:9px">Max Lat (ms)</label><input type="number" id="qMaxLat" value="3500" style="font-size:11px;padding:4px 7px"></div>
        <div><label style="font-size:9px">P2 Rounds</label><input type="number" id="qRounds" value="3" min="0" style="font-size:11px;padding:4px 7px"></div>
        <div><label style="font-size:9px">Sample/Subnet</label><input type="number" id="sampleSize" value="1" min="1" style="font-size:11px;padding:4px 7px"></div>
      </div>
    </div>
  </div>

  <!-- IP Input + Feed -->
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
    <div class="card">
      <div class="card-hd"><div>IP Ranges</div><span id="feedCount" style="color:var(--dim)"></span></div>
      <div class="card-bd" style="padding:10px">
        <textarea id="ipInput" rows="7" placeholder="104.16.0.0/12&#10;162.158.0.0/15&#10;Or single IPs..." style="resize:vertical"></textarea>
        <div style="display:flex;justify-content:space-between;align-items:center;margin-top:6px">
          <span style="font-size:10px;color:var(--dim);font-family:var(--font-mono)">CIDR or plain IPs</span>
          <input type="number" id="maxIPInput" placeholder="Max IPs (0=all)" style="width:140px;font-size:11px;padding:3px 7px">
        </div>
      </div>
    </div>
    <div class="card">
      <div class="card-hd">
        <div class="live-feed-hd" style="padding:0;border:none;background:none">
          <div class="dot dot-idle" id="feedDot"></div>
          <span>Live Feed</span>
        </div>
        <span id="feedCount2" style="color:var(--dim);font-size:10px"></span>
      </div>
      <div class="live-feed-body" id="liveFeed" style="height:160px">
        <div class="live-row live-row-scan"><span style="color:var(--dim)">›</span><span>Waiting to start...</span></div>
      </div>
    </div>
  </div>
</div>

<!-- ══ RESULTS PAGE ══ -->
<div id="page-results" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Results</h2><p id="resSummary" style="font-family:var(--font-mono);font-size:10px">No results yet</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="exportResults('csv')">↓ CSV</button>
      <button class="btn btn-sm" onclick="exportResults('json')">↓ JSON</button>
      <button class="btn btn-sm" onclick="copyAllPassed()">⎘ Copy All</button>
    </div>
  </div>

  <!-- IP Chips -->
  <div class="card" style="margin-bottom:12px">
    <div class="card-hd">
      <div>✓ Passed IPs</div>
      <span id="passedBadge" class="badge bg">0</span>
    </div>
    <div class="card-bd"><div class="ip-chips" id="ipChips"><span style="color:var(--dim);font-size:12px">No results</span></div></div>
  </div>

  <!-- Tabs -->
  <div class="tab-bar">
    <button class="tab active" onclick="switchTab('p2',this)">▶ Phase 2 — Deep Test</button>
    <button class="tab" onclick="switchTab('p1',this)">⚡ Phase 1 — Initial Scan</button>
  </div>

  <!-- Phase 2 table -->
  <div id="tab-p2" class="card">
    <div class="card-hd">
      <div>Phase 2 — Stability & Speed</div>
      <span id="p2CountBadge" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP Address</th><th>Score</th><th>Latency</th><th>Jitter</th><th>Pkt Loss</th><th>Download</th><th>Upload</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody id="p2Tbody"><tr><td colspan="10" style="text-align:center;color:var(--dim);padding:32px;font-family:var(--font-mono)">No Phase 2 results yet</td></tr></tbody>
      </table>
    </div>
  </div>

  <!-- Phase 1 table -->
  <div id="tab-p1" class="card" style="display:none">
    <div class="card-hd">
      <div>Phase 1 — Initial Scan (passed only)</div>
      <span id="p1CountBadge" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP Address</th><th>Latency</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody id="p1Tbody"><tr><td colspan="5" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<!-- ══ HISTORY PAGE ══ -->
<div id="page-history" class="page">
  <div class="phd">
    <div class="phd-l"><h2>History</h2><p>Previous scan sessions</p></div>
  </div>
  <div id="histList"><p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p></div>
</div>

<!-- ══ CONFIG PAGE ══ -->
<div id="page-config" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Settings</h2><p>Saved automatically to disk on save</p></div>
    <div class="phd-r">
      <button class="btn btn-success" onclick="saveConfig()">⬡ Save Settings</button>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>⚡ PHASE 1 — Initial Scan</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Threads</label><input type="number" id="cfgThreads" value="200" min="1"></div>
        <div class="f-row"><label>Timeout (seconds)</label><input type="number" id="cfgTimeout" value="8" min="1"></div>
        <div class="f-row"><label>Max Latency (ms)</label><input type="number" id="cfgMaxLat" value="3500"></div>
        <div class="f-row"><label>Retries</label><input type="number" id="cfgRetries" value="2" min="0"></div>
        <div class="f-row"><label>Max IPs (0 = all)</label><input type="number" id="cfgMaxIPs" value="0" min="0"></div>
        <div class="f-row"><label>Sample per Subnet</label><input type="number" id="cfgSampleSize" value="1" min="1"></div>
      </div>
      <div class="f-row"><label>Test URL</label><input type="text" id="cfgTestURL" value="https://www.gstatic.com/generate_204"></div>
      <label class="chk-row"><input type="checkbox" id="cfgShuffle" checked> Shuffle IPs before scan</label>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>◈ PHASE 2 — Deep Test</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Rounds (0 = disabled)</label><input type="number" id="cfgRounds" value="3" min="0"></div>
        <div class="f-row"><label>Interval (seconds)</label><input type="number" id="cfgInterval" value="5" min="1"></div>
        <div class="f-row"><label>Ping count (packet loss)</label><input type="number" id="cfgPLCount" value="5" min="1"></div>
        <div class="f-row"><label>Max Packet Loss % (-1 = off)</label><input type="number" id="cfgMaxPL" value="-1" min="-1" max="100"></div>
        <div class="f-row"><label>Min Download Mbps (0 = off)</label><input type="number" id="cfgMinDL" value="0" min="0" step="0.1"></div>
        <div class="f-row"><label>Min Upload Mbps (0 = off)</label><input type="number" id="cfgMinUL" value="0" min="0" step="0.1"></div>
      </div>
      <div style="display:flex;gap:20px;flex-wrap:wrap;margin-bottom:10px">
        <label class="chk-row"><input type="checkbox" id="cfgJitter"> Measure Jitter (RFC 3550)</label>
        <label class="chk-row"><input type="checkbox" id="cfgSpeed"> Speed Test (Phase 2 only)</label>
      </div>
      <div class="f-grid" id="speedURLs" style="display:none">
        <div class="f-row"><label>Download URL</label><input type="text" id="cfgDLURL" value="https://speed.cloudflare.com/__down?bytes=5000000"></div>
        <div class="f-row"><label>Upload URL</label><input type="text" id="cfgULURL" value="https://speed.cloudflare.com/__up"></div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>⬡ FRAGMENT</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Mode</label>
          <select id="cfgFragMode">
            <option value="manual">manual</option>
            <option value="auto">auto (tests all zones)</option>
            <option value="off">off</option>
          </select>
        </div>
        <div class="f-row"><label>Packets (manual mode)</label><input type="text" id="cfgFragPkts" value="tlshello"></div>
        <div></div>
        <div class="f-row"><label>Length (manual mode)</label><input type="text" id="cfgFragLen" value="10-20"></div>
        <div class="f-row"><label>Interval ms (manual mode)</label><input type="text" id="cfgFragInt" value="10-20"></div>
      </div>
      <div style="font-size:10px;color:var(--dim);font-family:var(--font-mono);margin-top:4px">
        Auto mode tests: tlshello · 1-3 · 1-5 · 1-10 · random
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-hd"><div>▸ XRAY</div></div>
    <div class="card-bd">
      <div class="f-grid-3">
        <div class="f-row"><label>Log Level</label>
          <select id="cfgXrayLog">
            <option value="none">none</option>
            <option value="error">error</option>
            <option value="warning">warning</option>
            <option value="info">info</option>
            <option value="debug">debug</option>
          </select>
        </div>
        <div class="f-row"><label>Mux Concurrency (-1 = off)</label><input type="number" id="cfgMuxConc" value="-1"></div>
        <div style="display:flex;align-items:flex-end;padding-bottom:11px"><label class="chk-row"><input type="checkbox" id="cfgMuxEnabled"> Enable Mux</label></div>
      </div>
    </div>
  </div>
</div>

<!-- ══ IMPORT PAGE ══ -->
<div id="page-import" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Import Config</h2><p>Paste a vless/vmess/trojan link</p></div>
    <div class="phd-r" id="clearProxyBtn" style="display:none">
      <button class="btn btn-danger btn-sm" onclick="clearSavedProxy()">✕ Remove Config</button>
    </div>
  </div>
  <div class="card">
    <div class="card-hd"><div>⬡ Proxy Link</div></div>
    <div class="card-bd">
      <div class="f-row"><label>vless:// or vmess:// or trojan://</label><textarea id="linkInput" rows="3" placeholder="vless://uuid@domain:443?..."></textarea></div>
      <button class="btn btn-primary" onclick="parseLink()">▶ Parse & Save</button>
    </div>
  </div>
  <div id="parsedResult" style="display:none" class="card">
    <div class="card-hd"><div>✓ Config Parsed</div></div>
    <div class="card-bd">
      <div class="parsed-box" id="parsedBox"></div>
    </div>
  </div>
</div>

<!-- ══ TUI LOG ══ -->
<div id="page-tui" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Live Log</h2><p>All scanner events</p></div>
    <div class="phd-r">
      <button class="btn btn-sm" onclick="clearTUI()">✕ Clear</button>
      <button class="btn btn-sm" id="btnAS" onclick="toggleAS()">↓ Auto-scroll</button>
    </div>
  </div>
  <div class="tui-wrap">
    <div class="tui-hd">
      <div class="tui-dots">
        <div class="tui-dot" style="background:#ff3366"></div>
        <div class="tui-dot" style="background:#ffe000"></div>
        <div class="tui-dot" style="background:#00ff88"></div>
      </div>
      <span style="margin-left:8px;font-size:11px;color:var(--tx2)">piyazche — scanner log</span>
      <span id="tuiStatus" style="margin-left:auto;color:var(--dim);font-size:10px">idle</span>
    </div>
    <div class="tui-body" id="tuiBody">
      <div class="tui-line"><span class="tui-t">--:--:--</span><span class="tui-info">Piyazche ready<span class="cursor"></span></span></div>
    </div>
  </div>
</div>

</div><!-- /main -->
</div><!-- /app -->

<script>
// ══ STATE ══
let ws=null,p1Results=[],p2Results=[],shodanIPs=[],tuiAS=true,viewingSession=false;
let feedRows=[],maxFeedRows=100,currentTab='p2';
let currentTheme='neon';

// ══ THEME ══
function toggleTheme(){
  currentTheme=currentTheme==='neon'?'newspaper':'neon';
  document.documentElement.setAttribute('data-theme',currentTheme==='newspaper'?'newspaper':'');
  document.getElementById('themeBtn').textContent=currentTheme==='neon'?'◈ NEWSPAPER':'◈ NEON NOIR';
  localStorage.setItem('pyz_theme',currentTheme);
}
(function(){
  const t=localStorage.getItem('pyz_theme');
  if(t==='newspaper'){currentTheme='newspaper';document.documentElement.setAttribute('data-theme','newspaper');document.getElementById('themeBtn').textContent='◈ NEON NOIR';}
})();

// ══ NAV ══
function nav(page,btn){
  document.querySelectorAll('.page').forEach(p=>p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(b=>b.classList.remove('active'));
  document.getElementById('page-'+page).classList.add('active');
  if(btn) btn.classList.add('active');
  else{const b=document.querySelector('[data-page="'+page+'"]');if(b)b.classList.add('active');}
  if(page==='results') refreshResults();
  if(page==='history') refreshHistory();
}

// ══ TABS ══
function switchTab(tab,btn){
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
  const{type,payload}=msg;
  switch(type){
    case 'status': setStatus(payload.status,payload.phase); break;
    case 'progress': onProgress(payload); break;
    case 'live_ip': addFeedRow(payload.ip,'scan'); break;
    case 'ip_result':
      if(payload.success) addFeedRow(payload.ip+' · '+payload.latency+'ms','ok');
      break;
    case 'tui': appendTUI(payload); break;
    case 'phase2_start':
      setStatus('scanning','phase2');
      addFeedRow('◈ Phase 2 starting — '+payload.count+' IPs','p2');
      break;
    case 'phase2_result':{
      const r=payload;
      const ul=r.UploadMbps>0?' ↑'+r.UploadMbps.toFixed(1)+'M':'';
      const dl=r.DownloadMbps>0?' ↓'+r.DownloadMbps.toFixed(1)+'M':'';
      const jt=r.JitterMs>0?' ~'+r.JitterMs.toFixed(0)+'ms':'';
      const rowTxt=r.IP+' · '+Math.round(r.AvgLatencyMs)+'ms'+jt+dl+ul;
      addFeedRow(rowTxt,r.Passed?'p2':'fail');
      if(!viewingSession){p2Results.push(r);renderP2();updatePassedChips();}
      break;
    }
    case 'scan_done':
      setStatus('done','');
      addFeedRow('✓ Scan complete — '+payload.passed+' passed','ok');
      if(!viewingSession){refreshResults();}
      refreshHistory();
      break;
    case 'error': appendTUI({t:now(),l:'err',m:payload.message}); break;
    case 'shodan_done':
      shodanIPs=payload.ips||[];
      appendTUI({t:now(),l:'ok',m:'Shodan: '+shodanIPs.length+' IPs found'});
      break;
  }
}

function now(){return new Date().toLocaleTimeString('en-US',{hour12:false});}

// ══ LIVE FEED ══
function addFeedRow(txt,type){
  const body=document.getElementById('liveFeed');
  const div=document.createElement('div');
  div.className='live-row live-row-'+type;
  div.innerHTML='<span style="color:var(--dim);font-size:9px;font-family:var(--font-mono);flex-shrink:0">'+now()+'</span><span>'+txt+'</span>';
  body.insertBefore(div,body.firstChild);
  feedRows.push(div);
  if(feedRows.length>maxFeedRows){const old=feedRows.shift();if(old.parentNode)old.parentNode.removeChild(old);}
  const c=document.getElementById('feedCount2');
  if(c) c.textContent=feedRows.length+' events';
}

// ══ TUI ══
function appendTUI(entry){
  const body=document.getElementById('tuiBody');
  const div=document.createElement('div');div.className='tui-line';
  const cls='tui-'+(entry.l==='ok'?'ok':entry.l==='err'?'err':entry.l==='warn'?'warn':entry.l==='p2'?'p2':entry.l==='scan'?'scan':'info');
  div.innerHTML='<span class="tui-t">'+entry.t+'</span><span class="'+cls+'">'+entry.m+'</span>';
  body.appendChild(div);
  if(tuiAS) body.scrollTop=body.scrollHeight;
  const st=document.getElementById('tuiStatus');if(st)st.textContent=entry.m.substring(0,50);
}
function clearTUI(){document.getElementById('tuiBody').innerHTML='';}
function toggleAS(){tuiAS=!tuiAS;document.getElementById('btnAS').textContent=tuiAS?'↓ Auto-scroll':'— Manual';}

// ══ STATUS ══
function setStatus(st,phase){
  const dot=document.getElementById('sDot'),txt=document.getElementById('sTxt'),ph=document.getElementById('sPhase');
  const pdot=document.getElementById('pDot');
  const scan=document.getElementById('nbScan');
  dot.className='dot dot-'+(st==='scanning'?'scan':st==='done'?'done':st==='paused'?'warn':'idle');
  if(pdot) pdot.className='dot dot-'+(st==='scanning'?'scan':st==='done'?'done':'idle');
  txt.textContent=st;
  ph.textContent=phase?'· '+phase:'';
  if(scan) scan.style.display=st==='scanning'?'':'none';
  document.getElementById('btnStart').style.display=st==='scanning'||st==='paused'?'none':'';
  document.getElementById('btnStop').style.display=st==='scanning'||st==='paused'?'':'none';
  document.getElementById('btnPause').style.display=st==='scanning'?'':'none';
  if(st==='idle'||st==='done'){document.getElementById('progBar').style.width='0%';}
  if(st==='scanning'&&phase==='phase2') document.getElementById('progBar').classList.add('p2');
  else document.getElementById('progBar').classList.remove('p2');
}

// ══ PROGRESS ══
function onProgress(p){
  const pct=p.total>0?Math.round(p.done/p.total*100):0;
  document.getElementById('progBar').style.width=pct+'%';
  document.getElementById('progPct').textContent=pct+'%';
  document.getElementById('progTxt').textContent=p.done+' / '+p.total;
  document.getElementById('stTotal').textContent=p.total||'—';
  document.getElementById('stDone').textContent=p.done||0;
  document.getElementById('stPass').textContent=p.succeeded||0;
  document.getElementById('stFail').textContent=p.failed||0;
  document.getElementById('stETA').textContent=p.eta||'—';
  document.getElementById('tbProgress').textContent=p.done+'/'+p.total+' · '+pct+'%';
  if(p.rate>0) document.getElementById('progRate').textContent=(p.rate||0).toFixed(1)+' IP/s';
}

// ══ SCAN ══
async function startScan(){
  const ipInput=document.getElementById('ipInput').value.trim();
  if(!ipInput){appendTUI({t:now(),l:'warn',m:'No IP ranges entered'});return;}
  const maxIPs=parseInt(document.getElementById('maxIPInput').value)||0;
  const quickSettings={
    threads:parseInt(document.getElementById('qThreads').value)||200,
    timeout:parseInt(document.getElementById('qTimeout').value)||8,
    maxLatency:parseInt(document.getElementById('qMaxLat').value)||3500,
    stabilityRounds:parseInt(document.getElementById('qRounds').value)||3,
    sampleSize:parseInt(document.getElementById('sampleSize').value)||1,
  };
  const btn=document.getElementById('btnStart');
  btn.disabled=true;
  viewingSession=false;
  const b=document.getElementById('sessionBanner');if(b)b.remove();
  p1Results=[];p2Results=[];
  feedRows=[];
  document.getElementById('liveFeed').innerHTML='<div class="live-row live-row-scan"><span style="color:var(--dim)">›</span><span>Scan started...</span></div>';
  document.getElementById('progBar').classList.remove('p2');
  const res=await fetch('/api/scan/start',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({quickSettings,ipRanges:ipInput,maxIPs})});
  const data=await res.json();
  btn.disabled=false;
  if(!data.ok){appendTUI({t:now(),l:'err',m:'Error: '+data.error});return;}
  setStatus('scanning','phase1');
  appendTUI({t:now(),l:'ok',m:'▶ Scan started — '+data.total+' IPs'});
}

async function stopScan(){
  await fetch('/api/scan/stop',{method:'POST'});
  setStatus('idle','');
}

async function pauseScan(){
  const res=await fetch('/api/scan/pause',{method:'POST'});
  const d=await res.json();
  if(d.message==='paused') setStatus('paused','');
  else setStatus('scanning','');
}

// ══ RESULTS ══
function refreshResults(){
  if(viewingSession) return;
  fetch('/api/results').then(r=>r.json()).then(data=>{
    p1Results=data.phase1||[];
    p2Results=data.phase2||[];
    renderP1();renderP2();updatePassedChips();
  });
}

function updatePassedChips(){
  const passed=(p2Results||[]).filter(r=>r.Passed);
  document.getElementById('resSummary').textContent=passed.length+' passed out of '+(p2Results||[]).length+' tested';
  document.getElementById('passedBadge').textContent=passed.length;
  document.getElementById('nbResults').textContent=passed.length;
  const chips=document.getElementById('ipChips');
  if(!passed.length){chips.innerHTML='<span style="color:var(--dim);font-size:12px">No results</span>';return;}
  chips.innerHTML=passed.map(r=>{
    const dl=r.DownloadMbps>0?' ↓'+r.DownloadMbps.toFixed(1):'';
    return '<div class="ip-chip" onclick="copyWithIP(\''+r.IP+'\')" title="Click: copy link with this IP">'+r.IP+'<span class="lat">'+Math.round(r.AvgLatencyMs)+'ms'+dl+'</span></div>';
  }).join('');
}

function renderP2(){
  const tbody=document.getElementById('p2Tbody');
  document.getElementById('p2CountBadge').textContent=(p2Results||[]).length+' IPs';
  if(!p2Results||!p2Results.length){
    tbody.innerHTML='<tr><td colspan="10" style="text-align:center;color:var(--dim);padding:32px;font-family:var(--font-mono)">No Phase 2 results yet</td></tr>';
    return;
  }
  tbody.innerHTML=p2Results.map((r,i)=>{
    const sc=r.StabilityScore||0;
    const scc=sc>=75?'var(--g)':sc>=50?'var(--y)':'var(--r)';
    const lc=r.AvgLatencyMs<=500?'var(--g)':r.AvgLatencyMs<=1500?'var(--y)':'var(--r)';
    const badge=r.Passed?'<span class="badge bg">PASS</span>':'<span class="badge br" title="'+(r.FailReason||'')+'">FAIL</span>';
    const dl=r.DownloadMbps||0;const ul=r.UploadMbps||0;
    const dlTxt=dl>0?dl.toFixed(1)+' M':'—';
    const ulTxt=ul>0?ul.toFixed(1)+' M':'—';
    const dlc=dl<=0?'var(--dim)':dl>=5?'var(--g)':dl>=1?'var(--y)':'var(--r)';
    const plc=(r.PacketLossPct||0)<=5?'var(--g)':(r.PacketLossPct||0)<=20?'var(--y)':'var(--r)';
    const jt=r.JitterMs||0;const jc=jt<=20?'var(--g)':jt<=80?'var(--y)':'var(--r)';
    return '<tr class="'+(r.Passed?'pass-row':'fail-row')+'">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700;font-size:12px">'+r.IP+'</td>'+
      '<td style="color:'+scc+';font-weight:700;font-size:14px">'+sc.toFixed(0)+'</td>'+
      '<td style="color:'+lc+'">'+Math.round(r.AvgLatencyMs||0)+'ms</td>'+
      '<td style="color:'+jc+'">'+(jt>0?jt.toFixed(0)+'ms':'—')+'</td>'+
      '<td style="color:'+plc+'">'+(r.PacketLossPct||0).toFixed(0)+'%</td>'+
      '<td style="color:'+dlc+'">'+dlTxt+'</td>'+
      '<td style="color:var(--tx2)">'+ulTxt+'</td>'+
      '<td>'+badge+'</td>'+
      '<td><div style="display:flex;gap:3px">'+
        '<button class="copy-btn" onclick="copyIP(\''+r.IP+'\')" title="Copy IP">⎘</button>'+
        '<button class="copy-btn" onclick="copyWithIP(\''+r.IP+'\')" title="Copy vless link with this IP">⬡</button>'+
      '</div></td></tr>';
  }).join('');
}

function renderP1(){
  const tbody=document.getElementById('p1Tbody');
  const succ=(p1Results||[]).filter(r=>r.success||r.Success);
  document.getElementById('p1CountBadge').textContent=(p1Results||[]).length+' IPs scanned · '+succ.length+' passed';
  if(!succ.length){
    tbody.innerHTML='<tr><td colspan="5" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr>';
    return;
  }
  tbody.innerHTML=succ.map((r,i)=>{
    const ip=r.ip||r.IP||'';const lat=r.latency_ms||r.LatencyMs||0;
    const lc=lat<=500?'var(--g)':lat<=1500?'var(--y)':'var(--r)';
    return '<tr class="p1-row">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700">'+ip+'</td>'+
      '<td style="color:'+lc+'">'+Math.round(lat)+'ms</td>'+
      '<td><span class="badge bg">OK</span></td>'+
      '<td><button class="copy-btn" onclick="copyIP(\''+ip+'\')" title="Copy IP">⎘</button></td>'+
    '</tr>';
  }).join('');
  if(p1Results.length>succ.length) tbody.innerHTML+='<tr><td colspan="5" style="text-align:center;color:var(--dim);padding:10px;font-size:10px">'+(p1Results.length-succ.length)+' failed IPs hidden</td></tr>';
}

function copyAllPassed(){
  const passed=(p2Results||[]).filter(r=>r.Passed).map(r=>r.IP);
  if(!passed.length) return;
  navigator.clipboard.writeText(passed.join('\n')).then(()=>appendTUI({t:now(),l:'ok',m:'⎘ '+passed.length+' IPs copied'}));
}

// ══ HISTORY ══
function refreshHistory(){
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const el=document.getElementById('histList');
    if(!sessions||!sessions.length){el.innerHTML='<p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p>';return;}
    el.innerHTML=sessions.map(s=>{
      const passed=(s.results||[]).filter(r=>r.Passed).length;
      const d=new Date(s.startedAt);
      return '<div class="hist-item" onclick="showSession(\''+s.id+'\')">'+
        '<div class="hist-n" style="color:'+(passed>0?'var(--g)':'var(--dim)')+'">'+passed+'</div>'+
        '<div class="hist-info">'+
          '<div style="color:var(--tx);font-weight:600;font-size:12px">'+s.totalIPs+' IPs · '+passed+' passed</div>'+
          '<div class="hist-date">'+d.toLocaleString()+' · '+s.duration+'</div>'+
        '</div>'+
        '<div style="color:var(--dim);font-size:10px">▶</div>'+
      '</div>';
    }).join('');
  });
}

function showSession(id){
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    const s=sessions.find(x=>x.id===id);if(!s)return;
    viewingSession=true;
    p2Results=s.results||[];p1Results=[];
    renderP2();renderP1();updatePassedChips();
    nav('results');
    // Session banner
    const existing=document.getElementById('sessionBanner');if(existing)existing.remove();
    const banner=document.createElement('div');
    banner.id='sessionBanner';banner.className='session-banner';
    const d=new Date(s.startedAt);
    banner.innerHTML='<span>📂 Viewing session: '+d.toLocaleString()+' — '+(s.results||[]).filter(r=>r.Passed).length+' passed</span>'+
      '<button onclick="clearSession()" style="background:var(--rd);border:1px solid var(--r);color:var(--r);padding:3px 10px;cursor:pointer;border-radius:3px;font-size:11px;font-family:var(--font-mono)">✕ Back to live</button>';
    document.getElementById('page-results').insertBefore(banner,document.getElementById('page-results').firstChild);
  });
}

function clearSession(){
  viewingSession=false;
  const b=document.getElementById('sessionBanner');if(b)b.remove();
  refreshResults();
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
    xray:{
      logLevel:document.getElementById('cfgXrayLog').value,
      mux:{enabled:document.getElementById('cfgMuxEnabled').checked,concurrency:parseInt(document.getElementById('cfgMuxConc').value)||-1}
    },
  };
  // Sync quick panel
  document.getElementById('qThreads').value=scanCfg.scan.threads;
  document.getElementById('qTimeout').value=scanCfg.scan.timeout;
  document.getElementById('qMaxLat').value=scanCfg.scan.maxLatency;
  document.getElementById('qRounds').value=scanCfg.scan.stabilityRounds;
  document.getElementById('sampleSize').value=scanCfg.scan.sampleSize;
  fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({scanConfig:JSON.stringify(scanCfg)})}).then(()=>{
    appendTUI({t:now(),l:'ok',m:'✓ Settings saved to disk'});
    updateConfigSummary(scanCfg.scan,scanCfg.fragment);
    nav('scan');
  });
}

function updateConfigSummary(s,f){
  const el=document.getElementById('configSummary');
  const frag=f?(' · frag:'+f.mode):'';
  el.innerHTML='threads:'+s.threads+' · timeout:'+s.timeout+'s · maxLat:'+s.maxLatency+'ms · rounds:'+s.stabilityRounds+frag+(s.speedTest?' · speed:ON':'');
}

// ══ IMPORT ══
async function parseLink(){
  const input=document.getElementById('linkInput').value.trim();if(!input)return;
  const res=await fetch('/api/config/parse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})});
  const data=await res.json();
  if(!data.ok){appendTUI({t:now(),l:'err',m:'Error: '+data.error});return;}
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
    (p.fp?'<br><span class="k">fingerprint: </span><span class="v">'+p.fp+'</span>':'');
  document.getElementById('parsedResult').style.display='block';
  appendTUI({t:now(),l:'ok',m:'✓ Config: '+p.address+' ('+p.method+'/'+p.type+')'});
}

function updateProxyChip(addr,method,type){
  document.getElementById('proxyChipTxt').textContent=addr+' · '+method+'/'+type;
  document.getElementById('proxyChip').style.display='inline-flex';
  document.getElementById('clearProxyBtn').style.display='';
  document.getElementById('configSummary').innerHTML='⬡ '+addr+' ('+method+'/'+type+')';
}

async function clearSavedProxy(){
  await fetch('/api/config/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({proxyConfig:''})});
  document.getElementById('proxyChip').style.display='none';
  document.getElementById('clearProxyBtn').style.display='none';
  document.getElementById('parsedResult').style.display='none';
  document.getElementById('configSummary').textContent='No config — import a proxy link first';
  appendTUI({t:now(),l:'warn',m:'Proxy config removed'});
}

function maskUUID(u){return!u||u.length<8?u:u.slice(0,8)+'••••••••';}

// ══ COPY / EXPORT ══
function exportResults(f){window.location.href='/api/results/export?format='+f;}
function copyIP(ip){
  navigator.clipboard.writeText(ip).then(()=>appendTUI({t:now(),l:'ok',m:'⎘ '+ip})).catch(()=>{
    const el=document.createElement('textarea');el.value=ip;document.body.appendChild(el);el.select();document.execCommand('copy');document.body.removeChild(el);
    appendTUI({t:now(),l:'ok',m:'⎘ '+ip});
  });
}
function copyWithIP(newIP){
  const rawLink=document.getElementById('linkInput').value.trim();
  if(!rawLink){copyIP(newIP);return;}
  try{
    const updated=rawLink.replace(/(@)([^:@\/?#\[\]]+)(:\d+)/,'$1'+newIP+'$3');
    navigator.clipboard.writeText(updated).then(()=>appendTUI({t:now(),l:'ok',m:'⬡ Link with '+newIP+' copied'}));
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
        const s=sc.scan||{},f=sc.fragment||{},x=sc.xray||{};
        const sv=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        const sc2=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.checked=!!v;};
        const ss=(id,v)=>{const el=document.getElementById(id);if(el&&v!=null)el.value=v;};
        if(s.threads!=null){sv('cfgThreads',s.threads);sv('qThreads',s.threads);}
        if(s.timeout!=null){sv('cfgTimeout',s.timeout);sv('qTimeout',s.timeout);}
        if(s.maxLatency!=null){sv('cfgMaxLat',s.maxLatency);sv('qMaxLat',s.maxLatency);}
        if(s.retries!=null) sv('cfgRetries',s.retries);
        if(s.maxIPs!=null) sv('cfgMaxIPs',s.maxIPs);
        if(s.sampleSize!=null){sv('cfgSampleSize',s.sampleSize);sv('sampleSize',s.sampleSize);}
        if(s.testUrl) sv('cfgTestURL',s.testUrl);
        if(s.shuffle!=null) sc2('cfgShuffle',s.shuffle);
        if(s.stabilityRounds!=null){sv('cfgRounds',s.stabilityRounds);sv('qRounds',s.stabilityRounds);}
        if(s.stabilityInterval!=null) sv('cfgInterval',s.stabilityInterval);
        if(s.packetLossCount!=null) sv('cfgPLCount',s.packetLossCount);
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
        updateConfigSummary(s,f);
      }catch(e){console.warn('load err',e);}
    }
    fetch('/api/tui/stream').then(r=>r.json()).then(data=>{
      if(data.lines) data.lines.forEach(l=>{try{appendTUI(JSON.parse(l));}catch(e){}});
    }).catch(()=>{});
  });
}

// ══ INIT ══
connectWS();
fetch('/api/status').then(r=>r.json()).then(d=>setStatus(d.status||'idle',d.phase||''));
loadSavedSettings();
</script>
</body>
</html>
`
