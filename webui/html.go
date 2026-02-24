package webui

const indexHTMLContent = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Piyazche Scanner</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&family=Space+Mono:wght@400;700&family=Bebas+Neue&family=Orbitron:wght@400;700;900&display=swap" rel="stylesheet">
<style>
/* ══════════════════════════════════════════════════
   THEME VARIABLES
══════════════════════════════════════════════════ */
:root{
  /* ── NEON NIGHT (default) ── */
  --bg:#050709;--bg2:#080c12;--bg3:#0d1220;--bg4:#121828;
  --bd:#0f1825;--bd2:#192235;--bd3:#223048;
  --tx:#c8d8f5;--tx2:#4a6080;--dim:#1e2d45;
  --g:#00ffaa;--gd:rgba(0,255,170,.07);--g2:#00cc88;
  --c:#38bfff;--cd:rgba(56,191,255,.07);--c2:#0099dd;
  --y:#ffd700;--yd:rgba(255,215,0,.07);
  --r:#ff2d6b;--rd:rgba(255,45,107,.07);
  --p:#c060ff;--pd:rgba(192,96,255,.07);
  --o:#ff7700;
  --rad:10px;--rad-sm:6px;--rad-xs:4px;
  --font-head:'Space Grotesk',sans-serif;
  --font-mono:'Space Mono',monospace;
  --glow-g:0 0 24px rgba(0,255,170,.35),0 0 8px rgba(0,255,170,.2);
  --glow-c:0 0 24px rgba(56,191,255,.35),0 0 8px rgba(56,191,255,.2);
  --glow-r:0 0 24px rgba(255,45,107,.35),0 0 8px rgba(255,45,107,.2);
  --glow-p:0 0 24px rgba(192,96,255,.35);
  --shadow:0 4px 20px rgba(0,0,0,.5);
}
[data-theme="day"]{
  /* ── DAY MODE ── */
  --bg:#f0f4fa;--bg2:#ffffff;--bg3:#e8eef8;--bg4:#dde5f2;
  --bd:#c8d5e8;--bd2:#b8c8e0;--bd3:#a0b5d0;
  --tx:#1a2540;--tx2:#5068a0;--dim:#8899bb;
  --g:#00aa55;--gd:rgba(0,170,85,.08);--g2:#008844;
  --c:#0077dd;--cd:rgba(0,119,221,.08);--c2:#0055aa;
  --y:#cc8800;--yd:rgba(204,136,0,.08);
  --r:#dd1144;--rd:rgba(221,17,68,.08);
  --p:#7700cc;--pd:rgba(119,0,204,.08);
  --o:#cc5500;
  --glow-g:0 2px 12px rgba(0,170,85,.2);
  --glow-c:0 2px 12px rgba(0,119,221,.2);
  --glow-r:0 2px 12px rgba(221,17,68,.2);
  --glow-p:0 2px 12px rgba(119,0,204,.2);
  --shadow:0 2px 12px rgba(0,0,0,.1);
}
*{margin:0;padding:0;box-sizing:border-box}
html{height:100%}
body{font-family:var(--font-head);background:var(--bg);color:var(--tx);height:100%;font-size:14px;line-height:1.5;overflow:hidden;transition:background .3s,color .3s}
.app{display:grid;grid-template-columns:200px 1fr;grid-template-rows:56px 1fr;height:100vh}

/* ══ TOPBAR ══ */
.topbar{
  grid-column:1/-1;
  background:var(--bg2);
  border-bottom:2px solid var(--bd2);
  display:flex;align-items:center;
  padding:0 18px;gap:14px;
  position:relative;z-index:100;
  box-shadow:var(--shadow);
}
.logo{
  font-family:'Orbitron',monospace;
  font-size:18px;letter-spacing:4px;font-weight:900;
  user-select:none;
  background:linear-gradient(90deg,var(--c),var(--g),var(--p));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
  filter:drop-shadow(0 0 8px rgba(56,191,255,.4));
}
[data-theme="day"] .logo{
  filter:drop-shadow(0 1px 3px rgba(0,119,221,.3));
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
.tb-right{margin-left:auto;display:flex;align-items:center;gap:8px}
.proxy-chip{
  display:inline-flex;align-items:center;gap:5px;
  padding:3px 10px;border-radius:5px;
  font-size:11px;font-family:var(--font-mono);
  background:var(--gd);border:1px solid rgba(0,255,170,.3);
  color:var(--g);cursor:pointer;transition:all .15s;
}
.proxy-chip:hover{background:rgba(0,255,170,.12)}
.theme-btn{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx2);padding:5px 13px;border-radius:20px;
  font-size:11px;cursor:pointer;font-family:var(--font-mono);
  transition:all .2s;letter-spacing:.5px;
  display:flex;align-items:center;gap:5px;
}
.theme-btn:hover{border-color:var(--c);color:var(--c);background:var(--cd)}
.theme-icon{font-size:13px}

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
  font-family:'Orbitron',monospace;font-size:18px;font-weight:700;
  letter-spacing:2px;
  background:linear-gradient(90deg,var(--tx),var(--c));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
}
[data-theme="day"] .phd-l h2{
  background:linear-gradient(90deg,var(--tx),var(--c));
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;
}
.phd-l p{font-size:11px;color:var(--dim);margin-top:3px}
.phd-r{display:flex;gap:7px;align-items:center;flex-shrink:0}

/* ══ STATS ROW ══ */
.stats-row{display:grid;grid-template-columns:repeat(5,1fr);gap:10px;margin-bottom:16px}
.stat-card{
  background:var(--bg2);border:1px solid var(--bd);
  border-radius:var(--rad-sm);padding:14px 16px;
  position:relative;overflow:hidden;
  transition:border-color .2s,transform .1s;
  cursor:default;
}
.stat-card:hover{transform:translateY(-1px);border-color:var(--bd3)}
.stat-card::before{
  content:'';position:absolute;top:0;left:0;right:0;height:2px;
  background:linear-gradient(90deg,var(--c),var(--g));
  opacity:0;transition:opacity .2s;
}
.stat-card:hover::before{opacity:1}
.stat-v{
  font-family:'Orbitron',monospace;font-weight:700;
  line-height:1;transition:color .3s,text-shadow .3s;
  font-size:28px;
  letter-spacing:-1px;
}
.stat-v.active{text-shadow:0 0 20px currentColor}
.stat-l{font-size:9px;color:var(--dim);margin-top:8px;letter-spacing:2px;text-transform:uppercase}

/* ══ PROGRESS ══ */
.prog-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:14px;box-shadow:var(--shadow)}
.prog-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 16px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono)}
.prog-hd-l{display:flex;align-items:center;gap:7px}
.prog-bd{padding:14px 16px}
.prog-meta{display:flex;justify-content:space-between;font-size:11px;color:var(--dim);margin-bottom:6px;font-family:var(--font-mono)}
.prog-pct{color:var(--c);font-weight:700;font-size:14px}
.prog-wrap{background:var(--bg);border-radius:3px;height:8px;overflow:hidden;margin-bottom:10px;border:1px solid var(--bd)}
.prog-bar{height:100%;background:linear-gradient(90deg,var(--c),var(--g));border-radius:3px;transition:width .5s cubic-bezier(.4,0,.2,1);width:0%;box-shadow:var(--glow-c);position:relative}
.prog-bar::after{
  content:'';position:absolute;top:0;left:0;right:0;bottom:0;
  background:linear-gradient(90deg,transparent,rgba(255,255,255,.3),transparent);
  animation:shimmer 1.5s infinite;
}
@keyframes shimmer{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
.prog-bar.p2{background:linear-gradient(90deg,var(--p),var(--c));box-shadow:var(--glow-p)}

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
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--rad);overflow:hidden;margin-bottom:12px;box-shadow:var(--shadow)}
.card-hd{background:var(--bg3);border-bottom:1px solid var(--bd);padding:8px 14px;font-size:10px;color:var(--dim);display:flex;align-items:center;justify-content:space-between;font-family:var(--font-mono);letter-spacing:.5px}
.card-bd{padding:14px}

/* ══ FORMS ══ */
input,select,textarea{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx);border-radius:var(--rad-xs);
  padding:7px 10px;font-family:var(--font-mono);
  font-size:12px;width:100%;outline:none;
  transition:border-color .15s,box-shadow .15s;
}
input:focus,select:focus,textarea:focus{border-color:var(--c);box-shadow:0 0 0 2px var(--cd)}
textarea{resize:vertical;min-height:60px}
select option{background:var(--bg3)}
label{display:block;font-size:10px;color:var(--dim);margin-bottom:4px;letter-spacing:.5px;text-transform:uppercase;font-family:var(--font-mono)}

/* ══ BUTTONS ══ */
.btn{
  display:inline-flex;align-items:center;gap:5px;
  padding:7px 16px;border-radius:var(--rad-sm);
  font-family:var(--font-mono);font-size:11px;
  cursor:pointer;border:1px solid transparent;
  font-weight:700;letter-spacing:.5px;
  transition:all .15s;
}
.btn-success{background:var(--gd);border-color:var(--g2);color:var(--g);box-shadow:0 0 12px rgba(0,255,170,.15)}
.btn-success:hover{background:rgba(0,255,170,.15);box-shadow:var(--glow-g)}
.btn-danger{background:var(--rd);border-color:var(--r);color:var(--r)}
.btn-danger:hover{background:rgba(255,45,107,.15);box-shadow:var(--glow-r)}
.btn-primary{background:var(--cd);border-color:var(--c2);color:var(--c);box-shadow:0 0 12px rgba(56,191,255,.15)}
.btn-primary:hover{background:rgba(56,191,255,.15);box-shadow:var(--glow-c)}
.btn-warn{background:var(--yd);border-color:var(--y);color:var(--y)}
.btn{background:var(--bg3);border-color:var(--bd2);color:var(--tx2)}
.btn:hover{border-color:var(--bd3);color:var(--tx)}
.btn-success,.btn-danger,.btn-primary,.btn-warn{background:var(--gd)}.btn-success,.btn-success:hover{background:var(--gd);border-color:var(--g2);color:var(--g)}
.btn-sm{padding:4px 10px;font-size:10px}
.btn:disabled{opacity:.4;cursor:not-allowed}
.copy-btn{
  background:var(--bg3);border:1px solid var(--bd2);
  color:var(--tx2);border-radius:3px;padding:2px 7px;
  font-size:11px;cursor:pointer;transition:all .12s;
}
.copy-btn:hover{border-color:var(--c);color:var(--c)}

/* Explicit button color overrides (must come after .btn) */
.btn-success-real{background:var(--gd)!important;border-color:var(--g2)!important;color:var(--g)!important}
.btn-success-real:hover{background:rgba(0,255,170,.15)!important;box-shadow:var(--glow-g)!important}
.btn-danger-real{background:var(--rd)!important;border-color:var(--r)!important;color:var(--r)!important}
.btn-primary-real{background:var(--cd)!important;border-color:var(--c2)!important;color:var(--c)!important}

/* ══ TABLE ══ */
.tbl-wrap{overflow-x:auto}
.tbl{width:100%;border-collapse:collapse;font-size:12px}
.tbl th{
  padding:9px 12px;text-align:left;
  font-size:9px;letter-spacing:1.5px;text-transform:uppercase;
  color:var(--dim);font-family:var(--font-mono);
  background:var(--bg3);border-bottom:1px solid var(--bd2);
  position:sticky;top:0;
}
.tbl td{padding:8px 12px;border-bottom:1px solid var(--bd);vertical-align:middle}
.tbl tr:hover td{background:var(--bg3)}
.tbl tr.pass-row:hover td{background:rgba(0,255,170,.04)}
.tbl tr.fail-row td{opacity:.7}
.tbl tr.p1-row:hover td{background:rgba(56,191,255,.04)}

/* ══ BADGES ══ */
.badge{display:inline-flex;align-items:center;padding:2px 8px;border-radius:3px;font-size:10px;font-family:var(--font-mono);font-weight:700;letter-spacing:.5px}
.bg{background:var(--gd);color:var(--g);border:1px solid rgba(0,255,170,.3)}
.br{background:var(--rd);color:var(--r);border:1px solid rgba(255,45,107,.3)}
.by{background:var(--yd);color:var(--y);border:1px solid rgba(255,215,0,.3)}

/* ══ IP CHIPS ══ */
.ip-chips{display:flex;flex-wrap:wrap;gap:6px}
.ip-chip{
  display:inline-flex;align-items:center;gap:6px;
  background:var(--cd);border:1px solid var(--c2);
  color:var(--c);border-radius:4px;padding:4px 10px;
  font-family:var(--font-mono);font-size:11px;
  cursor:pointer;transition:all .12s;
}
.ip-chip:hover{background:rgba(56,191,255,.15);box-shadow:var(--glow-c)}
.ip-chip .lat{color:var(--y);font-size:10px}

/* ══ TABS ══ */
.tab-bar{display:flex;gap:1px;margin-bottom:0;background:var(--bd);border-radius:var(--rad) var(--rad) 0 0;overflow:hidden}
.tab{
  flex:1;padding:9px 14px;border:none;background:var(--bg3);
  color:var(--tx2);font-family:var(--font-mono);font-size:11px;
  cursor:pointer;transition:all .12s;
}
.tab:hover{background:var(--bg2);color:var(--tx)}
.tab.active{background:var(--bg2);color:var(--c);border-bottom:2px solid var(--c)}

/* ══ FORM GRID ══ */
.f-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.f-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.f-row{display:flex;flex-direction:column;gap:4px;margin-bottom:10px}
.chk-row{display:flex;align-items:center;gap:6px;font-size:11px;font-family:var(--font-mono);color:var(--tx2);cursor:pointer;margin-bottom:8px}
.chk-row input{width:auto}

/* ══ TUI ══ */
.tui-wrap{background:var(--bg2);border:1px solid var(--bd2);border-radius:var(--rad);overflow:hidden;box-shadow:var(--shadow)}
.tui-hd{
  background:var(--bg3);padding:8px 14px;
  display:flex;align-items:center;gap:8px;
  border-bottom:1px solid var(--bd2);
}
.tui-dots{display:flex;gap:5px}
.tui-dot{width:10px;height:10px;border-radius:50%}
.tui-body{height:calc(100vh - 200px);overflow-y:auto;padding:12px 16px;font-family:var(--font-mono);font-size:12px}
.tui-body::-webkit-scrollbar{width:3px}
.tui-body::-webkit-scrollbar-thumb{background:var(--bd3)}
.tui-line{display:flex;gap:10px;padding:1px 0;animation:fadeIn .15s}
.tui-t{color:var(--dim);flex-shrink:0;font-size:10px}
.tui-ok{color:var(--g)}.tui-err{color:var(--r)}.tui-warn{color:var(--y)}
.tui-p2{color:var(--p)}.tui-scan{color:var(--c)}.tui-info{color:var(--tx2)}
.cursor{display:inline-block;width:7px;height:12px;background:var(--c);margin-left:2px;animation:blink 1s step-end infinite;vertical-align:text-bottom}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}

/* ══ HISTORY ══ */
.hist-item{
  display:flex;align-items:center;gap:16px;
  padding:14px 18px;background:var(--bg2);
  border:1px solid var(--bd);border-radius:var(--rad);
  margin-bottom:8px;cursor:pointer;
  transition:all .12s;
}
.hist-item:hover{border-color:var(--c);background:var(--bg3);transform:translateX(2px)}
.hist-n{font-family:'Orbitron',monospace;font-size:28px;font-weight:700;min-width:50px;text-align:center}
.hist-info{flex:1}
.hist-date{font-size:10px;color:var(--dim);font-family:var(--font-mono);margin-top:2px}

/* ══ SESSION BANNER ══ */
.session-banner{
  display:flex;justify-content:space-between;align-items:center;
  padding:9px 14px;background:var(--yd);border:1px solid var(--y);
  border-radius:var(--rad-sm);margin-bottom:12px;
  font-family:var(--font-mono);font-size:11px;color:var(--y);
}

/* ══ PARSED BOX ══ */
.parsed-box{font-family:var(--font-mono);font-size:12px;line-height:1.8}
.parsed-box .k{color:var(--dim)}.parsed-box .v{color:var(--c)}

/* ══ SCROLLBAR GLOBAL ══ */
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
    <button class="theme-btn" onclick="toggleTheme()" id="themeBtn">
      <span class="theme-icon" id="themeIcon">☀</span>
      <span id="themeTxt">DAY</span>
    </button>
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
    <span class="nav-badge" id="nbHistory">0</span>
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
      <button class="btn btn-success-real" id="btnStart" onclick="startScan()">▶ Start</button>
      <button class="btn btn-danger-real" id="btnStop" onclick="stopScan()" style="display:none">■ Stop</button>
      <button class="btn btn-warn" id="btnPause" onclick="pauseScan()" style="display:none">⏸</button>
    </div>
  </div>

  <!-- Stats -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-v" id="stTotal" style="color:var(--tx2)">—</div>
      <div class="stat-l">Total IPs</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stDone" style="color:var(--c)">—</div>
      <div class="stat-l">Scanned</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stPass" style="color:var(--g)">0</div>
      <div class="stat-l">Passed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stFail" style="color:var(--r)">0</div>
      <div class="stat-l">Failed</div>
    </div>
    <div class="stat-card">
      <div class="stat-v" id="stETA" style="color:var(--y);font-size:20px">—</div>
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
        <span id="progTxt" style="font-family:var(--font-mono)">0 / 0</span>
        <span class="prog-pct" id="progPct">0%</span>
      </div>
      <div class="prog-wrap"><div class="prog-bar" id="progBar"></div></div>
      <!-- Quick settings -->
      <div style="display:grid;grid-template-columns:repeat(5,1fr);gap:8px;margin-top:4px">
        <div><label>Threads</label><input type="number" id="qThreads" value="200" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Timeout (s)</label><input type="number" id="qTimeout" value="8" min="1" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Max Lat (ms)</label><input type="number" id="qMaxLat" value="3500" style="font-size:11px;padding:4px 7px"></div>
        <div><label>P2 Rounds</label><input type="number" id="qRounds" value="3" min="0" style="font-size:11px;padding:4px 7px"></div>
        <div><label>Sample/Subnet</label><input type="number" id="sampleSize" value="1" min="1" style="font-size:11px;padding:4px 7px"></div>
      </div>
    </div>
  </div>

  <!-- IP Input + Feed -->
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
    <div class="card">
      <div class="card-hd"><div>IP Ranges</div><span id="feedCount" style="color:var(--dim)"></span></div>
      <div class="card-bd" style="padding:10px">
        <textarea id="ipInput" rows="7" placeholder="104.16.0.0/12&#10;162.158.0.0/15&#10;Or single IPs..."></textarea>
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
  <div id="tab-p2" class="card" style="border-radius:0 0 var(--rad) var(--rad)">
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
  <div id="tab-p1" class="card" style="display:none;border-radius:0 0 var(--rad) var(--rad)">
    <div class="card-hd">
      <div>Phase 1 — Initial Scan (passed only)</div>
      <span id="p1CountBadge" style="color:var(--dim);font-size:10px"></span>
    </div>
    <div class="tbl-wrap">
      <table class="tbl">
        <thead><tr>
          <th>#</th><th>IP Address</th><th>Latency</th><th>Pkt Loss</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody id="p1Tbody"><tr><td colspan="6" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr></tbody>
      </table>
    </div>
  </div>
</div>

<!-- ══ HISTORY PAGE ══ -->
<div id="page-history" class="page">
  <div class="phd">
    <div class="phd-l"><h2>History</h2><p>Previous scan sessions — saved locally</p></div>
    <div class="phd-r">
      <button class="btn btn-sm btn-danger-real" onclick="clearHistory()">✕ Clear History</button>
    </div>
  </div>
  <div id="histList"><p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p></div>
</div>

<!-- ══ CONFIG PAGE ══ -->
<div id="page-config" class="page">
  <div class="phd">
    <div class="phd-l"><h2>Settings</h2><p>Saved automatically to disk on save</p></div>
    <div class="phd-r">
      <button class="btn btn-success-real" onclick="saveConfig()">⬡ Save Settings</button>
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
      <button class="btn btn-danger-real btn-sm" onclick="clearSavedProxy()">✕ Remove Config</button>
    </div>
  </div>
  <div class="card">
    <div class="card-hd"><div>⬡ Proxy Link</div></div>
    <div class="card-bd">
      <div class="f-row"><label>vless:// or vmess:// or trojan://</label><textarea id="linkInput" rows="3" placeholder="vless://uuid@domain:443?..."></textarea></div>
      <button class="btn btn-primary-real" onclick="parseLink()">▶ Parse & Save</button>
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
        <div class="tui-dot" style="background:#ffd700"></div>
        <div class="tui-dot" style="background:#00ffaa"></div>
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
let currentTheme='night';
// localStorage key for history
const LS_HISTORY='pyz_history_v2';
const LS_THEME='pyz_theme_v2';

// ══ THEME ══
function toggleTheme(){
  currentTheme=currentTheme==='night'?'day':'night';
  applyTheme();
  localStorage.setItem(LS_THEME,currentTheme);
}
function applyTheme(){
  if(currentTheme==='day'){
    document.documentElement.setAttribute('data-theme','day');
    document.getElementById('themeIcon').textContent='🌙';
    document.getElementById('themeTxt').textContent='NIGHT';
  } else {
    document.documentElement.removeAttribute('data-theme');
    document.getElementById('themeIcon').textContent='☀';
    document.getElementById('themeTxt').textContent='DAY';
  }
}
(function(){
  const t=localStorage.getItem(LS_THEME);
  if(t){currentTheme=t;}
  applyTheme();
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
      // Reset progress for phase2
      document.getElementById('progBar').classList.add('p2');
      document.getElementById('progBar').style.width='0%';
      document.getElementById('progPct').textContent='0%';
      break;
    case 'phase2_progress':{
      // FIX: server sends phase2_progress, not phase2_result
      const r=payload;
      const done=r.done||0,total=r.total||1;
      const pct=Math.round(done/total*100);
      document.getElementById('progBar').style.width=pct+'%';
      document.getElementById('progPct').textContent=pct+'%';
      document.getElementById('progTxt').textContent=done+' / '+total+' (P2)';
      document.getElementById('tbProgress').textContent='P2: '+done+'/'+total+' · '+pct+'%';
      const passed=r.passed;
      const dl=r.dl||'—';const ul=r.ul||'—';
      const rowTxt=r.ip+' · '+Math.round(r.latency||0)+'ms  loss:'+Math.round(r.loss||0)+'%  ↓'+dl+' ↑'+ul;
      addFeedRow(rowTxt,passed?'p2':'fail');
      // Build p2 result obj to show live in table
      if(!viewingSession){
        const existing=p2Results.findIndex(x=>x.IP===r.ip);
        const obj={
          IP:r.ip,AvgLatencyMs:r.latency||0,JitterMs:r.jitter||0,
          PacketLossPct:r.loss||0,StabilityScore:r.score||0,
          Passed:r.passed,FailReason:r.failReason||'',
          DownloadMbps:parseFloat(r.dl)||0,UploadMbps:parseFloat(r.ul)||0
        };
        if(existing>=0) p2Results[existing]=obj; else p2Results.push(obj);
        renderP2();updatePassedChips();
      }
      break;
    }
    case 'phase2_done':{
      // Full results from server after phase2 completes
      if(payload.results&&!viewingSession){
        p2Results=payload.results;
        renderP2();updatePassedChips();
      }
      break;
    }
    case 'scan_done':
      setStatus('done','');
      addFeedRow('✓ Scan complete — '+payload.passed+' passed','ok');
      if(!viewingSession){refreshResults();}
      // Save session to localStorage history
      saveSessionToHistory(payload);
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
  if(st==='idle'){
    document.getElementById('progBar').style.width='0%';
    document.getElementById('progPct').textContent='0%';
    document.getElementById('progTxt').textContent='0 / 0';
    document.getElementById('tbProgress').textContent='';
  }
  if(st==='scanning'&&phase==='phase2') document.getElementById('progBar').classList.add('p2');
  else if(st!=='scanning') document.getElementById('progBar').classList.remove('p2');

  // Animate stat numbers when active
  ['stTotal','stDone','stPass','stFail'].forEach(id=>{
    const el=document.getElementById(id);
    if(el) el.classList.toggle('active',st==='scanning');
  });
}

// ══ PROGRESS ══
function onProgress(p){
  const pct=p.total>0?Math.round(p.done/p.total*100):0;
  document.getElementById('progBar').style.width=pct+'%';
  document.getElementById('progPct').textContent=pct+'%';
  document.getElementById('progTxt').textContent=(p.done||0)+' / '+(p.total||0);
  // FIX: properly update stat cards with numbers
  setStatValue('stTotal',p.total||'—','var(--tx2)');
  setStatValue('stDone',p.done||0,'var(--c)');
  setStatValue('stPass',p.succeeded||p.passed||0,'var(--g)');
  setStatValue('stFail',p.failed||0,'var(--r)');
  if(p.eta) setStatValue('stETA',p.eta,'var(--y)');
  document.getElementById('tbProgress').textContent=(p.done||0)+'/'+(p.total||0)+' · '+pct+'%';
  if(p.rate>0) document.getElementById('progRate').textContent=(p.rate||0).toFixed(1)+' IP/s';
}

function setStatValue(id,val,color){
  const el=document.getElementById(id);
  if(!el) return;
  el.textContent=val;
  if(color) el.style.color=color;
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
  // Reset stat cards
  setStatValue('stTotal','—','var(--tx2)');
  setStatValue('stDone',0,'var(--c)');
  setStatValue('stPass',0,'var(--g)');
  setStatValue('stFail',0,'var(--r)');
  setStatValue('stETA','—','var(--y)');
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
    // FIX: handle both string dl/ul from live feed AND numeric from API
    let dl=0,ul=0;
    if(typeof r.DownloadMbps==='number') dl=r.DownloadMbps;
    else if(typeof r.DownloadMbps==='string') dl=parseFloat(r.DownloadMbps)||0;
    if(typeof r.UploadMbps==='number') ul=r.UploadMbps;
    else if(typeof r.UploadMbps==='string') ul=parseFloat(r.UploadMbps)||0;
    const dlTxt=dl>0?dl.toFixed(1)+' M':'—';
    const ulTxt=ul>0?ul.toFixed(1)+' M':'—';
    const dlc=dl<=0?'var(--dim)':dl>=5?'var(--g)':dl>=1?'var(--y)':'var(--r)';
    const pl=r.PacketLossPct||0;
    const plc=pl<=5?'var(--g)':pl<=20?'var(--y)':'var(--r)';
    const jt=r.JitterMs||0;const jc=jt<=20?'var(--g)':jt<=80?'var(--y)':'var(--r)';
    return '<tr class="'+(r.Passed?'pass-row':'fail-row')+'">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700;font-size:12px;font-family:var(--font-mono)">'+r.IP+'</td>'+
      '<td style="color:'+scc+';font-weight:700;font-size:14px;font-family:var(--font-mono)">'+sc.toFixed(0)+'</td>'+
      '<td style="color:'+lc+';font-family:var(--font-mono)">'+Math.round(r.AvgLatencyMs||0)+'ms</td>'+
      '<td style="color:'+jc+';font-family:var(--font-mono)">'+(jt>0?jt.toFixed(0)+'ms':'—')+'</td>'+
      '<td style="color:'+plc+';font-family:var(--font-mono)">'+pl.toFixed(0)+'%</td>'+
      '<td style="color:'+dlc+';font-family:var(--font-mono)">'+dlTxt+'</td>'+
      '<td style="color:var(--tx2);font-family:var(--font-mono)">'+ulTxt+'</td>'+
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
    tbody.innerHTML='<tr><td colspan="6" style="text-align:center;color:var(--dim);padding:32px">No Phase 1 results</td></tr>';
    return;
  }
  tbody.innerHTML=succ.map((r,i)=>{
    const ip=r.ip||r.IP||'';const lat=r.latency_ms||r.LatencyMs||0;
    const lc=lat<=500?'var(--g)':lat<=1500?'var(--y)':'var(--r)';
    // Phase 1 packet loss (may not be available, show — if 0)
    const pl=r.packet_loss_pct||r.PacketLossPct||0;
    const plTxt=pl>0?pl.toFixed(0)+'%':'—';
    const plc=pl<=0?'var(--dim)':pl<=5?'var(--g)':pl<=20?'var(--y)':'var(--r)';
    return '<tr class="p1-row">'+
      '<td style="color:var(--dim);font-size:10px">'+(i+1)+'</td>'+
      '<td style="color:var(--c);font-weight:700;font-family:var(--font-mono)">'+ip+'</td>'+
      '<td style="color:'+lc+';font-family:var(--font-mono)">'+Math.round(lat)+'ms</td>'+
      '<td style="color:'+plc+';font-family:var(--font-mono)">'+plTxt+'</td>'+
      '<td><span class="badge bg">OK</span></td>'+
      '<td><button class="copy-btn" onclick="copyIP(\''+ip+'\')" title="Copy IP">⎘</button></td>'+
    '</tr>';
  }).join('');
  if(p1Results.length>succ.length) tbody.innerHTML+='<tr><td colspan="6" style="text-align:center;color:var(--dim);padding:10px;font-size:10px">'+(p1Results.length-succ.length)+' failed IPs hidden</td></tr>';
}

function copyAllPassed(){
  const passed=(p2Results||[]).filter(r=>r.Passed).map(r=>r.IP);
  if(!passed.length) return;
  navigator.clipboard.writeText(passed.join('\n')).then(()=>appendTUI({t:now(),l:'ok',m:'⎘ '+passed.length+' IPs copied'}));
}

// ══ HISTORY — localStorage based ══
function saveSessionToHistory(payload){
  // Also fetch current results to store
  fetch('/api/results').then(r=>r.json()).then(data=>{
    let history=loadHistory();
    const session={
      id:Date.now().toString(),
      startedAt:new Date().toISOString(),
      totalIPs:payload.total||0,
      passed:payload.passed||0,
      duration:payload.duration||'',
      results:data.phase2||[],
      p1Results:data.phase1||[]
    };
    history.unshift(session);
    // Keep last 50 sessions
    if(history.length>50) history=history.slice(0,50);
    try{localStorage.setItem(LS_HISTORY,JSON.stringify(history));}catch(e){console.warn('History save error',e);}
    updateHistoryBadge(history.length);
  }).catch(()=>{
    // Save minimal info if results fetch fails
    let history=loadHistory();
    const session={
      id:Date.now().toString(),
      startedAt:new Date().toISOString(),
      totalIPs:payload.total||0,
      passed:payload.passed||0,
      duration:payload.duration||'',
      results:[],p1Results:[]
    };
    history.unshift(session);
    if(history.length>50) history=history.slice(0,50);
    try{localStorage.setItem(LS_HISTORY,JSON.stringify(history));}catch(e){}
    updateHistoryBadge(history.length);
  });
}

function loadHistory(){
  try{
    const raw=localStorage.getItem(LS_HISTORY);
    return raw?JSON.parse(raw):[];
  }catch(e){return [];}
}

function updateHistoryBadge(count){
  const el=document.getElementById('nbHistory');
  if(el) el.textContent=count>0?count:'';
}

function refreshHistory(){
  // First try server sessions, then fallback to localStorage
  fetch('/api/sessions').then(r=>r.json()).then(sessions=>{
    let merged=loadHistory();
    // Merge server sessions into local (server sessions may have more data)
    // We prefer local since they persist across restarts
    renderHistoryList(merged.length>0?merged:sessions);
  }).catch(()=>{
    renderHistoryList(loadHistory());
  });
}

function renderHistoryList(sessions){
  const el=document.getElementById('histList');
  updateHistoryBadge(sessions.length);
  if(!sessions||!sessions.length){
    el.innerHTML='<p style="color:var(--dim);font-family:var(--font-mono);font-size:12px">No scans yet</p>';
    return;
  }
  el.innerHTML=sessions.map(s=>{
    const passed=s.passed||(s.results||[]).filter(r=>r.Passed).length;
    const d=new Date(s.startedAt);
    const total=s.totalIPs||(s.results||[]).length;
    return '<div class="hist-item" onclick="showSession(\''+s.id+'\',\'local\')">'+
      '<div class="hist-n" style="color:'+(passed>0?'var(--g)':'var(--dim)')+'">'+passed+'</div>'+
      '<div class="hist-info">'+
        '<div style="color:var(--tx);font-weight:600;font-size:12px">'+total+' IPs · '+passed+' passed</div>'+
        '<div class="hist-date">'+d.toLocaleString()+(s.duration?' · '+s.duration:'')+'</div>'+
      '</div>'+
      '<div style="color:var(--dim);font-size:10px">▶</div>'+
    '</div>';
  }).join('');
}

function showSession(id,source){
  let sessions=loadHistory();
  let s=sessions.find(x=>x.id===id);
  if(!s){
    // Try server
    fetch('/api/sessions').then(r=>r.json()).then(srvSessions=>{
      const ss=srvSessions.find(x=>x.id===id);
      if(ss) _showSessionData(ss);
    });
    return;
  }
  _showSessionData(s);
}

function _showSessionData(s){
  viewingSession=true;
  p2Results=s.results||[];p1Results=s.p1Results||[];
  renderP2();renderP1();updatePassedChips();
  nav('results');
  const existing=document.getElementById('sessionBanner');if(existing)existing.remove();
  const banner=document.createElement('div');
  banner.id='sessionBanner';banner.className='session-banner';
  const d=new Date(s.startedAt);
  const passed=s.passed||(s.results||[]).filter(r=>r.Passed).length;
  banner.innerHTML='<span>📂 Viewing: '+d.toLocaleString()+' — '+passed+' passed</span>'+
    '<button onclick="clearSession()" style="background:var(--rd);border:1px solid var(--r);color:var(--r);padding:3px 10px;cursor:pointer;border-radius:3px;font-size:11px;font-family:var(--font-mono)">✕ Back to live</button>';
  document.getElementById('page-results').insertBefore(banner,document.getElementById('page-results').firstChild);
}

function clearSession(){
  viewingSession=false;
  const b=document.getElementById('sessionBanner');if(b)b.remove();
  refreshResults();
}

function clearHistory(){
  if(!confirm('Clear all history?')) return;
  localStorage.removeItem(LS_HISTORY);
  updateHistoryBadge(0);
  refreshHistory();
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
  // Load history badge
  const hist=loadHistory();
  updateHistoryBadge(hist.length);
}

// ══ INIT ══
connectWS();
fetch('/api/status').then(r=>r.json()).then(d=>setStatus(d.status||'idle',d.phase||''));
loadSavedSettings();
</script>
</body>
</html>
`
