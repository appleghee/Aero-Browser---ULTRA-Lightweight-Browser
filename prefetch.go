// PrefetchEngine — learns link click patterns per domain, prefetches high-probability pages
package main

import (
	"encoding/json"
	_ "embed"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed prefetch.js
var prefetchJS string

type PrefetchEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	sites   map[string]*sitePatterns
	stats   PrefetchStats
}

type sitePatterns struct {
	Patterns  map[string]*linkStat
	TotalHov  int
	TotalClk  int
	LastPrune time.Time
}

type linkStat struct {
	Href       string
	HoverCount int
	ClickCount int
	LastSeen   time.Time
}

type PrefetchStats struct {
	Domains  int   `json:"domains"`
	Patterns int   `json:"patterns"`
	Hovers   int64 `json:"hovers"`
	Clicks   int64 `json:"clicks"`
	Hits     int64 `json:"hits"`
	Enabled  bool  `json:"enabled"`
}

const (
	predictThreshold = 0.5 // minimum probability to trigger prefetch
	predictMinHov    = 2   // minimum hovers before considering prefetch
	predictMaxAge    = 10 * time.Minute
	predictPruneInt  = 5 * time.Minute
)

func NewPrefetchEngine(b *browser) *PrefetchEngine {
	return &PrefetchEngine{
		b:       b,
		enabled: true,
		sites:   make(map[string]*sitePatterns),
	}
}

func (p *PrefetchEngine) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.enabled {
		return
	}
	go p.pruneLoop()
}

func (p *PrefetchEngine) pruneLoop() {
	for {
		time.Sleep(predictPruneInt)
		p.mu.Lock()
		now := time.Now()
		for domain, sp := range p.sites {
			for href, st := range sp.Patterns {
				if now.Sub(st.LastSeen) > predictMaxAge {
					delete(sp.Patterns, href)
				}
			}
			if len(sp.Patterns) == 0 {
				delete(p.sites, domain)
			} else {
				sp.LastPrune = now
			}
		}
		p.mu.Unlock()
	}
}

// cleanDomain strips scheme/port/path for consistent domain key
func cleanDomain(raw string) string {
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimPrefix(raw, "https://")
	if idx := strings.Index(raw, "/"); idx > 0 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, ":"); idx > 0 {
		raw = raw[:idx]
	}
	return raw
}

func (p *PrefetchEngine) getOrCreate(domain, href string) (*sitePatterns, *linkStat) {
	sp, ok := p.sites[domain]
	if !ok {
		sp = &sitePatterns{Patterns: make(map[string]*linkStat), LastPrune: time.Now()}
		p.sites[domain] = sp
	}
	st, ok := sp.Patterns[href]
	if !ok {
		st = &linkStat{Href: href}
		sp.Patterns[href] = st
	}
	return sp, st
}

func (p *PrefetchEngine) HandleHover(domain, href string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	domain = cleanDomain(strings.TrimSpace(domain))
	href = strings.TrimSpace(href)
	if domain == "" || href == "" {
		return false
	}

	sp, st := p.getOrCreate(domain, href)
	st.HoverCount++
	st.LastSeen = time.Now()
	sp.TotalHov++
	p.stats.Hovers++

	prob := float64(st.ClickCount+1) / float64(st.HoverCount+2)
	return st.HoverCount >= predictMinHov && prob >= predictThreshold
}

func (p *PrefetchEngine) HandleClick(domain, href string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	domain = cleanDomain(strings.TrimSpace(domain))
	href = strings.TrimSpace(href)
	if domain == "" || href == "" {
		return
	}

	sp, st := p.getOrCreate(domain, href)
	st.ClickCount++
	st.LastSeen = time.Now()
	sp.TotalClk++
	p.stats.Clicks++
}

func (p *PrefetchEngine) HandleHit() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.Hits++
}

func (p *PrefetchEngine) Stats() PrefetchStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.Domains = len(p.sites)
	total := 0
	for _, sp := range p.sites {
		total += len(sp.Patterns)
	}
	p.stats.Patterns = total
	return p.stats
}

// --- API handlers ---

func (b *browser) handlePredictHover(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.prf == nil {
		writeError(w, 503, "prf not init")
		return
	}
	var body struct {
		Domain string `json:"domain"`
		Href   string `json:"href"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "bad request")
		return
	}
	prefetch := b.opt.prf.HandleHover(body.Domain, body.Href)
	if prefetch {
		log.Printf("[PRF] prefetch: domain=%s href=%.64s", body.Domain, body.Href)
	}
	writeJSON(w, map[string]bool{"prefetch": prefetch})
}

func (b *browser) handlePredictClick(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.prf == nil {
		writeError(w, 503, "prf not init")
		return
	}
	var body struct {
		Domain string `json:"domain"`
		Href   string `json:"href"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "bad request")
		return
	}
	b.opt.prf.HandleClick(body.Domain, body.Href)
	writeJSON(w, map[string]bool{"ok": true})
}

func (b *browser) handlePredictHit(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.prf == nil {
		writeError(w, 503, "prf not init")
		return
	}
	b.opt.prf.HandleHit()
	writeJSON(w, map[string]bool{"ok": true})
}

func (b *browser) handlePredictStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.prf == nil {
		writeError(w, 503, "prf not init")
		return
	}
	writeJSON(w, map[string]interface{}{
		"ok":    true,
		"stats": b.opt.prf.Stats(),
	})
}
