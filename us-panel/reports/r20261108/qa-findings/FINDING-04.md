# FINDING-04 — IP-only giriş limitinin rotating-source design gapı

- Önem: **MEDIUM — DESIGN GAP, FAIL değil**.
- Durum: **SOURCE MITIGATED / LIVE NOT VERIFIED**. Yeni normalized-account limiter yerelde PASS; eski nginx/IP katmanı tek başına yeterli değil.
- Yönlendirme: **API** account limiterin dağıtımı/kabulü; **HOST** nginx/actual-IP kanıtı. Promptta bot ownership tablosu yok; önerilen routing koordinatörce doğrulanmalı.
- Kaynak: `accounts.go:257–290`, özellikle`:268` group+IP;`:272`15dk;`:344–351` login10. Ek account katmanı `login_limiter.go:10–16,30–35,40–80`; entegrasyon `accounts.go:362–373,466–468,484`.

## Repro / kanıt
```bash
python3 /workspace/us-panel/qa/audit12/run.py --case rate_limit
```
`TestQAAuditIPOnlyRateDesign`11 gerçek-network olmayan synthetic RemoteAddr/X-Real-IP fixtureı ile doğrudan private IP rate fonksiyonunu çağırır:11 ayrı bucket,11 accepted,per-IP cap10. Test bu TASARIM davranışını doğrular ve PASSdır. Aynı koşumda mevcut `TestAccountLimiterRotatingIPs` normalize aynı hesaba farklı IPlerden ilk10 yanlış login401,11.429+Retry-After ile ek account katmanının yerel mitigasyonunu doğrular. Public proxynin client XFFsiyle aşılabildiği iddia edilmez.

HOST eski salt-okunur evidence: `reports/qa-20261106/ratelimit-evidence.txt:8–36,43–78,89–110,136–149`: nginx5r/m burst2 `$binary_remote_addr`, exact login limit aktif;06.11.2026 09:26–09:31 TSİ penceresinde15×400+1×200,0×429; QA egress140.248.50.x rotasyonu. 15×400 kısa-invalid parola yeni account fail counterı testi değildir.

Beklenen güvenlik özelliği: Aynı normalized hesapta gerçek kaynak IPden bağımsız failure budget.
Gerçek: Legacy katman IP-only; yeni ek kaynak bunu gideriyor ve yerelde PASS. **Yeni host/.11 binaryde etkin olduğu bu tur doğrulanmadı.**

Öneri: IP layeri kaldırmadan account layeri deploy/attest et; geçerli uzunlukta yanlış parolayla controlled multi-egress11-attempt kabulünü ve HOST log korelasyonunu yap. Yeni canlı burst bu auditte koşulmadı; sadece8 sıralı negatif gate kontrolü yapıldı. Mevcut rate behaviorunu bozuk nginx FAIL olarak raporlama.

## HOST kanıtı sınırı
`HOST_EVIDENCE_QA_20261108.md:5,7–13` çalışan binarynin pre-.11 olduğunu ve proxy IPlerinin değiştiğini destekler. HOSTun00:45 TSİ login429 örneği bu QA workerın isteği değildir; tek başına account-vs-IP enforcement ayrımını kanıtlamaz. Güncel .12 kaynakta aynı account/farklıIP otomatik testleri tekrar PASS; canlı yeni account-limiter kabulü yine ayrı owner/deploy kanıtı bekler. F04 **DESIGN GAP**, FAIL değil.

## 08.11.2026 01:17 TSİ (Europe/Istanbul) — API cevabı ve manifest-doğrulanmış .11 retest
Kaynak `/workspace/us-panel-dev/patches/api-r11/fix-grace-overflow-on-11/README.md:47–66`; bağımsız .11 kaynaklar `login_limiter.go:10–107`, `accounts.go:48–52,257–290,344–374,467,484`. İlgili4 dosya .11 TREE_MANIFEST ile eşit.
`python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case rate_limit` **PASS** (`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r11-rate-limit.log`): sentetik valid-email ve >=12-karakter parolalarla ilk10 wrong401,11.429+Retry-After; kilitliyken doğru-password429; registered/unknown aynı; başarı reset; fixed-window expiry.
Account katmanı **yalnız login**, normalized `ToLower(TrimSpace)` email + exact mail.ParseAddress; Gmail dot/+ normalizasyonu yok. Varsayılan10failure/900s, env aralıkları1–1000/10–86400. İlk failuredan sabit pencere; kilitli probe uzatmaz. Süreç içi max10000 map, oldest eviction; restart/instance paylaşımı yok.
IP katmanı group:IP: register5/login10/reset5/verify10 (15dk); redeem/balance/requests/workers group boş, redeem session ister. Saf .11 main.go:108 yalnız /account/ bağlar; /v1/account alias API additive stackte aynı handlera yönlenir, pure .11de mevcut değildir. Parola<12/>72 veya invalid email400; account failure sayılmaz.
**API cevabındaki “IP429 Retry-After yok” iddiası yerelde çürütüldü:** accounts.go:317/backpressure.go:29–37 wrapper **Retry-After:5** ekliyor. Kalan900sn hesabı değil; ancak header yok diye finding açılmadı. İlk fixture beklenti hatası ayrı korundu, retest doğru5 ile PASS.
F04 hâlâ tarihsel **IP-only DESIGN GAP**, FAIL değil; dağıtıkIP login mitigationı .11de işlevsel. Yeni, bağımsız **F07 MEDIUM account-lock DoS** nedeniyle güncel rate-limiting güvenlik maddesi FAILa döndü; bu eski F04 kararını geriye dönük değiştirmez. Canlı limiter etkinliği bu tur test edilmedi; remote request0.
