# FINDING-02 — repairLoop default HTTP client cross-port307 redirectini izliyor

- Önem: **MEDIUM — latent/helper scope**.
- Durum: **FIXED — bağımsız patched .11 üzerinde doğrulandı; .12 merge/release kabulü pending**. Mutable work2/.12 targeted testi de PASS; yayımlanmış release/deploy eşliği değil.
- Yönlendirme: **PANEL DEV** (bileşen temelli öneri; promptta bot routing tablosu yok).
- Kaynak: `work2/kiyas/term/repair_loop.go:262–274`; Authorization`:267–268`; default client`:270–273` CheckRedirect içermez. Model request proje gövdesini`:229–260`da oluşturur.

## Repro
```bash
python3 /workspace/us-panel/qa/audit12/run.py --case ssrf
```
İki yalnız-loopback httptest server; origin307 ile aynı hostnamein başka portundaki sinke yönlendirir. Sentetik proje+sentetik token kullanılır. Metadata/özel ağ/gerçek API/token yok.

Beklenen: Diğer panel/model clientları gibi redirect reddi; sinke istek, project body veya credential gitmemeli.
Gerçek: `redirect_contacted=true authorization_forwarded=true project_body_forwarded=true`; test FAIL/exit1. Tarihsel kanıt: `qa-raw/pre-host-source-retest/ssrf.log`, `panel-audit.log:11–14`. Güncel ssrf.log fix sonrası PASS.

## Kök neden / kapsam
Go default redirect policy aynı hostname altında port değişimini credential güven sınırı saymaz;307 POST gövdesini korur. Helper defaultu no-redirect policyyi uygulamıyor. Mevcut `model_proxy.go:57`, `gateway.go:112` ve GPU clientları redirectleri reddediyor; bu finding onların kontrolünü boşa çıkardığı iddiası değil.

**Reachability sınırı:** Mevcut dar call-site taramasında repairLoop kurucuları test dosyalarında bulundu (`qa-repair-call-sites.txt`); aktif GUI/public hook routeuna bağlandığı gösterilemedi. Bu nedenle canlı API SSRF/credential theft veya gerçek metadata okuması iddia edilmiyor. Unused/helper kod yeniden ürüne bağlanmadan önce düzeltilmeli.

## Önerilen fix / §13
Default ve caller-injected client kopyasında CheckRedirect=ErrUseLastResponse; aynı origin/scheme/port bağını koru, HTTPS downgrade reddet. Konfigüre edilmiş endpoint güven modelini açık tut. Testi düzeltmeden xfail yapma. QA yalnız qa/ reproducer yazdı; product fix ve owner sonrası retest **NOT DONE**. Fix sonrası aynı komut PASS; ilgili repair-loop ve gateway regresyonları korunmalı.

## Güncel kaynak düzeltmesi / QA retesti (08.11.2026 TSİ)
Paylaşılan ağaçta owner değişikliği görüldü: `release.go:11` artık 2026.11.06.12. `repair_loop.go:271–281` caller clientını kopyalar veya default client oluşturur; **her iki yolda CheckRedirect=ErrUseLastResponse** uygular. QA ürün kodu değiştirmedi. Yukarıdaki aynı otomatik test yeniden **PASS**, `qa-raw/ssrf.log`; güncel api_tokens testi de PASS. İlk fail kanıtı arşivde korunur. Önceki NOT DONE ifadesi tarihsel durumdur; şimdi fix+targeted-retest tamam. Productionda halen pre-.11 binary HOST kanıtı olduğundan source fixin live olduğu iddia edilmez.

## Bağımsız .11 patch doğrulaması (08.11.2026 00:58 TSİ)
Patch SHA `c31ac8b8b015184d7911c4eadf0abde6155f41cefa3da24da25fd5badabf40dc` doğrulandı. `dist/2026.11.06.11/TREE_MANIFEST.sha256`205 dosyası birebir kontrol edilerek USnin term.orig ağacından QA scratchına kopyalandı; yalnız kopyaya patch uygulandı. Orijinal kaynak205 hash sonunda değişmemişti.
QA testi önce **FAIL true/true/true**, sonra **PASS false/false/false** (redirect_contacted/authorization_forwarded/project_body_forwarded). Owner default/caller client guardları ve US repro scripti (yalnız çıktı yolları QAya yönlendirilerek) de doğrulandı.
Kanıt `qa-raw/patch-r11-verify/verification-results.json`; her koşumun ayrı RUN_DIRi ilgili baseline/patched logunda. Güncel repro:
```bash
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case ssrf
```
Bağımsız Go deps: hookta yalnız kiyas/licensesrv + kiyas/cmd/us-license-hook var; repair_loop kök desktop package main yalnız EXE yolundadır. Linux/Windows GoFiles doğrulaması mevcut. Hook binary bu helperı içermez; .11 hook deployu açısından F02 engel değildir. US/PANEL DEV kaynakları değiştirilmedi.

## 08.11.2026 01:17 TSİ (Europe/Istanbul) — .12 kaynak/artifact entegrasyonu bağımsız doğrulandı
PANEL DEV cevabı `/workspace/us-panel-dev/PANEL_S2_REVIEW_R12_20261108.md:165–175` rapor kanıtı olarak okundu. `dist/2026.11.06.12/TREE_MANIFEST.sha256`209/209, SHA256SUMS5/5 yerelde doğrulandı; binary çalıştırılmadı. Pinned/v12den yalnız QA scratchına kopyalandı.
`python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term --case ssrf` **PASS**, redirect_contacted/authorization_forwarded/project_body_forwarded **false/false/false**. Kanıt `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r12-ssrf.log`, `r12-artifact-sha-verification.json`. Önceki .12-merge-pending durumunun source/artifact entegrasyon kısmı artık doğrulanmıştır; native EXE/live deploy kabulü değildir. **F02 FIXED — .12 source verified.**
