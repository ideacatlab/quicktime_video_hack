#!/usr/bin/env python3
# control-latency — measure the WDA control path on a pod so we can pinpoint + fix command lag.
# Breaks latency into: WDA /status (transport + WDA responsiveness baseline), a tap via the pod
# admin control (full command path), and warm-vs-cold (first command after idle). Run on the pod.
# Usage: control-latency.py <udid> <clientPort> [N]   (env POD_ADMIN=http://127.0.0.1:7799)
import sys, os, time, json, urllib.request

udid = sys.argv[1]; port = sys.argv[2]; N = int(sys.argv[3]) if len(sys.argv) > 3 else 20
ADMIN = os.environ.get("POD_ADMIN", "http://127.0.0.1:7799")

def timed(fn):
    t = time.time()
    try:
        fn()
    except Exception:
        return None
    return (time.time() - t) * 1000

def wda_status():
    urllib.request.urlopen(f"http://127.0.0.1:{port}/status", timeout=12).read()

def tap():
    req = urllib.request.Request(f"{ADMIN}/admin/control/tap?udid={udid}",
        data=json.dumps({"coords": {"normX": 0.5, "normY": 0.5}}).encode(),
        headers={"content-type": "application/json"}, method="POST")
    urllib.request.urlopen(req, timeout=12).read()

def report(name, fn, n, gap=0.0):
    xs = []
    for _ in range(n):
        v = timed(fn)
        if v is not None:
            xs.append(v)
        if gap:
            time.sleep(gap)
    if not xs:
        print(f"  {name:32s} ALL FAILED"); return
    xs.sort()
    p = lambda q: xs[min(len(xs) - 1, int(len(xs) * q))]
    print(f"  {name:32s} p50={p(.5):6.0f}ms  p95={p(.95):6.0f}ms  max={max(xs):6.0f}ms  n={len(xs)}")

print(f"=== control latency {udid[:8]} port {port}, {N} samples ===")
report("WDA /status (transport+wda)", wda_status, N)
report("tap warm (full command)", tap, N)
print("  (idle 8s -> cold)"); time.sleep(8)
report("tap cold (1 after idle)", tap, 1)
report("tap warm again", tap, 5)
