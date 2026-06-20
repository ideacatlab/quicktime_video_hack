#!/usr/bin/env node
// playwright-verify — load the QVH dashboard/stream in a real headless browser, force
// playback, and report whether video actually DECODES (videoWidth>0) + capture console
// errors + a screenshot. The end-to-end client-side check (catches MIME/CORS/hls.js bugs
// that pod-side segment checks miss).
// Usage: node playwright-verify.js <url> [out.png] [waitSeconds]
//   needs: npm i playwright  (browsers via `npx playwright install chromium`)
const { chromium } = require('playwright');
(async () => {
  const url = process.argv[2] || 'https://qvh.turbocat.dk';
  const out = process.argv[3] || '/tmp/qvh_verify.png';
  const waitS = parseInt(process.argv[4] || '14', 10);
  const browser = await chromium.launch({ args: ['--autoplay-policy=no-user-gesture-required'] });
  const page = await browser.newPage({ viewport: { width: 900, height: 1300 } });
  const logs = [];
  page.on('console', m => logs.push('C/' + m.type() + ': ' + m.text()));
  page.on('pageerror', e => logs.push('PAGEERR: ' + e.message));
  page.on('requestfailed', r => logs.push('REQFAIL: ' + r.url().split('/').pop() + ' ' + ((r.failure() || {}).errorText || '')));
  await page.goto(url, { waitUntil: 'load', timeout: 30000 });
  // force any <video> to play
  const deadline = Date.now() + waitS * 1000;
  let info = {};
  while (Date.now() < deadline) {
    info = await page.evaluate(async () => {
      const v = document.querySelector('video');
      if (!v) return { noVideo: true };
      try { await v.play(); } catch (e) {}
      return { rs: v.readyState, vw: v.videoWidth, vh: v.videoHeight, ct: +v.currentTime.toFixed(2), paused: v.paused, err: v.error ? v.error.code : null, buffered: v.buffered.length };
    });
    if (info.vw > 0 && info.rs >= 2) break;
    await page.waitForTimeout(1000);
  }
  const verdict = (info.vw > 0 && info.rs >= 2) ? 'PASS-video-decoding' : 'FAIL-no-video';
  console.log('VERDICT: ' + verdict);
  console.log('VIDEO_STATE: ' + JSON.stringify(info));
  if (logs.length) { console.log('--- browser logs (last 15) ---'); logs.slice(-15).forEach(l => console.log('  ' + l)); }
  await page.screenshot({ path: out });
  console.log('screenshot: ' + out);
  await browser.close();
  process.exit(verdict.startsWith('PASS') ? 0 : 1);
})().catch(e => { console.error('ERR', e.message); process.exit(2); });
