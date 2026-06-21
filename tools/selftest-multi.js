#!/usr/bin/env node
// selftest-multi — autonomous E2E browser test across MANY phones + the QVH app.
// For each UDID: loads dev.aliremote.com/#/phone/<udid> with the operator's session,
// verifies login + that the MJPEG <img>#remoteVideo actually renders (naturalWidth>0 &&
// complete), and times how long until first render. Emits one machine-readable JSON line
// per phone (prefixed RESULT:) plus a final SUMMARY: line, so an agent can self-test a
// whole pod without a human. Also probes the QVH Cloudflare app once.
//   Usage: NODE_PATH=/tmp/pw/node_modules node selftest-multi.js <udid1> <udid2> ...
//          (or pass a file of udids:  --file /tmp/udids.txt)
// Cookies: /tmp/pw_cookies.json (live operator auth — local-only, never commit).
const { chromium } = require('playwright');
const fs = require('fs');

let udids = process.argv.slice(2);
const fi = udids.indexOf('--file');
if (fi !== -1) udids = fs.readFileSync(udids[fi + 1], 'utf8').split(/\s+/).filter(Boolean);
if (!udids.length) udids = ['3fe1e72be1eb12b818654882fa3cbe939c97a4c2'];

const cookies = JSON.parse(fs.readFileSync('/tmp/pw_cookies.json', 'utf8'));

(async () => {
  const browser = await chromium.launch({ args: ['--autoplay-policy=no-user-gesture-required'] });
  const ctx = await browser.newContext({ viewport: { width: 1400, height: 900 } });
  let ck = 0;
  for (const c of cookies) {
    const cc = { ...c };
    if (cc.sameSite === 'None' && !cc.secure) cc.sameSite = 'Lax';
    try { await ctx.addCookies([cc]); ck++; } catch (e) { try { await ctx.addCookies([{ ...cc, sameSite: 'Lax' }]); ck++; } catch (e2) {} }
  }
  console.error('cookies injected:', ck + '/' + cookies.length, '| phones:', udids.length);

  const results = [];
  for (let i = 0; i < udids.length; i++) {
    const udid = udids[i];
    const r = { udid: udid.slice(0, 8), loggedIn: false, render: false, w: 0, h: 0, renderMs: null, fps: null, err: null };
    // Fresh page per phone: MJPEG is a long-lived <img> connection and the SPA's
    // hash-route nav leaks it, so a reused page hits the browser's ~6-per-host
    // connection cap and phones 7+ falsely report render:false. A new page (closed
    // after) tears the connection down and gives each phone a clean slot.
    const page = await ctx.newPage();
    try {
      await page.goto('https://dev.aliremote.com/#/phone/' + udid, { waitUntil: 'load', timeout: 40000 });
      // poll up to 12s for the stream <img> to actually decode a frame
      const t0 = Date.now();
      let frame = null;
      for (let t = 0; t < 24; t++) {
        frame = await page.evaluate(() => {
          const loggedIn = !/login|sign|auth/i.test(location.hash + location.pathname);
          const img = document.querySelector('#remoteVideo') || document.querySelector('img[src*="stream"]');
          const w = img ? (img.naturalWidth || img.videoWidth || 0) : 0;
          const h = img ? (img.naturalHeight || img.videoHeight || 0) : 0;
          const ok = !!img && (img.complete !== false) && w > 0;
          return { loggedIn, has: !!img, w, h, ok };
        });
        if (frame.loggedIn === false) break;
        if (frame.ok) { r.renderMs = Date.now() - t0; break; }
        await page.waitForTimeout(500);
      }
      r.loggedIn = frame.loggedIn;
      r.render = frame.ok;
      r.w = frame.w; r.h = frame.h;
      // measure refresh: sample naturalWidth/src change isn't reliable for MJPEG; instead
      // count decoded-frame updates over 2s via the img's decode timestamps.
      if (frame.ok) {
        const fr = await page.evaluate(async () => {
          const img = document.querySelector('#remoteVideo') || document.querySelector('img[src*="stream"]');
          if (!img) return 0;
          let n = 0; const seen = new Set();
          const t0 = performance.now();
          while (performance.now() - t0 < 2000) {
            // MJPEG multipart updates the img in place; use a cheap heuristic: decode() resolves per current frame
            try { await img.decode(); n++; } catch (e) {}
            await new Promise(r => setTimeout(r, 50));
          }
          return Math.round(n / 2);
        });
        r.fps = fr;
      }
    } catch (e) { r.err = e.message.slice(0, 80); }
    results.push(r);
    console.log('RESULT:' + JSON.stringify(r));
    if (i === 0) await page.screenshot({ path: '/tmp/st_dash.png' });
    else if (i === udids.length - 1) await page.screenshot({ path: '/tmp/st_dash_last.png' });
    await page.close();
  }

  // QVH app probe
  const page = await ctx.newPage();
  let qvh = { err: null };
  try {
    await page.goto('https://qvh.turbocat.dk/', { waitUntil: 'load', timeout: 30000 });
    await page.waitForTimeout(3500);
    qvh = await page.evaluate(() => ({
      title: document.title, cards: document.querySelectorAll('.card').length,
      count: (document.getElementById('count') || {}).textContent || '',
      sample: (document.body.innerText || '').slice(0, 70).replace(/\n/g, ' ')
    }));
    await page.screenshot({ path: '/tmp/st_qvh.png' });
  } catch (e) { qvh.err = e.message.slice(0, 80); }
  console.log('QVH:' + JSON.stringify(qvh));

  const rendered = results.filter(r => r.render).length;
  const renderMsVals = results.filter(r => r.renderMs != null).map(r => r.renderMs);
  const avgMs = renderMsVals.length ? Math.round(renderMsVals.reduce((a, b) => a + b, 0) / renderMsVals.length) : null;
  console.log('SUMMARY:' + JSON.stringify({ total: results.length, loggedIn: results.filter(r => r.loggedIn).length, rendered, avgRenderMs: avgMs, broken: results.filter(r => !r.render).map(r => r.udid) }));
  await browser.close();
})().catch(e => { console.error('HARNESS ERR', e.message); process.exit(2); });
