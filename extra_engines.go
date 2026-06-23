// DRA, MCS, CBL, UEE, HFS, RCM — remaining Genesis engines
package main

import (
	_ "embed"
	"net/http"
	"sync"
	"time"
)

// =========================================================================
// DRA — Dynamic Resource Adjustment
// =========================================================================

//go:embed dra.js
var draJS string

type DRAEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   DRAStats
}

type DRAStats struct {
	Throttled int    `json:"throttled"`
	Budget    string `json:"budget"`
	Status    string `json:"status"`
}

const draGatherJS = `(function(){var s=window.__mbDRA;if(!s)return{throttled:0,budget:'100%',status:'n/a'};return{throttled:s._throttled||0,budget:Math.round((s._budget||100))+'%',status:'ok'};})()`

func NewDRAEngine(b *browser) *DRAEngine {
	return &DRAEngine{b: b, enabled: true}
}

func (d *DRAEngine) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.enabled {
		return
	}
	d.b.syncExec(draJS)
}

func (d *DRAEngine) Gather() *DRAStats {
	var s DRAStats
	if err := d.b.syncUnwrapInto(draGatherJS, 5*time.Second, &s); err != nil {
		return &d.stats
	}
	d.mu.Lock()
	d.stats = s
	d.mu.Unlock()
	return &s
}

// =========================================================================
// MCS — Micro-Controller Scheduler
// =========================================================================

//go:embed mcs.js
var mcsJS string

type MCSEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   MCSStats
}

type MCSStats struct {
	Deferred int    `json:"deferred"`
	Executed int    `json:"executed"`
	Status   string `json:"status"`
}

const mcsGatherJS = `(function(){var s=window.__mbMCS;if(!s)return{deferred:0,executed:0,status:'n/a'};return{deferred:s._def||0,executed:s._exe||0,status:'ok'};})()`

func NewMCSEngine(b *browser) *MCSEngine {
	return &MCSEngine{b: b, enabled: true}
}

func (m *MCSEngine) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.enabled {
		return
	}
	m.b.syncExec(mcsJS)
}

func (m *MCSEngine) Gather() *MCSStats {
	var s MCSStats
	if err := m.b.syncUnwrapInto(mcsGatherJS, 5*time.Second, &s); err != nil {
		return &m.stats
	}
	m.mu.Lock()
	m.stats = s
	m.mu.Unlock()
	return &s
}

// =========================================================================
// CBL — Content-Based Loading
// =========================================================================

//go:embed cbl.js
var cblJS string

type CBLEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   CBLStats
}

type CBLStats struct {
	Deferred    int    `json:"deferred"`
	Prioritized int    `json:"prioritized"`
	Status      string `json:"status"`
}

const cblGatherJS = `(function(){var s=window.__mbCBL;if(!s)return{deferred:0,prioritized:0,status:'n/a'};return{deferred:s._def||0,prioritized:s._pri||0,status:'ok'};})()`

func NewCBLEngine(b *browser) *CBLEngine {
	return &CBLEngine{b: b, enabled: true}
}

func (c *CBLEngine) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return
	}
	c.b.syncExec(cblJS)
}

func (c *CBLEngine) Gather() *CBLStats {
	var s CBLStats
	if err := c.b.syncUnwrapInto(cblGatherJS, 5*time.Second, &s); err != nil {
		return &c.stats
	}
	c.mu.Lock()
	c.stats = s
	c.mu.Unlock()
	return &s
}

// =========================================================================
// UEE — Unified Event Engine
// =========================================================================

//go:embed uee.js
var ueeJS string

type UEEEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   UEEStats
}

type UEEStats struct {
	Delegated int    `json:"delegated"`
	Saved     int    `json:"saved"`
	Status    string `json:"status"`
}

const ueeGatherJS = `(function(){var s=window.__mbUEE;if(!s)return{delegated:0,saved:0,status:'n/a'};return{delegated:s._del||0,saved:s._sav||0,status:'ok'};})()`

func NewUEEEngine(b *browser) *UEEEngine {
	return &UEEEngine{b: b, enabled: true}
}

func (u *UEEEngine) Start() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.enabled {
		return
	}
	u.b.syncExec(ueeJS)
}

func (u *UEEEngine) Gather() *UEEStats {
	var s UEEStats
	if err := u.b.syncUnwrapInto(ueeGatherJS, 5*time.Second, &s); err != nil {
		return &u.stats
	}
	u.mu.Lock()
	u.stats = s
	u.mu.Unlock()
	return &s
}

// =========================================================================
// HFS — Heat-File System
// =========================================================================

//go:embed hfs.js
var hfsJS string

type HFSEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   HFSStats
}

type HFSStats struct {
	HotFiles int    `json:"hotFiles"`
	Cached   int    `json:"cached"`
	Status   string `json:"status"`
}

const hfsGatherJS = `(function(){var s=window.__mbHFS;if(!s)return{hotFiles:0,cached:0,status:'n/a'};return{hotFiles:s._hot||0,cached:Object.keys(s._store||{}).length,status:'ok'};})()`

func NewHFSEngine(b *browser) *HFSEngine {
	return &HFSEngine{b: b, enabled: true}
}

func (h *HFSEngine) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.enabled {
		return
	}
	h.b.syncExec(hfsJS)
}

func (h *HFSEngine) Gather() *HFSStats {
	var s HFSStats
	if err := h.b.syncUnwrapInto(hfsGatherJS, 5*time.Second, &s); err != nil {
		return &h.stats
	}
	h.mu.Lock()
	h.stats = s
	h.mu.Unlock()
	return &s
}

// =========================================================================
// RCM — Resource Cost Model
// =========================================================================

//go:embed rcm.js
var rcmJS string

type RCMEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   RCMStats
}

type RCMStats struct {
	Models    int    `json:"models"`
	Threshold string `json:"threshold"`
	Status    string `json:"status"`
}

const rcmGatherJS = `(function(){var s=window.__mbRCM;if(!s)return{models:0,threshold:'auto',status:'n/a'};return{models:Object.keys(s._models||{}).length,threshold:s._thresh||'auto',status:'ok'};})()`

func NewRCMEngine(b *browser) *RCMEngine {
	return &RCMEngine{b: b, enabled: true}
}

func (r *RCMEngine) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return
	}
	r.b.syncExec(rcmJS)
}

func (r *RCMEngine) Gather() *RCMStats {
	var s RCMStats
	if err := r.b.syncUnwrapInto(rcmGatherJS, 5*time.Second, &s); err != nil {
		return &r.stats
	}
	r.mu.Lock()
	r.stats = s
	r.mu.Unlock()
	return &s
}

// =========================================================================
// API handlers
// =========================================================================

func (b *browser) handleDRAStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.dra == nil {
		writeError(w, 503, "DRA not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.dra.Gather()})
}

func (b *browser) handleMCSStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.mcs == nil {
		writeError(w, 503, "MCS not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.mcs.Gather()})
}

func (b *browser) handleCBLStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.cbl == nil {
		writeError(w, 503, "CBL not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.cbl.Gather()})
}

func (b *browser) handleUEEStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.uee == nil {
		writeError(w, 503, "UEE not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.uee.Gather()})
}

func (b *browser) handleHFSStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.hfs == nil {
		writeError(w, 503, "HFS not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.hfs.Gather()})
}

func (b *browser) handleRCMStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.rcm == nil {
		writeError(w, 503, "RCM not init")
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "stats": b.opt.rcm.Gather()})
}
