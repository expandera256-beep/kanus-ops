package licensesrv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// outboxSection returns the persisted "outbox" JSON of a store file (or a
// byte-for-byte backup of it, as the backup script makes with install -m 640).
func outboxSection(t *testing.T, file string) string {
	t.Helper()
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return string(doc["outbox"])
}

func backupCopy(t *testing.T, src string) string {
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "licenses.json.20261106T000000Z")
	if err := os.WriteFile(dst, raw, 0o640); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestPrivacyOutboxNoRawKeyLogsNoEmail(t *testing.T) {
	h, path := setup(t)
	post(h, payload, sign("whsec", payload), "w1")
	keys := h.Store.st.Orders["555"]
	if len(keys) == 0 {
		t.Fatal("anahtar yok")
	}
	email := h.Store.st.Licenses[keys[0]].Email
	local := email[:strings.Index(email, "@")]
	noLeak := func(where, s string) {
		t.Helper()
		for _, k := range keys {
			if strings.Contains(s, k) {
				t.Fatalf("%s içinde ham lisans anahtarı", where)
			}
		}
	}
	// 1) pending delivery mail persisted (and its backup)
	noLeak("outbox (bekleyen)", outboxSection(t, path))
	noLeak("yedek outbox (bekleyen)", outboxSection(t, backupCopy(t, path)))
	// 2) dry-run log: no email, no key
	var logs []string
	h.Store.FlushOutbox(DryRunSender{Log: func(s string) { logs = append(logs, s) }}, time.Now().Add(time.Hour))
	all := strings.Join(logs, "\n")
	if len(logs) == 0 || strings.Contains(all, email) || strings.Contains(all, local+"@") {
		t.Fatalf("dry-run logu e-postayı açık yazıyor: %q", all)
	}
	noLeak("log", all)
	// 3) real delivery still carries the full key(s) to the sender
	f := &fakeSender{}
	h.Store.FlushOutbox(f, time.Now().Add(48*time.Hour))
	if len(f.got) != 1 {
		t.Fatalf("teslim e-postası gönderilmedi: %d", len(f.got))
	}
	for _, k := range keys {
		if !strings.Contains(f.got[0].Body, k) {
			t.Fatal("teslim e-postasında tam anahtar yok (teslim bozuldu)")
		}
	}
	// 4) after send: outbox and backup still free of raw keys
	noLeak("outbox (gönderildi)", outboxSection(t, path))
	noLeak("yedek outbox (gönderildi)", outboxSection(t, backupCopy(t, path)))
}

func TestPrivacyRevokedMailMaskedKey(t *testing.T) {
	h, path := setup(t)
	post(h, payload, sign("whsec", payload), "w1")
	keys := append([]string(nil), h.Store.st.Orders["555"]...)
	if c := postTopic(h, "orders/cancelled", `{"id":555}`, "c1"); c != 200 {
		t.Fatal(c)
	}
	ob := h.Store.Outbox()
	if len(ob) == 0 || !ob[0].Dead {
		t.Fatal("iptal edilen teslim e-postası ölü değil")
	}
	sec := outboxSection(t, path)
	for _, k := range keys {
		if strings.Contains(sec, k) || strings.Contains(outboxSection(t, backupCopy(t, path)), k) {
			t.Fatal("iptal/iade sonrası outbox/yedekte ham anahtar")
		}
		if !strings.Contains(ob[0].Body, maskLicenseKey(k)) || !strings.HasPrefix(maskLicenseKey(k), "us1.…") {
			t.Fatalf("maskeli anahtar yok: %q", ob[0].Body)
		}
	}
	// Nothing is sent for a cancelled order.
	f := &fakeSender{}
	h.Store.FlushOutbox(f, time.Now().Add(time.Hour))
	if len(f.got) != 0 {
		t.Fatal("iptal edilen sipariş için anahtar gönderildi")
	}
}

func TestPrivacyMaskHelpers(t *testing.T) {
	if got := maskEmail("demir@gmail.com"); got != "d***@gmail.com" {
		t.Fatal(got)
	}
	if got := maskEmail("x"); strings.Contains(got, "x") {
		t.Fatal(got)
	}
	k := "us1.abcdefgh.ijklWXYZ"
	if got := maskLicenseKey(k); got != "us1.…WXYZ" {
		t.Fatal(got)
	}
}
