package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// SiteType classifies what kind of page the user is on.
type SiteType int

const (
	SiteGeneral SiteType = iota
	SiteSearch
	SiteCode
	SiteVideo
	SiteSocial
	SiteNews
	SiteEcommerce
	SiteEmail
)

func (s SiteType) String() string {
	switch s {
	case SiteSearch:
		return "search"
	case SiteCode:
		return "code"
	case SiteVideo:
		return "video"
	case SiteSocial:
		return "social"
	case SiteNews:
		return "news"
	case SiteEcommerce:
		return "ecommerce"
	case SiteEmail:
		return "email"
	default:
		return "general"
	}
}

// Adapt is the intelligent site-aware engine orchestrator.
type Adapt struct {
	mu         sync.RWMutex
	currentURL string
	siteType   SiteType

	disabled        map[string]bool
	totalClassified int
}

// per-site engine blacklists — engines that contribute nothing to that site type.
var siteEngineBlacklist = map[SiteType][]string{
	SiteSearch: {
		"dna", "hbm", "avp", "pce", "upm", "ncg", "dra", "mcs", "cbl", "uee", "hfs", "rcm",
		"lod", "pvc", "rhd", "ehs", "rpc", "crg", "domCompress",
	},
	SiteCode: {
		"avp", "ehs", "rpc", "crg", "pce",
	},
	SiteVideo: {
		"domCompress", "lod", "pvc",
	},
	SiteSocial: {
		"pce", "ncg",
	},
	SiteNews: {
		"pce", "ncg",
	},
	SiteEcommerce: {
		// everything useful
	},
	SiteEmail: {
		"dna", "hbm", "avp", "pce", "upm", "ncg", "dra", "mcs", "cbl", "uee", "hfs", "rcm",
		"lod", "pvc", "rhd", "ehs", "rpc", "crg", "domCompress", "avp",
	},
}

func NewAdapt() *Adapt {
	return &Adapt{
		disabled: make(map[string]bool),
	}
}

func classifySite(rawURL string) SiteType {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return SiteGeneral
	}
	host := u.Hostname()
	host = strings.ToLower(host)

	match := func(domain string) bool {
		return host == domain || strings.HasSuffix(host, "."+domain)
	}

	// Email (must be checked BEFORE generic search e.g. mail.google.com)
	if match("mail.google.com") ||
		match("outlook.com") || match("live.com") || host == "outlook.live.com" ||
		match("protonmail.com") ||
		match("zoho.com") ||
		match("fastmail.com") {
		return SiteEmail
	}

	// Search engines
	if match("google.com") || match("google.co") ||
		match("duckduckgo.com") ||
		match("bing.com") ||
		match("yahoo.com") ||
		match("baidu.com") ||
		match("yandex.com") ||
		match("ecosia.org") ||
		match("qwant.com") ||
		host == "search.brave.com" ||
		host == "www.startpage.com" {
		return SiteSearch
	}

	// Code/Dev
	if match("github.com") || match("github.io") ||
		match("gitlab.com") ||
		match("bitbucket.org") ||
		match("stackoverflow.com") ||
		match("stackexchange.com") ||
		match("npmjs.com") ||
		match("pypi.org") ||
		match("codepen.io") ||
		match("codesandbox.io") ||
		match("replit.com") ||
		match("readthedocs.io") {
		return SiteCode
	}

	// Video
	if match("youtube.com") || host == "youtu.be" ||
		match("twitch.tv") ||
		match("vimeo.com") ||
		match("dailymotion.com") ||
		match("netflix.com") ||
		match("hulu.com") ||
		match("spotify.com") ||
		match("tiktok.com") {
		return SiteVideo
	}

	// Social
	if match("facebook.com") || match("fb.com") ||
		match("twitter.com") || host == "x.com" || match("x.com") ||
		match("reddit.com") ||
		match("instagram.com") ||
		match("linkedin.com") ||
		match("discord.com") || match("discord.gg") ||
		match("telegram.org") ||
		match("whatsapp.com") ||
		host == "t.me" {
		return SiteSocial
	}

	// News
	if match("cnn.com") ||
		match("bbc.com") || match("bbc.co.uk") ||
		match("nytimes.com") ||
		match("reuters.com") ||
		match("bloomberg.com") ||
		match("medium.com") ||
		match("substack.com") {
		return SiteNews
	}

	// E-commerce
	if match("amazon.com") || match("amazon.co.uk") || match("amazon.de") || match("amazon.fr") || match("amazon.co.jp") ||
		match("ebay.com") ||
		match("etsy.com") ||
		match("walmart.com") ||
		match("bestbuy.com") ||
		match("target.com") ||
		match("aliexpress.com") || match("alibaba.com") ||
		match("shopee.com") || match("shopee.sg") ||
		match("lazada.com") || match("lazada.sg") ||
		match("taobao.com") ||
		match("tmall.com") {
		return SiteEcommerce
	}

	return SiteGeneral
}

// OnNavigate classifies a URL and updates the disabled engine set.
// Returns the new site type.
func (a *Adapt) OnNavigate(rawURL string) SiteType {
	rawURL = strings.TrimSpace(rawURL)
	st := classifySite(rawURL)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.currentURL = rawURL
	a.siteType = st
	a.totalClassified++

	// Rebuild disabled set
	a.disabled = make(map[string]bool)
	if blacklist, ok := siteEngineBlacklist[st]; ok {
		for _, name := range blacklist {
			a.disabled[name] = true
		}
	}

	log.Printf("[ADAPT] site=%s url=%s disabled=%d/%d engines",
		st, truncate(rawURL, 80), len(a.disabled), len(engineNames))

	return st
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ShouldRun returns true if an engine should execute for the current page.
func (a *Adapt) ShouldRun(engine string) bool {
	a.mu.RLock()
	disabled := a.disabled[engine]
	a.mu.RUnlock()
	return !disabled
}

// Profile returns the current adapt state.
func (a *Adapt) Profile() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	disabledList := make([]string, 0, len(a.disabled))
	for name := range a.disabled {
		disabledList = append(disabledList, name)
	}

	return map[string]interface{}{
		"siteType":   a.siteType.String(),
		"currentURL": a.currentURL,
		"disabled":   disabledList,
		"count":      len(disabledList),
		"classified": a.totalClassified,
	}
}

// engineNames is the canonical list of all known engines.
var engineNames = []string{
	"hlrc", "uhe", "autotune", "gcCtl",
	"dna", "hbm", "avp", "domCompress", "ncg", "pce", "upm",
	"dra", "mcs", "cbl", "uee", "hfs", "rcm",
	"lod", "pvc", "rhd", "ehs", "rpc", "crg",
	"qse", "vd", "quick", "netq", "csso", "media", "cache", "tuner",
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

func (b *browser) handleAdaptStats(w http.ResponseWriter, r *http.Request) {
	if b.opt == nil || b.opt.adapt == nil {
		writeError(w, 503, "adapt not initialized")
		return
	}
	writeJSON(w, map[string]interface{}{
		"ok":    true,
		"adapt": b.opt.adapt.Profile(),
	})
}
