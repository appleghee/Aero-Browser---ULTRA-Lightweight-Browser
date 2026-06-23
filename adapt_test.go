package main

import (
	"testing"
)

func TestClassifySite(t *testing.T) {
	tests := []struct {
		url      string
		expected SiteType
	}{
		// Empty / invalid
		{"", SiteGeneral},
		{"not a url", SiteGeneral},
		{"http://", SiteGeneral},

		// Search
		{"https://www.google.com/search?q=foo", SiteSearch},
		{"https://google.com", SiteSearch},
		{"https://duckduckgo.com/?q=test", SiteSearch},
		{"https://search.brave.com", SiteSearch},
		{"https://www.startpage.com", SiteSearch},

		// Email (must be detected before generic search)
		{"https://mail.google.com", SiteEmail},
		{"https://outlook.live.com", SiteEmail},
		{"https://protonmail.com", SiteEmail},

		// Code
		{"https://github.com/foo", SiteCode},
		{"https://gitlab.com/foo", SiteCode},
		{"https://stackoverflow.com/questions", SiteCode},

		// Video
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", SiteVideo},
		{"https://youtu.be/dQw4w9WgXcQ", SiteVideo},
		{"https://twitch.tv/lirik", SiteVideo},
		{"https://www.tiktok.com/@user", SiteVideo},

		// Social
		{"https://www.facebook.com", SiteSocial},
		{"https://x.com/elonmusk", SiteSocial},
		{"https://twitter.com/elonmusk", SiteSocial},
		{"https://t.me/durov", SiteSocial},

		// News
		{"https://www.cnn.com", SiteNews},
		{"https://www.bbc.com/news", SiteNews},
		{"https://medium.com/@user", SiteNews},

		// Ecommerce
		{"https://www.amazon.com/dp/B08N5WRWNW", SiteEcommerce},
		{"https://www.ebay.com", SiteEcommerce},
		{"https://shopee.com", SiteEcommerce},

		// General (fallback)
		{"https://example.com", SiteGeneral},
		{"https://myblog.org", SiteGeneral},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := classifySite(tt.url)
			if result != tt.expected {
				t.Errorf("classifySite(%q) = %s, want %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestAdaptOnNavigateAndShouldRun(t *testing.T) {
	a := NewAdapt()

	st := a.OnNavigate("https://google.com/search")
	if st != SiteSearch {
		t.Fatalf("expected SiteSearch, got %s", st)
	}
	if a.ShouldRun("dna") {
		t.Error("dna should be disabled on search")
	}
	if !a.ShouldRun("hlrc") {
		t.Error("hlrc should be enabled on search")
	}

	a.OnNavigate("https://github.com/golang/go")
	if a.ShouldRun("avp") {
		t.Error("avp should be disabled on code")
	}
	if !a.ShouldRun("domCompress") {
		t.Error("domCompress should be enabled on code")
	}

	a.OnNavigate("https://example.com")
	if !a.ShouldRun("dna") {
		t.Error("no engine should be disabled on general")
	}
	if cnt := len(a.Profile()["disabled"].([]string)); cnt != 0 {
		t.Errorf("expected 0 disabled, got %d", cnt)
	}
}

func TestAdaptProfile(t *testing.T) {
	a := NewAdapt()
	a.OnNavigate("https://mail.google.com")
	profile := a.Profile()
	if profile["siteType"] != "email" {
		t.Errorf("expected email, got %s", profile["siteType"])
	}
	disabled := profile["disabled"].([]string)
	if len(disabled) == 0 {
		t.Error("expected some engines disabled on email")
	}
}

func TestAdaptSameTypeNoRebuild(t *testing.T) {
	a := NewAdapt()
	a.OnNavigate("https://google.com")
	firstCnt := len(a.Profile()["disabled"].([]string))
	a.OnNavigate("https://www.bing.com")
	secondCnt := len(a.Profile()["disabled"].([]string))
	if firstCnt != secondCnt {
		t.Errorf("disabled set size changed from %d to %d without type change", firstCnt, secondCnt)
	}
	if a.Profile()["classified"].(int) != 2 {
		t.Error("classified count should be 2")
	}
}
