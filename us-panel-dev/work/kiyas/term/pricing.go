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
	if doc.USD > 0 {
		return " " + itoa(doc.USD) + " USD"
	}
	return ""
}

// pricingSummary: paneldeki fiyat paragrafı.
func pricingSummary() string {
	m, h := monthlyOffer(), hourlyOffer()
	var packs []string
	for _, o := range licensesrv.Catalog {
		if strings.HasPrefix(o.SKU, "US-CREDIT-") {
			packs = append(packs, fmt.Sprintf("%d istem %s", o.Plan.Prompts, licensesrv.TRYText(o.PriceTRY)))
		}
	}
	return fmt.Sprintf("Aylık abonelik %s, %d istem / %d gün. Saatlik kullanım %s, 1 saat işlem veya açık panel. İstem paketleri: %s. Fiyatlar KDV dahil TL; USD karşılığı ≈%.2f TL/USD ile bilgi amaçlıdır.",
		m.PriceText(), m.Plan.Prompts, m.Plan.Days, h.PriceText(), strings.Join(packs, ", "), licensesrv.USDTRYRef)
}

const pricingMarker = "{{US_FIYAT}}"

func renderGUIPage(page []byte) []byte {
	return bytes.ReplaceAll(page, []byte(pricingMarker), []byte(pricingSummary()))
}
