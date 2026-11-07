# FINDING-01 — Proje yürütmesi sibling workspacei aynı OS kimliğiyle okuyabiliyor

- Önem: **HIGH — koşullu multi-user/untrusted-code tasarım açığı**. Trusted single-user desktopta başka müşteri verisi sızdığı iddiası değildir.
- Durum: **OPEN / FAIL** (§12 Dosya izolasyonu; §6 kullanıcı başına workspace izolasyonu gereksinimi).
- Yönlendirme: **PANEL DEV** birincil; shared-host sandbox/UID altyapısı için **HOST/SRE**. Bunlar bileşene göre önerilen routingdir; verilen prompt bot-isim/sahiplik tablosu içermiyor, koordinatör doğrulamalı.
- Kaynak: `work2/kiyas/term/safe_exec.go:17–19` sandbox olmadığını açıkça söyler;`:176–208` yalnız cwd confinement;`:223–258` host interpreter, aynı UID/environment HOME; child için mount namespace/chroot/user boundary yok. `file_tools.go:44–105` API yolu guardı çocuk processin doğrudan OS dosya erişimine uygulanmaz.

## Repro (tamamı yerel/sentetik)
```bash
python3 /workspace/us-panel/qa/audit12/run.py --case file_isolation
```
Test: `qa/audit12/panel_audit_test.go:TestQAAuditExecSiblingIsolation`. Temp parent altında user-a ve user-b oluşturulur; user-b/synthetic.txt sadece test canarysi taşır. user-a içindeki Python proje scripti product `safeRun` ile çalışır. HOME, /workspace/secrets veya gerçek kullanıcı dosyası okunmaz.

Beklenen: Untrusted multi-user execution boundary sibling dosyayı okumayı reddeder.
Gerçek: `synthetic_sibling_read=true state=success`; test FAIL/exit1. Kanıt: `qa-raw/file_isolation.log`, toplu `qa-raw/panel-audit.log:7–10`.

## Kök neden / etki
Cwd/argv/env/resource limit OS filesystem access policy değildir. Aynı servis UIDsiyle çalışan izinli Python/Go/node kodu sibling rootu doğrudan açabilir. Etki yalnız aynı-UID/shared-filesystem ve untrusted-code senaryosu için doğrulandı; başka OS kullanıcısı/root escalation veya production müşteri sızıntısı gösterilmedi. Bu davranış yorumlarda bilindiği için “yeni gizli exploit” olarak sunulmuyor; multi-user kabul kriterini karşılamayan somut gap.

## Önerilen düzeltme / §13 handoff
PANEL DEV+HOST untrusted yürütmeyi per-user UID/container/microVM sandboxa taşımalı; sadece ilgili workspace mountu, secrets/HOME erişimsiz, uygun network policy. Desktop trusted-project modunu ayrı ve açık tanımlamalı. File API guardlarını korumalı. QA ürün kaynağını değiştirmedi. Owner fix sonrası aynı negative test PASS olmalı; command/path/env regresyonları ve go test/-race/vet yeniden koşulmalı. Fix ve fix sonrası retest **NOT DONE / owner bekleniyor**.

## HOST kanıtı ve güncel tekrar
`HOST_EVIDENCE_QA_20261108.md:31–32`: hook servisleri ayrı uslic/uslic-test UIDleriyle çalışıyor; sunucuda panelin untrusted kod yürütmesi henüz yok (HOST-provided). Bu nedenle bulgu production hooklar arası okuma olarak yorumlanmaz; aynı-UID multi-user/untrusted panel gereksinimine koşulludur. Güncel .12 kaynakta `run.py --case file_isolation` yeniden **FAIL** (`qa-raw/file_isolation.log`). UID/container/mount politikası PANEL DEV + HOST/SRE tasarım kararı bekler.

## Bağımsız patched .11 doğrulaması
205-file manifest doğrulanmış .11 QA kopyasında owner patchi sonrası aynı test **FAIL / synthetic_sibling_read=true**. Ownerın iki ucuz guardı PASS; bulguyu kapatmaz. `qa-raw/patch-r11-verify/verification-results.json`. Repro:
```bash
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case file_isolation
```
`go list -deps` ile safe_exec/file_tools kök package mainin hook dependency graphında olmadığı, Linux+Windows desktop GoFiles içinde olduğu doğrulandı: **EXE-only**, hook binaryye ait bir cross-user isolation açığı olarak raporlanmaz. Mimari multi-user tasarım kararı halen PANEL DEV + HOST/SREye aittir.

## 08.11.2026 01:17 TSİ (Europe/Istanbul) — .12 yeniden repro
PANEL DEV raporu `PANEL_S2_REVIEW_R12_20261108.md:161–171`: mimari açık, guardlar sandbox değil. Manifest209/209 doğrulanmış QA .12 scratchında `run.py --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term --case file_isolation` yeniden **FAIL / synthetic_sibling_read=true**. `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r12-file-isolation.log`. safe_exec.go:17–19,176–258; aynı-UID OS boundary yok. **OPEN**, HIGH koşullu, PANEL DEV + HOST/SRE. Windowsun native exploit/runtime kabulü bu testten çıkarılmaz; hook dependency kapsamı değişmedi.

## 08.11.2026 01:26 TSİ (Europe/Istanbul) — Platforma göre risk kararı / R13 testi
**Windows: ACCEPTED-RISK — Demir'in 08.11.2026 01:15 TSİ kararı, KAN.US üzerinden iletildi (QA doğrudan teyit etmedi); sürüm notuna yazılacak. Linux: OPEN, .13'te Landlock ile doğrulanacak.** Kaynak: koordinatör steering08.11.2026 01:20 TSİ; doğrudan kullanıcı teyidi veya QA native kanıtı diye sunulmaz.
KAN.US: Windows writes blocked via Low IL in .13 — **to be verified**. Yazma engeli okuma sandboxı değildir. QA native Windows testi bu boxta **NOT TESTED**; fixture cross-compile PASS sadece derlenebilirliktir. Product/release owner sürüm notunu eklemeli; QA product notlarını değiştirmedi. Bu karar F07/diğer Windows TOCTOU bulgularını kapsamaz.
Linux kernel6.12.113+ Landlock ABI6 mevcut. Yeni R13 testi inside-positive-control ister ve sibling read denial doğrular; unsupported kernel SKIP+reason, false PASS yok. .12 ve F02-yamalı .11 baseline suiteinde Linux **FAIL**: inside_control=true/sibling_read_allowed=true. `.13` gerçek retesti bekler. Komutlar `reports/r20261108/QA_REGRESSION_R13_COMMANDS.md`; kanıt `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r13-baseline-validation.json`.
