# Hyperspeed Browser v3.2.0-alpha

## IO Cascade — Phase 2 (UI)

### IO Cascade

Replaced the rAF-based `_classify` loop in LOD with an **IntersectionObserver + `content-visibility:auto`** cascade:

- **LOD0** (viewport + 1.5×): Full render — no change
- **LOD1** (1.5×–4×): `content-visibility:auto` + `contain-intrinsic-size:<stored>` — browser skips painting, keeps layout slot
- **LOD2** (4×–8×): `content-visibility:auto` + `contain-intrinsic-size:1px 1px` — minimal placeholder, browser manages rendering
- **LOD3** (8×+): DOM removal — same as before

**Why:** The old rAF classify loop scanned every element every frame (even when nothing scrolled). Browser-native `content-visibility` does the same work in the compositor thread with zero JS cost per frame. The 2s interval is only for LOD3 transitions (deep off-screen), which is a cold path.

**Impact:** Zero JS classify overhead on idle pages. LOD1-2 elements get browser-native render skipping with guaranteed scroll-position restoration (no manual `_inflate` for content-visibility'd elements).

### API Endpoints

| Route | Method | Description |
|-------|--------|-------------|
| `GET /api/ioc/stats` | GET | LOD stats with `cascade` field (content-visibility count) |

### Files Changed

- `lod.js` — Rewritten: rAF loop → 2s ticker, LOD1-2 uses content-visibility:auto + contain-intrinsic-size
- `lod.go` — Added `Cascade int` to LODStats, `handleIOCStats` handler
- `main.go` — `/api/ioc/stats` route + API root listing
- `startpage.html` — IOC cascade stats row + polling

### Next

- Full alpha: per-element `contain-intrinsic-size` smoothing (avoid 1px snap on restore)
- Beta: UHE integration — hot elements skip content-visibility cascade
- Stable: HMR sensor for scroll velocity-based cascade
