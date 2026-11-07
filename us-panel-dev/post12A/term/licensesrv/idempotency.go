package licensesrv

// POST12-A retry idempotency (çift tahsilat önleme). Demir onayı: 08.11.2026
// 01:11 TSİ. Protokol: POST12_A_PROTOCOL_PROPOSAL_TR.md.
//
// Bir mantıksal istek X-US-Request-ID ile tanımlanır. Kayıt kimliği global
// tutulur ve ilk istekteki asıl (principal) + gövde hash'ine bağlanır; aynı
// kimlik farklı asıl/gövde ile gelirse 409 idempotency_mismatch. Kayıtlar store
// JSON'unda (atomik yazım) kalıcıdır, yeniden başlatmadan sağ çıkar.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	idemPending   = "pending"
	idemCommitted = "committed"
	idemReleased  = "released"

	idemKindChat    = "chat"
	idemKindReserve = "reserve"
)

// IdempotencyParkTTL: başarısız bir denemeden sonra hold'un aynı kimlikli
// yeniden deneme için tutulduğu süre. EXE backoff'u (100/300 ms) ve yeniden
// bağlanma süresinden çok uzun, ama bakiyeyi uzun süre kilitlemeyecek kadar kısa.
const IdempotencyParkTTL = 2 * time.Minute

// IdempotencyRetention: terminal (committed/released) kayıtların tombstone
// olarak saklandığı süre. Bu süre içinde aynı kimlik yeni ücretli işe dönüşmez.
const IdempotencyRetention = 24 * time.Hour

var (
	errIdemMismatch   = errors.New("idempotency_mismatch")
	errIdemTerminal   = errors.New("request_terminal")
	errIdemInProgress = errors.New("request_in_progress")
	errIdemReleased   = errors.New("reservation_released")
)

// headerRequestID: kanonik X-US-Request-ID; Idempotency-Key alias (API/SDK
// uyumu). İkisi farklıysa ok=false → 400 idempotency_header_mismatch.
func headerRequestID(r *http.Request) (string, bool) {
	a := strings.TrimSpace(r.Header.Get("X-US-Request-ID"))
	b := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if a != "" && b != "" && a != b {
		return "", false
	}
	if (r.Header.Get("X-US-Request-ID") != "" && a == "") || (r.Header.Get("Idempotency-Key") != "" && b == "") {
		return "", false
	}
	if a == "" {
		a = b
	}
	return a, true
}

// IdempotencyRecord persist edilir. Principal ham lisans anahtarı ya da oturum
// token'ı DEĞİLDİR: "lic:"+sha256(anahtar) veya "acct:"+hesap kimliği.
type IdempotencyRecord struct {
	Kind          string    `json:"kind"`
	Principal     string    `json:"principal"`
	BodyHash      string    `json:"bodyHash"`
	ReservationID string    `json:"reservationId,omitempty"`
	State         string    `json:"state"`
	Attempts      int       `json:"attempts"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// validRequestID: 16-64 karakter [A-Za-z0-9_-]. EXE 32 hex üretir.
func validRequestID(id string) bool {
	if len(id) < 16 || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func hashBytes(kind string, b []byte) string {
	h := sha256.New()
	h.Write([]byte("kanus-idem-v1\n" + kind + "\n"))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// canonicalBodyHash: JSON yeniden kodlanır (Go map anahtarları sıralar), yani
// boşluk/alan sırası farkı aynı mantıksal isteği değiştirmez.
func canonicalBodyHash(kind string, raw []byte) string {
	var v any
	if json.Unmarshal(raw, &v) == nil {
		if c, err := json.Marshal(v); err == nil {
			return hashBytes(kind, c)
		}
	}
	return hashBytes(kind, raw)
}

// principalLocked: oturum yenilemesinden etkilenmeyen kimlik.
func (s *Store) principalLocked(token string, now time.Time) string {
	if strings.HasPrefix(token, "usa1.") {
		if a := s.accountForTokenLocked(token, now); a != nil {
			return "acct:" + a.ID
		}
		return ""
	}
	if _, ok := s.st.Licenses[token]; ok {
		return "lic:" + digest(token)
	}
	return ""
}

// recordStateLocked: kaydın durumunu ledger'dan türetir (revoke/expire gibi
// yollar kayda dokunmadan hold'u kapatabilir).
func (s *Store) recordStateLocked(rec *IdempotencyRecord) string {
	if rec.State != idemPending || rec.ReservationID == "" {
		return rec.State
	}
	if _, ok := s.st.Reservations[rec.ReservationID]; ok {
		return idemPending
	}
	if done, ok := s.st.Settlements[rec.ReservationID]; ok {
		if done.Commit {
			return idemCommitted
		}
		return idemReleased
	}
	return idemReleased
}

// idemCheckLocked: salt okuma ön denetimi. Uyuşmazlık, terminal ya da uçuşta
// olan kayıt için hata döner; GPU wake/probe'dan ÖNCE çağrılır.
func (s *Store) idemCheckLocked(kind, id, principal, bodyHash string, now time.Time) (*IdempotencyRecord, error) {
	rec := s.st.Idempotency[id]
	if rec == nil {
		return nil, nil
	}
	if rec.Kind != kind || rec.Principal != principal || rec.BodyHash != bodyHash {
		return nil, errIdemMismatch
	}
	switch s.recordStateLocked(rec) {
	case idemCommitted:
		return nil, errIdemTerminal
	case idemReleased:
		// API uyumu (d): serbest bırakılmış (TTL/park/iptal) rezervasyonun
		// tekrarı yeni kredi kilitlemez.
		return nil, errIdemReleased
	}
	if s.inflight[id] {
		return nil, errIdemInProgress
	}
	return rec, nil
}

// IdemPrecheck chat proxy için salt okuma ön denetimi.
func (s *Store) idemPrecheck(token, id string, raw []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.principalLocked(token, now)
	if p == "" {
		return ErrInvalid
	}
	_, err := s.idemCheckLocked(idemKindChat, id, p, canonicalBodyHash(idemKindChat, raw), now)
	return err
}

// reserveModelIdem: aynı kimlikte park edilmiş canlı hold varsa onu yeniden
// kullanır (ekstra hold yok), yoksa reserveModel ile yeni hold açıp kayda bağlar.
// Kayıt ve hold tek transaction içinde yazılır.
func (s *Store) reserveModelIdem(token, id string, raw []byte, budget time.Duration, now time.Time) (string, int, bool, error) {
	s.mu.Lock()
	p := s.principalLocked(token, now)
	if p == "" {
		s.mu.Unlock()
		return "", 0, false, ErrInvalid
	}
	bh := canonicalBodyHash(idemKindChat, raw)
	s.expireLocked(now)
	rec, err := s.idemCheckLocked(idemKindChat, id, p, bh, now)
	if err != nil {
		s.mu.Unlock()
		return "", 0, false, err
	}
	if rec != nil && s.recordStateLocked(rec) == idemPending {
		r := s.st.Reservations[rec.ReservationID]
		l := s.st.Licenses[r.Key]
		if l == nil || l.Revoked || !now.Before(l.ExpiresAt) {
			s.mu.Unlock()
			return "", 0, false, ErrInvalid
		}
		err := s.transactionLocked(func() error {
			r.Expires = now.Add(budget + 5*time.Minute)
			rec.Attempts++
			rec.UpdatedAt = now
			return nil
		})
		if err == nil {
			s.markInflightLocked(id)
		}
		s.mu.Unlock()
		return rec.ReservationID, r.Amount, l.Plan.Kind == "use", err
	}
	s.mu.Unlock()
	// Yeni hold. reserveModel kendi kilidini alır; ardından kaydı bağlarız.
	// Arada aynı kimlikle eşzamanlı ikinci bir istek gelirse bağlama aşamasında
	// yakalanır ve fazladan hold anında serbest bırakılır.
	resID, held, hourly, err := s.reserveModel(token, budget, now)
	if err != nil {
		return "", 0, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur := s.st.Idempotency[id]; cur != nil || s.inflight[id] {
		_ = s.releaseLocked(resID, now)
		return "", 0, false, errIdemInProgress
	}
	err = s.transactionLocked(func() error {
		s.st.Idempotency[id] = &IdempotencyRecord{Kind: idemKindChat, Principal: p, BodyHash: bh, ReservationID: resID, State: idemPending, Attempts: 1, CreatedAt: now, UpdatedAt: now}
		return nil
	})
	if err != nil {
		_ = s.releaseLocked(resID, now)
		return "", 0, false, err
	}
	s.markInflightLocked(id)
	return resID, held, hourly, nil
}

func (s *Store) markInflightLocked(id string) {
	if s.inflight == nil {
		s.inflight = map[string]bool{}
	}
	s.inflight[id] = true
}

// releaseLocked: yalnız henüz kayda bağlanmamış hold'u bırakır.
func (s *Store) releaseLocked(resID string, now time.Time) error {
	r := s.st.Reservations[resID]
	if r == nil {
		return nil
	}
	return s.transactionLocked(func() error {
		if l := s.st.Licenses[r.Key]; l != nil {
			l.Reserved -= r.Amount
		}
		s.st.Settlements[resID] = Settlement{CreatedAt: r.CreatedAt, Key: r.Key, SettledAt: now}
		delete(s.st.Reservations, resID)
		return nil
	})
}

// finishModelIdem: deneme bitti. committed=true ise settleAmount zaten
// yapılmıştır; kayıt terminal olur. Aksi halde hold kısa TTL ile park edilir
// (ücret yok) ve aynı kimlikli yeniden deneme onu paylaşır.
func (s *Store) finishModelIdem(id, resID string, committed bool, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inflight, id)
	rec := s.st.Idempotency[id]
	if rec == nil || rec.ReservationID != resID {
		return
	}
	_ = s.transactionLocked(func() error {
		rec.UpdatedAt = now
		if committed {
			rec.State = idemCommitted
			return nil
		}
		if r := s.st.Reservations[resID]; r != nil {
			if park := now.Add(IdempotencyParkTTL); park.Before(r.Expires) {
				r.Expires = park
			}
		}
		return nil
	})
}

// reserveIdem: /v1/credits/reserve için. Aynı kimlik + asıl + gövde → aynı
// rezervasyon kimliği (ACK kaybı). Commit edilmiş → 409 request_terminal; bırakılmış → 409 reservation_released.
func (s *Store) reserveIdem(key, id string, ttl time.Duration, now time.Time) (string, bool, error) {
	bh := hashBytes(idemKindReserve, []byte("amount=1"))
	s.mu.Lock()
	p := s.principalLocked(key, now)
	if p == "" {
		s.mu.Unlock()
		return "", false, ErrInvalid
	}
	s.expireLocked(now)
	rec, err := s.idemCheckLocked(idemKindReserve, id, p, bh, now)
	if err == nil && rec != nil {
		s.mu.Unlock()
		return rec.ReservationID, true, nil
	}
	s.mu.Unlock()
	if err != nil {
		return "", err != errIdemMismatch, err
	}
	resID, err := s.Reserve(key, 1, ttl, now)
	if err != nil {
		return "", false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur := s.st.Idempotency[id]; cur != nil {
		_ = s.releaseLocked(resID, now)
		if cur.Kind == idemKindReserve && cur.Principal == p && cur.BodyHash == bh && s.recordStateLocked(cur) == idemPending {
			return cur.ReservationID, true, nil
		}
		return "", false, errIdemInProgress
	}
	err = s.transactionLocked(func() error {
		s.st.Idempotency[id] = &IdempotencyRecord{Kind: idemKindReserve, Principal: p, BodyHash: bh, ReservationID: resID, State: idemPending, Attempts: 1, CreatedAt: now, UpdatedAt: now}
		return nil
	})
	if err != nil {
		_ = s.releaseLocked(resID, now)
		return "", false, err
	}
	return resID, false, nil
}

// IsIdempotentReplay: salt okuma. API'nin günlük kota sayacı (api-0006
// consumeAPIKeyQuotaLocked) bunu kimlik doğrulamada çağırır; true ise günlük
// kota DÜŞÜLMEZ (dakikalık rate sayılabilir). Durum değiştirmez.
func (s *Store) IsIdempotentReplay(token, requestID, path string, body []byte) bool {
	if !validRequestID(requestID) {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := s.st.Idempotency[requestID]
	if rec == nil || rec.Principal != s.principalLocked(token, time.Now()) {
		return false
	}
	if rec.Kind == idemKindReserve && path == "/v1/credits/reserve" {
		return true
	}
	return rec.Kind == idemKindChat && path != "/v1/credits/reserve" && rec.BodyHash == canonicalBodyHash(idemKindChat, body)
}

// SweepIdempotency: yetim temizliği. Başlangıçta ve periyodik çağrılır.
// Süresi dolmuş hold'lar (park edilmiş ya da çöken süreçten kalan) serbest
// bırakılır; committed asla release'e çevrilmez; retention'ı geçen terminal
// kayıtlar silinir. Dönen değer: değişen kayıt sayısı.
func (s *Store) SweepIdempotency(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := 0
	_ = s.transactionLocked(func() error {
		s.expireLocked(now)
		for id, rec := range s.st.Idempotency {
			st := s.recordStateLocked(rec)
			if st != rec.State {
				rec.State = st
				rec.UpdatedAt = now
				changed++
			}
			if st != idemPending && now.Sub(rec.UpdatedAt) > IdempotencyRetention && !s.inflight[id] {
				delete(s.st.Idempotency, id)
				changed++
			}
		}
		return nil
	})
	return changed
}

// writeIdemError: makine okunur hata gövdesi; kimlik/anahtar içermez.
func writeIdemError(w http.ResponseWriter, err error) bool {
	status, code := 0, ""
	switch err {
	case errIdemMismatch:
		status, code = 409, "idempotency_mismatch"
	case errIdemTerminal:
		status, code = 409, "request_terminal"
	case errIdemReleased:
		status, code = 409, "reservation_released"
	case errIdemInProgress:
		status, code = 409, "request_in_progress"
		w.Header().Set("Retry-After", "1")
	default:
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-US-Idempotency-Error", code)
	if err != errIdemMismatch && err != errIdemInProgress {
		w.Header().Set("X-US-Replayed", "true")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
	return true
}

func writeBadRequestID(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
