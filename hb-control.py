#!/usr/bin/env python3
"""Hyperspeed Browser external control tool.

Usage:
  hb-control.py status           Show dashboard
  hb-control.py profile <name>   Set optimizer mode
  hb-control.py toggle <key>     Toggle setting on/off
  hb-control.py optimize         Run full optimization
  hb-control.py snapshot         DOM snapshot
  hb-control.py eval <js>        Execute JS in browser
  hb-control.py discover         Find browser connection
"""

import sys, os, json, time, urllib.request, urllib.error

def find_browser():
    port_file = os.path.join(os.environ.get('TEMP', os.environ.get('TMPDIR', '/tmp')), 'hyperspeed-browser.port')
    if not os.path.exists(port_file):
        return None
    with open(port_file) as f:
        lines = [l.strip() for l in f.readlines() if l.strip()]
    if len(lines) < 2:
        return None
    return {'port': int(lines[0]), 'token': lines[1]}

def api(b, method, path, body=None):
    url = f'http://127.0.0.1:{b["port"]}{path}'
    req = urllib.request.Request(url, method=method, data=json.dumps(body).encode() if body else None,
        headers={'Content-Type': 'application/json', 'X-API-Token': b['token']})
    try:
        with urllib.request.urlopen(req, timeout=5) as r:
            return json.loads(r.read())
    except urllib.error.HTTPError as e:
        return json.loads(e.read()) if e.code != 404 else {'error': 'not found'}
    except Exception as e:
        return {'error': str(e)}

def cmd_status(b):
    info = api(b, 'GET', '/api/info')
    metrics = api(b, 'GET', '/api/opt/metrics')
    vd = api(b, 'GET', '/api/vd/snapshot')
    crg = api(b, 'GET', '/api/crg/snapshot')
    m = metrics.get('metrics', {}) if isinstance(metrics, dict) else {}
    print(f'URL:        {info.get("currentURL","?")}')
    print(f'Score:      {m.get("score","-")}')
    print(f'Load:       {m.get("loadTimeMs","-")}ms')
    print(f'Requests:   {m.get("requestCount","-")}')
    print(f'DOM nodes:  {m.get("domNodeCount","-")}')
    print(f'Mem:        {m.get("memoryUsageMB","-")}MB')
    print(f'VD avg:     {vd.get("avgVD","-")}')
    print(f'CRG cache:  {crg.get("cacheHits","-")} hits')

def cmd_profile(b, name):
    r = api(b, 'POST', '/api/opt/profile', {'profile': name})
    print(f'Profile: {name} -> {"OK" if r.get("ok") else "FAIL"}')

def cmd_toggle(b, key):
    valid = {'lazy': 'lazyImages', 'defer': 'deferJS', 'tracker': 'blockTrackers', 'cache': 'smartCache',
             'lazyimages': 'lazyImages', 'deferjs': 'deferJS', 'blocktrackers': 'blockTrackers', 'smartcache': 'smartCache'}
    k = valid.get(key.lower(), key)
    r = api(b, 'POST', '/api/opt/toggle', {'key': k, 'val': True})
    print(f'{k}: {"OK" if r.get("ok") else "FAIL"}')

def cmd_optimize(b):
    r = api(b, 'POST', '/api/opt/run')
    print(f'Optimize: {"OK" if r.get("ok") else "FAIL"}')

def cmd_snapshot(b):
    r = api(b, 'GET', '/api/snapshot')
    if isinstance(r, dict) and 'result' in r:
        print(f'Snapshot: {len(r["result"])} nodes')
    else:
        print(f'Snapshot: {r}')

def cmd_eval(b, js):
    r = api(b, 'POST', '/api/eval', {'js': js})
    print(f'Eval: {r}')

def cmd_discover(b):
    print(f'Port:  {b["port"]}')
    print(f'Token: {b["token"]}')
    info = api(b, 'GET', '/api/info')
    print(f'Title: {info.get("currentURL","?")}')

def main():
    cmds = {
        'status': cmd_status, 'profile': cmd_profile, 'toggle': cmd_toggle,
        'optimize': cmd_optimize, 'snapshot': cmd_snapshot, 'eval': cmd_eval,
        'discover': cmd_discover,
    }
    if len(sys.argv) < 2 or sys.argv[1] not in cmds:
        print(__doc__)
        sys.exit(1)
    b = find_browser()
    if not b:
        print('Browser not running.')
        sys.exit(1)
    cmd = sys.argv[1]
    args = sys.argv[2:]
    if cmd in ('profile', 'toggle', 'eval') and not args:
        print(f'{cmd} needs argument')
        sys.exit(1)
    cmds[cmd](b, *args)

if __name__ == '__main__':
    main()
