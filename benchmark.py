import sys, json, urllib.request, time, os, tempfile

sys.stdout.reconfigure(encoding='utf-8')

port_file = os.path.join(tempfile.gettempdir(), "hyperspeed-browser.port")
if not os.path.exists(port_file):
    print("Browser chưa chạy. Mở Hyperspeed Browser trước.")
    sys.exit(1)
with open(port_file) as f:
    lines = f.read().strip().splitlines()
    port = lines[0].strip()
    token = lines[1].strip() if len(lines) > 1 else ""

base = f"http://127.0.0.1:{port}"

def api(method, path, data=None):
    url = f"{base}{path}"
    headers = {"X-API-Token": token}
    if data is not None:
        headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=json.dumps(data).encode(), headers=headers, method=method)
    else:
        req = urllib.request.Request(url, method=method)
        for k, v in headers.items():
            req.add_header(k, v)
    with urllib.request.urlopen(req) as r:
        return json.loads(r.read())

SITES = [
    ("Google", "https://google.com"),
    ("GitHub", "https://github.com"),
    ("Wikipedia", "https://en.wikipedia.org"),
]

print("\n=== Hyperspeed Browser Benchmark ===\n")
print(f"{'Site':<15} {'Load(ms)':<12} {'DOM(ms)':<12} {'Paint(ms)':<12} {'Reqs':<8} {'Mem(MB)':<10} {'Score':<8}")
print("-" * 80)

results = []
for name, url in SITES:
    print(f"\nNavigating to {name}...")
    try:
        nav = api("POST", "/api/navigate", {"url": url})
        time.sleep(3)
        info = api("GET", "/api/info")
        for _ in range(10):
            m = api("GET", "/api/opt/metrics")
            if m.get("metrics") and m["metrics"].get("loadTimeMs", 0) > 0:
                break
            time.sleep(1)
        if m.get("metrics"):
            met = m["metrics"]
            print(f"{name:<15} {met.get('loadTimeMs',0):<12.0f} {met.get('domReadyMs',0):<12.0f} {met.get('firstPaintMs',0):<12.0f} {met.get('requestCount',0):<8} {met.get('memoryUsageMB',0):<10.1f} {met.get('score',0):<8.0f}")
            results.append(met)
        else:
            print(f"{name:<15} FAIL (no metrics)")
    except Exception as e:
        print(f"{name:<15} ERROR: {e}")

if results:
    print("\n" + "=" * 80)
    print("\nBENCHMARK RESULTS\n")
    avg = {k: sum(r.get(k, 0) for r in results) / len(results) for k in results[0] if isinstance(results[0][k], (int, float))}
    print(f"Average Load: {avg.get('loadTimeMs',0):.0f} ms")
    print(f"Average DOM:  {avg.get('domReadyMs',0):.0f} ms")
    print(f"Average Reqs: {avg.get('requestCount',0):.0f}")
    print(f"Average Mem:  {avg.get('memoryUsageMB',0):.1f} MB")
    print(f"Average Score:{avg.get('score',0):.0f}/100")
    print("\nDone.")
