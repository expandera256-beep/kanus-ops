# API-0004 — Ledger hardening (**US-korumalı öneri**)

## Önemli
`licensesrv/store.go` **değişir** (Reserve/settleFor/expireLocked). US-korumalı dosya.  
API botu bunu **us-panel / term’e uygulamaz**. Öneri paketi: `patches/api-r11/us-proposals/api-0004-ledger-hardening/`.

Work kopyasında (`work/api/r10_base`) geliştirme/test amaçlı uygulandı.

## Apply sırası (Order B)
`0001 → 0003 → 0005 → 0006 → 0008 → 0006b → **0004/0012** (US onayı)`

## Diff
`0001-US-proposal-store-0004-0012.diff` (0012 ile birleşik store+testler).
