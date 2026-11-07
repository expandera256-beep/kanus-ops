package i18n

import (
	"strings"
	"testing"
	"time"
)

// TestFourLocaleParity: TR/EN/DE/RU are shipped and have identical keys.
func TestFourLocaleParity(t *testing.T) {
	b := mustLoad(t)
	have := strings.Join(b.Languages(), ",")
	for _, c := range []string{"de", "en", "ru", "tr"} {
		if !strings.Contains(","+have+",", ","+c+",") {
			t.Fatalf("locale %s missing (have %s)", c, have)
		}
	}
	for _, c := range []string{"de", "ru"} {
		for _, ref := range []string{"tr", "en"} {
			if m := b.MissingKeys(ref, c); len(m) > 0 {
				t.Errorf("%s missing %v (ref %s)", c, m, ref)
			}
			if m := b.MissingKeys(c, ref); len(m) > 0 {
				t.Errorf("%s has extra %v (ref %s)", c, m, ref)
			}
		}
		// every value is localized (not an empty or English copy), except
		// brand/product names and a few technical words.
		same := map[string]bool{"brand.name": true, "product.name": true, "gui.nav.terminal": true,
			"gui.terminal.title": true, "gui.nav.monitor": true, "gui.nav.studio": true}
		for k, v := range b.cats[c].msgs {
			if v == "" {
				t.Errorf("%s/%s empty", c, k)
			}
			if !same[k] && v == b.cats["en"].msgs[k] {
				t.Errorf("%s/%s is an English copy: %q", c, k, v)
			}
		}
	}
}

func TestFormattingDeRu(t *testing.T) {
	b := mustLoad(t)
	de, ru := b.For("de"), b.For("ru")
	cases := map[string]string{
		de.Money(123450, "TRY"):                         "1.234,50 ₺",
		ru.Money(123450, "TRY"):                         "1\u00a0234,50 ₺",
		de.Number(1234567.891, 2):                       "1.234.567,89",
		ru.Number(1234567.891, 2):                       "1\u00a0234\u00a0567,89",
		de.T("err.network"):                             "Der Server ist nicht erreichbar. Prüfe deine Verbindung und versuche es erneut.",
		ru.T("ui.credits", "count", ru.Number(1500, 0)): "Осталось кредитов: 1\u00a0500",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
	ist := time.FixedZone("+03", 3*3600)
	ts := time.Date(2026, 11, 6, 6, 5, 0, 0, time.UTC)
	if g := de.DateTime(ts, ist); g != "06.11.2026 09:05 +03" {
		t.Error(g)
	}
	if g := ru.Date(ts, ist); g != "06.11.2026" {
		t.Error(g)
	}
	if ru.Dir() != "ltr" || de.Dir() != "ltr" {
		t.Error("dir")
	}
}
