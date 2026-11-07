package i18n

// Approximate foreign-currency display next to TRY prices (display only).
// Payments and the ledger stay TRY; nothing here feeds a charge or credit.
// Input is a kanus.fx/1 snapshot built from TCMB today.xml ForexSelling.

import (
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Freshness policy (identical on site, cron and panel): the TCMB bulletin
// date is a calendar day in Türkiye time (Europe/Istanbul, fixed UTC+3
// since 2016; a fixed zone avoids tzdata on Windows). With today = now in
// that zone, a snapshot is valid for today-5 .. today inclusive; any future
// date or today-6 and older is rejected.
const FXMaxAgeDays = 5

// FXZone is Türkiye time used to interpret the bulletin date and "today".
var FXZone = time.FixedZone("TRT", 3*3600)

// FXFresh reports whether bulletin date d (any zone; only its Y-M-D in
// FXZone counts) is within the policy for the injected clock now.
func FXFresh(d, now time.Time) bool {
	dy, dm, dd := d.In(FXZone).Date()
	ny, nm, nd := now.In(FXZone).Date()
	days := int(time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC).Sub(time.Date(dy, dm, dd, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	return days >= 0 && days <= FXMaxAgeDays
}

// FXSnapshot holds TRY per 1 EUR / 1 USD on the TCMB bulletin date
// (Date is midnight of that day in FXZone).
type FXSnapshot struct {
	Date     time.Time
	EUR, USD float64
}

var fxDateRe = regexp.MustCompile(`^\d{4}-\d\d-\d\d$`)

// fxRateOK: finite and within [1, 1000] TRY per unit.
func fxRateOK(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 1 && v <= 1000 }

// ParseFX validates a kanus.fx/1 document against the injected clock now.
// Any problem returns an error and callers must then show TRY only.
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
	t, err := time.ParseInLocation("2006-01-02", d.Date, FXZone)
	if err != nil {
		return z, errors.New("fx: date")
	}
	if !FXFresh(t, now) {
		return z, errors.New("fx: stale or future")
	}
	rate := func(k string) (float64, error) {
		var v float64
		r, ok := d.Rates[k]
		if !ok || len(r) == 0 || r[0] == '"' || json.Unmarshal(r, &v) != nil || !fxRateOK(v) {
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

// FXMaxTRYMinor bounds the converted amount (10^15 kuruş = 10 trillion TL);
// larger inputs hide FX instead of overflowing.
const FXMaxTRYMinor = int64(1e15)

// fxConvert returns round-half-away-from-zero(tryMinor / rate) in foreign
// minor units using exact rational arithmetic on the rate's shortest
// decimal form (e.g. 55.9609). ok=false for invalid input or overflow.
func fxConvert(tryMinor int64, rate float64) (int64, bool) {
	if tryMinor < 0 || tryMinor > FXMaxTRYMinor || !fxRateOK(rate) {
		return 0, false
	}
	r, ok := new(big.Rat).SetString(strconv.FormatFloat(rate, 'g', -1, 64))
	if !ok || r.Sign() <= 0 {
		return 0, false
	}
	q := new(big.Rat).Quo(new(big.Rat).SetInt64(tryMinor), r)
	num, den := q.Num(), q.Denom() // q >= 0
	n := new(big.Int).Add(new(big.Int).Mul(num, big.NewInt(2)), den)
	n.Quo(n, new(big.Int).Mul(den, big.NewInt(2)))
	if !n.IsInt64() {
		return 0, false
	}
	return n.Int64(), true
}

var fxNotes = map[string]string{
	"tr": "Yaklaşık tutar; ödeme TL ile alınır. Kur: TCMB döviz satış, %s.",
	"en": "Approximate amount; payment is charged in Turkish lira. Rate: CBRT forex selling, %s.",
	"de": "Ungefährer Betrag; bezahlt wird in türkischen Lira. Kurs: TCMB-Devisenverkaufskurs vom %s.",
	"ru": "Сумма приблизительная; оплата производится в турецких лирах. Курс: ЦБ Турции, продажа валюты, %s.",
}

// ApproxFX returns "≈ 337,74 € · ≈ 377,45 $" style text and the localized
// note for a TRY amount in kuruş. Invalid snapshot (zero, non-finite,
// out of range) or amount (negative, > FXMaxTRYMinor) -> "", "".
// Rounding: half away from zero to the cent, exact decimal arithmetic.
func (l *Localizer) ApproxFX(tryMinor int64, s FXSnapshot) (approx, note string) {
	if s.Date.IsZero() {
		return "", ""
	}
	eur, ok1 := fxConvert(tryMinor, s.EUR)
	usd, ok2 := fxConvert(tryMinor, s.USD)
	if !ok1 || !ok2 {
		return "", ""
	}
	approx = "≈ " + l.Money(eur, "EUR") + " · ≈ " + l.Money(usd, "USD")
	n, ok := fxNotes[l.lang]
	if !ok {
		n = fxNotes["en"]
	}
	return approx, strings.Replace(n, "%s", l.Date(s.Date, FXZone), 1)
}
