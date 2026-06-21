# QVH testing & diagnostic tooling (ideacatlab)

Fast feedback loop for the QuickTime-over-USB stream POC. Designed to be **read-mostly**
and **sequential** (never hammer usbmuxd in parallel — that causes the very contention
that breaks arming).

## Tools

### `qvh-probe.sh <udid> [seconds] [qvh_path]`
Runs `qvh record` briefly and classifies ONE phone's AV arming state:
| state | meaning | action |
|---|---|---|
| `FRESH-VIDEO` | H.264 frames flowing (CVRP+FEED) | streamable ✅ |
| `AUDIO-ONLY-stuck` | CWPA/audio but no video | stuck half-open AV session |
| `TIME-ONLY-stuck` | only PING + SYNC_TIME | partial arm |
| `NO-AV-IFACE` | on config-4 / no Valeria iface | not in mode-2 |
| `NO-PING/dead` | no response | dead / USB-contended |

### `qvh-reliability.sh <udid> [N] [clientPort]`
N probes (optionally home-press/wake first) → state distribution + **FRESH-VIDEO success
rate**. This is the metric to drive iteration on the arming fix.

### `fleet-av-scan.sh [max] ` (env `POD_ADMIN`)
Scans every attached phone (sequentially) → fleet-wide AV-state summary. Answers "how many
phones can stream right now".

### `playwright-verify.js <url> [out.png] [waitS]`
Loads the dashboard/stream in headless Chromium, force-plays, and verifies the video
actually **decodes** client-side (`videoWidth>0`) — catches MIME/CORS/hls.js bugs that
pod-side segment checks miss. Captures console errors + a screenshot.
`npm i playwright && npx playwright install chromium`, then `node playwright-verify.js https://qvh.turbocat.dk`.

## Iteration loop
1. `fleet-av-scan.sh` → baseline (how many FRESH vs stuck).
2. Apply a change (e.g. `QVH_RESET_MODE_TRANSITION=1` for the SET_MODE 1→2 re-arm, or a
   usbmuxd fork change).
3. `qvh-reliability.sh <udid> 10 <port>` on a stuck phone → success-rate delta.
4. `playwright-verify.js` → confirm end-to-end client playback.
5. Record the result in `FINDINGS_IDEACATLAB.md`.

## ⚠️ Run on a dedicated rig, not a control-serving pod
Repeated probing + direct libusb access contends with usbmuxd and degrades the WDA control
path (measured: control lag + `guess_mode` failing so mode-2 stops applying fleet-wide).
Use a spare Linux box + 2–3 jailbroken iOS≤16 phones where usbmuxd can be stopped.

## selftest-multi.js — many-phone autonomous E2E (machine-readable)
Loops `selftest.js`'s dashboard check over a LIST of phones and emits one `RESULT:{json}`
line per phone (`render`, `w/h`, `renderMs`, `fps`) plus a final `SUMMARY:{json}` and a
`QVH:{json}` probe — so an agent can self-test a whole pod and diff runs.
`NODE_PATH=/tmp/pw/node_modules node tools/selftest-multi.js <udid…>` or `--file /tmp/udids.txt`.
Screenshots: `/tmp/st_dash.png` (first phone), `/tmp/st_dash_last.png`, `/tmp/st_qvh.png`.
**Gotcha learned (2026-06-21):** do NOT run this concurrently with heavy Appium load
(admin `/admin/screenshot` polling, control-verify sweeps). WDA's screenshot RPC and the
MJPEG pump share the testmanagerd path, so concurrent screenshotting starves the stream
and makes healthy phones report `render:false`. Re-test suspects in isolation before
calling a stream broken (5/5 false-positives flipped to `render:true` when run alone).

### Control-verify (pod-side, loopback)
To prove control *lands* (not just that the UI sends it), on the pod loop over
`:7799/state` attached phones: md5 a `/admin/screenshot`, POST `/admin/control/swipe`
`{"direction":"left"}` (capture `%{http_code}` + `%{time_total}` with `-o file` so the
JSON body doesn't corrupt the parse), sleep 1.5s, md5 again → `200`+`"ok":true` = WDA
executed it; md5 delta = visible effect. 2026-06-21 run: 20/21 LAND, median 0.84s.
A `409 no-session` = phone attached but with no live Appium session (control+screenshot
both fail) — a real stuck-attached fault distinct from a dead stream.

## selftest.js — autonomous E2E browser test (with operator session)
Drives `dev.aliremote.com` phone pages (auth, stream render, tap-on-screen) **and** `qvh.turbocat.dk`
in headless Chromium using the operator's logged-in session — so the agent self-tests without a human.
Setup (one-time): extract Firefox cookies →
`cp ~/.mozilla/firefox/<profile>/cookies.sqlite /tmp/ffc.sqlite` then a small sqlite→Playwright-JSON
converter writing `/tmp/pw_cookies.json` (accessToken + communicationToken are the auth cookies; QVH app
needs none). Run: `NODE_PATH=/tmp/pw/node_modules node tools/selftest.js <udid>` → reports auth/stream/
control + writes `/tmp/st_dash.png`, `/tmp/st_qvh.png`. **Note:** `/tmp/pw_cookies.json` holds live auth —
local-only, never commit; rotate the session if leaked.
