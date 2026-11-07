package i18n

import (
	"strings"
	"testing"
	"time"
)

const fxGood = `{"schema":"kanus.fx/1","source":"TCMB today.xml ForexSelling (döviz satış)","tcmb_tarih":"06.11.2026","date":"2026-11-06","base":"TRY","rates":{"USD":50.0724,"EUR":55.9609}}`

var fxNow = time.Date(2026, 11, 8, 0, 37, 0, 0, time.UTC)

func TestFXParseGood(t *testing.T) {
	s, err := ParseFX([]byte(fxGood), fxNow)
	if err != nil || s.EUR != 55.9609 || s.USD != 50.0724 || s.Date.Format("2006-01-02") != "2026-11-06" {
		t.Fatalf("%v %+v", err, s)
	}
}

func TestFXParseRejects(t *testing.T) {
	for name, raw := range map[string]string{
		"empty": ``, "json": `{`, "schema": strings.Replace(fxGood, "kanus.fx/1", "x", 1),
		"base": strings.Replace(fxGood, `"TRY"`, `"USD"`, 1), "date": strings.Replace(fxGood, "2026-11-06", "06.11.2026", 2),
		"string rate": strings.Replace(fxGood, `50.0724`, `"50.0724"`, 1), "zero": strings.Replace(fxGood, `50.0724`, `0`, 1),
		"missing": strings.Replace(fxGood, `,"EUR":55.9609`, ``, 1), "extra": strings.Replace(fxGood, `}}`, `,"GBP":60}}`, 1),
		"cross": strings.Replace(fxGood, `55.9609`, `155.9609`, 1), "huge": strings.Repeat(" ", 5000) + fxGood,
	} {
		if _, err := ParseFX([]byte(raw), fxNow); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := ParseFX([]byte(fxGood), fxNow.Add(10*24*time.Hour)); err == nil {
		t.Error("stale accepted")
	}
	if _, err := ParseFX([]byte(fxGood), fxNow.Add(-5*24*time.Hour)); err == nil {
		t.Error("future accepted")
	}
}

func TestFXApprox4Langs(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	s, _ := ParseFX([]byte(fxGood), fxNow)
	want := map[string][2]string{
		"tr": {"≈ 337,74 € · ≈ 377,45 $", "ödeme TL ile alınır"},
		"en": {"≈ 337.74 € · ≈ $377.45", "charged in Turkish lira"},
		"de": {"≈ 337,74 € · ≈ 377,45 $", "türkischen Lira"},
		"ru": {"≈ 337,74 € · ≈ 377,45 $", "турецких лирах"},
	}
	for lang, w := range want {
		a, n := b.For(lang).ApproxFX(1890000, s)
		if a != w[0] || !strings.Contains(n, w[1]) || !strings.Contains(n, "2026") {
			t.Errorf("%s: %q %q", lang, a, n)
		}
	}
	if _, n := b.For("tr").ApproxFX(1890000, s); !strings.Contains(n, "06.11.2026") {
		t.Errorf("TR note must show the TCMB day in Türkiye time: %q", n)
	}
	if a, n := b.For("tr").ApproxFX(1890000, FXSnapshot{}); a != "" || n != "" {
		t.Error("zero snapshot must hide FX")
	}
}
