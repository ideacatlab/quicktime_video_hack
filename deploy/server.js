// ALI QuickTime Live — multi-phone stream server (ideacatlab POC)
// Serves a dashboard + per-phone HLS with CORRECT mime (video/mp2t) and no-cache,
// and an API to list phones (from the pod ledger), start/stop a phone's QVH capture
// on demand, and wake its screen via WDA. No npm deps — Node built-ins only.
//
// Runs under systemd `qvh-server.service` (Restart=always, RestartSec=3, enabled at
// boot) so a crash self-heals in <5s. RESILIENCE RULE for cleanup scripts: kill only
// the per-capture children (`qvh record …` / `ffmpeg …`), NEVER `systemctl stop
// qvh-server` or `pkill node` — that takes the dashboard offline (was the 502 cause).
// Liveness probe: GET /healthz (cheap, no pod call). A per-capture watchdog reaps any
// half-dead capture (ffmpeg died but qvh alive, or vice-versa) so the slot frees.
const http = require('http');
const { spawn, execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const SERVE = '/root/qvh-stream/serve';
const QVH = '/root/qvh-test/qvh';
const POD = { host: '127.0.0.1', port: 7799 };
const captures = {}; // udid -> {qvh, ffmpeg, fifo, ka, startedAt}
const startedAt = Date.now();

fs.mkdirSync(SERVE, { recursive: true });

// Reap half-dead captures: if either child (qvh or ffmpeg) has exited but the entry
// lingers, tear the whole capture down so the slot frees and a restart can re-arm.
setInterval(() => {
  for (const udid of Object.keys(captures)) {
    const c = captures[udid];
    const dead = (c.qvh && c.qvh.exitCode !== null) || (c.ffmpeg && c.ffmpeg.exitCode !== null);
    if (dead) { try { stopCapture(udid); } catch (e) {} }
  }
}, 10000).unref();

function podState(cb) {
  http.get({ ...POD, path: '/state', timeout: 6000 }, r => {
    let d = ''; r.on('data', c => d += c); r.on('end', () => { try { cb(JSON.parse(d)); } catch (e) { cb(null); } });
  }).on('error', () => cb(null));
}
function phones(cb) {
  podState(s => {
    if (!s) return cb([]);
    let v = s.devices || s; v = Array.isArray(v) ? v : Object.values(v);
    cb(v.filter(x => x && x.udid).map(x => ({
      udid: x.udid, short: x.udid.slice(0, 8), slot: x.portIndex,
      state: (x.state || {}).kind, clientPort: (x.state || {}).clientPort,
      streaming: !!captures[x.udid]
    })).sort((a, b) => (a.slot || 0) - (b.slot || 0)));
  });
}
function wake(udid, clientPort) { // home-button press: wakes a slept display (Wake button + initial)
  if (!clientPort) return;
  const req = http.request({ host: '127.0.0.1', port: clientPort, path: '/wda/homescreen', method: 'POST', headers: { 'content-type': 'application/json' }, timeout: 5000 }, r => r.resume());
  req.on('error', () => {}); req.on('timeout', () => req.destroy()); req.end('{}');
}
function tap(udid) { // gentle status-bar tap: keeps an AWAKE screen awake without leaving the current app
  const body = JSON.stringify({ coords: { normX: 0.5, normY: 0.01 } });
  const req = http.request({ ...POD, path: '/admin/control/tap?udid=' + udid, method: 'POST', headers: { 'content-type': 'application/json', 'content-length': Buffer.byteLength(body) }, timeout: 5000 }, r => r.resume());
  req.on('error', () => {}); req.on('timeout', () => req.destroy()); req.end(body);
}
function startCapture(udid, clientPort) {
  if (captures[udid]) return;
  const dir = path.join(SERVE, udid);
  fs.mkdirSync(dir, { recursive: true });
  for (const f of fs.readdirSync(dir)) if (/seg_|stream\.m3u8/.test(f)) try { fs.unlinkSync(path.join(dir, f)); } catch (e) {}
  const fifo = '/tmp/qvh_' + udid + '.h264';
  const afifo = '/tmp/qvh_' + udid + '.pcm';
  for (const f of [fifo, afifo]) { try { fs.unlinkSync(f); } catch (e) {} try { execSync('mkfifo ' + f); } catch (e) {} }
  // mux QVH H.264 video (fifo) + QVH raw PCM audio (s16le 48kHz stereo, from the AFMT handshake) -> HLS w/ AAC.
  const ff = spawn('ffmpeg', ['-y', '-loglevel', 'error', '-fflags', '+genpts',
    '-r', '30', '-f', 'h264', '-i', fifo,
    '-f', 's16le', '-ar', '48000', '-ac', '2', '-i', afifo,
    '-map', '0:v:0', '-map', '1:a:0',
    '-c:v', 'copy',
    '-c:a', 'aac', '-b:a', '128k', '-ar', '48000', '-ac', '2',
    '-f', 'hls', '-hls_time', '1', '-hls_list_size', '8',
    '-hls_flags', 'delete_segments+omit_endlist+independent_segments', '-hls_segment_filename', path.join(dir, 'seg_%05d.ts'),
    path.join(dir, 'stream.m3u8')], { stdio: 'ignore' });
  // QVH writes H.264 to fifo + raw PCM audio to afifo. NO_RECYCLE = catch the natural CWPA (reboot-fresh devices).
  const qv = spawn(QVH, ['record', fifo, afifo, '--udid=' + udid], { stdio: 'ignore', env: { ...process.env, QVH_NO_RECYCLE: '1' } });
  const ka = setInterval(() => tap(udid), 15000);
  captures[udid] = { qvh: qv, ffmpeg: ff, fifo, afifo, ka, startedAt: Date.now() };
  wake(udid, clientPort);
  qv.on('exit', () => { const c = captures[udid]; if (c && c.qvh === qv) { try { c.ffmpeg.kill(); } catch (e) {} clearInterval(c.ka); delete captures[udid]; } });
}
function stopCapture(udid) {
  const c = captures[udid]; if (!c) return;
  clearInterval(c.ka);
  try { c.qvh.kill('SIGKILL'); } catch (e) {}
  try { c.ffmpeg.kill('SIGKILL'); } catch (e) {}
  try { fs.unlinkSync(c.fifo); } catch (e) {}
  try { fs.unlinkSync(c.afifo); } catch (e) {}
  delete captures[udid];
}

const MIME = { '.m3u8': 'application/vnd.apple.mpegurl', '.ts': 'video/mp2t', '.html': 'text/html; charset=utf-8', '.js': 'application/javascript', '.css': 'text/css', '.json': 'application/json' };
const NOCACHE = { 'cache-control': 'no-store, no-cache, must-revalidate', 'access-control-allow-origin': '*' };

http.createServer((req, res) => {
  const u = new URL(req.url, 'http://x');
  const p = u.pathname;
  if (p === '/healthz') { res.writeHead(200, { 'content-type': 'application/json', ...NOCACHE }); res.end(JSON.stringify({ ok: true, captures: Object.keys(captures).length, uptimeS: Math.round((Date.now() - startedAt) / 1000) })); return; }
  if (p === '/api/phones') { phones(list => { res.writeHead(200, { 'content-type': 'application/json', ...NOCACHE }); res.end(JSON.stringify(list)); }); return; }
  if (p === '/api/start') { startCapture(u.searchParams.get('udid'), u.searchParams.get('port')); res.writeHead(200, NOCACHE); res.end('{"ok":true}'); return; }
  if (p === '/api/stop') { stopCapture(u.searchParams.get('udid')); res.writeHead(200, NOCACHE); res.end('{"ok":true}'); return; }
  if (p === '/api/wake') { wake(u.searchParams.get('udid'), u.searchParams.get('port')); res.writeHead(200, NOCACHE); res.end('{"ok":true}'); return; }
  let fp = path.join(SERVE, p === '/' ? '/index.html' : decodeURIComponent(p));
  if (!fp.startsWith(SERVE)) { res.writeHead(403); res.end(); return; }
  fs.readFile(fp, (e, data) => {
    if (e) { res.writeHead(404, NOCACHE); res.end('not found'); return; }
    res.writeHead(200, { 'content-type': MIME[path.extname(fp)] || 'application/octet-stream', ...NOCACHE });
    res.end(data);
  });
}).listen(8099, '127.0.0.1', () => console.log('qvh multi-phone server on 127.0.0.1:8099'));
