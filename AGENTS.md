# AGENTS.md — Hyperspeed Browser

## Build & Run

```powershell
$env:CGO_ENABLED=1; $env:CC="gcc"
go build -ldflags="-s -w -H=windowsgui" -o hyperspeed-browser.exe .
```

Windows-only. Requires **MinGW-w64** (`C:\mingw64\bin\gcc`) + **WebView2 Runtime** (Windows 11 includes it).

No tests, no linter, no formatter config in repo. Default `go build` is the only verification step.

## Architecture

Single package `main`, single binary. Everything lives in `main.go` + engine files (`*.go`, same package).

Data flow: `main()` creates `browser{WebView}` → binds Go→JS via `w.Bind("goNavigate", ...)` → starts HTTP API goroutine → `w.Init(...)` injects all bootstrap JS → `w.Run()`.

## JS Injection Conventions

All bootstrap JS is **merged into one `w.Init(...)` call** at `main.go:200`. Adding new page-loaded behavior requires appending to that concat. Inline `const` strings (overlayJS, toolbarJS, runtimeJS, etc.) are the convention — do not extract to external files unless also updating `//go:embed`.

DOM-injected JS (post-load) goes through `b.w.Dispatch(func() { b.w.Eval(...) })` (see `injectTurboLoop` at `main.go:323`). Always guard with `window.__mbXxx` flags to prevent re-execution.

## Sync Eval Bridge

`syncEval(js, timeout)` → dispatches `__evalCb(id, JSON.stringify(result))` → callback into `evalCallback`. This is the only safe way to get JS results back. Never call `w.Eval` directly for values that need to return to Go.

`syncExec(js)` is fire-and-forget (no return).

## Runtime Server

- Binds `127.0.0.1:0` (random port)
- Port + 32-byte hex token written to `%TEMP%\hyperspeed-browser.port` (line 2 = token)
- CORS: `Access-Control-Allow-Origin: *` (localhost-only binding limits exposure)
- Auth: single `X-API-Token` header check, no expiry, no rotation

## Platform Assumptions

- `handleScreenshot()` (`main.go:853`) shells out to **PowerShell** + inline C# P/Invoke `PrintWindow`. Windows-only, no fallback.
- `startpage.html` is rendered as `data:text/html,...` for `hyperspeed://console`. Template tokens `{{APITOKEN}}` and `{{APIPORT}}` are replaced at startup.

## IO Cascade (v3.2 alpha)

`lod.js` now uses `content-visibility:auto` + `contain-intrinsic-size` for LOD1-2 instead of innerHTML stripping + rAF classify loop. The browser's compositor handles off-screen render skipping natively. A 2s interval handles LOD3 transitions only.

- LOD1: `content-visibility:auto` with stored element dimensions (browser skips paint, keeps layout)
- LOD2: `content-visibility:auto` with 1px intrinsic size (minimal placeholder)
- LOD3: same as before (DOM removal + placeholder div)

Stats endpoint: `GET /api/ioc/stats` — returns LOD stats with `cascade` field (elements using content-visibility).

## Key Structural Gotchas

- `Optimizer` struct owns 18 engine sub-objects; many are started in goroutines (`uhe.Start()`, `hlrc.Start()`, `autotune.Start()`, `gcctl.Start()`). Adding a new engine requires wiring into `NewOptimizer` + registering API routes in `startAPI()`.
- `b.opt` is nil-guarded in every handler — new endpoints must follow the same pattern.
- History and browse history are protected by `b.mu`; eval dispatch runs on WebView thread and must not block.
