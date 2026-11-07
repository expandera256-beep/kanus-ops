package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed gui.html
var guiPage []byte

const guiAddr = "127.0.0.1:4477"

var launchPanel = func(string) {}
var notifyUser = func(string) {}

type guiServer struct {
	mu sync.Mutex
	s  *session
}

var errPanelReady = errors.New("panel hazır")

func serveGUI(s *session) error {
	if err := listenPanel(s); err != nil {
		if errors.Is(err, errPanelReady) {
			return nil
		}
		notifyUser("Panel açılamadı.\n" + err.Error() + "\nhttp://" + guiAddr + "/")
		return err
	}
	page := "http://" + guiAddr + "/"
	openWhenReady(page)
	select {}
}

func listenPanel(s *session) error {
	if s.modelCh == nil {
		s.modelCh = make(chan modelResult, 4)
	}
	ln, err := net.Listen("tcp", guiAddr)
	if err != nil {
		if openExisting("http://" + guiAddr + "/") {
			return errPanelReady
		}
		return err
	}
	go func() { _ = http.Serve(ln, guiMux(s)) }()
	if autoTunnelOnThisPlatform() { startAutoTunnel(s) }
	go func() {
		for {
			snap := probe()
			unlock := s.hold()
			s.noteLife(snap)
			s.resWait = false
			unlock()
			time.Sleep(3 * time.Second)
		}
	}()
	go func() {
		tick := time.NewTicker(time.Second)
		for now := range tick.C {
			unlock := s.hold()
			s.accrueOpen(now)
			unlock()
		}
	}()
	println("US http://" + guiAddr + "/")
	println("klasör " + s.root)
	s.noteOpen(s.root)
	return nil
}

func openExisting(url string) bool {
	if !panelPage(url) {
		return false
	}
	launchPanel(url)
	return true
}

func panelPage(url string) bool {
	client := http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	return resp.StatusCode == http.StatusOK && bytes.Contains(body, []byte("<title>US</title>"))
}

func openWhenReady(url string) {
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 50; i++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				launchPanel(url)
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	launchPanel(url)
}

func guiMux(s *session) http.Handler {
	g := &guiServer{s: s}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(renderGUIPage(guiPage))
	})
	mux.HandleFunc("/api/state", g.state)
	mux.HandleFunc("/api/act", g.act)
	mux.HandleFunc("/view", g.viewFile)
	return mux
}

func (g *guiServer) state(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "yöntem yok", http.StatusMethodNotAllowed)
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	unlock := g.s.hold()
	defer unlock()
	g.s.drainIO()
	writeJSON(w, g.snapshot())
}

func (g *guiServer) act(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "yöntem yok", http.StatusMethodNotAllowed)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "http" || !strings.EqualFold(u.Host, r.Host) {
			http.Error(w, "Kaynak doğrulanamadı.", http.StatusForbidden)
			return
		}
	}
	var body struct {
		Op     string `json:"op"`
		Text   string `json:"text"`
		Path   string `json:"path"`
		Number int    `json:"n"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 24<<20)).Decode(&body) != nil {
		http.Error(w, "İstek okunamadı veya çok büyük.", http.StatusBadRequest)
		return
	}
	if body.Op == "task" && len(body.Text) > longInputLimit {
		http.Error(w, "İstem 8 MiB sınırını aşıyor; daha küçük bölümlere ayırın.", http.StatusRequestEntityTooLarge)
		return
	}
	g.mu.Lock()
	unlock := g.s.hold()
	defer func() {
		unlock()
		g.mu.Unlock()
	}()
	g.s.drainIO()
	switch body.Op {
	case "task":
		if g.s.calling {
			http.Error(w, "Model isteği zaten sürüyor.", http.StatusConflict)
			return
		}
		g.s.noteLog("istem", "kullanıcı: "+body.Text)
		if len(body.Text) <= longTaskThreshold && wantsSingleExe(body.Text) {
			root := g.s.root
			unlock()
			g.mu.Unlock()
			note := buildSingleExe(root)
			g.mu.Lock()
			unlock = g.s.hold()
			g.s.keepBuild(note)
			break
		}
		if len(body.Text) <= longTaskThreshold && reaPrompt(body.Text) {
			root := g.s.root
			unlock()
			g.mu.Unlock()
			note := runREA(root, "", body.Text)
			g.mu.Lock()
			unlock = g.s.hold()
			g.s.keepREA(note)
			break
		}
		text := body.Text
		if len(g.s.attached) > 0 {
			text = strings.TrimSpace(text + "\nek: " + strings.Join(g.s.attached, ", "))
		}
		g.s.draft = text
		g.s.cursor = len([]rune(text))
		g.s.writing = true
		g.s.notice = "Model çağrılıyor."
		unlock()
		g.mu.Unlock()
		g.s.submit()
		g.s.finishFlow()
		g.mu.Lock()
		unlock = g.s.hold()
	case "model-stop":
		if g.s.calling && g.s.longActive {
			g.s.longStop = true
			g.s.notice = "Geçerli parça bitince işlem duracak; sonuç uygulanmayacak."
		}
	case "rea":
		root := g.s.root
		unlock()
		g.mu.Unlock()
		note := runREA(root, body.Path, body.Text)
		g.mu.Lock()
		unlock = g.s.hold()
		g.s.keepREA(note)
	case "tunnel-retry":
		if !autoTunnelOnThisPlatform() { http.Error(w, "Otomatik tünel yalnızca Windows sürümünde kullanılabilir.", http.StatusBadRequest); return }
		startAutoTunnel(g.s)
	case "gateway":
		if g.s.calling {
			http.Error(w, "İstek devam ederken bağlantı değiştirilemez.", http.StatusConflict)
			return
		}
		base := strings.TrimRight(strings.TrimSpace(body.Text), "/")
		if err := validateGateway(base); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := os.Setenv("US_GATEWAY_URL", base); err != nil {
			http.Error(w, "Bağlantı ayarlanamadı.", http.StatusInternalServerError)
			return
		}
		if len(body.Path) > 256 || strings.ContainsAny(body.Path, "\r\n") {
			http.Error(w, "Model adı geçersiz.", http.StatusBadRequest)
			return
		}
		if err := os.Setenv("US_MODEL_NAME", strings.TrimSpace(body.Path)); err != nil {
			http.Error(w, "Model adı ayarlanamadı.", http.StatusInternalServerError)
			return
		}
		g.s.notice = "Bağlantı adresi güncellendi."
	case "key":
		g.s.key = body.Text
		g.s.keyCursor = len([]rune(body.Text))
	case "worker":
		switch body.Text {
		case "q5", "flash", "a100", "q8":
			g.s.retarget(workerID(body.Text))
		default:
			g.s.retarget("q8")
		}
	case "select":
		if body.Number >= 0 && body.Number < len(modes) {
			g.s.selected = body.Number
			g.s.scroll = 0
		}
	case "chat-new":
		g.s.newChat()
	case "chat-use":
		g.s.useChat(body.Text)
	case "chat-delete":
		g.s.deleteChat()
	case "browse":
		g.s.notice = ""
		if body.Text == "" || body.Text == "yerel" {
			g.s.showLocalPlaces()
		} else if body.Text != "" {
			g.s.browse = body.Text
			g.s.picking = true
			g.s.loadBrowse()
		}
	case "enter":
		g.s.enterPlace(filepath.Join(g.s.root, strings.Trim(body.Text, `/\`)), false)
	case "goto":
		g.s.enterPlace(body.Text, false)
	case "up":
		g.s.notice = ""
		parent := filepath.Dir(g.s.browse)
		if parent == g.s.browse || parent == "." {
			g.s.showLocalPlaces()
		} else {
			g.s.browse = parent
			g.s.picking = true
			g.s.loadBrowse()
		}
	case "pick":
		if body.Number >= 0 && body.Number < len(g.s.picks) {
			g.s.pick = body.Number
			g.s.activatePick()
		}
	case "refresh":
	case "read":
		g.s.readFile(body.Path)
		g.s.noteLog("işlem", "Dosya okundu: "+body.Path+". Panelde açıldı. Gövde kayda geçmedi.")
	case "save":
		g.s.saveFile(body.Path, body.Text)
		g.s.noteLog("işlem", "Dosya kaydedildi: "+body.Path+". Gövde kayda geçmedi.")
	case "append":
		g.s.appendFile(body.Path, body.Text)
	case "find":
		g.s.findInFolder(body.Text)
	case "track":
		g.s.addTrack(body.Text, body.Number)
	case "note":
		g.s.addNote(body.Text)
	case "review":
		g.s.reviewNow()
	case "run":
		root := g.s.root
		unlock()
		g.mu.Unlock()
		line := g.s.startProgram(root, "")
		g.mu.Lock()
		unlock = g.s.hold()
		g.s.project.RunLine = line
		g.s.noteLog("işlem", "Program çalıştırıldı.")
		g.s.notice = ""
	case "stop":
		g.s.runGen++
		g.s.stopProgram()
		g.s.runStatus = "durum  durdu"
		g.s.noteLog("işlem", "Program durduruldu.")
		g.s.notice = ""
	case "license":
		g.s.notice = g.s.acceptLicense(body.Text, time.Now())
	case "mint":
		if !g.s.isAdmin() {
			g.s.notice = "Yönetim kapalı."
			break
		}
		token, err := mintLicense(body.Text, time.Now())
		if err != nil {
			g.s.notice = err.Error()
			break
		}
		g.s.licenseNote = token
		g.s.notice = "Lisans üretildi."
	case "check":
		if !g.s.isAdmin() {
			g.s.notice = "Yönetim kapalı."
			break
		}
		doc, err := parseLicense(body.Text, time.Now())
		if err != nil {
			g.s.notice = err.Error()
			break
		}
		state := licenseState{}
		if strings.TrimSpace(body.Text) == strings.TrimSpace(g.s.ledger.Token) {
			state = g.s.ledger
		}
		g.s.notice = describeLicense(doc, state)
	case "use":
		g.s.useProject(body.Number)
	case "attach":
		g.s.attach(body.Path)
	case "detach":
		g.s.detach(body.Path)
	case "plugin":
		g.s.addPlugin(body.Text)
	case "repo":
		g.s.repoLine = repoStatus(g.s.root)
		g.s.notice = ""
		g.s.noteLog("işlem", "depo\n"+g.s.repoLine)
	case "github":
		root := g.s.root
		unlock()
		g.mu.Unlock()
		line := githubStatus(root)
		g.mu.Lock()
		unlock = g.s.hold()
		g.s.hubLine = line
		g.s.notice = ""
		g.s.noteLog("işlem", "github\n"+line)
	case "render":
		if view, ok := visualPath(body.Path); ok {
			g.s.visual = view
		} else {
			g.s.visual = "/view?engine=tuval"
		}
		g.s.notice = ""
		g.s.noteLog("işlem", "tuval "+g.s.visual)
	case "skill":
		switch strings.ToLower(strings.TrimSpace(body.Text)) {
		case "rea":
			root, open := g.s.root, g.s.openPath
			unlock()
			g.mu.Unlock()
			note := runREA(root, open, "inspect-artifact")
			g.mu.Lock()
			unlock = g.s.hold()
			g.s.keepREA(note)
		case "tek exe", "exe":
			root := g.s.root
			unlock()
			g.mu.Unlock()
			note := buildSingleExe(root)
			g.mu.Lock()
			unlock = g.s.hold()
			g.s.keepBuild(note)
		default:
			g.s.notice = g.s.runSkill(body.Text)
		}
	case "auto":
		root := g.s.root
		unlock()
		g.mu.Unlock()
		repo := repoStatus(root)
		hub := githubStatus(root)
		g.mu.Lock()
		unlock = g.s.hold()
		g.s.repoLine = repo
		g.s.hubLine = hub
		g.s.visual = "/view?engine=tuval"
		g.s.review = trimRunes(repo+"\n"+hub, 700)
		g.s.notice = ""
		g.s.noteLog("işlem", "bakış")
		g.s.noteLog("düşünce", g.s.review)
	default:
		http.Error(w, "işlem yok", http.StatusBadRequest)
		return
	}
	writeJSON(w, g.snapshot())
}

func (s *session) drainIO() {
	if s.modelCh != nil {
		for {
			select {
			case result := <-s.modelCh:
				s.applyModel(result)
			default:
				goto runs
			}
		}
	}
runs:
	if s.procDone == nil {
		return
	}
	for {
		select {
		case note := <-s.procDone:
			s.applyProgram(note)
		default:
			return
		}
	}
}

type guiSnap struct {
	Worker    string        `json:"worker"`
	Selected  int           `json:"selected"`
	Summary   string        `json:"summary"`
	Rows      []guiRow      `json:"rows"`
	Root      string        `json:"root"`
	Browse    string        `json:"browse"`
	Picking   bool          `json:"picking"`
	Picks     []guiPick     `json:"picks"`
	Ask       string        `json:"ask"`
	Notice    string        `json:"notice"`
	Calling   bool          `json:"calling"`
	LongActive bool         `json:"longActive"`
	Tunnel    string        `json:"tunnel"`
	RunStatus string        `json:"runStatus"`
	KeyMask   string        `json:"keyMask"`
	KeySet    bool          `json:"keySet"`
	DoorOK    bool          `json:"doorOk"`
	DoorMS    int           `json:"doorMs"`
	Ollama    bool          `json:"ollama"`
	DoorError string        `json:"doorError"`
	Trial     string        `json:"trial"`
	Steps     []guiStep     `json:"steps"`
	Tree      []string      `json:"tree"`
	OpenPath  string        `json:"openPath"`
	OpenText  string        `json:"openText"`
	Hits      []string      `json:"hits"`
	Tracks    []trackItem   `json:"tracks"`
	Notes     []string      `json:"notes"`
	Review    string        `json:"review"`
	Rea       string        `json:"rea"`
	Tools     []string      `json:"tools"`
	Theory    string        `json:"theory"`
	Projects  []projectCard `json:"projects"`
	Plugins   []pluginCard  `json:"plugins"`
	Attached  []string      `json:"attached"`
	Repo      string        `json:"repo"`
	GitHub    string        `json:"github"`
	Visual    string        `json:"visual"`
	Journal   []journalItem `json:"journal"`
	Chats     []chatHead    `json:"chats"`
	ChatID    string        `json:"chatId"`
	Turns     []chatTurn    `json:"turns"`
	CallLine  string        `json:"callLine"`
	LoadLine  string        `json:"loadLine"`
	IdleLine  string        `json:"idleLine"`
	IdleLeft  int           `json:"idleLeft"`
	Allowance string        `json:"allowance"`
	License   string        `json:"licenseNote"`
	Admin     bool          `json:"admin"`
}

type guiRow struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Stance string `json:"stance"`
	Egress int    `json:"egress"`
	Calls  int    `json:"calls"`
}

type chatHead struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type guiPick struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type guiStep struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Kind  string `json:"kind"`
}

func (g *guiServer) snapshot() guiSnap {
	s := g.s
	s.ensureDesk()
	v := s.view()
	rows := make([]guiRow, 0, len(v.rows))
	for _, row := range v.rows {
		rows = append(rows, guiRow{ID: string(row.ID), Name: row.Name, Stance: row.Stance, Egress: row.EgressBytes, Calls: row.ModelCalls})
	}
	picks := make([]guiPick, 0, len(s.picks))
	for _, item := range s.picks {
		picks = append(picks, guiPick{Kind: item.kind, Label: item.label})
	}
	steps := make([]guiStep, 0, len(s.flow))
	for _, item := range s.flow {
		steps = append(steps, guiStep{Title: item.title, Body: item.body, Kind: item.kind})
	}
	ask := ""
	if s.ask != nil {
		ask = s.ask.question
	}
	callLine, loadLine, idleLine, idleLeft := s.lifeLines(time.Now())
	s.ensureChats()
	heads := make([]chatHead, 0, len(s.chats))
	for _, item := range s.chats {
		heads = append(heads, chatHead{ID: item.ID, Title: item.Title})
	}
	turns := append([]chatTurn(nil), s.chat...)
	return guiSnap{
		Worker:    string(s.worker),
		Selected:  s.selected,
		Summary:   v.summary,
		Rows:      rows,
		Root:      s.root,
		Browse:    s.browse,
		Picking:   s.picking,
		Picks:     picks,
		Ask:       ask,
		Notice:    s.notice,
		Calling:   s.calling,
		LongActive: s.longActive,
		Tunnel:    autoTunnel.status(),
		RunStatus: s.runStatus,
		KeyMask:   maskKey(s.key, len([]rune(s.key)), false),
		KeySet:    s.key != "",
		DoorOK:    s.resources.DoorOK,
		DoorMS:    s.resources.DoorMS,
		Ollama:    s.resources.Ollama,
		DoorError: s.resources.Error,
		Trial:     s.trial.Line,
		Steps:     steps,
		Tree:      workspaceNames(s.root),
		OpenPath:  s.openPath,
		OpenText:  s.openText,
		Hits:      s.hits,
		Tracks:    s.tracks,
		Notes:     s.notes,
		Review:    s.review,
		Rea:       s.reaNote,
		Tools:     usTools,
		Theory:    usTheory,
		Projects:  s.projects,
		Plugins:   s.pluginView(),
		Attached:  s.attached,
		Repo:      s.repoLine,
		GitHub:    s.hubLine,
		Visual:    s.visual,
		Journal:   s.journal,
		Chats:     heads,
		ChatID:    s.chatID,
		Turns:     turns,
		CallLine:  callLine,
		LoadLine:  loadLine,
		IdleLine:  idleLine,
		IdleLeft:  idleLeft,
		Allowance: s.allowanceLine(time.Now()),
		License:   s.licenseNote,
		Admin:     s.isAdmin(),
	}
}

func (g *guiServer) viewFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "yöntem yok", http.StatusMethodNotAllowed)
		return
	}
	g.mu.Lock()
	unlock := g.s.hold()
	defer func() {
		unlock()
		g.mu.Unlock()
	}()
	if r.URL.Query().Get("engine") == "tuval" {
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		_, _ = w.Write([]byte(boardSVG(filepath.Base(g.s.root), g.s.tracks)))
		return
	}
	full, err := g.s.safePath(r.URL.Query().Get("path"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var ctype string
	switch strings.ToLower(filepath.Ext(full)) {
	case ".svg":
		ctype = "image/svg+xml; charset=utf-8"
	case ".html", ".htm":
		ctype = "text/html; charset=utf-8"
	case ".png":
		ctype = "image/png"
	default:
		http.Error(w, "görsel değil", http.StatusUnsupportedMediaType)
		return
	}
	info, err := os.Stat(full)
	if err != nil || info.IsDir() || info.Size() > 256*1024 {
		http.NotFound(w, r)
		return
	}
	body, err := os.ReadFile(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", ctype)
	_, _ = w.Write(body)
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}
