package licensesrv

import (
	"encoding/json"
	"testing"
)

func TestCreditUnitPriceMonotonic(t *testing.T) {
	prev := 1e18
	n := 0
	for _, o := range Catalog {
		if o.Plan.Kind != "sub" {
			continue
		}
		if o.UnitTRY() >= prev {
			t.Fatalf("%s birim %.2f >= önceki %.2f", o.SKU, o.UnitTRY(), prev)
		}
		prev = o.UnitTRY()
		n++
	}
	if n != 7 {
		t.Fatal(n)
	}
}

func TestLivePlansJSON(t *testing.T) {
	var m map[string]Plan
	if err := json.Unmarshal([]byte(LivePlansJSON()), &m); err != nil || len(m) != 8 {
		t.Fatal(err, m)
	}
	if p := m["15424693010660"]; p.Kind != "sub" || p.Prompts != 200 || p.Days != 30 {
		t.Fatal(p)
	}
	if p := m["15424693108964"]; p.Kind != "use" || p.Seconds != 3600 {
		t.Fatal(p)
	}
	if p := m["15428084170980"]; p.Prompts != 10 || p.Credits != 10 {
		t.Fatal(p)
	}
	if TRYText(18900) != "18.900 TL" || TRYText(149) != "149 TL" {
		t.Fatal(TRYText(18900))
	}
}
