#!/usr/bin/env node
// selftest — autonomous E2E test of the ALI dashboard (with the operator's session) + the QVH app.
// Drives dev.aliremote.com phone pages (auth, stream render, tap-on-screen) and qvh.turbocat.dk,
// reports findings + screenshots, so the agent can self-test without a human in the loop.
// Usage: node selftest.js <udid>   (cookies from /tmp/pw_cookies.json)
const { chromium } = require('playwright');
const fs = require('fs');
const cookies = JSON.parse(fs.readFileSync('/tmp/pw_cookies.json', 'utf8'));
const udid = process.argv[2] || '3fe1e72be1eb12b818654882fa3cbe939c97a4c2';

(async () => {
  const browser = await chromium.launch({ args: ['--autoplay-policy=no-user-gesture-required'] });
  const ctx = await browser.newContext({ viewport: { width: 1500, height: 950 } });
  // inject session cookies (robust: per-cookie, fix bad sameSite combos)
  let ok = 0;
  for (const c of cookies) {
    const cc = { ...c };
    if (cc.sameSite === 'None' && !cc.secure) cc.sameSite = 'Lax';
    try { await ctx.addCookies([cc]); ok++; }
    catch (e) { try { await ctx.addCookies([{ ...cc, sameSite: 'Lax' }]); ok++; } catch (e2) {} }
  }
  console.log('cookies injected:', ok + '/' + cookies.length);
  const page = await ctx.newPage();
  const logs = [];
  page.on('console', m => logs.push('C/' + m.type() + ':' + m.text().slice(0, 140)));
  page.on('pageerror', e => logs.push('PAGEERR:' + e.message.slice(0, 140)));
  page.on('requestfailed', r => { const u = r.url(); if (/aliremote|turbocat|stream/.test(u)) logs.push('REQFAIL:' + u.split('/').slice(2, 4).join('/') + ' ' + ((r.failure() || {}).errorText || '')); });

  // ===== 1. Dashboard phone page =====
  console.log('\n=== DASHBOARD: dev.aliremote.com/#/phone/' + udid.slice(0, 8) + ' ===');
  await page.goto('https://dev.aliremote.com/#/phone/' + udid, { waitUntil: 'load', timeout: 45000 });
  await page.waitForTimeout(9000);
  const dash = await page.evaluate(() => {
    const loggedIn = !/login|sign|auth/i.test(location.hash + location.pathname);
    const img = document.querySelector('#remoteVideo') || document.querySelector('img[src*="stream"]');
    const stream = img ? { tag: img.tagName, src: (img.src || img.currentSrc || '').slice(0, 90), w: img.naturalWidth || img.videoWidth, h: img.naturalHeight || img.videoHeight, complete: img.complete } : null;
    return { url: location.href, loggedIn, stream, bodyLen: document.body.innerText.length };
  });
  console.log('url:', dash.url);
  console.log('loggedIn:', dash.loggedIn, '| stream:', JSON.stringify(dash.stream));
  await page.screenshot({ path: '/tmp/st_dash.png' });
  // tap test: click the phone screen + watch for a control request
  if (dash.stream) {
    const reqs = [];
    page.on('request', r => { if (/control|tap|command|ws/i.test(r.url())) reqs.push(r.url().split('?')[0].slice(0, 60)); });
    try { await page.click('#remoteVideo', { position: { x: 100, y: 200 }, timeout: 4000 }); } catch (e) {}
    await page.waitForTimeout(2500);
    console.log('tap -> control requests seen:', reqs.length ? reqs.slice(0, 3) : '(none — control likely over WS)');
  }

  // ===== 2. QVH Cloudflare app =====
  console.log('\n=== QVH APP: qvh.turbocat.dk ===');
  await page.goto('https://qvh.turbocat.dk/', { waitUntil: 'load', timeout: 30000 });
  await page.waitForTimeout(4000);
  const qvh = await page.evaluate(() => ({
    title: document.title, cards: document.querySelectorAll('.card').length,
    streaming: (document.getElementById('count') || {}).textContent || '',
    sample: (document.body.innerText || '').slice(0, 80).replace(/\n/g, ' ')
  }));
  console.log(JSON.stringify(qvh));
  await page.screenshot({ path: '/tmp/st_qvh.png' });

  console.log('\n=== browser logs (last 12) ==='); logs.slice(-12).forEach(l => console.log(' ', l));
  await browser.close();
})().catch(e => { console.error('HARNESS ERR', e.message); process.exit(2); });
