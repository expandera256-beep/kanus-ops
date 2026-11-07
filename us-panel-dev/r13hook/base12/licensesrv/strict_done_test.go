package licensesrv

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const contentEv = "data: {\"choices\":[{\"delta\":{\"content\":\"merhaba\"}}]}\n\n"
const stopEv = "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"

func strictCase(t *testing.T, strict bool, events string) (charged bool, body string) {
	t.Helper()
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, events)
	})
	p.StrictDone = strict
	before, _ := st.License(key)
	w := proxyRequest(p, key, proxyInput, context.Background())
	after, _ := st.License(key)
	if after.Reserved != 0 {
		t.Fatalf("rezervasyon kaldı: %+v", after)
	}
	return after.Credits == before.Credits-1, w.Body.String()
}

func TestStrictDonePolicy(t *testing.T) {
	if c, b := strictCase(t, true, contentEv+stopEv+"data: [DONE]\n\n"); !c || !strings.Contains(b, "[DONE]") {
		t.Fatal("[DONE] varken ücretlendirilmedi")
	}
	n0 := MissingDoneCount()
	if c, b := strictCase(t, true, contentEv+stopEv); c || strings.Contains(b, "data: [DONE]") {
		t.Fatalf("varsayılan (strict): stop ama [DONE] yok → serbest bırakılmalı; ücret=%v", c)
	}
	if MissingDoneCount() != n0+1 {
		t.Fatal("eksik [DONE] sayacı artmadı")
	}
	if c, _ := strictCase(t, false, contentEv+stopEv); !c {
		t.Fatal("US_STRICT_DONE=0: stop tamamlanma sayılmalı")
	}
	if c, _ := strictCase(t, true, contentEv); c {
		t.Fatal("yarıda kesilen akış ücretlendirildi")
	}
	if c, _ := strictCase(t, true, contentEv+"data: [DONE]\n\n"); c {
		t.Fatal("finish olmadan [DONE] ücretlendirildi (strict)")
	}
}

func TestStrictDoneDefaultFromEnv(t *testing.T) {
	t.Setenv("US_STRICT_DONE", "")
	p, _ := NewModelProxy(nil, "", "", "m", 45*60e9, 1, 2)
	if !p.StrictDone {
		t.Fatal("varsayılan strict değil")
	}
	t.Setenv("US_STRICT_DONE", "0")
	p, _ = NewModelProxy(nil, "", "", "m", 45*60e9, 1, 2)
	if p.StrictDone {
		t.Fatal("US_STRICT_DONE=0 uygulanmadı")
	}
}
