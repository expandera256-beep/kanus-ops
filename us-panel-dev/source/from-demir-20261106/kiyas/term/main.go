package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "--window" {
		runAppWindow(os.Args[2])
		return
	}
	dump := flag.Bool("dump", false, "karşılaştırmayı bir kez yaz ve çık")
	gui := flag.Bool("gui", false, "tarayıcı arayüzünü aç")
	workerFlag := flag.String("worker", "q8", "q8, q5, flash veya a100")
	taskFlag := flag.String("task", "", "görevi bir kez yerelde yürüt")
	rootFlag := flag.String("root", "", "çalışma klasörü")
	flag.Parse()
	worker := workerID("q8")
	switch *workerFlag {
	case "q5":
		worker = "q5"
	case "flash":
		worker = "flash"
	case "a100":
		worker = "a100"
	}
	state := newView(worker)
	state.trial = runTrial()
	state.resWait = true
	state.root = localFolder(*rootFlag)
	state.browse = state.root
	state.asUS = namedUS(os.Args[0])
	// A downloaded copy is often named US (1).exe. On Windows that file is
	// still the panel: there is no console window to close, and closing one
	// was what dropped the listener and made the browser report connection refused.
	panel := *gui || state.asUS || (runtime.GOOS == "windows" && strings.TrimSpace(*taskFlag) == "" && !*dump)
	if panel {
		if err := serveGUI(&state); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if strings.TrimSpace(*taskFlag) != "" {
		state.draft = *taskFlag
		state.submit()
		state.finishFlow()
		fmt.Println(render(state.view()))
		state.stopProgram()
		return
	}
	if *dump || !isTerminal() {
		state.resources = probe()
		fmt.Println(render(state.view()))
		return
	}
	if err := runTUI(&state); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func localFolder(chosen string) string {
	chosen = strings.Trim(strings.TrimSpace(chosen), `"'`)
	if folderOK(chosen) {
		return chosen
	}
	if wd, err := os.Getwd(); err == nil && folderOK(wd) && !ephemeralDir(wd) {
		return wd
	}
	if home, err := os.UserHomeDir(); err == nil && folderOK(home) {
		return home
	}
	return "."
}

func ephemeralDir(path string) bool {
	clean := strings.ToLower(filepath.Clean(path))
	return strings.Contains(clean, string(filepath.Separator)+"tmp"+string(filepath.Separator)) || strings.HasPrefix(clean, string(filepath.Separator)+"tmp") || strings.Contains(clean, `\temp\`) || strings.HasSuffix(clean, `\temp`) || strings.Contains(clean, `/tmp/`) || clean == "/tmp"
}

func folderOK(path string) bool {
	if strings.TrimSpace(path) == "" || path == "." {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir() && !refusedFolder(path)
}

func namedUS(path string) bool {
	path = strings.ReplaceAll(path, `\`, "/")
	base := strings.ToLower(path)
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	base = strings.TrimSuffix(base, ".exe")
	return base == "us"
}

type session struct {
	worker            workerID
	selected          int
	scroll            int
	trial             trialResult
	resources         resourceSnap
	resWait           bool
	draft             string
	cursor            int
	writing           bool
	picking           bool
	root              string
	browse            string
	picks             []pickItem
	pick              int
	project           projectRun
	ran               bool
	notice            string
	flow              []event
	shown             int
	flowing           bool
	flowAt            time.Time
	lintReady         trialResult
	follow            bool
	pending           []byte
	key               string
	keyCursor         int
	keying            bool
	modelCh           chan modelResult
	modelSeq          int
	modelCtx          context.Context
	modelCancel       context.CancelFunc
	testCancel        context.CancelFunc
	testing           bool
	lastTaskID        string
	appliedModel      int
	sendModel         func(worker workerID, key, wire string) modelResult
	proc              *os.Process
	procDone          chan programNote
	runStatus         string
	runGen            int
	ask               *clarification
	program           string
	modelQuestion     string
	modelAnswer       string
	pendingKind       string
	pendingSubject    string
	pendingFresh      bool
	pendingTask       string
	siteSrv           *http.Server
	asUS              bool
	lock              *sync.Mutex
	depth             int
	openPath          string
	openText          string
	openRoot          string
	openHash          string
	openEditable      bool
	openRevision      int
	hits              []string
	tracks            []trackItem
	notes             []string
	review            string
	reaNote           string
	projects          []projectCard
	plugins           []pluginCard
	attached          []string
	repoLine          string
	hubLine           string
	visual            string
	deskReady         bool
	journal           []journalItem
	chat              []chatTurn
	chats             []chatThread
	chatID            string
	mem               jobMemory
	admin             bool
	ledger            licenseState
	licenseNote       string
	openMark          time.Time
	calling           bool
	callStart         time.Time
	callStop          time.Time
	tracked           workerID
	loading           bool
	seenAwake         bool
	loadStart         time.Time
	loadStop          time.Time
	lastBeat          time.Time
	liveChars         int
	modelStage        string
	longActive        bool
	longStop          bool
	reviewBeforeApply bool
	proposal          *fileProposal
	testReport        string
	accountGateway    string
	accountHistory    []accountModelRequest
}

type resourceSnap struct {
	DoorOK  bool
	DoorMS  int
	Workers []workerStatus
	Ollama  bool
	Error   string
	Managed []accountWorkerState
}

func newView(worker workerID) session {
	if base := os.Getenv("US_DESK_BASE"); base != "" {
		os.Setenv("US_DESK", base)
	}
	isolateSharedDesk()
	root, err := os.Getwd()
	if err != nil || root == "" {
		root = "."
	}
	return session{worker: worker, selected: 2, writing: true, root: root, browse: root, follow: true, procDone: make(chan programNote, 1), lock: &sync.Mutex{}}
}

func (s session) view() view {
	scene := makeScenario(0)
	rows := runAll(s.worker, scene)
	return view{
		worker:    s.worker,
		selected:  s.selected,
		scroll:    s.scroll,
		rows:      rows,
		summary:   compare(rows),
		task:      scene.Task,
		addition:  scene.Addition,
		trial:     s.trial,
		draft:     s.draft,
		cursor:    s.cursor,
		writing:   s.writing,
		picking:   s.picking,
		project:   s.root,
		browse:    s.browse,
		picks:     s.picks,
		pick:      s.pick,
		user:      s.project,
		ran:       s.ran,
		asUS:      s.asUS,
		notice:    s.notice,
		flow:      s.flow,
		shown:     s.shown,
		follow:    s.follow,
		doorOK:    s.resources.DoorOK,
		doorMS:    s.resources.DoorMS,
		workers:   s.resources.Workers,
		ollama:    s.resources.Ollama,
		probe:     s.resources.Error,
		keyMask:   maskKey(s.key, s.keyCursor, s.keying),
		keying:    s.keying,
		runStatus: s.runStatus,
		asking:    s.ask != nil,
		width:     100,
		height:    36,
	}
}

func probe() resourceSnap {
	if commercialRequired() { return resourceSnap{Error: "Sunucu durumu için hesabınıza giriş yapın."} }
	client := http.Client{Timeout: 4 * time.Second}
	started := time.Now()
	snap := resourceSnap{}
	if activeGateway() == tunnelURL && strings.TrimSpace(os.Getenv("US_MODEL_NAME")) == tunnelModel {
		snap.DoorOK = modelAvailable(tunnelURL, &client)
		snap.DoorMS = int(time.Since(started).Milliseconds())
		if !snap.DoorOK {
			snap.Error = "DeepSeek bağlantısı hazır değil."
		}
		return snap
	}
	response, err := client.Get(activeGateway() + "/api/durum")
	snap.DoorMS = int(time.Since(started).Milliseconds())
	if err != nil {
		snap.Error = "Kapıya ulaşılamadı. Sayım yerelde durur."
	} else {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		if response.StatusCode == 200 {
			var parsed struct {
				Workers []workerStatus `json:"workers"`
			}
			if json.Unmarshal(body, &parsed) == nil {
				snap.DoorOK = true
				snap.Workers = parsed.Workers
			} else {
				snap.Error = "Kapı cevabı okunamadı."
			}
		} else {
			snap.Error = "Kapıya ulaşılamadı. Sayım yerelde durur."
		}
	}
	ollama, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err == nil {
		ollama.Body.Close()
		snap.Ollama = ollama.StatusCode == 200
	}
	return snap
}

func (s *session) hold() func() {
	if s.lock == nil {
		return func() {}
	}
	s.lock.Lock()
	s.depth++
	return func() {
		s.depth--
		s.lock.Unlock()
	}
}

func (s *session) withoutLock(fn func()) {
	if s.lock == nil || s.depth == 0 {
		fn()
		return
	}
	depth := s.depth
	for s.depth > 0 {
		s.depth--
		s.lock.Unlock()
	}
	fn()
	for i := 0; i < depth; i++ {
		s.lock.Lock()
		s.depth++
	}
}

func runTUI(s *session) error {
	enableWindowsVT()
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, old)
	defer s.stopProgram()
	fmt.Fprint(os.Stdout, "\033[?1049h\033[?25l\033[?2004h")
	defer fmt.Fprint(os.Stdout, "\033[?2004l\033[?25h\033[?1049l")

	probes := make(chan resourceSnap, 4)
	go func() { probes <- probe() }()
	if s.modelCh == nil {
		s.modelCh = make(chan modelResult, 4)
	}
	modelCh := s.modelCh
	keys := make(chan []byte, 8)
	go readKeys(keys)

	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	for {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || width <= 0 {
			width, height = 80, 24
		}
		unlock := s.hold()
		s.advanceFlow(time.Now())
		snap := *s
		unlock()
		paint(snap, width, height)
		select {
		case snap := <-probes:
			unlock = s.hold()
			s.resources = snap
			s.resWait = false
			unlock()
		case result := <-modelCh:
			unlock = s.hold()
			s.applyModel(result)
			unlock()
		case note := <-s.procDone:
			unlock = s.hold()
			s.applyProgram(note)
			unlock()
		case key, ok := <-keys:
			if !ok {
				return nil
			}
			unlock = s.hold()
			events, rest := parseKeys(append(s.pending, key...))
			if len(rest) == 1 && rest[0] == 27 {
				events = append(events, keyEvent{kind: "escape"})
				rest = nil
			}
			s.pending = rest
			stop := false
			for _, event := range events {
				if !handleEvent(s, event, probes) {
					stop = true
					break
				}
			}
			unlock()
			if stop {
				return nil
			}
		case <-tick.C:
		}
	}
}

func readKeys(keys chan<- []byte) {
	buf := make([]byte, 16)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			item := make([]byte, n)
			copy(item, buf[:n])
			keys <- item
		}
		if err != nil {
			close(keys)
			return
		}
	}
}

func handleEvent(s *session, ev keyEvent, probes chan<- resourceSnap) bool {
	switch ev.kind {
	case "ctrlc":
		return false
	case "ctrlp":
		s.openPicker()
		return true
	case "escape":
		if s.picking {
			s.picking = false
			s.writing = true
			return true
		}
		if s.keying {
			s.keying = false
			s.writing = true
			return true
		}
		if s.ask != nil {
			s.ask = nil
			s.notice = "Soru kapatıldı."
			s.writing = true
			return true
		}
		s.writing = !s.writing
		return true
	case "tab":
		if s.asUS || s.picking {
			return true
		}
		if !s.writing {
			s.writing = true
		}
		s.keying = !s.keying
		return true
	case "enter":
		if s.picking {
			s.activatePick()
			return true
		}
		if s.keying {
			s.keying = false
			return true
		}
		s.submit()
		return true
	case "backspace":
		if s.picking {
			if parent := filepath.Dir(s.browse); parent != s.browse {
				s.browse = parent
				s.pick = 0
				s.loadBrowse()
			}
			return true
		}
		if s.keying {
			s.keyBackspace()
			return true
		}
		if s.writing {
			s.backspace()
		}
		return true
	case "clear":
		if s.picking {
			return true
		}
		if s.keying {
			s.key = ""
			s.keyCursor = 0
			return true
		}
		if s.writing {
			s.draft = ""
			s.cursor = 0
		}
		return true
	case "rune":
		if s.picking {
			s.jumpPick(ev.r)
			return true
		}
		if s.keying {
			if ev.r != '\n' && ev.r != '\r' {
				s.keyInsert(ev.r)
			}
			return true
		}
		if s.writing {
			s.insert(ev.r)
			return true
		}
		return s.command(ev.r, probes)
	case "up":
		if s.picking {
			if s.pick > 0 {
				s.pick--
			}
			return true
		}
		s.follow = false
		if s.scroll > 0 {
			s.scroll--
		}
	case "down":
		if s.picking {
			if s.pick < len(s.picks)-1 {
				s.pick++
			}
			return true
		}
		s.follow = false
		s.scroll++
	case "left":
		if s.picking {
			return true
		}
		if s.keying {
			if s.keyCursor > 0 {
				s.keyCursor--
			}
			return true
		}
		if s.writing {
			if s.cursor > 0 {
				s.cursor--
			}
			return true
		}
		if s.selected > 0 {
			s.selected--
			s.scroll = 0
		}
	case "right":
		if s.picking {
			return true
		}
		if s.keying {
			if s.keyCursor < len([]rune(s.key)) {
				s.keyCursor++
			}
			return true
		}
		if s.writing {
			if s.cursor < len([]rune(s.draft)) {
				s.cursor++
			}
			return true
		}
		if s.selected < len(modes)-1 {
			s.selected++
			s.scroll = 0
		}
	}
	return true
}

func (s *session) command(r rune, probes chan<- resourceSnap) bool {
	switch r {
	case 'q', 'Q':
		return false
	case '8':
		s.retarget("q8")
		s.scroll = 0
	case '5':
		s.retarget("q5")
		s.scroll = 0
	case 'f', 'F':
		s.retarget("flash")
		s.scroll = 0
	case 'a', 'A':
		s.retarget("a100")
		s.scroll = 0
	case 'r', 'R':
		if probes != nil {
			go func() { probes <- probe() }()
		}
	case 'j', 'J':
		s.follow = false
		s.scroll++
	case 'k', 'K':
		s.follow = false
		if s.scroll > 0 {
			s.scroll--
		}
	case 'h', 'H':
		if s.selected > 0 {
			s.selected--
			s.scroll = 0
		}
	case 'l', 'L':
		if s.selected < len(modes)-1 {
			s.selected++
			s.scroll = 0
		}
	case '1':
		s.selected = 0
		s.scroll = 0
	case '2':
		s.selected = 1
		s.scroll = 0
	case '3':
		s.selected = 2
		s.scroll = 0
	}
	return true
}

func paint(s session, width, height int) {
	v := s.view()
	v.width = width
	v.height = height
	if s.resWait && v.probe == "" && !v.doorOK {
		v.probe = "Kapıya bakılıyor."
	}
	fmt.Fprint(os.Stdout, "\033[H\033[J")
	fmt.Fprint(os.Stdout, render(v))
}
