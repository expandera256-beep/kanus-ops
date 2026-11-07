package licensesrv

import (
	"encoding/json"
	"strconv"
)

// Fiyat tablosu: tek kaynak. İstemci metinleri, US_PLANS_JSON taslağı ve test
// mağazası ürün dosyası buradan üretilir. Ayrıntı: reports/FIYATLANDIRMA_TR.md
// Fiyatlar KDV dahil TL. Döviz gösterimi yalnız doğrulanmış güncel FX ile
// istemci tarafında eklenebilir; katalog sabit bir kur sağlamaz.

type Offer struct {
	SKU      string `json:"sku"`
	Title    string `json:"title"`
	PriceTRY int    `json:"price_try"`
	LiveID   int64  `json:"live_product_id,omitempty"` // canlı shop.kanus.org ürün kimliği
	Plan     Plan   `json:"plan"`
}

// Units: kredi/istem paketi için istem sayısı; saatlik için 0.
func (o Offer) Units() int { return o.Plan.Prompts }

func (o Offer) UnitTRY() float64 {
	if o.Plan.Prompts == 0 {
		return float64(o.PriceTRY)
	}
	return float64(o.PriceTRY) / float64(o.Plan.Prompts)
}

var Catalog = []Offer{
	{SKU: "US-CREDIT-1", Title: "Kanus AI — 1 Model İsteği", PriceTRY: 149, LiveID: 15428080632036, Plan: Plan{Kind: "sub", Days: 90, Prompts: 1, Credits: 1}},
	{SKU: "US-CREDIT-5", Title: "Kanus AI — 5 Model İsteği", PriceTRY: 695, LiveID: 15428080664804, Plan: Plan{Kind: "sub", Days: 90, Prompts: 5, Credits: 5}},
	{SKU: "US-CREDIT-10", Title: "Kanus AI — 10 Model İsteği", PriceTRY: 1290, LiveID: 15428084170980, Plan: Plan{Kind: "sub", Days: 90, Prompts: 10, Credits: 10}},
	{SKU: "US-CREDIT-25", Title: "Kanus AI — 25 Model İsteği", PriceTRY: 2990, LiveID: 15432761344228, Plan: Plan{Kind: "sub", Days: 180, Prompts: 25, Credits: 25}},
	{SKU: "US-CREDIT-50", Title: "Kanus AI — 50 Model İsteği", PriceTRY: 5490, LiveID: 15432761409764, Plan: Plan{Kind: "sub", Days: 180, Prompts: 50, Credits: 50}},
	{SKU: "US-CREDIT-100", Title: "Kanus AI — 100 Model İsteği", PriceTRY: 9900, LiveID: 15432761442532, Plan: Plan{Kind: "sub", Days: 365, Prompts: 100, Credits: 100}},
	{SKU: "US-MONTHLY-200", Title: "Kanus AI — 1 Aylık Erişim / 200 İstem", PriceTRY: 18900, LiveID: 15424693010660, Plan: Plan{Kind: "sub", Days: 30, Prompts: 200, Credits: 200}},
	{SKU: "US-HOUR-1", Title: "Kanus AI — 1 Saatlik Kullanım", PriceTRY: 1990, LiveID: 15424693108964, Plan: Plan{Kind: "use", Days: 30, Seconds: 3600}},
}

func OfferBySKU(sku string) (Offer, bool) {
	for _, o := range Catalog {
		if o.SKU == sku {
			return o, true
		}
	}
	return Offer{}, false
}

// LivePlansJSON: canlı product_id -> plan (US_PLANS_JSON taslağı).
func LivePlansJSON() string {
	m := map[string]Plan{}
	for _, o := range Catalog {
		if o.LiveID != 0 {
			m[strconv.FormatInt(o.LiveID, 10)] = o.Plan
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// TRYText: 18900 -> "18.900 TL"
func TRYText(v int) string {
	s := strconv.Itoa(v)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += "."
		}
		out += string(c)
	}
	return out + " TL"
}

func (o Offer) PriceText() string {
	return TRYText(o.PriceTRY)
}
