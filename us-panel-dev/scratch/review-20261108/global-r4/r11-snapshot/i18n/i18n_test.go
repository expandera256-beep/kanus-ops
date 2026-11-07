package i18n

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func mustLoad(t *testing.T) *Bundle {
	t.Helper()
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNoMissingKeysBetweenLocales(t *testing.T) {
	b := mustLoad(t)
	for _, a := range b.Languages() {
		for _, c := range b.Languages() {
			if m := b.MissingKeys(a, c); len(m) > 0 {
				t.Errorf("%s missing keys from %s: %v", c, a, m)
			}
		}
	}
}

func TestPlaceholdersMatch(t *testing.T) {
	b := mustLoad(t)
	re := regexp.MustCompile(`\{[a-z_]+\}`)
	tr := b.cats["tr"]
	for _, lang := range b.Languages() {
		for k, v := range tr.msgs {
			got := re.FindAllString(b.cats[lang].msgs[k], -1)
			want := re.FindAllString(v, -1)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("%s/%s placeholders %v != %v", lang, k, got, want)
			}
		}
	}
}

func TestForbiddenSloganAbsent(t *testing.T) {
	b := mustLoad(t)
	bad := []string{"yerli teknoloji", "küresel vizyon", "built in türkiye", "designed for the world", "yerli ve milli", "homegrown", "türkiye merkezli"}
	for _, lang := range b.Languages() {
		for k, v := range b.cats[lang].msgs {
			lv := strings.ToLower(v)
			for _, s := range bad {
				if strings.Contains(lv, s) {
					t.Errorf("%s/%s contains forbidden phrase %q", lang, k, s)
				}
			}
		}
	}
}

func TestSloganSingleKey(t *testing.T) {
	b := mustLoad(t)
	if got := b.For("tr").T("brand.slogan"); got != "Fikirlerini yazılıma dönüştür." {
		t.Fatal(got)
	}
	if got := b.For("en").T("brand.slogan"); got != "Turn your ideas into software." {
		t.Fatal(got)
	}
}

func TestNegotiate(t *testing.T) {
	b := mustLoad(t)
	cases := []struct{ pref, al, want string }{
		{"", "", "tr"},
		{"en", "tr-TR", "en"},
		{"EN_us", "", "en"},
		{"xx", "en-GB,en;q=0.8", "en"},
		{"", "de-DE,de;q=0.9,en;q=0.5", "de"},
		{"", "ru-RU,ru;q=0.9", "ru"},
		{"", "ja-JP,ja;q=0.9,en;q=0.5", "en"},
		{"", "fr;q=0.9, tr;q=0.95, en;q=0.1", "tr"},
		{"", "en;q=0", "tr"},
		{"", "*", "tr"},
	}
	for _, c := range cases {
		if got := b.Negotiate(c.pref, c.al); got != c.want {
			t.Errorf("Negotiate(%q,%q)=%q want %q", c.pref, c.al, got, c.want)
		}
	}
}

func TestFallbackAndPlaceholders(t *testing.T) {
	b := mustLoad(t)
	if err := b.Add("zz", []byte(`{"_meta":{"name":"Test","dir":"rtl"},"ui.login":"ZZ"}`)); err != nil {
		t.Fatal(err)
	}
	defer func() { b.mu.Lock(); delete(b.cats, "zz"); b.mu.Unlock() }()
	l := b.For("zz")
	if l.Dir() != "rtl" {
		t.Fatal("rtl not honored")
	}
	if l.T("ui.login") != "ZZ" {
		t.Fatal("own key")
	}
	if l.T("ui.logout") != "Sign out" {
		t.Fatal("fallback to en failed:", l.T("ui.logout"))
	}
	if l.T("no.such") != "no.such" {
		t.Fatal("unknown key should echo key")
	}
	if got := b.For("tr").T("ui.credits", "count", 42); got != "Kalan kredi: 42" {
		t.Fatal(got)
	}
	if b.For("qq").Lang() != "tr" {
		t.Fatal("unknown lang should map to default")
	}
}

func TestAddRejectsBadLocale(t *testing.T) {
	b := &Bundle{cats: map[string]*Catalog{}}
	if b.Add("x", []byte(`{`)) == nil {
		t.Fatal("bad json accepted")
	}
	if b.Add("x", []byte(`{"_meta":{"dir":"up"}}`)) == nil {
		t.Fatal("bad dir accepted")
	}
	if b.Add("x", []byte(`{"k":1}`)) == nil {
		t.Fatal("non-string accepted")
	}
}

func TestFormatting(t *testing.T) {
	b := mustLoad(t)
	tr, en := b.For("tr"), b.For("en")
	cases := map[string]string{
		tr.Money(123450, "TRY"):   "1.234,50 ₺",
		en.Money(123450, "USD"):   "$1,234.50",
		en.Money(-500, "GBP"):     "£-5.00",
		tr.Money(100, "EUR"):      "1,00 €",
		en.Money(100, "EUR"):      "1.00 €",
		en.Money(100, "XYZ"):      "1.00 XYZ",
		tr.Number(1234567.891, 2): "1.234.567,89",
		en.Number(999, 0):         "999",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
	ist := time.FixedZone("+03", 3*3600)
	ts := time.Date(2026, 11, 6, 6, 5, 0, 0, time.UTC)
	if g := tr.DateTime(ts, ist); g != "06.11.2026 09:05 +03" {
		t.Error(g)
	}
	if g := en.Date(ts, ist); g != "Nov 6, 2026" {
		t.Error(g)
	}
}

func TestConcurrentUse(t *testing.T) {
	b := mustLoad(t)
	done := make(chan bool)
	for i := 0; i < 8; i++ {
		go func() {
			for j := 0; j < 200; j++ {
				_ = b.For("en").T("ui.login")
				_ = b.Negotiate("", "en")
			}
			done <- true
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
}
