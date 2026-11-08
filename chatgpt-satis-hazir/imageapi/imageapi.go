// Package imageapi implements the /v1/image HTTP contract (r2) as a
// standalone, compilable package. It does not import licensesrv and does
// not define Store, Mail, or Plan types.
//
// Production New() has no rate table, no ready mark, and no generator:
// it never opens a GPU/VM and never charges a real ledger. Numeric
// rates are not shipped; tests may install a fixture rate table that is
// not a product tariff.
package imageapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxBodyBytes    = 32 * 1024
	retention       = 24 * time.Hour
	idempotencyMin  = 16
	idempotencyMax  = 128
	defaultModel    = "klein"
	defaultSize     = "1024x1024"
	defaultN        = 1
	retentionPolicy = "authorized_24h"
	kleinRepo       = "black-forest-labs/FLUX.2-klein-4B"
	kleinRev        = "e7b7dc27f91deacad38e78976d1f2b499d76a294"
	schnellRepo     = "black-forest-labs/FLUX.1-schnell"
	schnellRev      = "741f7c3ce8b383c54771c7003378a50191e9efe9"
)

var allowedSizes = map[string][2]int{
	"1024x1024": {1024, 1024},
	"1344x768":  {1344, 768},
	"768x1344":  {768, 1344},
}

var unsupportedFields = map[string]struct{}{
	"steps": {}, "num_inference_steps": {}, "guidance": {}, "guidance_scale": {},
	"true_cfg_scale": {}, "dtype": {}, "width": {}, "height": {}, "batch": {},
	"batch_size": {}, "scheduler": {}, "image": {}, "init_image": {}, "mask": {},
	"strength": {}, "lora": {}, "adapter": {}, "checkpoint": {},
	"negative_prompt": {}, "response_format": {}, "quality": {}, "style": {},
	"amount": {}, "ttl": {}, "credits": {}, "price": {}, "user": {},
	"idempotency_key": {}, "fallback": {}, "allow_fallback": {},
}

// License is the caller-supplied entitlement view. It is not licensesrv.Store.
type License struct {
	ID     string
	Kind   string // "sub" or "use"
	Active bool
}

// Directory looks up a bearer token. Production wiring to licensesrv is out of
// this package (blocked on missing Store/Mail/Plan in the panel tree).
type Directory interface {
	Lookup(bearer string) (License, bool)
}

// MapDirectory is an in-memory Directory for tests and local harnesses.
type MapDirectory map[string]License

func (m MapDirectory) Lookup(bearer string) (License, bool) {
	lic, ok := m[bearer]
	return lic, ok
}

// RateKey is (model, size). An empty Rates map means the product tariff is PENDING.
type RateKey struct {
	Model, Size string
}

// EngineSnapshot is a read-only admission view. Reading it must not wake a GPU.
type EngineSnapshot struct {
	State        string // e.g. COLD, IMAGE_READY
	ModelStaged  map[string]bool
	SlotReady    bool
	DailyLeftSec int
	MaxUptimeOK  bool
}

// Generator produces encoded image bytes. The default is nil (no GPU).
type Generator func(model, prompt, size string, seed uint32) (contentType string, body []byte, err error)

// Server serves the image HTTP surface.
type Server struct {
	dir            Directory
	rates          map[RateKey]int // nil/empty => PENDING, no charge
	snap           EngineSnapshot
	gen            Generator
	acceptColdSwap bool
	bootIntent     func(jobID string) // optional hook; default records only
	now            func() time.Time
	hmacKey        []byte

	mu        sync.Mutex
	jobs      map[string]*job
	byIdem    map[string]*job // licenseID + "\x00" + key
	outputs   map[string]*output
	activeLic map[string]string // licenseID -> running jobID
}

type job struct {
	ID            string
	LicenseID     string
	IdemKey       string
	Fingerprint   string
	Status        string
	PartialReason string
	Model         string
	Size          string
	Prompt        string
	N             int
	Seed          *uint32
	ChosenSeed    uint32
	Created       time.Time
	BootIntent    bool
	ReservationID string // never serialized
	Reserved      int
	Committed     int
	Released      int
	Outputs       []outputMeta
	Terminal      bool
	Failed        *apiError
	ReplayStatus  int
}

type output struct {
	Meta      outputMeta
	Body      []byte
	LicenseID string
	Created   time.Time
}

type outputMeta struct {
	Index       int    `json:"index"`
	Ref         string `json:"ref"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Bytes       int    `json:"bytes"`
	SHA256      string `json:"sha256"`
	Seed        uint32 `json:"seed"`
}

type Options struct {
	Directory      Directory
	Rates          map[RateKey]int
	Snapshot       EngineSnapshot
	Generator      Generator
	AcceptColdSwap bool
	OnBootIntent   func(jobID string)
	Now            func() time.Time
}

// New returns a server. Defaults refuse generation: no tariff, no ready mark, no GPU.
func New(opt Options) *Server {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.Snapshot.ModelStaged == nil {
		opt.Snapshot.ModelStaged = map[string]bool{"klein": false, "schnell": false}
	}
	if opt.Snapshot.State == "" {
		opt.Snapshot.State = "COLD"
	}
	if opt.Snapshot.DailyLeftSec == 0 {
		opt.Snapshot.DailyLeftSec = 3600
	}
	if !opt.Snapshot.MaxUptimeOK && opt.Snapshot.State == "IMAGE_READY" {
		opt.Snapshot.MaxUptimeOK = true
	}
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return &Server{
		dir:            opt.Directory,
		rates:          opt.Rates,
		snap:           opt.Snapshot,
		gen:            opt.Generator,
		acceptColdSwap: opt.AcceptColdSwap,
		bootIntent:     opt.OnBootIntent,
		now:            opt.Now,
		hmacKey:        key,
		jobs:           map[string]*job{},
		byIdem:         map[string]*job{},
		outputs:        map[string]*output{},
		activeLic:      map[string]string{},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-US-Metered", "1")
	if s.snap.State != "" {
		w.Header().Set("X-US-Image-State", s.snap.State)
	}
	switch {
	case r.URL.Path == "/v1/image":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			s.writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "Yalnız POST kabul edilir", "", "", nil)
			return
		}
		s.postImage(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/image/jobs/"):
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			s.writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "Yalnız GET kabul edilir", "", "", nil)
			return
		}
		s.getJob(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/image/outputs/"):
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			s.writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "Yalnız GET kabul edilir", "", "", nil)
			return
		}
		s.getOutput(w, r)
	default:
		http.NotFound(w, r)
	}
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   any    `json:"field"`
	Reason  any    `json:"reason"`
	Retry   any    `json:"retry_after"`
}

func (s *Server) writeErr(w http.ResponseWriter, status int, code, msg, field, reason string, retry *int) {
	if retry != nil {
		w.Header().Set("Retry-After", strconv.Itoa(*retry))
	}
	w.Header().Set("Content-Type", "application/json")
	var f, rn, ra any
	if field != "" {
		f = field
	}
	if reason != "" {
		rn = reason
	}
	if retry != nil {
		ra = *retry
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   apiError{Code: code, Message: msg, Field: f, Reason: rn, Retry: ra},
		"credits": map[string]int{"reserved": 0, "committed": 0, "released": 0},
	})
}

func (s *Server) postImage(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if !isJSONContentType(ct) {
		s.writeErr(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type application/json olmalı", "", "", nil)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		s.writeErr(w, http.StatusBadRequest, "invalid_json", "Gövde okunamadı", "", "", nil)
		return
	}
	if len(body) > maxBodyBytes {
		s.writeErr(w, http.StatusRequestEntityTooLarge, "body_too_large", "Gövde 32 KiB sınırını aşıyor", "", "", nil)
		return
	}
	req, ferr := parseImageRequest(body)
	if ferr != nil {
		s.writeErr(w, ferr.status, ferr.code, ferr.msg, ferr.field, ferr.reason, nil)
		return
	}
	idem := r.Header.Get("Idempotency-Key")
	if idem == "" {
		s.writeErr(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key başlığı zorunludur", "", "", nil)
		return
	}
	if !validIdempotencyKey(idem) {
		s.writeErr(w, http.StatusBadRequest, "invalid_idempotency_key", "Idempotency-Key biçimi geçersiz", "", "", nil)
		return
	}

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		s.writeErr(w, http.StatusUnauthorized, "license", "Authorization Bearer gerekli", "", "", nil)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token == "" || s.dir == nil {
		s.writeErr(w, http.StatusUnauthorized, "license", "Authorization Bearer gerekli", "", "", nil)
		return
	}
	lic, ok := s.dir.Lookup(token)
	if !ok {
		s.writeErr(w, http.StatusUnauthorized, "license", "Lisans bulunamadı", "", "", nil)
		return
	}
	if !lic.Active {
		s.writeErr(w, http.StatusForbidden, "license_invalid", "Lisans aktif değil", "", "", nil)
		return
	}
	if lic.Kind == "use" {
		s.writeErr(w, http.StatusForbidden, "plan_kind_not_supported", "Saatlik use planı image ile henüz eşlenmedi", "", "", nil)
		return
	}

	rate, ok := s.lookupRate(req.Model, req.Size)
	if !ok {
		retry := 5
		s.writeErr(w, http.StatusServiceUnavailable, "reservation_failed", "Kredi tarifesi henüz yok; rezervasyon yapılamaz", "", "rate_table_pending", &retry)
		return
	}
	need := rate * req.N
	_ = need

	fp := s.fingerprint(req, false)
	idemID := lic.ID + "\x00" + idem

	s.mu.Lock()
	existing := s.byIdem[idemID]
	if existing != nil {
		if existing.Fingerprint != fp {
			s.mu.Unlock()
			s.writeErr(w, http.StatusConflict, "idempotency_key_reused", "Aynı anahtar farklı istekle kullanıldı", "", "", nil)
			return
		}
		s.replayLocked(w, existing)
		s.mu.Unlock()
		return
	}

	if running, busy := s.activeLic[lic.ID]; busy && running != "" {
		s.mu.Unlock()
		retry := 1
		s.writeErr(w, http.StatusTooManyRequests, "queue_full", "Lisans başına en çok bir eşzamanlı iş", "", "queue_full", &retry)
		return
	}

	if s.snap.DailyLeftSec < 0 {
		s.mu.Unlock()
		retry := 3600
		s.writeErr(w, http.StatusServiceUnavailable, "not_ready", "Günlük runtime yetersiz", "", "daily_runtime", &retry)
		return
	}
	if !s.snap.MaxUptimeOK && s.snap.State == "IMAGE_READY" {
		s.mu.Unlock()
		retry := 600
		s.writeErr(w, http.StatusServiceUnavailable, "not_ready", "max_uptime yetersiz", "", "max_uptime", &retry)
		return
	}
	if !s.snap.ModelStaged[req.Model] {
		s.mu.Unlock()
		retry := 3600
		s.writeErr(w, http.StatusServiceUnavailable, "not_ready", "Model ready değil veya sahnelenmedi", "", "model_not_staged", &retry)
		return
	}

	j := &job{
		ID:            "imgjob_" + randomID(),
		LicenseID:     lic.ID,
		IdemKey:       idem,
		Fingerprint:   fp,
		Model:         req.Model,
		Size:          req.Size,
		Prompt:        req.Prompt,
		N:             req.N,
		Seed:          req.Seed,
		Created:       s.now(),
		ReservationID: "rsv_" + randomID(),
		Reserved:      need,
	}
	s.jobs[j.ID] = j
	s.byIdem[idemID] = j

	readySync := s.snap.State == "IMAGE_READY" && s.snap.SlotReady && s.gen != nil
	if readySync {
		s.activeLic[lic.ID] = j.ID
		s.mu.Unlock()
		s.runSync(j, req)
		s.mu.Lock()
		delete(s.activeLic, lic.ID)
		s.replayLocked(w, j)
		s.mu.Unlock()
		return
	}

	// Cold/swap: product allows POST to request boot/swap; this package
	// records intent only and does not start a GPU/VM.
	if s.acceptColdSwap && (s.snap.State == "COLD" || s.snap.State == "CHAT_READY") {
		j.Status = "starting"
		j.BootIntent = true
		s.jobs[j.ID] = j
		s.mu.Unlock()
		if s.bootIntent != nil {
			s.bootIntent(j.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v1/image/jobs/"+j.ID)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": j.ID, "status": j.Status})
		return
	}

	delete(s.jobs, j.ID)
	delete(s.byIdem, idemID)
	s.mu.Unlock()
	retry := 300
	s.writeErr(w, http.StatusServiceUnavailable, "not_ready", "Boot yetkisi yok veya kabul edilemiyor", "", "boot_pending", &retry)
}

func (s *Server) replayLocked(w http.ResponseWriter, j *job) {
	w.Header().Set("Idempotent-Replay", "true")
	if !j.Terminal {
		w.Header().Set("Content-Type", "application/json")
		if j.Status == "queued" || j.Status == "starting" || j.Status == "running" {
			w.Header().Set("Location", "/v1/image/jobs/"+j.ID)
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": j.ID, "status": j.Status})
			return
		}
	}
	if j.Failed != nil && j.ReplayStatus >= 400 {
		s.writeErr(w, j.ReplayStatus, j.Failed.Code, j.Failed.Message, "", fmt.Sprint(j.Failed.Reason), nil)
		return
	}
	s.writeJobJSON(w, j, http.StatusOK)
}

func (s *Server) runSync(j *job, req parsedReq) {
	seed := j.ChosenSeed
	if req.Seed != nil {
		seed = *req.Seed
	} else {
		var b [4]byte
		_, _ = rand.Read(b[:])
		seed = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}
	j.ChosenSeed = seed
	wh := allowedSizes[req.Size]
	var outs []outputMeta
	for i := 0; i < req.N; i++ {
		is := seed + uint32(i)
		ct, body, err := s.gen(req.Model, req.Prompt, req.Size, is)
		if err != nil {
			j.Status = "failed"
			j.Terminal = true
			j.Released = j.Reserved
			j.Committed = 0
			j.ReplayStatus = http.StatusBadGateway
			j.Failed = &apiError{Code: "engine_error", Message: "Motor hatası", Reason: "engine_error"}
			return
		}
		sum := sha256.Sum256(body)
		ref := "imgout_" + randomID()
		meta := outputMeta{
			Index: i, Ref: ref, ContentType: ct,
			Width: wh[0], Height: wh[1], Bytes: len(body),
			SHA256: hex.EncodeToString(sum[:]), Seed: is,
		}
		s.mu.Lock()
		s.outputs[ref] = &output{Meta: meta, Body: body, LicenseID: j.LicenseID, Created: s.now()}
		s.mu.Unlock()
		outs = append(outs, meta)
	}
	rate, _ := s.lookupRate(req.Model, req.Size)
	j.Outputs = outs
	j.Status = "succeeded"
	j.Terminal = true
	j.Committed = rate * len(outs)
	j.Released = j.Reserved - j.Committed
	j.ReplayStatus = http.StatusOK
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	lic, ok := s.authLicense(w, r)
	if !ok {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/image/jobs/")
	if id == "" || strings.Contains(id, "/") {
		s.writeErr(w, http.StatusUnauthorized, "license", "İş bulunamadı", "", "", nil)
		return
	}
	s.mu.Lock()
	j := s.jobs[id]
	if j == nil || j.LicenseID != lic.ID {
		s.mu.Unlock()
		// Do not reveal other licenses' jobs.
		s.writeErr(w, http.StatusUnauthorized, "license", "İş bulunamadı", "", "", nil)
		return
	}
	// Poll must not request boot or refresh idle; BootIntent is left as-is.
	if !j.Terminal {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": j.ID, "status": j.Status})
		s.mu.Unlock()
		return
	}
	s.writeJobJSON(w, j, http.StatusOK)
	s.mu.Unlock()
}

func (s *Server) getOutput(w http.ResponseWriter, r *http.Request) {
	lic, ok := s.authLicense(w, r)
	if !ok {
		return
	}
	ref := strings.TrimPrefix(r.URL.Path, "/v1/image/outputs/")
	s.mu.Lock()
	o := s.outputs[ref]
	if o == nil || o.LicenseID != lic.ID {
		s.mu.Unlock()
		s.writeErr(w, http.StatusUnauthorized, "license", "Çıktı bulunamadı", "", "", nil)
		return
	}
	if s.now().Sub(o.Created) > retention {
		s.mu.Unlock()
		// Access not promised after 24h; object is not auto-deleted.
		s.writeErr(w, http.StatusUnauthorized, "license", "Çıktı erişim penceresi doldu", "", "", nil)
		return
	}
	body := o.Body
	ct := o.Meta.ContentType
	s.mu.Unlock()
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-US-Metered", "1")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (s *Server) authLicense(w http.ResponseWriter, r *http.Request) (License, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		s.writeErr(w, http.StatusUnauthorized, "license", "Authorization Bearer gerekli", "", "", nil)
		return License{}, false
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if s.dir == nil {
		s.writeErr(w, http.StatusUnauthorized, "license", "Authorization Bearer gerekli", "", "", nil)
		return License{}, false
	}
	lic, ok := s.dir.Lookup(token)
	if !ok || !lic.Active {
		s.writeErr(w, http.StatusUnauthorized, "license", "Lisans bulunamadı", "", "", nil)
		return License{}, false
	}
	return lic, true
}

func (s *Server) writeJobJSON(w http.ResponseWriter, j *job, status int) {
	alias := j.Model
	repo, rev, steps, guidance := modelParams(alias)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object":          "image.generation",
		"id":              j.ID,
		"idempotency_key": j.IdemKey,
		"status":          j.Status,
		"partial_reason":  nil,
		"model":           map[string]any{"alias": alias, "repo": repo, "revision": rev},
		"params":          map[string]any{"steps": steps, "guidance": guidance, "dtype": "bf16", "cuda_batch": 1},
		"size":            j.Size,
		"n_requested":     j.N,
		"n_produced":      len(j.Outputs),
		"seed":            j.ChosenSeed,
		"outputs":         j.Outputs,
		"fallback":        map[string]any{"used": false, "attempts": 0, "reason": nil},
		"credits": map[string]any{
			"unit": "credits", "reserved": j.Reserved, "committed": j.Committed, "released": j.Released,
		},
		"timing_ms": map[string]int{"admission": 0, "inference_total": 0, "encode_persist": 0, "total": 0},
		"retention": map[string]string{"policy": retentionPolicy},
	})
}

func modelParams(alias string) (repo, rev string, steps, guidance int) {
	if alias == "schnell" {
		return schnellRepo, schnellRev, 4, 0
	}
	return kleinRepo, kleinRev, 4, 1
}

func (s *Server) lookupRate(model, size string) (int, bool) {
	if len(s.rates) == 0 {
		return 0, false
	}
	v, ok := s.rates[RateKey{Model: model, Size: size}]
	return v, ok
}

func (s *Server) fingerprint(req parsedReq, seedMissingAsEmpty bool) string {
	_ = seedMissingAsEmpty
	seedPart := "yok"
	if req.Seed != nil {
		seedPart = strconv.FormatUint(uint64(*req.Seed), 10)
	}
	mac := hmac.New(sha256.New, s.hmacKey)
	fmt.Fprintf(mac, "%s\n%s\n%s\n%d\n%s", req.Model, req.Prompt, req.Size, req.N, seedPart)
	return hex.EncodeToString(mac.Sum(nil))
}

// JobBootIntent reports whether POST recorded a boot/swap intent. Tests use
// this to prove POST may request boot without the poll path doing so.
func (s *Server) JobBootIntent(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[id]
	if j == nil {
		return false
	}
	return j.BootIntent
}

func (s *Server) reservationIDExposed(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[id]
	if j == nil {
		return ""
	}
	return j.ReservationID
}

func randomID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
