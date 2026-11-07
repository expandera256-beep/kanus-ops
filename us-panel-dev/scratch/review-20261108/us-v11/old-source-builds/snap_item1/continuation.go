package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const continuationLimit = 4
const continuationAnchorBytes = 96
const continuationOutputLimit = 128 << 10
const continuationSystem = "Kesilmiş bir metni sürdürüyorsun. İstenen köprüyü eksiksiz tekrarla ve hemen ardından yalnızca devamı yaz. Kod bloğu, açıklama veya yeni başlık ekleme."
const compactRepairSystem = "Tek bir dosyada küçük ve güvenli bir değişiklik yap. Yalnızca istenen hedef dosyayı DOSYA yol satırı ve eksiksiz gövdesiyle ver. Başka dosya üretme."

// A multi-file repair can exceed a single response. Retry against one fully
// read file and reject any reply that names a different or additional target.
func (s *session) retrySingleRepair(original string, allowed map[string]bool, send func(string, int) modelResult) modelResult {
	return s.retrySingleRepairWithLimit(original, allowed, send, promptWireBudget)
}

func (s *session) retrySingleRepairWithLimit(original string, allowed map[string]bool, send func(string, int) modelResult, promptLimit int) modelResult {
	var targets []string
	for name, complete := range allowed { if complete { targets = append(targets, name) } }
	if len(targets) == 0 { return modelResult{body: "Tam okunmuş dosya yok; değişiklik uygulanmadı."} }
	sort.Strings(targets)
	target := targets[0]
	wire := "Önceki yanıt çok uzundu. Yalnızca aşağıda tamamen okunmuş tek hedef dosyada en küçük gerekli değişikliği yap; dosyayı eksiksiz ver. Başka dosyaya dokunma.\nTEK DOSYA HEDEFİ:\n" + target + "\nORİJİNAL İSTEK VE OKUMALAR:\n" + original
	if len(wire)+len(compactRepairSystem) > promptLimit {
		return modelResult{body: "Tek dosyalık deneme bağlama sığmadı; dosyalar uygulanmadı."}
	}
	s.longActive = true
	if why := s.longContinuationError(); why != "" { return modelResult{body: why} }
	s.modelStage = "Tek dosya için yeniden deneniyor: " + target
	s.notice = s.modelStage
	s.noteLog("işlem", "Yanıt sınırı aşıldı; tam okunan tek dosya için yeniden deneniyor: " + target)
	result := send(wire, -4)
	if result.limited { result = s.continueLimitedResponse(wire, result, send) }
	if !result.ok { return result }
	count := 0
	for _, line := range strings.Split(expandFences(result.body), "\n") {
		trim := strings.TrimSpace(line)
		name, ok := strings.CutPrefix(trim, "DOSYA ")
		if !ok { name, ok = strings.CutPrefix(trim, "DOSYA:") }
		if !ok { continue }
		count++
		_, rel, valid := replyTarget(s.root, name)
		if !valid || rel != target || count > 1 {
			return modelResult{body: "Model tek hedef dosya sınırına uymadı; dosyalar uygulanmadı.", elapsed: result.elapsed}
		}
	}
	return result
}

// Limited output is never applied. It is joined only after the next complete
// or limited answer reproduces the exact trailing anchor. A mismatch, timeout,
// cancellation, secret or exhausted continuation limit leaves all files alone.
func (s *session) continueLimitedResponse(original string, first modelResult, send func(string, int) modelResult) modelResult {
	full := first.partial
	elapsed := first.elapsed
	if full == "" || !utf8.ValidString(full) || redactLine(full) != full {
		return modelResult{body: "Kesilen yanıt güvenle sürdürülemedi; dosyalar uygulanmadı.", elapsed: elapsed}
	}
	s.longActive = true
	previousStage := s.modelStage
	defer func() { s.modelStage = previousStage }()
	brief, _ := clipUTF8(original, 1200)
	for attempt := 1; attempt <= continuationLimit; attempt++ {
		if why := s.longContinuationError(); why != "" {
			return modelResult{body: why, elapsed: elapsed}
		}
		anchor := longTail(full, continuationAnchorBytes)
		history := longTail(full, 1800)
		wire := "Kesilen yanıtı tamamla. Son kesitin içeriğini yeniden yorumlama. Cevabının ilk karakterinden itibaren aşağıdaki KÖPRÜYÜ birebir kopyala; hemen ardından kesildiği yerden devam et. Başka açıklama veya işaret ekleme.\nİSTEK BAŞI:\n" + brief + "\nSON KESİT:\n" + history + "\nKÖPRÜ:\n" + anchor
		if len(wire)+len(continuationSystem) > promptWireBudget {
			return modelResult{body: "Devam isteği bağlama sığmadı; dosyalar uygulanmadı.", elapsed: elapsed}
		}
		s.modelStage = fmt.Sprintf("Yanıt sürdürülüyor · %d/%d", attempt, continuationLimit)
		s.notice = s.modelStage
		s.noteLog("işlem", fmt.Sprintf("Yanıt sınırına ulaşıldı; devam bölümü %d/%d isteniyor.", attempt, continuationLimit))
		next := send(wire, -3)
		elapsed += next.elapsed
		if why := s.longContinuationError(); why != "" {
			return modelResult{body: why, elapsed: elapsed}
		}
		if !next.ok && !next.limited {
			return modelResult{body: "Yanıt devam ettirilemedi: " + next.body + " Dosyalar uygulanmadı.", elapsed: elapsed}
		}
		piece := next.raw
		if next.limited { piece = next.partial }
		if piece == "" && next.ok { piece = next.body }
		if !utf8.ValidString(piece) || redactLine(piece) != piece || !strings.HasPrefix(piece, anchor) {
			return modelResult{body: "Model devamın başlangıcını doğrulayamadı; dosyalar uygulanmadı.", elapsed: elapsed}
		}
		suffix := strings.TrimPrefix(piece, anchor)
		if suffix == "" || strings.HasPrefix(suffix, anchor) || len(full)+len(suffix) > continuationOutputLimit {
			return modelResult{body: "Model geçerli bir devam üretmedi veya yanıt çok büyük; dosyalar uygulanmadı.", elapsed: elapsed}
		}
		full += suffix
		if next.ok {
			s.noteLog("işlem", "Yanıtın devamı doğrulandı; sonuç hazırlanıyor.")
			return modelResult{ok: true, body: strings.TrimSpace(full), raw: full, elapsed: elapsed}
		}
	}
	return modelResult{body: "Yanıt devam sınırına ulaştı; dosyalar uygulanmadı. Görevi daha küçük parçalara bölün.", elapsed: elapsed}
}
