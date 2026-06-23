// PCE — Page Change Engine (mutation batching)
package main

import (
	_ "embed"
	"net/http"
	"sync"
	"time"
)

//go:embed pce.js
var pceJS string

type PCEEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   PCEStats
}

type PCEStats struct {
	Batched  int    `json:"batched"`
	Flushed  int    `json:"flushed"`
	AvgBatch int    `json:"avgBatch"`
	Status   string `json:"status"`
}

const pceGatherJS = `(function(){var s=window.__mbPCE;if(!s)return{batched:0,flushed:0,avgBatch:0,status:'n/a'};return{batched:s._batched||0,flushed:s._flushed||0,avgBatch:s._batched>0?Math.round(s._flushed/s._batched):0,status:'ok'};})()`

func NewPCEEngine(b *browser) *PCEEngine {
	return &PCEEngine{b: b, enabled: true}
}

func (p *PCEEngine) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.enabled {
		return
	}
	p.b.syncExec(pceJS)
}

func (p *PCEEngine) Gather() *PCEStats {
	var s PCEStats
	if err := p.b.syncUnwrapInto(pceGatherJS, 5*time.Second, &s); err != nil {
		return &p.stats
	}
	p.mu.Lock()
	p.stats = s
	p.mu.Unlock()
	return &s
}

func (b *browser) handlePCEStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.pce == nil {
		writeError(w, 503, "PCE not init")
		return
	}
	s := b.opt.pce.Gather()
	writeJSON(w, map[string]interface{}{"ok": true, "stats": s})
}
