package main

import (
	"encoding/json"
	"fmt"
	"kiyas/licensesrv"
	"os"
)

func main() {
	b, _ := json.MarshalIndent(map[string]any{"currency": "TRY", "note": "TEST mağazası içindir; canlı fiyatlar değiştirilmedi. Fiyatlar KDV dahil.", "products": licensesrv.Catalog}, "", "  ")
	os.WriteFile(os.Args[1]+".json", b, 0644)
	f, _ := os.Create(os.Args[1] + ".csv")
	defer f.Close()
	fmt.Fprintln(f, "Handle,Title,Vendor,Type,Published,Variant SKU,Variant Price,Variant Requires Shipping,Variant Taxable,Status")
	for _, o := range licensesrv.Catalog {
		fmt.Fprintf(f, "%s,\"%s\",Kanus,Dijital,TRUE,%s,%d.00,FALSE,TRUE,draft\n", o.SKU, o.Title, o.SKU, o.PriceTRY)
	}
	fmt.Println(licensesrv.LivePlansJSON())
	for _, o := range licensesrv.Catalog {
		fmt.Printf("%s %d %.2f\n", o.SKU, o.PriceTRY, o.UnitTRY())
	}
}
