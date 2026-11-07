package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

type view struct {
	worker    workerID
	selected  int
	scroll    int
	rows      []runResult
	summary   string
	task      string
	addition  string
	trial     trialResult
	draft     string
	cursor    int
	writing   bool
	picking   bool
	project   string
	browse    string
	picks     []pickItem
	pick      int
	user      projectRun
	ran       bool
	notice    string
	flow      []event
	shown     int
	follow    bool
	doorOK    bool
	doorMS    int
	workers   []workerStatus
	ollama    bool
	probe     string
	keyMask   string
	keying    bool
	runStatus string
	asking    bool
	asUS      bool
	width     int
	height    int
}

type workerStatus struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Awake bool   `json:"awake"`
}

type event struct {
	title string
	meta  string
	body  string
	kind  string
}

var (
	ink    = lipgloss.Color("#0c0e13")
	paper  = lipgloss.Color("#f3efe7")
	muted  = lipgloss.Color("#9a9488")
	line   = lipgloss.Color("#2c313c")
	copper = lipgloss.Color("#e7a06a")
	blue   = lipgloss.Color("#8eb6ff")
	green  = lipgloss.Color("#7dcea0")
	warn   = lipgloss.Color("#f0a3a3")
	barBg  = lipgloss.Color("#171b24")
	cardBg = lipgloss.Color("#1c2433")
	pickBg = lipgloss.Color("#173028")
	diffBg = lipgloss.Color("#12261c")
)

func grouped(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	raw := fmt.Sprintf("%d", n)
	var out []byte
	for i := 0; i < len(raw); i++ {
		if i > 0 && (len(raw)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, raw[i])
	}
	return sign + string(out)
}

func render(v view) string {
	width := v.width
	if width < 48 {
		width = 48
	}
	height := v.height
	if height < 18 {
		height = 18
	}
	row := selectedRow(v)
	top := lipgloss.JoinVertical(lipgloss.Left, band(titleBlock(v), workerBadge(v), width, barBg), resourceLine(v, width), projectLine(v, width))
	if v.runStatus != "" {
		top = lipgloss.JoinVertical(lipgloss.Left, top, statusLine(v.runStatus, width))
	}
	if v.probe != "" {
		top = lipgloss.JoinVertical(lipgloss.Left, top, lipgloss.NewStyle().Foreground(warn).MaxWidth(width).Render(v.probe))
	}
	if v.notice != "" {
		top = lipgloss.JoinVertical(lipgloss.Left, top, lipgloss.NewStyle().Foreground(warn).MaxWidth(width).Render(v.notice))
	}
	stance := stanceBlock(row, width)
	if v.ran && strings.TrimSpace(v.user.Task) != "" {
		line := v.user.Reading
		if line == "" {
			line = v.user.Task
		}
		stance = lipgloss.NewStyle().Bold(true).Foreground(paper).MaxWidth(width).Render(line)
	} else if v.ran {
		stance = lipgloss.NewStyle().Bold(true).Foreground(copper).MaxWidth(width).Render(row.Name)
	}
	top = lipgloss.JoinVertical(lipgloss.Left, top, stance)
	bottomParts := []string{}
	if strings.TrimSpace(v.user.Task) == "" || (!v.writing && !v.picking && !v.keying) {
		bottomParts = append(bottomParts, rail(v, width), summaryBlock(v, width, height))
	}
	bottomParts = append(bottomParts, promptBox(v, width))
	if !v.picking && !v.asUS {
		bottomParts = append(bottomParts, keyLine(v, width))
	}
	bottomParts = append(bottomParts, band(keyHint(v), "", width, barBg))
	bottom := lipgloss.JoinVertical(lipgloss.Left, bottomParts...)
	room := height - lipgloss.Height(top) - lipgloss.Height(bottom)
	if room < 0 {
		room = 0
	}
	diff := ""
	diffLines := 0
	if !v.picking && !v.ran && room >= 8 {
		diff = diffBlock(v.addition, width)
		diffLines = lipgloss.Height(diff) + 1
		if diffLines >= room {
			diff = ""
			diffLines = 0
		}
	}
	items := shownEvents(v)
	start := v.scroll
	budget := room - diffLines
	if budget < 0 {
		budget = 0
	}
	if v.follow {
		start = flowScroll(items, width, budget)
	}
	var transcript string
	if v.picking {
		transcript = pickerBlock(v, width, room)
	} else if len(items) == 0 {
		transcript = holdBlock(v, width, budget)
	} else {
		transcript = transcriptBlock(items, start, width, budget)
		if v.shown < len(v.flow) {
			transcript = appendFit(transcript, lipgloss.NewStyle().MaxWidth(width).Foreground(muted).Render("İş akışı sürüyor."), budget)
		}
	}
	parts := []string{top}
	if transcript != "" {
		parts = append(parts, transcript)
	}
	if diff != "" {
		parts = append(parts, diff)
	}
	parts = append(parts, bottom)
	return fillScreen(lipgloss.JoinVertical(lipgloss.Left, parts...), width, height)
}

func titleBlock(v view) string {
	mark := lipgloss.NewStyle().Background(barBg).Foreground(green).Bold(true).Render("✦")
	brand := " Kıyas "
	if v.asUS {
		brand = " US "
	}
	name := lipgloss.NewStyle().Background(barBg).Foreground(paper).Bold(true).Render(brand)
	label := "YAZ"
	badgeBG := green
	if v.asking {
		label = "SORU"
		badgeBG = blue
	} else if v.picking {
		label = "PROJE"
		badgeBG = blue
	} else if v.keying {
		label = "ANAHTAR"
		badgeBG = copper
	} else if !v.writing {
		label = "KOMUT"
		badgeBG = copper
	}
	badge := lipgloss.NewStyle().Background(badgeBG).Foreground(ink).Bold(true).Render(" " + label + " ")
	return mark + name + badge
}

func workerBadge(v view) string {
	label := strings.ToUpper(string(v.worker))
	if v.worker == "a100" {
		label = "KAN"
	}
	if label == "" {
		label = "Q8"
	}
	return lipgloss.NewStyle().Background(barBg).Foreground(copper).Bold(true).Render(label)
}

func keyHint(v view) string {
	text := "enter yürüt    tab anahtar    ^P proje    esc komut    ^C çık"
	if v.asUS {
		text = "enter yürüt    ^P proje    esc komut    ^C çık"
	}
	if v.asking {
		text = "cevap yaz    enter gönder    esc soruyu kapat    ^C çık"
	} else if v.picking {
		text = "enter aç    bs yukarı    esc kapat    ^C çık"
	} else if v.keying {
		text = "enter göreve dön    tab görev    ^U sil    ^C çık"
	} else if !v.writing {
		text = "1 2 3 işleyiş    8/5 işçi    j k    q çık    esc yaz"
	}
	return lipgloss.NewStyle().Background(barBg).Foreground(muted).Render(text)
}

func keyLine(v view, width int) string {
	mask := v.keyMask
	if mask == "" {
		mask = "boş"
	}
	color := muted
	if v.keying {
		color = copper
	}
	return lipgloss.NewStyle().MaxWidth(width).Foreground(color).Render("anahtar  " + mask + "  ·  tab")
}

func maskKey(key string, cursor int, keying bool) string {
	runes := []rune(key)
	n := len(runes)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > n {
		cursor = n
	}
	if n == 0 {
		if keying {
			return "boş ▌"
		}
		return "boş"
	}
	visible := n
	if visible > 12 {
		visible = 12
	}
	start := 0
	if n > visible {
		start = cursor - visible
		if cursor == n {
			start = n - visible
		}
		if start < 0 {
			start = 0
		}
		if start > n-visible {
			start = n - visible
		}
	}
	end := start + visible
	if end > n {
		end = n
	}
	var b strings.Builder
	if start > 0 {
		b.WriteString("…")
	}
	for i := start; i < end; i++ {
		if keying && i == cursor {
			b.WriteString("▌")
		}
		b.WriteString("●")
	}
	if keying && cursor >= end {
		b.WriteString("▌")
	}
	return b.String()
}

func statusLine(status string, width int) string {
	color := copper
	if strings.Contains(status, "durdu") {
		color = warn
	} else if strings.Contains(status, "bitti") {
		color = green
	}
	return lipgloss.NewStyle().MaxWidth(width).Foreground(color).Bold(true).Render("● " + status)
}

func projectLine(v view, width int) string {
	label := "(seçilmedi)"
	if v.project != "" {
		label = v.project
	}
	return lipgloss.NewStyle().MaxWidth(width).Foreground(blue).Render("▣  proje  " + redactLine(label))
}

func selectedRow(v view) runResult {
	if v.selected >= 0 && v.selected < len(v.rows) {
		return v.rows[v.selected]
	}
	if len(v.rows) == 0 {
		return runResult{Name: "Kıyas", Stance: "Görev yok."}
	}
	return v.rows[0]
}

func stanceBlock(row runResult, width int) string {
	name := lipgloss.NewStyle().Bold(true).Foreground(copper).Render(row.Name)
	stance := lipgloss.NewStyle().Width(width).Foreground(muted).Italic(true).Render(row.Stance)
	return lipgloss.JoinVertical(lipgloss.Left, name, stance)
}

func resourceLine(v view, width int) string {
	doorText := "kapı yok"
	doorOn := false
	if v.doorOK {
		doorText = fmt.Sprintf("kapı %d ms", v.doorMS)
		doorOn = true
	}
	parts := []string{chip(doorText, doorOn, green)}
	if len(v.workers) == 0 {
		parts = append(parts, chip("işçiler bakılıyor", false, muted))
	}
	for _, item := range v.workers {
		state := "uykuda"
		on := false
		if item.Awake {
			state = "uyanık"
			on = true
		}
		name := item.Name
		if name == "" {
			name = item.ID
		}
		if item.ID == string(v.worker) {
			name = "▸ " + name
		}
		parts = append(parts, chip(name+" "+state, on, copper))
	}
	ollama := "Ollama yok"
	if v.ollama {
		ollama = "Ollama dinliyor"
	}
	parts = append(parts, chip(ollama, v.ollama, blue))
	if v.trial.Line != "" {
		label := "deneme durdu"
		on := false
		if v.trial.OK {
			label = "deneme temiz"
			on = true
		}
		parts = append(parts, chip(label, on, green))
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(strings.Join(parts, "  "))
}

func chip(label string, on bool, color lipgloss.Color) string {
	mark := "○"
	fg := muted
	style := lipgloss.NewStyle().Foreground(fg)
	if on {
		mark = "●"
		style = lipgloss.NewStyle().Foreground(color).Background(cardBg)
	}
	return style.Render(" " + mark + " " + label + " ")
}

func shownEvents(v view) []event {
	if v.shown <= 0 || len(v.flow) == 0 {
		return nil
	}
	n := v.shown
	if n > len(v.flow) {
		n = len(v.flow)
	}
	return v.flow[:n]
}

func holdBlock(v view, width, room int) string {
	if room < 1 || width < 1 {
		return ""
	}
	text := "Emir bekleniyor. Yaz, enter ile istek okunur.\n› D diskinde yeni bir klasörde panel kur\n› … yapan proje kur\n› bu klasörde ne var"
	if v.ran {
		text = "Düşünce başlıyor."
	}
	block := lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(room).Foreground(muted).Render(text)
	lines := strings.Split(block, "\n")
	if len(lines) > room {
		lines = lines[:room]
	}
	return strings.Join(lines, "\n")
}

func appendFit(current, extra string, room int) string {
	if extra == "" || room < 1 {
		return current
	}
	if current == "" {
		if lipgloss.Height(extra) <= room {
			return extra
		}
		return ""
	}
	joined := lipgloss.JoinVertical(lipgloss.Left, current, extra)
	if lipgloss.Height(joined) <= room {
		return joined
	}
	return current
}

func flowScroll(items []event, width, room int) int {
	if len(items) == 0 || room < 1 {
		return 0
	}
	for start := 0; start < len(items); start++ {
		if transcriptHeight(items[start:], width) <= room {
			return start
		}
	}
	return len(items) - 1
}

func transcriptHeight(items []event, width int) int {
	used := 0
	for _, item := range items {
		used += lipgloss.Height(eventBlock(item, width))
	}
	return used
}

func fixtureEvents(v view) []event {
	if v.selected < 0 || v.selected >= len(v.rows) {
		return nil
	}
	row := v.rows[v.selected]
	out := []event{{title: "görev", body: v.task, kind: "user"}}
	if row.ID == modeAyrik {
		out = append(out,
			event{title: "environment", meta: "yerel", body: "işletim linux · kabuk bash · Europe/Istanbul · ayrik://", kind: "yerel"},
			event{title: "workspace", meta: "yerel", body: "pricing.php dizinde. Gövde yerelde, tele çıkmaz.", kind: "yerel"},
		)
	}
	for _, item := range row.Turns {
		kind := "uzak"
		meta := "uzak " + grouped(item.SentBytes) + " B"
		if item.Where == "yerel" {
			kind = "yerel"
			meta = "yerel"
			if item.KeptBytes > 0 {
				meta = "yerel " + grouped(item.KeptBytes) + " B kaldı"
			}
		}
		body := item.Note
		if row.ID == modeAyrik && item.Where == "uzak" {
			body += " dosya gövdesi yok."
		}
		if item.Leaked {
			body += " Anahtar bu turda gitti."
			kind = "uyari"
		}
		if item.Overflow {
			body += " Gönderim tavanı aşıldı."
		}
		if row.ID == modeAyrik && v.trial.Ran && strings.Contains(item.Title, "php -l") && v.trial.Line != "" {
			body = v.trial.Line
			if v.trial.OK {
				body += " · çıktı modele gitmedi."
			}
		}
		out = append(out, event{title: item.Title, meta: meta, body: body, kind: kind})
	}
	return out
}

func transcriptBlock(items []event, scroll, width, room int) string {
	if room < 1 || len(items) == 0 {
		return ""
	}
	start := scroll
	if start < 0 {
		start = 0
	}
	if start > len(items) {
		start = len(items)
	}
	var blocks []string
	used := 0
	for _, item := range items[start:] {
		block := eventBlock(item, width)
		need := lipgloss.Height(block)
		if used+need > room {
			break
		}
		blocks = append(blocks, block)
		used += need
	}
	if len(blocks) == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func eventBlock(item event, width int) string {
	if item.kind == "user" {
		return lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(3).Foreground(paper).Bold(true).Render("› " + item.body)
	}
	color := copper
	switch item.kind {
	case "yerel":
		color = green
	case "soru":
		color = blue
	case "uyari":
		color = warn
	}
	head := lipgloss.NewStyle().Foreground(color).Bold(true).Render("┃ " + item.title)
	meta := lipgloss.NewStyle().Foreground(muted).Render(item.meta)
	row := head
	if item.meta != "" && lipgloss.Width(head)+2+lipgloss.Width(meta) <= width {
		row = spread(head, meta, width)
	}
	body := lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(4).Foreground(muted).Render("  " + item.body)
	return lipgloss.JoinVertical(lipgloss.Left, row, body)
}

func diffBlock(addition string, width int) string {
	raw := strings.Split(strings.TrimRight(addition, "\n"), "\n")
	if len(raw) > 3 {
		raw = raw[:3]
	}
	lines := []string{lipgloss.NewStyle().Foreground(muted).Render("pricing.php  +" + fmt.Sprint(len(raw)))}
	for _, item := range raw {
		lines = append(lines, lipgloss.NewStyle().MaxWidth(width).Background(diffBg).Foreground(green).Render(" + "+item))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func rail(v view, width int) string {
	accents := []lipgloss.Color{copper, blue, green}
	parts := make([]string, 0, len(v.rows))
	for i, row := range v.rows {
		label := fmt.Sprintf("%s  %s B · %d", row.Name, grouped(row.EgressBytes), row.ModelCalls)
		style := lipgloss.NewStyle().Foreground(accents[i%3])
		if i == v.selected {
			label = "▸ " + label
			style = style.Background(cardBg).Bold(true)
		}
		parts = append(parts, style.Render(" "+label+" "))
	}
	if width >= 96 {
		return lipgloss.NewStyle().MaxWidth(width).Render(strings.Join(parts, "   "))
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(parts, "\n"))
}

func summaryBlock(v view, width, height int) string {
	limit := 3
	if height < 30 {
		limit = 2
	}
	if v.ran {
		return ""
	}
	return lipgloss.NewStyle().Width(width).MaxHeight(limit).Foreground(paper).Render(v.summary)
}

func promptBox(v view, width int) string {
	inner := width - 4
	if inner < 8 {
		inner = 8
	}
	shown := composerText(v.draft, v.cursor, v.writing && !v.keying, inner-2)
	if v.draft == "" && v.asking && v.writing && !v.keying {
		shown = "▌ cevap yaz"
	}
	text := lipgloss.NewStyle().Width(inner).Foreground(paper).MaxHeight(2).Render("› " + shown)
	border := green
	if v.asking || v.picking {
		border = blue
	} else if v.keying || !v.writing {
		border = line
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(inner).
		Render(text)
	if lipgloss.Width(box) > width || lipgloss.Height(box) > 5 {
		return lipgloss.NewStyle().Width(width).MaxHeight(2).Foreground(paper).Render("› " + shown)
	}
	return box
}

func composerText(draft string, cursor int, writing bool, budget int) string {
	if budget < 4 {
		budget = 4
	}
	runes := []rune(draft)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	if len(runes) == 0 {
		if writing {
			return "▌ istem yaz"
		}
		return "görev yok"
	}
	mark := ""
	if writing {
		mark = "▌"
	}
	start := 0
	if cursor+len([]rune(mark)) > budget {
		start = cursor - budget + 1
		if start < 0 {
			start = 0
		}
	}
	end := len(runes)
	body := string(runes[start:cursor]) + mark + string(runes[cursor:end])
	shown := []rune(body)
	if len(shown) > budget {
		shown = shown[:budget]
	}
	if start > 0 {
		return "…" + string(shown)
	}
	return string(shown)
}

func pickerBlock(v view, width, room int) string {
	if room < 1 {
		return ""
	}
	lines := []string{lipgloss.NewStyle().Foreground(paper).MaxWidth(width).Render(redactLine(v.browse))}
	for i, item := range v.picks {
		if len(lines) >= room {
			break
		}
		label := item.label
		if i == v.pick {
			label = "▸ " + label
		} else {
			label = "  " + label
		}
		style := lipgloss.NewStyle().Foreground(muted).MaxWidth(width)
		if i == v.pick {
			label = "▸ " + item.label
			style = lipgloss.NewStyle().Foreground(green).Background(pickBg).Bold(true).Width(width).MaxWidth(width)
		}
		lines = append(lines, style.Render(label))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func receiptLine(elapsed time.Duration, called bool) string {
	ms := elapsed.Milliseconds()
	if ms < 1 {
		ms = 1
	}
	span := fmt.Sprintf("%d ms", ms)
	if ms >= 1000 {
		tenths := (ms + 50) / 100
		span = fmt.Sprintf("%d,%d sn", tenths/10, tenths%10)
	}
	if called {
		return span + " · liste 3 TL · ödenecek 0,00 TL"
	}
	return span + " · ödenecek 0,00 TL"
}

func planText(run projectRun, sent bool) string {
	plan := "dosya gövdesi yok. Model çağrılmadı."
	if run.OrderBody != "" && !sent {
		plan = "dosya gövdesi yok. İstem yerelde yapıldı."
	}
	if sent {
		plan = "dosya gövdesi yok. Plan işçiye gidiyor."
	}
	if run.Note != "" {
		plan = run.Note + " " + plan
	}
	if run.Steps != "" {
		return run.Steps + "\n" + plan
	}
	return plan
}

func taskFlow(run projectRun, sent bool) []event {
	base := filepathBase(run.Root)
	files := "dizinde dosya yok"
	if len(run.Files) > 0 {
		shown := run.Files
		if len(shown) > 6 {
			shown = append(append([]string{}, shown[:6]...), "ve daha")
		}
		files = strings.Join(shown, " · ")
	}
	lint := run.Lint
	if lint == "" {
		lint = "kontrol yok"
	}
	lint += " · " + receiptLine(run.Elapsed, false)
	funcs := run.Funcs
	if len(funcs) == 0 {
		funcs = []string{"eşleşen işlev yok"}
	}
	if len(funcs) > 4 {
		funcs = funcs[:4]
	}
	events := []event{
		{title: "görev", body: redactLine(run.Task), kind: "user"},
		{title: "düşünce", meta: "yerel", body: thoughtText(sent, run.Reading), kind: "yerel"},
		{title: "environment", meta: "yerel", body: "işletim yerel · çalışma " + base + " · Europe/Istanbul", kind: "yerel"},
		{title: "workspace", meta: "yerel", body: files, kind: "yerel"},
		{title: "plan", meta: "yerel", body: planText(run, sent), kind: "yerel"},
		{title: "işlev", meta: "yerel", body: strings.Join(funcs, "\n"), kind: "yerel"},
		{title: "php -l", meta: "yerel", body: lint, kind: "yerel"},
	}
	if strings.TrimSpace(run.OrderBody) != "" {
		title := run.OrderTitle
		if title == "" {
			title = "istem"
		}
		kind := "yerel"
		if title == "soru" {
			kind = "soru"
		}
		events = append(events, event{title: title, meta: "yerel", body: run.OrderBody, kind: kind})
	}
	if strings.TrimSpace(run.RunLine) != "" {
		events = append(events, event{title: "çalıştır", meta: "yerel", body: run.RunLine, kind: "yerel"})
	}
	return events
}

func thoughtText(sent bool, reading string) string {
	text := "Emir alındı."
	if reading != "" {
		text += " " + reading
	}
	if sent {
		return text + " Model işi yapıyor."
	}
	return text + " Adımlar yazılıyor."
}

func filepathBase(path string) string {
	base := path
	if i := strings.LastIndexAny(path, `/\`); i >= 0 && i+1 < len(path) {
		base = path[i+1:]
	}
	if base == "" {
		return path
	}
	return base
}

func band(left, right string, width int, bg lipgloss.Color) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return lipgloss.NewStyle().Background(bg).MaxWidth(width).Render(left)
	}
	pad := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", gap))
	return left + pad + right
}

func spread(left, right string, width int) string {
	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 1 {
		return lipgloss.NewStyle().MaxWidth(width).Render(left)
	}
	return left + strings.Repeat(" ", space) + right
}

func fillScreen(content string, width, height int) string {
	bg := lipgloss.NewStyle().Background(ink).Foreground(paper).Width(width).MaxWidth(width)
	clip := lipgloss.NewStyle().MaxWidth(width)
	lines := strings.Split(content, "\n")
	flat := make([]string, 0, len(lines))
	for _, item := range lines {
		for _, piece := range strings.Split(clip.Render(item), "\n") {
			flat = append(flat, piece)
		}
	}
	if len(flat) > height {
		flat = flat[:height]
	}
	for len(flat) < height {
		flat = append(flat, "")
	}
	for i, item := range flat {
		flat[i] = bg.Render(item)
	}
	return strings.Join(flat, "\n")
}
