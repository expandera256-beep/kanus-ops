# US önerisi — api-0004 ledger hardening

**US-korumalı:** `licensesrv/store.go` değişir. us-panel / term ağacına API botu **yazmaz**.

## Ne
- `applyReservedDelta`: Reserved &lt; 0 olamaz; ihlalde hata, panic yok.
- settle/expire yollarında guard; happy-path Reserve→Settle sonucu aynı.
- Testler: `licensesrv/ledger_hardening_test.go` (ayrı dosya).

## Risk
Corrupt ledger’da artık hata (önceden sessiz negatif). Üretim happy path değişmez.

## Uygulama
`store-and-tests.diff` (0004+0012 birleşik; 0012 ile aynı dosya). Önce Order B API yamaları, sonra bu diff.
