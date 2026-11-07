package main

import (
	"kiyas/licensesrv"
	"strings"
	"testing"
	"time"
)

// An authenticated Ed25519 FX source is not present in .12. No price may
// retain the previous hardcoded conversion while that prerequisite is absent.
func TestPost12PricingTLOnlyWithoutVerifiedFX(t *testing.T) {
	text := pricingSummary()
	if strings.Contains(text, "USD") || strings.Contains(text, "≈") || !strings.Contains(text, "18.900 TL") || !strings.Contains(text, "9.900 TL") {
		t.Fatalf("unverified FX or missing TRY: %s", text)
	}
	for _, o := range licensesrv.Catalog {
		if o.PriceText() != licensesrv.TRYText(o.PriceTRY) {
			t.Fatalf("offer includes unverified FX: %s", o.SKU)
		}
	}
}
func TestPost12LegacyUSDNotDisplayedWithoutVerifiedFX(t *testing.T) {
	if got := priceNote(licenseDoc{USD: 99}); got != "" {
		t.Fatalf("legacy dollar approximation shown without verified rate: %s", got)
	}
	if got := priceNote(licenseDoc{TRY: 149, USD: 99}); got != " 149 TL" {
		t.Fatal(got)
	}
}
func TestPost12LicenseMintOmitsUnverifiedDollarPrice(t *testing.T) {
	t.Setenv("US_LICENSE_SECRET", "post12-test-only")
	token, err := mintLicense("sub", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	doc, err := parseLicense(token, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if doc.USD != 0 || doc.TRY != monthlyOffer().PriceTRY {
		t.Fatalf("minted unverified dollar amount: USD=%d TRY=%d", doc.USD, doc.TRY)
	}
}
