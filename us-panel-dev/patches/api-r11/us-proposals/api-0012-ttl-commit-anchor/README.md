# US önerisi — api-0012 TTL commit anchor

**Demir kararı: seçenek B.** US-korumalı `licensesrv/store.go` değişir.

## Davranış
- Commit, rezervasyon `Expires` + grace içinde kabul.
- Varsayılan grace **600s**; env `US_CREDIT_COMMIT_GRACE_SECONDS`.
- **Üst sınır 900s** (15 dk, KAN.US): 900 üstü → **900'e kırpılır** (varsayılana düşmez).
- Geçersiz/negatif/parse edilemeyen → **600**. `0` = legacy anında expire.
- Grace bitince expire reserved serbest (**debit yok**); geç commit çift düşmez.

## Testler
`licensesrv/ttl_commit_anchor_test.go` (ayrı dosya).  
FAIL_before ham çıktı: `patches/api-0012-ttl-commit-anchor/logs/FAIL_before_grace.txt` (EXIT 1).

## Diff
`store-and-tests.diff` (0004 guard’ları + 0012 grace; birleşik).
