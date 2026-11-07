package main

import (
	"bytes"
	"html"
	"net/http"
	"strings"

	"kiyas/i18n"
	"kiyas/licensesrv"
)

// Tüm fiyat metinleri licensesrv.Catalog'dan gelir; burada sabit fiyat yok.

func offerOrPanic(sku string) licensesrv.Offer {
	o, ok := licensesrv.OfferBySKU(sku)
	if !ok {
		panic("fiyat tablosunda yok: " + sku)
	}
	return o
}

func monthlyOffer() licensesrv.Offer { return offerOrPanic("US-MONTHLY-200") }
func hourlyOffer() licensesrv.Offer  { return offerOrPanic("US-HOUR-1") }

func priceNote(doc licenseDoc) string {
	if doc.TRY > 0 {
		return " " + licensesrv.TRYText(doc.TRY)
	}
	if doc.USD > 0 {
		return " " + itoa(doc.USD) + " USD"
	}
	return ""
}

// pricingSummary: paneldeki fiyat paragrafı. TL tutarları
// licensesrv.Catalog'dan; yaklaşık EUR/USD yalnız doğrulanmış /v1/fx ile.
// The server-rendered page keeps the TR markup default (like every other
// data-i18n text); the GUI switches it via /api/pricing after applyLocale.
func pricingSummary() string { return pricingSummaryFor("tr") }

func pricingSummaryFor(lang string) string {
	b, err := i18n.Load()
	if err != nil {
		return ""
	}
	l := b.For(lang)
	snap, fxOK := currentFX()
	tl := func(v int) string { return l.T("price.try", "amount", l.Number(float64(v), 0)) }
	withFX := func(v int) string {
		if !fxOK {
			return tl(v)
		}
		a, _ := l.ApproxFX(int64(v)*100, snap)
		return tl(v) + " (" + a + ")"
	}
	m, h := monthlyOffer(), hourlyOffer()
	var packs []string
	for _, o := range licensesrv.Catalog {
		if strings.HasPrefix(o.SKU, "US-CREDIT-") {
			packs = append(packs, l.T("price.pack", "prompts", o.Plan.Prompts, "price", tl(o.PriceTRY)))
		}
	}
	out := l.T("price.summary", "monthly", withFX(m.PriceTRY), "prompts", m.Plan.Prompts, "days", m.Plan.Days, "hourly", withFX(h.PriceTRY), "packs", strings.Join(packs, ", "))
	if fxOK {
		_, note := l.ApproxFX(int64(m.PriceTRY)*100, snap)
		out += " " + note
	}
	return out
}

// pricingHandler: GET /api/pricing?lang=xx -> {"text": ...} for live
// language switching (public catalog text only; no state change).
func pricingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		guiError(w, "err.gui.method", http.StatusMethodNotAllowed)
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = langPref.Get()
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, map[string]string{"text": pricingSummaryFor(lang)})
}

const pricingMarker = "{{US_FIYAT}}"

func renderGUIPage(page []byte) []byte {
	out := bytes.ReplaceAll(page, []byte(pricingMarker), []byte(html.EscapeString(pricingSummary())))
	if commercialRequired() && !hookLicenseMode() {
		out = bytes.ReplaceAll(out, []byte(`<form id="licenseform" class="row">`), []byte(`<form id="licenseform" class="row" hidden>`))
	}
	if hookLicenseMode() {
		for _, pair := range [][2]string{{`<form id="gatewayform">`, `<form id="gatewayform" hidden>`}, {`id="accountrequests"`, `id="accountrequests" hidden`}, {`id="accountworkers"`, `id="accountworkers" hidden`}} {
			out = bytes.ReplaceAll(out, []byte(pair[0]), []byte(pair[1]))
		}
		out = bytes.ReplaceAll(out, []byte(`<form id="accountform">`), []byte(`<p>Bu sürüm Go lisans anahtarı kullanır; Python hesap girişi kapalıdır.</p><form id="accountform" hidden>`))
	}
	return out
}
