// DOM Compression — binary-serialized DOM snapshot transport
package main

import (
	_ "embed"
	"net/http"
	"sync"
	"time"
)

//go:embed dom_compress.js
var domCompressJS string

type DOMCompressEngine struct {
	b       *browser
	mu      sync.Mutex
	enabled bool
	stats   DOMCompressStats
}

type DOMCompressStats struct {
	LastSize   int    `json:"lastSize"`
	Compressed int    `json:"compressed"`
	Ratio      string `json:"ratio"`
	Nodes      int    `json:"nodes"`
	Status     string `json:"status"`
}

const domCompressGatherJS = `(function(){
var s=window.__mbDOMC;if(!s)return{lastSize:0,compressed:0,ratio:'0%',nodes:0,status:'n/a'};
var r={lastSize:s._lastSize||0,compressed:s._comp||0,nodes:s._nodes||0,status:'ok'};
r.ratio=s._lastSize>0?Math.round(s._comp/s._lastSize*100)+'%':'0%';
return r;
})()`

func NewDOMCompressEngine(b *browser) *DOMCompressEngine {
	return &DOMCompressEngine{b: b, enabled: true}
}

func (d *DOMCompressEngine) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.enabled {
		return
	}
	d.b.syncExec(domCompressJS)
}

func (d *DOMCompressEngine) Gather() *DOMCompressStats {
	var s DOMCompressStats
	if err := d.b.syncUnwrapInto(domCompressGatherJS, 5*time.Second, &s); err != nil {
		return &d.stats
	}
	d.mu.Lock()
	d.stats = s
	d.mu.Unlock()
	return &s
}

func (b *browser) handleDOMCompressStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.domCompress == nil {
		writeError(w, 503, "DOM Compress not init")
		return
	}
	s := b.opt.domCompress.Gather()
	writeJSON(w, map[string]interface{}{"ok": true, "stats": s})
}
