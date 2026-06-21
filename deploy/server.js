// ALI QuickTime Live — multi-phone stream server (ideacatlab POC)
// Serves a dashboard + per-phone HLS with CORRECT mime (video/mp2t) and no-cache,
// and an API to list phones (from the pod ledger), start/stop a phone's QVH capture
// on demand, and wake its screen via WDA. No npm deps — Node built-ins only.
//
// DECOUPLED A/V: ONE qvh writes video->h264 fifo AND audio->.wav FILE. The VIDEO pipeline
// (qvh fifo -> ffmpeg -> stream.m3u8) is untouched and never gated on audio. A SEPARATE,
// independent AUDIO pipeline reads the .wav FILE (file reads never block qvh) -> feeder ->
// audio ffmpeg -> audio.m3u8 (AAC). If audio ever stalls/dies, video is completely unaffected.
const http = require('http');
const { spawn, execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const SERVE = '/root/qvh-stream/serve';
const QVH = '/root/qvh-test/qvh';
const FEEDER = '/root/qvh-stream/audio_feeder.py';
const POD = { host: '127.0.0.1', port: 7799 };
const captures = {}; // udid -> {qvh, ffmpeg, fifo, ka, afeeder, affmpeg, fedfifo, wav, startedAt}

fs.mkdirSync(SERVE, { recursive: true });

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
function wake(udid, clientPort) {
  if (!clientPort) return;
  const req = http.request({ host: '127.0.0.1', port: clientPort, path: '/wda/homescreen', method: 'POST', headers: { 'content-type': 'application/json' }, timeout: 5000 }, r => r.resume());
  req.on('error', () => {}); req.on('timeout', () => req.destroy()); req.end('{}');
}
function tap(udid) {
  const body = JSON.stringify({ coords: { normX: 0.5, normY: 0.01 } });
  const req = http.request({ ...POD, path: '/admin/control/tap?udid=' + udid, method: 'POST', headers: { 'content-type': 'application/json', 'content-length': Buffer.byteLength(body) }, timeout: 5000 }, r => r.resume());
  req.on('error', () => {}); req.on('timeout', () => req.destroy()); req.end(body);
}
function startCapture(udid, clientPort) {
  if (captures[udid]) return;
  const dir = path.join(SERVE, udid);
  fs.mkdirSync(dir, { recursive: true });
  for (const f of fs.readdirSync(dir)) if (/seg_|stream\.m3u8|aud_|audio\.m3u8/.test(f)) try { fs.unlinkSync(path.join(dir, f)); } catch (e) {}
  const fifo = '/tmp/qvh_' + udid + '.h264';
  const wav = '/tmp/qvh_' + udid + '.wav';
  const fedfifo = '/tmp/qvh_' + udid + '_fed.pcm';
  try { fs.unlinkSync(fifo); } catch (e) {}
  try { fs.unlinkSync(wav); } catch (e) {}
  try { execSync('mkfifo ' + fifo); } catch (e) {}
  // ---- VIDEO pipeline (unchanged, never gated on audio) ----
  const ff = spawn('ffmpeg', ['-y', '-loglevel', 'error', '-fflags', '+genpts', '-r', '30', '-f', 'h264', '-i', fifo,
    '-an', '-vf', 'scale=-2:1024,format=yuv420p', '-c:v', 'libx264', '-preset', 'veryfast', '-tune', 'zerolatency',
    '-profile:v', 'high', '-g', '30', '-keyint_min', '30', '-f', 'hls', '-hls_time', '1', '-hls_list_size', '8',
    '-hls_flags', 'delete_segments+omit_endlist+independent_segments', '-hls_segment_filename', path.join(dir, 'seg_%05d.ts'),
    path.join(dir, 'stream.m3u8')], { stdio: 'ignore' });
  // ONE qvh: video -> fifo, audio -> wav FILE (file write never blocks -> video safe)
  const qv = spawn(QVH, ['record', fifo, wav, '--udid=' + udid], { stdio: 'ignore' });
  const ka = setInterval(() => tap(udid), 15000);
  // ---- AUDIO pipeline (SEPARATE & independent; a failure here cannot touch video) ----
  let afeeder = null, affmpeg = null;
  try {
    try { fs.unlinkSync(fedfifo); } catch (e) {}
    execSync('mkfifo ' + fedfifo);
    afeeder = spawn('python3', [FEEDER, wav, fedfifo], { stdio: 'ignore' });
    afeeder.on('error', () => {});
    affmpeg = spawn('ffmpeg', ['-y', '-loglevel', 'error', '-thread_queue_size', '4096', '-f', 's16le', '-ar', '48000', '-ac', '2', '-i', fedfifo,
      '-c:a', 'aac', '-b:a', '128k', '-f', 'hls', '-hls_time', '1', '-hls_list_size', '8',
      '-hls_flags', 'delete_segments+omit_endlist', '-hls_segment_filename', path.join(dir, 'aud_%05d.ts'),
      path.join(dir, 'audio.m3u8')], { stdio: 'ignore' });
    affmpeg.on('error', () => {});
  } catch (e) { /* audio failure must never affect video */ }
  captures[udid] = { qvh: qv, ffmpeg: ff, fifo, wav, ka, afeeder, affmpeg, fedfifo, startedAt: Date.now() };
  wake(udid, clientPort);
  qv.on('exit', () => { const c = captures[udid]; if (c && c.qvh === qv) { try { c.ffmpeg.kill(); } catch (e) {} try { c.afeeder && c.afeeder.kill(); } catch (e) {} try { c.affmpeg && c.affmpeg.kill(); } catch (e) {} clearInterval(c.ka); delete captures[udid]; } });
}
function stopCapture(udid) {
  const c = captures[udid]; if (!c) return;
  clearInterval(c.ka);
  try { c.qvh.kill('SIGKILL'); } catch (e) {}
  try { c.ffmpeg.kill('SIGKILL'); } catch (e) {}
  try { c.afeeder && c.afeeder.kill('SIGKILL'); } catch (e) {}
  try { c.affmpeg && c.affmpeg.kill('SIGKILL'); } catch (e) {}
  try { fs.unlinkSync(c.fifo); } catch (e) {}
  try { fs.unlinkSync(c.fedfifo); } catch (e) {}
  delete captures[udid];
}

const MIME = { '.m3u8': 'application/vnd.apple.mpegurl', '.ts': 'video/mp2t', '.html': 'text/html; charset=utf-8', '.js': 'application/javascript', '.css': 'text/css', '.json': 'application/json' };
const NOCACHE = { 'cache-control': 'no-store, no-cache, must-revalidate', 'access-control-allow-origin': '*' };

http.createServer((req, res) => {
  const u = new URL(req.url, 'http://x');
  const p = u.pathname;
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
