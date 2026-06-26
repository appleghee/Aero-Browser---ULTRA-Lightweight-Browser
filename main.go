package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	ultralightui "github.com/YindSoft/ultralight-ebitengine-port"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

//go:embed cacert.pem
var cacert embed.FS

const (
	navH = 40
	winW = 1280
	winH = 800
)

var (
	colNavBg   = color.RGBA{0xE8, 0xE8, 0xE8, 0xFF}
	colBorder  = color.RGBA{0xCC, 0xCC, 0xCC, 0xFF}
	colURLBg   = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	colURLText = color.RGBA{0x22, 0x22, 0x22, 0xFF}
	colPh      = color.RGBA{0x99, 0x99, 0x99, 0xFF}
	colBtn     = color.RGBA{0x44, 0x44, 0x44, 0xFF}
	colDim     = color.RGBA{0xBB, 0xBB, 0xBB, 0xFF}
	colLoad    = color.RGBA{0x66, 0x99, 0xDD, 0xFF}
	colBench   = color.RGBA{0x88, 0x88, 0x88, 0xFF}

	ddgHome   = "https://duckduckgo.com/"
	ddgSearch = "https://duckduckgo.com/?q="

	whitePix *ebiten.Image
)

func init() {
	whitePix = ebiten.NewImage(1, 1)
	whitePix.Fill(color.White)
}

type App struct {
	ui         *ultralightui.UltralightUI
	curURL     string
	urlInput   string
	urlFocused bool
	loading    atomic.Bool
	loadStart  time.Time

	history []string
	histIdx int
	navChan chan string
	apiPort int
	font    font.Face

	pageOpts ebiten.DrawImageOptions
	rectOpts ebiten.DrawImageOptions
}

type Game struct{ app *App }

func main() {
	app := &App{
		curURL:  ddgHome,
		history: make([]string, 0, 16),
		navChan: make(chan string, 32),
	}
	app.pageOpts.GeoM.Translate(0, navH)
	app.rectOpts.ColorScale.Scale(1, 1, 1, 1)

	tt, _ := opentype.Parse(goregular.TTF)
	app.font, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 13, DPI: 96})

	app.spawnView(app.curURL)
	go app.startAPI()

	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("Ultra-Browser v4.0.1-ultra")
	ebiten.SetRunnableOnUnfocused(false)
	ebiten.SetVsyncEnabled(true)
	if err := ebiten.RunGame(&Game{app: app}); err != nil {
		log.Fatal(err)
	}
}

func (app *App) setupUI(ui *ultralightui.UltralightUI) {
	ui.OnMessage = func(msg string) {
		var m struct {
			Action string `json:"a"`
			URL    string `json:"u"`
		}
		if json.Unmarshal([]byte(msg), &m) == nil && m.Action == "loaded" {
			app.loading.Store(false)
			if m.URL != "" {
				app.curURL = m.URL
			}
		}
	}
	ui.Eval(`(function(){
var r=function(){go.send(JSON.stringify({a:"loaded",u:location.href}))};
if(document.readyState==='complete'||document.readyState==='interactive')setTimeout(r,50);
else document.addEventListener('DOMContentLoaded',r);
})()`)
	app.ui = ui
}

func (app *App) spawnView(url string) {
	if app.ui != nil {
		app.ui.Close()
	}
	ui, err := ultralightui.NewFromURL(winW, winH-navH, url, nil)
	if err != nil {
		log.Printf("[NAV] error: %v", err)
		return
	}
	ui.SetBounds(0, navH, winW, winH-navH)
	app.setupUI(ui)
	app.curURL = url
	app.loading.Store(true)
	app.loadStart = time.Now()
	log.Printf("[NAV] %s", url)
}

func (g *Game) Update() error {
	app := g.app

	select {
	case cmd := <-app.navChan:
		switch cmd {
		case "__back__":
			app.goBack()
		case "__forward__":
			app.goForward()
		case "__reload__":
			app.navigate(app.curURL)
		default:
			app.navigate(cmd)
		}
	default:
		if app.loading.Load() && time.Since(app.loadStart) > 3*time.Second {
			app.loading.Store(false)
		}
	}

	if app.ui != nil {
		if app.urlFocused {
			ultralightui.ClearFocus()
		} else {
			app.ui.SetFocus()
		}
		app.ui.Update()
	}

	app.handleInput()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xF5, 0xF5, 0xF5, 0xFF})
	if tex := g.app.ui.GetTexture(); tex != nil {
		screen.DrawImage(tex, &g.app.pageOpts)
	}
	g.app.drawNavBar(screen)
}

func (g *Game) Layout(int, int) (int, int) { return winW, winH }

// --- Navigation ---

func (app *App) navigate(raw string) {
	url := normalizeURL(strings.TrimSpace(raw))
	if url == "" || url == app.curURL {
		return
	}
	if app.histIdx < len(app.history)-1 {
		app.history = app.history[:app.histIdx+1]
	}
	app.history = append(app.history, url)
	if len(app.history) > 100 {
		app.history = app.history[50:]
		app.histIdx -= 50
	}
	app.histIdx = len(app.history) - 1
	app.spawnView(url)
}

func (app *App) goBack() {
	if app.histIdx <= 0 {
		return
	}
	app.histIdx--
	app.spawnView(app.history[app.histIdx])
}

func (app *App) goForward() {
	if app.histIdx >= len(app.history)-1 {
		return
	}
	app.histIdx++
	app.spawnView(app.history[app.histIdx])
}

// --- Input ---

func (app *App) handleInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if my < navH {
			app.handleNavBarClick(mx, my)
		} else {
			app.urlFocused = false
		}
	}

	if app.urlFocused {
		for _, r := range ebiten.AppendInputChars(nil) {
			app.urlInput += string(r)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			app.navigate(app.urlInput)
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			app.urlFocused = false
			app.urlInput = ""
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(app.urlInput) > 0 {
			app.urlInput = app.urlInput[:len(app.urlInput)-1]
		}
		return
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyL) {
			app.urlFocused = true
			app.urlInput = app.curURL
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			app.navigate(app.curURL)
			return
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			app.goBack()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			app.goForward()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		app.navigate(app.curURL)
	}
}

func (app *App) handleNavBarClick(mx, _ int) {
	const (bX = 8; bW = 24; fX = bX + bW + 8; rX = fX + bW + 8; uX = rX + bW + 12; uW = winW - uX - 8)
	switch {
	case mx >= bX && mx < bX+bW:
		app.goBack()
	case mx >= fX && mx < fX+bW:
		app.goForward()
	case mx >= rX && mx < rX+bW:
		app.navigate(app.curURL)
	case mx >= uX && mx < uX+uW:
		app.urlFocused = true
		app.urlInput = app.curURL
	default:
		app.urlFocused = false
	}
}

// --- Nav Bar ---

func (app *App) drawNavBar(screen *ebiten.Image) {
	app.fillRect(screen, 0, 0, winW, navH, colNavBg)
	app.fillRect(screen, 0, navH-1, winW, 1, colBorder)

	drawBtn(screen, 8, "◀", app.histIdx > 0, app.font)
	drawBtn(screen, 40, "▶", app.histIdx < len(app.history)-1, app.font)
	drawBtn(screen, 72, "↻", true, app.font)

	uX := 108
	uW := winW - 116
	app.fillRect(screen, float64(uX), 6, float64(uW), navH-12, colURLBg)
	app.fillRect(screen, float64(uX), 6, float64(uW), navH-12, colBorder)

	display := app.curURL
	txtCol := colURLText
	switch {
	case app.urlFocused:
		display = app.urlInput
		if display == "" {
			display = "Type URL..."
			txtCol = colPh
		}
	case app.loading.Load():
		display = "⟳ Loading..."
		txtCol = colLoad
	}
	text.Draw(screen, display, app.font, uX+6, navH-14, txtCol)

	if !app.loading.Load() && !app.loadStart.IsZero() {
		if d := time.Since(app.loadStart); d < 10*time.Second {
			bench := fmt.Sprintf("%dms", d.Milliseconds())
			text.Draw(screen, bench, app.font, uX+uW-len(bench)*8-4, navH-14, colBench)
		}
	}
}

func drawBtn(screen *ebiten.Image, x int, label string, enabled bool, f font.Face) {
	c := colBtn
	if !enabled {
		c = colDim
	}
	text.Draw(screen, label, f, x, navH-14, c)
}

func (app *App) fillRect(screen *ebiten.Image, x, y, w, h float64, cl color.Color) {
	r, g, b, a := cl.RGBA()
	app.rectOpts.GeoM.Reset()
	app.rectOpts.GeoM.Scale(w, h)
	app.rectOpts.GeoM.Translate(x, y)
	app.rectOpts.ColorScale.Reset()
	app.rectOpts.ColorScale.Scale(float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(a)/65535)
	screen.DrawImage(whitePix, &app.rectOpts)
}

// --- URL ---

func normalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "about:blank" {
		return raw
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "about:") {
		return raw
	}
	if strings.Contains(raw, ".") || strings.Contains(raw, "/") || (strings.Contains(raw, ":") && !strings.Contains(raw, " ")) {
		return "https://" + raw
	}
	return ddgSearch + strings.ReplaceAll(raw, " ", "+")
}

// --- HTTP API ---

func (app *App) startAPI() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Printf("[api] error: %v", err)
		return
	}
	app.apiPort = listener.Addr().(*net.TCPAddr).Port
	log.Printf("[api] listening on 127.0.0.1:%d", app.apiPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/navigate", app.apiNavigate)
	mux.HandleFunc("/api/back", app.apiBack)
	mux.HandleFunc("/api/forward", app.apiForward)
	mux.HandleFunc("/api/reload", app.apiReload)
	mux.HandleFunc("/api/info", app.apiInfo)
	mux.HandleFunc("/api/eval", app.apiEval)
	http.Serve(listener, cors(mux))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Token")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *App) apiNavigate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "POST required"})
		return
	}
	var b struct{ URL string }
	json.NewDecoder(r.Body).Decode(&b)
	app.navChan <- b.URL
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "url": b.URL})
}

func (app *App) apiBack(w http.ResponseWriter, _ *http.Request) {
	app.navChan <- "__back__"
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (app *App) apiForward(w http.ResponseWriter, _ *http.Request) {
	app.navChan <- "__forward__"
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (app *App) apiReload(w http.ResponseWriter, _ *http.Request) {
	app.navChan <- "__reload__"
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (app *App) apiInfo(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"url":   app.curURL,
		"title": app.curURL,
	})
}

func (app *App) apiEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "POST required"})
		return
	}
	var b struct{ JS string }
	json.NewDecoder(r.Body).Decode(&b)
	if app.ui != nil {
		app.ui.Eval(b.JS)
	}
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func randToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
