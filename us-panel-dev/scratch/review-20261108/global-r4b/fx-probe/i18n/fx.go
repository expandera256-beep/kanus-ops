package i18n

// Approximate foreign-currency display next to TRY prices (display only).
// Payments and the ledger stay TRY; nothing here feeds a charge or credit.
// Input is a kanus.fx/1 snapshot built from TCMB today.xml ForexSelling.

import (
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strings"
	"time"
)

// FXMaxAge: older snapshots are ignored (weekends + bank holidays).
const FXMaxAge = 5 * 24 * time.Hour

// FXSnapshot holds TRY per 1 EUR / 1 USD on the TCMB bulletin date.
type FXSnapshot struct {
	Date     time.Time
	EUR, USD float64
}

var fxDateRe = regexp.MustCompile(`^\d{4}-\d\d-\d\d$`)

// ParseFX validates a kanus.fx/1 document. Any problem returns an error and
// callers must then show TRY only.
func ParseFX(raw []byte, now time.Time) (FXSnapshot, error) {
	var z FXSnapshot
	if len(raw) == 0 || len(raw) > 4096 {
		return z, errors.New("fx: size")
	}
	var d struct {
		Schema string                     `json:"schema"`
		Base   string                     `json:"base"`
		Date   string                     `json:"date"`
		Rates  map[string]json.RawMessage `json:"rates"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return z, errors.New("fx: json")
	}
	if d.Schema != "kanus.fx/1" || d.Base != "TRY" || !fxDateRe.MatchString(d.Date) || len(d.Rates) != 2 {
		return z, errors.New("fx: schema")
	}
	t, err := time.Parse("2006-01-02", d.Date)
	if err != nil {
		return z, errors.New("fx: date")
	}
	if age := now.Sub(t); age < -36*time.Hour || age > FXMaxAge+24*time.Hour {
		return z, errors.New("fx: stale")
	}
	rate := func(k string) (float64, error) {
		var v float64
		r, ok := d.Rates[k]
		if !ok || len(r) == 0 || r[0] == '"' || json.Unmarshal(r, &v) != nil || math.IsNaN(v) || v < 1 || v > 1000 {
			return 0, errors.New("fx: rate " + k)
		}
		return v, nil
	}
	eur, err := rate("EUR")
	if err != nil {
		return z, err
	}
	usd, err := rate("USD")
	if err != nil {
		return z, err
	}
	if x := eur / usd; x < 0.8 || x > 1.6 {
		return z, errors.New("fx: cross")
	}
	return FXSnapshot{Date: t, EUR: eur, USD: usd}, nil
}

var fxNotes = map[string]string{
	"tr": "Yaklaşık tutar; ödeme TL ile alınır. Kur: TCMB döviz satış, %s.",
	"en": "Approximate amount; payment is charged in Turkish lira. Rate: CBRT forex selling, %s.",
	"de": "Ungefährer Betrag; bezahlt wird in türkischen Lira. Kurs: TCMB-Devisenverkaufskurs vom %s.",
	"ru": "Сумма приблизительная; оплата производится в турецких лирах. Курс: ЦБ Турции, продажа валюты, %s.",
}

// ApproxFX returns "≈ 337,74 € · ≈ 377,45 $" style text and the localized
// note for a TRY amount in kuruş. ok=false (zero snapshot) -> "", "".
func (l *Localizer) ApproxFX(tryMinor int64, s FXSnapshot) (approx, note string) {
	if s.EUR <= 0 || s.USD <= 0 || s.Date.IsZero() || tryMinor < 0 {
		return "", ""
	}
	tl := float64(tryMinor) / 100
	approx = "≈ " + l.Money(int64(math.Round(tl/s.EUR*100)), "EUR") + " · ≈ " + l.Money(int64(math.Round(tl/s.USD*100)), "USD")
	n, ok := fxNotes[l.lang]
	if !ok {
		n = fxNotes["en"]
	}
	return approx, strings.Replace(n, "%s", l.Date(s.Date, time.UTC), 1)
}
