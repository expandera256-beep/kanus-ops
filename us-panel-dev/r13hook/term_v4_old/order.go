package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type orderAction struct {
	kind    string
	subject string
}

type intent struct {
	kind     string
	subject  string
	fresh    bool
	drive    string
	question string
	reading  string
}

func classifyOrder(task string) orderAction {
	in := readIntent(task)
	switch in.kind {
	case "plan", "kur", "kdv", "ekle", "calistir":
		return orderAction{kind: in.kind, subject: in.subject}
	default:
		return orderAction{}
	}
}

func readIntent(task string) intent {
	fresh := wantsFreshFolder(task)
	drive := driveLetter(task)
	imperative := hasPlan(task) || hasKur(task) || hasBareKur(task) || hasEkle(task) || wantsRun(task)
	if looksLikeQuestion(task) && !imperative {
		return intent{kind: "soru", subject: cleanSubject(task), reading: "Soru okundu."}
	}
	if hasPlan(task) {
		return intent{kind: "plan", subject: cleanSubject(task), fresh: fresh, drive: drive, reading: "İstek okundu: plan çizilecek."}
	}
	if strings.Contains(strings.ToLower(task), "kdv") {
		return intent{kind: "kdv", subject: "kdv", reading: "İstek okundu: kdv yerelde yazılacak."}
	}
	if hasKur(task) || hasBareKur(task) {
		subject := subjectOfKur(task)
		place := "seçili klasör"
		if drive != "" {
			place = drive + " diski"
		}
		return intent{kind: "kur", subject: subject, fresh: fresh || drive != "", drive: drive, reading: "İstek okundu: " + subject + " kurulacak · " + place + "."}
	}
	if subject, ok := subjectOfAdd(task); ok {
		return intent{kind: "ekle", subject: subject, drive: drive, reading: "İstek okundu: " + subject + " eklenecek."}
	}
	if wantsRun(task) {
		return intent{kind: "calistir", subject: launchExtra(task), reading: "İstek okundu: program çalışacak."}
	}
	if wantsApp(task) {
		return intent{kind: "calistir", subject: "", reading: "Dosyadaki uygulama çalıştırılacak."}
	}
	if looksLikeQuestion(task) {
		return intent{kind: "soru", subject: cleanSubject(task), reading: "Soru okundu."}
	}
	if strings.TrimSpace(task) != "" {
		return intent{kind: "komut", subject: cleanSubject(task), reading: "Komut alındı."}
	}
	return intent{}
}

func wantsApp(task string) bool {
	folded := foldText(task)
	if strings.Contains(folded, "gerceklestir") || strings.Contains(folded, "dosyada yatan") {
		return true
	}
	if strings.Contains(folded, "dosyadaki") && (strings.Contains(folded, "uygulama") || strings.Contains(folded, "program")) {
		return true
	}
	for _, word := range splitWords(task) {
		folded = foldText(word)
		if strings.HasPrefix(folded, "uygula") && !strings.HasPrefix(folded, "uygulama") {
			return true
		}
	}
	return false
}

func subjectOfKur(task string) string {
	if at := findBounded(task, " yapan "); at >= 0 {
		return stripPlace(string([]rune(task)[:at]))
	}
	if hasKur(task) {
		return subjectOfCreate(task)
	}
	return subjectOfBare(task)
}

func hasBareKur(task string) bool {
	for _, word := range splitWords(task) {
		switch foldText(word) {
		case "kur", "kurulsun", "olustur", "olusturun", "olusturalim":
			return true
		}
	}
	return false
}

func subjectOfBare(task string) string {
	letter := strings.ToLower(driveLetter(task))
	var kept []string
	for _, word := range splitWords(task) {
		folded := foldText(word)
		trimmed := foldText(strings.Trim(word, `:\/`))
		if letter != "" && trimmed == letter {
			continue
		}
		if stopWord(folded) || bareVerb(folded) {
			continue
		}
		kept = append(kept, word)
	}
	if len(kept) == 0 {
		return "uygulama"
	}
	return cleanSubject(strings.Join(kept, " "))
}

func stopWord(folded string) bool {
	switch folded {
	case "yeni", "bir", "klasor", "klasorde", "klasore", "klasorune", "klasorunde", "dizin", "dizinde", "dizine",
		"disk", "diskinde", "diskine", "diske", "diskindeki",
		"surucu", "surucude", "surucusunde", "surucuye", "surucunde",
		"icinde", "icine", "uzerinde", "lutfen", "bana", "benim", "icin",
		"proje", "projesi", "projesini", "projeyi", "projeye", "yapan",
		"ve", "ile", "bu", "su":
		return true
	}
	return false
}

func bareVerb(folded string) bool {
	switch folded {
	case "kur", "kurulsun", "olustur", "olusturun", "olusturalim", "ekle", "calistir", "baslat", "run", "start", "yaz", "ciz", "planla":
		return true
	}
	return strings.HasPrefix(folded, "ekley") || strings.HasPrefix(folded, "planlay")
}

func looksLikeQuestion(task string) bool {
	if strings.HasSuffix(strings.TrimSpace(task), "?") {
		return true
	}
	for _, word := range splitWords(task) {
		switch foldText(word) {
		case "ne", "neden", "nasil", "nicin", "kim", "hangi", "kac", "nerede", "nereden", "nedir", "midir", "mi", "miyim", "misin":
			return true
		}
	}
	return false
}

func driveLetter(task string) string {
	words := splitWords(task)
	for i, word := range words {
		trimmed := strings.Trim(foldText(word), `:\/`)
		if len([]rune(trimmed)) != 1 || trimmed[0] < 'a' || trimmed[0] > 'z' {
			continue
		}
		if strings.Contains(word, ":") {
			return strings.ToUpper(trimmed)
		}
		if i+1 < len(words) {
			next := foldText(words[i+1])
			if strings.HasPrefix(next, "disk") || strings.HasPrefix(next, "suruc") {
				return strings.ToUpper(trimmed)
			}
		}
	}
	return ""
}

var driveLookup = statDrive

func collectDrives(hinted []string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(letter string, fromMask bool) {
		letter = strings.ToUpper(strings.Trim(letter, `:\/`))
		if len(letter) != 1 || letter[0] < 'A' || letter[0] > 'Z' || seen[letter] {
			return
		}
		path, ok := statDrive(letter)
		if !ok {
			if !fromMask {
				return
			}
			path = letter + `:\`
		}
		seen[letter] = true
		out = append(out, path)
	}
	for _, path := range hinted {
		if path != "" {
			add(path[:1], true)
		}
	}
	for letter := 'A'; letter <= 'Z'; letter++ {
		add(string(letter), false)
	}
	return out
}

func statDrive(letter string) (string, bool) {
	letter = strings.ToUpper(strings.Trim(letter, `:\/`))
	if len(letter) != 1 {
		return "", false
	}
	for _, path := range []string{letter + `:\`, letter + ":/"} {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, true
		}
	}
	return "", false
}

func launchExtra(task string) string {
	runes := []rune(task)
	for i := range runes {
		word := wordAt(runes, i)
		folded := foldText(word)
		if folded == "calistir" || folded == "baslat" || folded == "run" || folded == "start" {
			return strings.TrimSpace(string(runes[i+len([]rune(word)):]))
		}
	}
	return ""
}

func hasPlan(task string) bool {
	if hasPhrase(task, "plan çiz", "plan ciz", "plan yaz", "plan çıkar", "plan cikar", "için plan", "icin plan") {
		return true
	}
	for _, word := range splitWords(task) {
		folded := foldText(word)
		if folded == "planla" || strings.HasPrefix(folded, "planlay") {
			return true
		}
	}
	return false
}

func hasKur(task string) bool {
	return hasPhrase(task,
		"yapan proje",
		"projesini kur",
		"projesi kur",
		"projeyi kur",
		"proje kur",
		"projesini oluştur",
		"projesini olustur",
		"projesi oluştur",
		"projesi olustur",
		"projeyi oluştur",
		"projeyi olustur",
		"proje oluştur",
		"proje olustur",
	)
}

func subjectOfCreate(task string) string {
	if at := findBounded(task, " yapan "); at >= 0 {
		return stripPlace(string([]rune(task)[:at]))
	}
	for _, mark := range []string{"projesini kur", "projesi kur", "projeyi kur", "proje kur", "projesini oluştur", "projesini olustur", "projesi oluştur", "projesi olustur", "projeyi oluştur", "projeyi olustur", "proje oluştur", "proje olustur"} {
		at := findBounded(task, mark)
		if at < 0 {
			continue
		}
		runes := []rune(task)
		before := cleanSubject(string(runes[:at]))
		after := cleanSubject(string(runes[at+len([]rune(mark)):]))
		if before != "uygulama" {
			return stripPlace(before)
		}
		if after != "uygulama" {
			return stripPlace(after)
		}
		return "proje"
	}
	return "uygulama"
}

func stripPlace(subject string) string {
	subject = cleanSubject(subject)
	for {
		cut := false
		for _, prefix := range []string{"yeni bir klasörde ", "yeni bir klasorde ", "yeni klasörde ", "yeni klasorde ", "yeni klasör ", "yeni klasor ", "yeni dizin ", "bir klasörde ", "bir klasorde "} {
			if findBounded(subject, prefix) != 0 {
				continue
			}
			runes := []rune(subject)
			n := len([]rune(prefix))
			if n > len(runes) {
				continue
			}
			subject = cleanSubject(string(runes[n:]))
			cut = true
			break
		}
		if !cut {
			return subject
		}
	}
}

func wantsFreshFolder(task string) bool {
	return hasPhrase(task, "yeni klasörde", "yeni klasorde", "yeni klasör", "yeni klasor", "yeni dizin", "yeni bir klasörde", "yeni bir klasorde", "yeni bir klasör", "yeni bir klasor")
}

func isRootPath(root string) bool {
	clean := filepath.Clean(root)
	if clean == string(filepath.Separator) {
		return true
	}
	return volumeRoot(clean, filepath.VolumeName(clean))
}

func volumeRoot(clean, vol string) bool {
	if vol == "" {
		return false
	}
	rest := strings.TrimPrefix(clean, vol)
	rest = strings.Trim(rest, `/\`)
	return rest == ""
}

func systemDirName(name string) bool {
	switch strings.ToLower(name) {
	case "$recycle.bin", "system volume information", "$winreagent", "config.msi", "$sysreset", "$windows.~bt", "$windows.~ws", "windows", "program files", "program files (x86)", "perflogs", "recovery":
		return true
	}
	return false
}

func subjectOfAdd(task string) (string, bool) {
	if !hasEkle(task) {
		return "", false
	}
	runes := []rune(task)
	if at := findBounded(task, "projeye "); at >= 0 {
		rest := runes[at+len([]rune("projeye ")):]
		end := len(rest)
		for i := range rest {
			if foldText(string(rest[i:])) == "" {
				continue
			}
			word := wordAt(rest, i)
			folded := foldText(word)
			if folded == "ekle" || strings.HasPrefix(folded, "ekley") {
				end = i
				break
			}
		}
		return cleanSubject(string(rest[:end])), true
	}
	end := len(runes)
	for i := range runes {
		word := wordAt(runes, i)
		folded := foldText(word)
		if folded == "ekle" || strings.HasPrefix(folded, "ekley") {
			end = i
			break
		}
	}
	return cleanSubject(string(runes[:end])), true
}

func hasEkle(task string) bool {
	for _, word := range splitWords(task) {
		folded := foldText(word)
		if folded == "ekle" || strings.HasPrefix(folded, "ekley") {
			return true
		}
	}
	return false
}

func createProject(root, subject string, fresh bool, source string) (string, string) {
	subject = cleanSubject(subject)
	slug := slugOf(subject)
	if dir, ok := ourProjectDir(root, subject); ok {
		return dir, subject + " zaten kurulu."
	}
	dir := root
	if fresh || !dirEmpty(root) {
		dir = filepath.Join(root, slug)
	}
	if !insideRoot(root, dir) && dir != root {
		return "", "proje yolu dışarı çıkıyor."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "proje klasörü açılmadı."
	}
	mainPath := filepath.Join(dir, "main.py")
	if _, err := os.Stat(mainPath); err == nil {
		return "", "main.py duruyor. Üstüne yazılmadı."
	}
	body := projectMain(subject)
	if strings.TrimSpace(source) != "" {
		body = source
	}
	if err := os.WriteFile(mainPath, []byte(body), 0o644); err != nil {
		return "", "main.py yazılamadı."
	}
	readme := "# " + subject + "\n\nİstem: " + subject + " yapan proje kuruldu.\n\nÇalıştırma: `python3 main.py`\n"
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644); err != nil {
		return "", "README.md yazılamadı."
	}
	where := "main.py · README.md"
	if dir != root {
		where = dir + " · " + where
	}
	return dir, subject + " kuruldu. davranış: " + behaviorLabel(behaviorOf(subject)) + ". " + where
}

func addFeature(root, subject string) (string, string) {
	subject = cleanSubject(subject)
	name := identOf(subject)
	if mainPath := findOurMain(root); mainPath != "" {
		body, err := os.ReadFile(mainPath)
		if err != nil {
			return "", "main.py okunamadı."
		}
		if strings.Contains(string(body), "def "+name+"(") {
			return "", subject + " zaten projede."
		}
		next, ok := insertFeature(string(body), name, subject, behaviorIn(string(body)))
		if !ok {
			return "", "parça main.py içine yazılamadı."
		}
		if err := os.WriteFile(mainPath, []byte(next), 0o644); err != nil {
			return "", "main.py yazılamadı."
		}
		appendReadme(filepath.Dir(mainPath), subject)
		return filepath.Dir(mainPath), subject + " eklendi. main.py"
	}
	path := filepath.Join(root, name+".py")
	if !insideRoot(root, path) {
		return "", "parça yolu dışarı çıkıyor."
	}
	if _, err := os.Stat(path); err == nil {
		return "", subject + " zaten projede."
	}
	file := "# kıyas ek: " + subject + "\nBASLIK = " + pyString(subject) + "\n\ndef " + name + "(girdi=\"\"):\n    return girdi or BASLIK\n\nif __name__ == \"__main__\":\n    print(" + name + "())\n"
	if err := os.WriteFile(path, []byte(file), 0o644); err != nil {
		return "", name + ".py yazılamadı."
	}
	return "", subject + " ayrı dosya yazıldı, programa bağlanmadı. " + name + ".py"
}

func writePlan(root, task string, run projectRun) string {
	var b strings.Builder
	b.WriteString("# Plan\n\n")
	b.WriteString("İstem: " + cleanSubject(task) + "\n\n")
	b.WriteString("## Durum\n")
	count := 0
	for _, file := range run.Files {
		if file == "dizinde dosya yok" {
			continue
		}
		if count == 6 {
			b.WriteString("- ve daha\n")
			break
		}
		b.WriteString("- " + file + "\n")
		count++
	}
	if count == 0 {
		b.WriteString("- Dizinde dosya yok. İlk adım: proje kur.\n")
	}
	if len(run.Funcs) > 0 {
		b.WriteString("\n## İşlevler\n")
		n := 0
		for _, fn := range run.Funcs {
			if fn == "eşleşen işlev yok" || strings.TrimSpace(fn) == "" {
				continue
			}
			if n == 4 {
				break
			}
			b.WriteString("- " + fn + "\n")
			n++
		}
	}
	b.WriteString("\n## Adımlar\n")
	for i, step := range planSteps(root, run) {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}
	path := filepath.Join(root, "PLAN.md")
	if !insideRoot(root, path) && path != filepath.Join(root, "PLAN.md") {
		return "plan yolu dışarı çıkıyor."
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "PLAN.md yazılamadı."
	}
	return "PLAN.md yazıldı. " + strings.TrimSpace(strings.Split(b.String(), "\n")[2])
}

func ourProjectDir(root, subject string) (string, bool) {
	want := "# kıyas: " + subject
	for _, dir := range []string{root, filepath.Join(root, slugOf(subject))} {
		if dir != root && !insideRoot(root, dir) {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, "main.py"))
		if err != nil {
			continue
		}
		first, _, _ := strings.Cut(string(body), "\n")
		if first == want {
			return dir, true
		}
	}
	return "", false
}

func constructionSteps(in intent) string {
	if in.kind == "" && in.reading == "" {
		return ""
	}
	line := "Okuma: " + in.kind
	if line == "Okuma: " {
		line = "Okuma: cümle"
	}
	if in.subject != "" && !needsJob(in.subject) {
		line += " · " + in.subject
	}
	if in.fresh {
		line += " · yeni klasör"
	}
	kurgu := "Kurgu: sıradaki adım yazılıyor."
	switch {
	case in.kind == "kur" && knownJob(in.subject):
		kurgu = "Kurgu: " + behaviorLabel(behaviorOf(in.subject)) + " yerelde yazılıyor."
	case in.kind == "kur" && !needsJob(in.subject):
		kurgu = "Kurgu: " + in.subject + " için main.py yazılıyor."
	case in.kind == "kur":
		kurgu = "Kurgu: iş sorulacak."
	case in.kind == "plan":
		kurgu = "Kurgu: PLAN.md yazılıyor."
	case in.kind == "ekle":
		kurgu = "Kurgu: parça ekleniyor."
	case in.kind == "calistir":
		kurgu = "Kurgu: program çalıştırılacak."
	case in.kind == "komut":
		kurgu = "Kurgu: komut uygulanacak."
	case in.kind == "kdv":
		kurgu = "Kurgu: kdv yerelde yazılacak."
	}
	return line + "\n" + kurgu
}

func knownJob(subject string) bool {
	switch behaviorOf(subject) {
	case "hesap", "not":
		return true
	default:
		return false
	}
}

func needsJob(subject string) bool {
	switch foldText(subject) {
	case "", "uygulama", "proje":
		return true
	default:
		return false
	}
}

func behaviorOf(subject string) string {
	folded := foldText(subject)
	switch {
	case folded == "hesap" || strings.Contains(folded, "hesap makin"):
		return "hesap"
	case folded == "not" || strings.Contains(folded, "not tut"):
		return "not"
	default:
		return "baslik"
	}
}

func behaviorLabel(behavior string) string {
	switch behavior {
	case "hesap", "not":
		return behavior
	default:
		return "başlık"
	}
}

func behaviorIn(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# davranış: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# davranış: "))
		}
	}
	return "baslik"
}

func projectMain(subject string) string {
	behavior := behaviorOf(subject)
	var body string
	switch behavior {
	case "hesap":
		body = hesapBody()
	case "not":
		body = notBody()
	default:
		body = "def calistir(girdi=\"\"):\n    return girdi or BASLIK\n"
	}
	return "# kıyas: " + subject + "\n# davranış: " + behavior + "\nBASLIK = " + pyString(subject) + "\n\n" + body + "\nKAYIT = [\"calistir\"]\n\nif __name__ == \"__main__\":\n    import sys\n    girdi = \" \".join(sys.argv[1:])\n    g = globals()\n    for ad in KAYIT:\n        print(g[ad](girdi))\n"
}

func hesapBody() string {
	return strings.Join([]string{
		"def calistir(girdi=\"\"):",
		"    ifade = (girdi or \"2 + 40\").strip()",
		"    try:",
		"        sonuc = hesapla(ifade)",
		"    except Exception:",
		"        return \"hesap: ifade okunamadı\"",
		"    return \"\".join(ifade.split()) + \" = \" + str(sonuc)",
		"",
		"def hesapla(ifade):",
		"    s = \"\".join(ifade.split())",
		"    i = 0",
		"    def peek():",
		"        return s[i] if i < len(s) else \"\"",
		"    def eat(ch):",
		"        nonlocal i",
		"        if peek() == ch:",
		"            i += 1",
		"            return True",
		"        return False",
		"    def factor():",
		"        nonlocal i",
		"        if eat(\"(\"):",
		"            value = expr()",
		"            if not eat(\")\"):",
		"                raise ValueError(\"parantez\")",
		"            return value",
		"        start = i",
		"        while i < len(s) and s[i].isdigit():",
		"            i += 1",
		"        if start == i:",
		"            raise ValueError(\"sayi\")",
		"        return int(s[start:i])",
		"    def term():",
		"        value = factor()",
		"        while True:",
		"            if eat(\"*\"):",
		"                value *= factor()",
		"            elif eat(\"/\"):",
		"                bolen = factor()",
		"                if bolen == 0:",
		"                    raise ValueError(\"sifir\")",
		"                neg = (value < 0) ^ (bolen < 0)",
		"                value = abs(value) // abs(bolen)",
		"                if neg:",
		"                    value = -value",
		"            else:",
		"                break",
		"        return value",
		"    def expr():",
		"        value = term()",
		"        while True:",
		"            if eat(\"+\"):",
		"                value += term()",
		"            elif eat(\"-\"):",
		"                value -= term()",
		"            else:",
		"                break",
		"        return value",
		"    value = expr()",
		"    if i != len(s):",
		"        raise ValueError(\"kalan\")",
		"    return value",
		"",
	}, "\n")
}

func notBody() string {
	return strings.Join([]string{
		"def calistir(girdi=\"\"):",
		"    from pathlib import Path",
		"    dosya = Path(__file__).with_name(\"notlar.txt\")",
		"    satir = girdi.strip()",
		"    if satir:",
		"        with dosya.open(\"a\", encoding=\"utf-8\") as kayit:",
		"            kayit.write(satir + \"\\n\")",
		"    if not dosya.exists():",
		"        return \"not tutan: kayıt yok\"",
		"    satirlar = [satir for satir in dosya.read_text(encoding=\"utf-8\").splitlines() if satir.strip()]",
		"    if not satirlar:",
		"        return \"not tutan: kayıt yok\"",
		"    return \"not tutan: \" + \" · \".join(satirlar[-3:])",
		"",
	}, "\n")
}

func insertFeature(src, name, subject, behavior string) (string, bool) {
	const marker = "KAYIT = ["
	at := strings.Index(src, marker)
	if at < 0 {
		return src, false
	}
	end := strings.Index(src[at:], "]")
	if end < 0 {
		return src, false
	}
	list := src[at+len(marker) : at+end]
	fn := featureBlock(name, subject, behavior)
	updated := marker + strings.TrimRight(list, " \n") + ", \"" + name + "\""
	return src[:at] + fn + updated + src[at+end:], true
}

func featureBlock(name, subject, behavior string) string {
	if behavior == "not" && name == "arama" {
		return strings.Join([]string{
			"def arama(girdi=\"\"):",
			"    from pathlib import Path",
			"    dosya = Path(__file__).with_name(\"notlar.txt\")",
			"    sorgu = girdi.strip().casefold()",
			"    if not sorgu:",
			"        return \"arama: sorgu yok\"",
			"    if not dosya.exists():",
			"        return \"arama: yok\"",
			"    bulunan = [satir for satir in dosya.read_text(encoding=\"utf-8\").splitlines() if sorgu in satir.casefold()]",
			"    if not bulunan:",
			"        return \"arama: yok\"",
			"    return \"arama\\n\" + \"\\n\".join(bulunan[:3])",
			"",
			"",
		}, "\n")
	}
	return fmt.Sprintf("def %s(girdi=\"\"):\n    return %s\n\n", name, pyString(subject))
}

func planSteps(root string, run projectRun) []string {
	steps := []string{"Dosya gövdesi yerelde kalır."}
	runnable := false
	for _, file := range run.Files {
		fields := strings.Fields(file)
		if len(fields) == 0 || file == "dizinde dosya yok" {
			continue
		}
		base := strings.ToLower(filepath.Base(fields[0]))
		switch base {
		case "main.py", "app.py", "run.py", "manage.py", "dockerfile":
			runnable = true
		}
	}
	behavior, hasArama := projectBehavior(root)
	if behavior == "not" && !hasArama {
		steps = append(steps, "Eksik parça: projeye arama ekle.")
	}
	if !runnable {
		steps = append(steps, "İlk adım: … yapan proje kur.")
	}
	if runnable {
		steps = append(steps, "Doğrulamak için: çalıştır.")
	}
	return steps
}

func projectBehavior(root string) (string, bool) {
	behavior := ""
	hasArama := false
	walkProject(root, func(path string, entry os.DirEntry) {
		if entry.IsDir() || entry.Name() != "main.py" {
			return
		}
		body, err := os.ReadFile(path)
		if err != nil || !strings.HasPrefix(string(body), "# kıyas: ") {
			return
		}
		if behavior == "" {
			behavior = behaviorIn(string(body))
		}
		if strings.Contains(string(body), "def arama(") {
			hasArama = true
		}
	})
	return behavior, hasArama
}

func findOurMain(root string) string {
	var found string
	walkProject(root, func(path string, entry os.DirEntry) {
		if found != "" || entry.IsDir() || entry.Name() != "main.py" {
			return
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return
		}
		if strings.HasPrefix(string(body), "# kıyas: ") && strings.Contains(string(body), "KAYIT = [") {
			found = path
		}
	})
	return found
}

func appendReadme(dir, subject string) {
	path := filepath.Join(dir, "README.md")
	body, err := os.ReadFile(path)
	line := "\n- " + subject + "\n"
	if err != nil {
		_ = os.WriteFile(path, []byte("# Proje\n"+line), 0o644)
		return
	}
	if strings.Contains(string(body), line) || strings.Contains(string(body), "- "+subject) {
		return
	}
	_ = os.WriteFile(path, append(body, []byte(line)...), 0o644)
}

func dirEmpty(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		return false
	}
	return true
}

func cleanSubject(text string) string {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, " .,:;-")
	text = redactLine(text)
	fields := strings.Fields(text)
	text = strings.Join(fields, " ")
	runes := []rune(text)
	if len(runes) > 80 {
		text = strings.TrimSpace(string(runes[:80]))
	}
	if text == "" || text == "uygulama" && strings.TrimSpace(text) == "" {
		return "uygulama"
	}
	if text == "" {
		return "uygulama"
	}
	return text
}

func slugOf(subject string) string {
	name := identOf(subject)
	return strings.ReplaceAll(name, "_", "-")
}

func identOf(subject string) string {
	var b strings.Builder
	prev := false
	for _, r := range foldText(subject) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			prev = false
			continue
		}
		if !prev {
			b.WriteByte('_')
			prev = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		out = "uygulama"
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "is_" + out
	}
	if len(out) > 40 {
		out = strings.Trim(out[:40], "_")
	}
	if out == "" {
		return "uygulama"
	}
	return out
}

func pyString(text string) string {
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, `"`, `\"`)
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	return `"` + text + `"`
}

func hasPhrase(text string, phrases ...string) bool {
	for _, phrase := range phrases {
		if findBounded(text, phrase) >= 0 {
			return true
		}
	}
	return false
}

func findBounded(text, mark string) int {
	hay := []rune(foldText(text))
	needle := []rune(foldText(mark))
	if len(needle) == 0 || len(hay) < len(needle) {
		return -1
	}
	for i := 0; i+len(needle) <= len(hay); i++ {
		match := true
		for j := range needle {
			if hay[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match && phraseBounded(hay, i, i+len(needle)) {
			return i
		}
	}
	return -1
}

func phraseBounded(hay []rune, start, end int) bool {
	if start > 0 && wordRune(hay[start-1]) && wordRune(hay[start]) {
		return false
	}
	if end < len(hay) && wordRune(hay[end-1]) && wordRune(hay[end]) {
		return false
	}
	return true
}

func wordRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_'
}

func foldText(text string) string {
	return string(foldRunes(text))
}

func foldRunes(text string) []rune {
	in := []rune(text)
	out := make([]rune, len(in))
	for i, r := range in {
		out[i] = foldRune(r)
	}
	return out
}

func foldRune(r rune) rune {
	switch r {
	case 'Ç', 'ç':
		return 'c'
	case 'Ğ', 'ğ':
		return 'g'
	case 'I', 'İ', 'ı', 'i':
		return 'i'
	case 'Ö', 'ö':
		return 'o'
	case 'Ş', 'ş':
		return 's'
	case 'Ü', 'ü':
		return 'u'
	}
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

func splitWords(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '.' || r == ',' || r == ':' || r == ';'
	})
}

func wordAt(runes []rune, i int) string {
	if i < 0 || i >= len(runes) {
		return ""
	}
	j := i
	for j < len(runes) {
		r := runes[j]
		if r == ' ' || r == '\t' || r == '\n' || r == '.' || r == ',' || r == ':' || r == ';' {
			break
		}
		j++
	}
	return string(runes[i:j])
}
