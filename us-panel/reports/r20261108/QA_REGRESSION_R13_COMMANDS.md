# §12 — R13 tek-komut yerel regresyon

Tarih: 08.11.2026 01:26 TSİ (Europe/Istanbul). Hazırlayan QA. **Local-only; remote request0; api-test authenticated chat/completions/GPU çağrısı yok.** Ürün/VPS kaynağı değiştirilmez, yalnız Go overlay ve sentetik temp fixturelar kullanılır; case-parallelism1 (max2 sınırını aşmaz).

## Komutlar
Linux/box, .13 **Go source module** hazır olduğunda (artifact-only dist dizini değil):
```bash
/workspace/us-panel/qa/audit12/regress_r13.sh --tree /absolute/path/to/r13/term
```
Cross-platform entry (native Windows makinesinde de bu Python girişini kullanın; Go+Python3 gerekir):
```bash
python3 /workspace/us-panel/qa/audit12/regress_r13.py --tree /absolute/path/to/r13/term
```
Yalnız plan/kernel ABI sorgusu, Go test/HTTP yok:
```bash
/workspace/us-panel/qa/audit12/regress_r13.sh --tree /absolute/path/to/r13/term --dry-run
```
Native Windows pathleri o makinenin yerel checkout/scratchına göre verilmelidir; box pathi Windows makinesinde var sayılmaz. ZoneInfo Europe/Istanbul desteği Python ortamında gerekir. Registered computer erişimi veya cross-host çalıştırma bu hazırlıkta yapılmadı.

## Beş kontrol ve dürüst sonuç semantiği
| Kontrol | Beklenen | Repro / sınır |
|---|---|---|
|F02 SSRF|FIXED / PASS|Mevcut audit ssrf: redirect_contacted/authorization_forwarded/project_body_forwarded=false. Loopback synthetic HTTP only.|
|F04 account limiter|İşlevsel PASS|>=12 karakter sentetik wrong-password10×401,11.429+Retry-After; doğru parola lock sırasında429; registered/unknown ve normalized email. Saf .11/.12 login handlerı, remote yok.|
|F07 lock DoS|Açıkken gerçek FAIL|10 unauthenticated failure sonrası doğru-password200 güvenlik beklentisi;429 görürse FAIL/exit1. Xfail/accepted-risk yapılmaz.|
|F01 Linux Landlock boundary|.13te sibling read denied + inside positive control|Kernel ABI read-only sorgulanır. ENOSYS/EOPNOTSUPP/EPERM/unknown-arch gibi unavailable/blocked durumda **SKIP+reason**, asla PASS değil. ABI varsa gerçek safeRun childı workspace içini okuyabilmeli, synthetic sibling okuması PermissionError ile reddedilmeli; bütün komutun çalışmaması PASS sayılamaz.|
|Windows Low IL write block|Native çalışabiliyorsa source-function probe; yoksa **NOT TESTED**|Child token integrity RID<=0x1000 + inside-read control + sibling write denied + parent fixture unchanged şart. Yalnız cross-compile PASSı runtime PASS değildir. Bu yazma testi okuma izolasyonu kanıtı değildir.|

**F04/F07 politika gerilimi:** aynı akışta F04 mevcut lock sözleşmesi doğru-password429 beklerken F07 bunun DoS sonucunu FAIL sayar. Aynı request için iki beklenti birlikte PASS olamaz. API recovery/step-up/risk politikası veya açık risk kararıyla sözleşmeyi netleştirmeli; QA suite bu çelişkiyi gizlemez. F07nin risk kabulü verilmedi.

Her invocation `reports/r20261108/qa-raw/r13-runs/<TSI-timestamp>-<unique>/` altında summary.json/kernel probe/raw Go JSONL/case linkleri üretir. İç `run.py` de ayrı timestamped dizin açar, eski kanıt üzerine yazılmaz. Go test skip/zero-tests exit0 **PASS yapılmaz**. Timeout BLOCKED; build/harness errors FAIL kanıtıyla görünür.
Exit0 yalnız tüm requested kontroller PASS (veya açık DRY RUN); exit1 herhangi FAIL; exit2 FAIL yok ama SKIP/NOT TESTED/BLOCKED nedeniyle **PARTIAL**. Normal tek platformda karşı platform NOT TESTED kalabilir; bu tam release PASSı değildir.

## Gerçek baseline doğrulaması (.13 öncesi)
- Manifest209/209 doğrulanmış QA .12 source copy: `/workspace/us-panel/qa/build/qa-r12-odheei69/term`.
- İkinci koşum .11 limiter dosyaları manifest-eşit, **F02 owner patchi uygulanmış** QA kopyası: `/workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term`; saf .11 artifactı diye etiketlenmedi.
```bash
/workspace/us-panel/qa/audit12/regress_r13.sh --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term
/workspace/us-panel/qa/audit12/regress_r13.sh --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term
```
İkisi de gerçekten koşuldu: **2 PASS /2 FAIL /1 NOT TESTED**, exit1. F02/F04 PASS; F07 doğru-password429 ile FAIL; F01 Linux gerçek sibling read ile FAIL. Kernel **Linux6.12.113+, Landlock ABI6**: destek var, fakat .11/.12 safeRun henüz izolasyon sağlamıyor. F01 logunda **inside_control=true, sibling_read_allowed=true, sibling_read_denied=false**; compile/fixture hatası değil. .13 uygulaması yeniden koşulmadan PASS iddiası yok.

Kanıtlar:
- `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r13-baseline-validation.json` (iki gerçek summary yolu).
- `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r13-suite-current-r12.log`, `r13-suite-current-patched-r11.log`.
- Kernel-unsupported/EPERM/unknown-arch ve suite-SKIP guardları **6 unit test PASS**, `r13-probe-unit-final.log`.
- Windows fixture embedded Python syntax + GOOS=windows cross-compile PASS, `r13-windows-compile-only.log`; **native runtime NOT TESTED**.
- .13 dist varlığı bu belgenin hazırlanma anında: **False**. Varlık tek başına hazır release/manifest/işlev kabulü değildir; bu görevde .13 gerçek source/runtime kabulü yapılmadı.

## F01 platform kararı / provenance
**Windows: ACCEPTED-RISK — Demir'in08.11.2026 01:15 TSİ kararı, KAN.US üzerinden iletildi (QA doğrudan teyit etmedi); sürüm notuna yazılacak. Linux: OPEN, .13te Landlock ile doğrulanacak.** Kaynak: koordinatörün08.11.2026 01:20 TSİ steering aktarımı.
KAN.USa göre .13 Windows yazmaları Low IL ile engellenecek/engelleniyor: **to be verified**, native test olmadan PASS değil. Low IL yazma engeli sibling-read izolasyonu sağlamış diye sunulmaz. Sürüm notunu product/release owner yazmalı; QA product release-notes dosyasını değiştirmedi. Kabul yalnız F01 Windows riskine ilişkindir; Linux, F07 veya PANEL DEV'in diğer Windows TOCTOU/.12 hook bulgularını kapatmaz.
