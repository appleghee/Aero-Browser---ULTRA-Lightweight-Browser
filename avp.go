// AVP — Adaptive Viewport Predictor (scroll prediction)
package main

import (
	_ "embed"
	"net/http"
	"sync"
	"time"
)

//go:embed avp.js
var avpJS string

type AVPEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   AVPStats
}

type AVPStats struct {
	Velocity    float64 `json:"velocity"`
	Direction   int     `json:"direction"`
	Predictions int     `json:"predictions"`
	Hits        int     `json:"hits"`
	Status      string  `json:"status"`
}

const avpGatherJS = `(function(){var s=window.__mbAVP;if(!s)return{velocity:0,direction:0,predictions:0,hits:0,status:'n/a'};return{velocity:s.velocity,direction:s.direction,predictions:s._pred||0,hits:s._hits||0,status:'ok'};})()`

func NewAVPEngine(b *browser) *AVPEngine {
	return &AVPEngine{b: b, enabled: true}
}

func (a *AVPEngine) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.enabled {
		return
	}
	a.b.syncExec(avpJS)
}

func (a *AVPEngine) Gather() *AVPStats {
	var s AVPStats
	if err := a.b.syncUnwrapInto(avpGatherJS, 5*time.Second, &s); err != nil {
		return &a.stats
	}
	a.mu.Lock()
	a.stats = s
	a.mu.Unlock()
	return &s
}

func (b *browser) handleAVPStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.avp == nil {
		writeError(w, 503, "AVP not init")
		return
	}
	s := b.opt.avp.Gather()
	writeJSON(w, map[string]interface{}{"ok": true, "stats": s})
}
