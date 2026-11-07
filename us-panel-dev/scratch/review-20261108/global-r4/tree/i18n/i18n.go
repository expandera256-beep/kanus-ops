// Package i18n provides locale files, language negotiation, fallback and
// locale-aware date/number/currency formatting for US Panel.
// Adding a language (de, fr, es, ar ...) only requires locales/<code>.json;
// "_meta.dir" = "rtl" marks right-to-left languages.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed locales/*.json
var localeFS embed.FS

const Default = "tr"
const Fallback = "en"

type Meta struct {
	Name     string `json:"name"`
	Dir      string `json:"dir"`
	BCP47    string `json:"bcp47"`
	Currency string `json:"currency"`
}

type Catalog struct {
	Code string
	Meta Meta
	msgs map[string]string
}

type Bundle struct {
	mu   sync.RWMutex
	cats map[string]*Catalog
}

var (
	defOnce sync.Once
	def     *Bundle
	defErr  error
)

// Load returns the embedded bundle.
func Load() (*Bundle, error) {
	defOnce.Do(func() { def, defErr = loadFS() })
	return def, defErr
}

func loadFS() (*Bundle, error) {
	ents, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil, err
	}
	b := &Bundle{cats: map[string]*Catalog{}}
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := localeFS.ReadFile(path.Join("locales", e.Name()))
		if err != nil {
			return nil, err
		}
		if err := b.Add(strings.TrimSuffix(e.Name(), ".json"), raw); err != nil {
			return nil, err
		}
	}
	for _, c := range []string{Default, Fallback} {
		if b.cats[c] == nil {
			return nil, fmt.Errorf("i18n: required locale %q missing", c)
		}
	}
	return b, nil
}

// Add parses a locale JSON document and registers it.
func (b *Bundle) Add(code string, raw []byte) error {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("i18n: %s: %w", code, err)
	}
	c := &Catalog{Code: normalize(code), msgs: map[string]string{}}
	for k, v := range doc {
		if k == "_meta" {
			if err := json.Unmarshal(v, &c.Meta); err != nil {
				return fmt.Errorf("i18n: %s meta: %w", code, err)
			}
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("i18n: %s key %s: not a string", code, k)
		}
		c.msgs[k] = s
	}
	if c.Meta.Dir == "" {
		c.Meta.Dir = "ltr"
	}
	if c.Meta.Dir != "ltr" && c.Meta.Dir != "rtl" {
		return fmt.Errorf("i18n: %s: bad dir %q", code, c.Meta.Dir)
	}
	b.mu.Lock()
	b.cats[c.Code] = c
	b.mu.Unlock()
	return nil
}

func normalize(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	tag = strings.ReplaceAll(tag, "_", "-")
	if i := strings.IndexByte(tag, '-'); i > 0 {
		tag = tag[:i]
	}
	return tag
}

// Languages returns supported codes sorted.
func (b *Bundle) Languages() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, len(b.cats))
	for k := range b.cats {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (b *Bundle) has(code string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cats[code] != nil
}

// Negotiate picks a language: explicit user preference first, then the
// Accept-Language header (q-values honored), then Default.
func (b *Bundle) Negotiate(userPref, acceptLanguage string) string {
	if p := normalize(userPref); p != "" && b.has(p) {
		return p
	}
	type cand struct {
		code string
		q    float64
		pos  int
	}
	var cs []cand
	for i, part := range strings.Split(acceptLanguage, ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		code := normalize(fields[0])
		if code == "" || code == "*" {
			continue
		}
		q := 1.0
		for _, f := range fields[1:] {
			f = strings.TrimSpace(f)
			if strings.HasPrefix(f, "q=") {
				if v, err := strconv.ParseFloat(f[2:], 64); err == nil {
					q = v
				}
			}
		}
		if q > 0 {
			cs = append(cs, cand{code, q, i})
		}
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].q > cs[j].q })
	for _, c := range cs {
		if b.has(c.code) {
			return c.code
		}
	}
	return Default
}

// Localizer is bound to one language.
type Localizer struct {
	b    *Bundle
	lang string
}

func (b *Bundle) For(lang string) *Localizer {
	l := normalize(lang)
	if !b.has(l) {
		l = Default
	}
	return &Localizer{b: b, lang: l}
}

func (l *Localizer) Lang() string { return l.lang }

func (l *Localizer) Meta() Meta {
	l.b.mu.RLock()
	defer l.b.mu.RUnlock()
	return l.b.cats[l.lang].Meta
}

// Dir returns "ltr" or "rtl" for the HTML dir attribute.
func (l *Localizer) Dir() string { return l.Meta().Dir }

// T translates key with {name} placeholders. Missing keys fall back to
// Fallback then Default, then the key itself (never empty).
func (l *Localizer) T(key string, args ...any) string {
	l.b.mu.RLock()
	s, ok := l.b.cats[l.lang].msgs[key]
	if !ok {
		for _, fb := range []string{Fallback, Default} {
			if c := l.b.cats[fb]; c != nil {
				if s, ok = c.msgs[key]; ok {
					break
				}
			}
		}
	}
	l.b.mu.RUnlock()
	if !ok {
		return key
	}
	for i := 0; i+1 < len(args); i += 2 {
		s = strings.ReplaceAll(s, "{"+fmt.Sprint(args[i])+"}", fmt.Sprint(args[i+1]))
	}
	return s
}

// MissingKeys lists keys present in the reference locale but absent in lang.
func (b *Bundle) MissingKeys(ref, lang string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []string
	r, c := b.cats[ref], b.cats[lang]
	if r == nil || c == nil {
		return nil
	}
	for k := range r.msgs {
		if _, ok := c.msgs[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

type numFmt struct{ group, decimal string }

var numFormats = map[string]numFmt{
	"tr": {".", ","}, "de": {".", ","}, "es": {".", ","},
	"fr": {"\u202f", ","}, "en": {",", "."}, "ar": {",", "."},
	"ru": {"\u00a0", ","},
}

type curFmt struct {
	symbol string
	before bool
	space  bool
}

var currencies = map[string]curFmt{
	"TRY": {"₺", true, false}, "USD": {"$", true, false}, "EUR": {"€", false, true},
	"GBP": {"£", true, false}, "AED": {"AED", true, true}, "SAR": {"SAR", true, true},
	"INR": {"₹", true, false}, "RUB": {"₽", false, true},
}

func (l *Localizer) nf() numFmt {
	if f, ok := numFormats[l.lang]; ok {
		return f
	}
	return numFormats["en"]
}

// Number formats v with the given decimals using locale separators.
func (l *Localizer) Number(v float64, decimals int) string {
	f := l.nf()
	neg := v < 0
	v = math.Abs(v)
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	var g strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			g.WriteString(f.group)
		}
		g.WriteRune(r)
	}
	out := g.String()
	if frac != "" {
		out += f.decimal + frac
	}
	if neg {
		out = "-" + out
	}
	return out
}

// Money formats an amount in minor units (e.g. kuruş/cents).
func (l *Localizer) Money(minor int64, currency string) string {
	currency = strings.ToUpper(currency)
	n := l.Number(float64(minor)/100, 2)
	c, ok := currencies[currency]
	if !ok {
		return n + " " + currency
	}
	if l.lang == "tr" || l.lang == "de" || l.lang == "fr" || l.lang == "es" || l.lang == "ru" {
		// Continental style: amount then symbol, e.g. 1.234,50 ₺ / 1.234,50 €
		return n + " " + c.symbol
	}
	if c.before {
		if c.space {
			return c.symbol + " " + n
		}
		return c.symbol + n
	}
	return n + " " + c.symbol
}

var dateLayouts = map[string][2]string{
	"tr": {"02.01.2006", "02.01.2006 15:04"},
	"de": {"02.01.2006", "02.01.2006 15:04"},
	"ru": {"02.01.2006", "02.01.2006 15:04"},
	"fr": {"02/01/2006", "02/01/2006 15:04"},
	"es": {"02/01/2006", "02/01/2006 15:04"},
	"en": {"Jan 2, 2006", "Jan 2, 2006 3:04 PM"},
	"ar": {"2006-01-02", "2006-01-02 15:04"},
}

func (l *Localizer) layouts() [2]string {
	if d, ok := dateLayouts[l.lang]; ok {
		return d
	}
	return dateLayouts["en"]
}

// Date formats t in loc (nil = t's own zone).
func (l *Localizer) Date(t time.Time, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}
	return t.Format(l.layouts()[0])
}

// DateTime formats t with time and zone abbreviation.
func (l *Localizer) DateTime(t time.Time, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}
	return t.Format(l.layouts()[1]) + " " + t.Format("MST")
}
