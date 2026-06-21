// UPM — User Presence Model (idle detection)
package main

import (
	_ "embed"
	"net/http"
	"sync"
	"time"
)

//go:embed upm.js
var upmJS string

type UPMEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   UPMStats
}

type UPMStats struct {
	State     string `json:"state"`
	IdleSec   int    `json:"idleSec"`
	TotalIdle int    `json:"totalIdle"`
	Status    string `json:"status"`
}

const upmGatherJS = `(function(){var s=window.__mbUPM;if(!s)return{state:'unknown',idleSec:0,totalIdle:0,status:'n/a'};return{state:s.state,idleSec:Math.round((Date.now()-s._last)/1000),totalIdle:s._total||0,status:'ok'};})()`

func NewUPMEngine(b *browser) *UPMEngine {
	return &UPMEngine{b: b, enabled: true}
}

func (u *UPMEngine) Start() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.enabled {
		return
	}
	u.b.syncExec(upmJS)
}

func (u *UPMEngine) Gather() *UPMStats {
	var s UPMStats
	if err := u.b.syncUnwrapInto(upmGatherJS, 5*time.Second, &s); err != nil {
		return &u.stats
	}
	u.mu.Lock()
	u.stats = s
	u.mu.Unlock()
	return &s
}

func (b *browser) handleUPMStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.upm == nil {
		writeError(w, 503, "UPM not init")
		return
	}
	s := b.opt.upm.Gather()
	writeJSON(w, map[string]interface{}{"ok": true, "stats": s})
}
