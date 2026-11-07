package main

import (
	"bytes"
	"fmt"
	"strings"

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
	return ""
}

// pricingSummary uses GLOBAL patch 08's TRY-only fallback wording.
// No FX approximation is shown until a trusted signed rate is available.
func pricingSummary() string {
	m, h := monthlyOffer(), hourlyOffer()
	var packs []string
	for _, o := range licensesrv.Catalog {
		if strings.HasPrefix(o.SKU, "US-CREDIT-") {
			packs = append(packs, fmt.Sprintf("%d istem %s", o.Plan.Prompts, licensesrv.TRYText(o.PriceTRY)))
		}
	}
	return fmt.Sprintf("Aylık abonelik %s, %d istem / %d gün. Saatlik kullanım %s, 1 saat işlem veya açık panel. İstem paketleri: %s. Fiyatlar KDV dahil TL.",
		m.PriceText(), m.Plan.Prompts, m.Plan.Days, h.PriceText(), strings.Join(packs, ", "))
}

const pricingMarker = "{{US_FIYAT}}"

func renderGUIPage(page []byte) []byte {
	out := bytes.ReplaceAll(page, []byte(pricingMarker), []byte(pricingSummary()))
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
