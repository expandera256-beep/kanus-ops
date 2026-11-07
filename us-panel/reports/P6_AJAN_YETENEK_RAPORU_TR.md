# §6 AI Ajan Yetenekleri: Boşluk Analizi ve Geliştirme Raporu
Tarih: 08.11.2026 (Europe/Istanbul). Taban: .11 (`r12_prev/term_11`, TREE_MANIFEST doğrulandı). Çalışma kopyası: `/workspace/us-panel-dev/p6/term`. Canlı servis çağrısı ve dağıtım yapılmadı.
Yama: `/workspace/us-panel-dev/patches/p6/p6-project-kind-symbols-on-11.diff` (sha256 5194340b…efe5e). Yama .11'e karşı; .12'ye rebase gerekecek.

## Başlangıç durumu: boşluk tablosu
| Yetenek | Durum (.11) | Kanıt |
|---|---|---|
| Büyük proje bağlam yönetimi | Kısmi | context_limits.go:17-72, long_context.go:23-93 (bölme, özet) |
| Dosya önceliklendirme | Var (yalnız yol adına göre) | project_map.go:13 relevantProjectPaths |
| Kod sembol arama | **Eksik** (yalnız ilk 6 imza) | project.go:685 signatures |
| Kalıcı görev belleği | Var | task_state.go:103/155 (MAC'li), task_history.go:87, journal.go:254 |
| Uzun görev devam ettirme | Var | long_checkpoint.go:14-20, task_state.go:191 runPersistentRepair |
| Güvenli dosya yazma | Var | file_tools.go:45 resolve, change_history.go:30 noSymlinkPath |
| Atomic değişiklikler | Var | file_tools.go:158, transaction_recovery.go:21-42 |
| Değişiklik önizlemesi | Var | proposal.go:131 proposalPreview |
| Rollback | Var | change_history.go:75 undoLastChange, transaction_recovery.go:42 |
| Gerçek terminal çıktısı | Var | safe_exec.go:157 safeRun; project_test_runner.go:24 |
| Proje türü tespiti | **Kısmi** (yalnız go.mod/package.json, yoksa Python) | project_test_runner.go:24-50 |
| Doğru test aracı seçimi | **Hatalı**: go.mod'suz Go / package.json'suz JS projesinde Python unittest seçiliyordu | aynı |
| Model hatası kurtarma | Kısmi (tek onarım tekrarı; geçici hata için yeniden deneme yok) | continuation.go:18-22, model.go:324, gateway.go:131 |
| Streaming yanıtları | Var | model.go:2330 streamIdleReader, 2359 readModelStreamPolicy |
| Kullanıcı başına çalışma alanı izolasyonu | Kısmi / NOT TESTED | Masaüstü EXE tek kullanıcılı, proje kökü bazlı; sunucu tarafında kullanıcı başına çalışma alanı yok |
| Ajan döngüsü (istem→analiz→keşif→plan→kod→değişiklik→derleme→test→düzeltme→yeniden test→çıktı) | Var; test adımı yukarıdaki hatadan etkileniyordu | project.go:275 runOrder, repair_loop.go |
| Önceki hatalar (model cevap vermedi / bağlam doldu / dosya gönderilemedi…) | Gerçek modelle NOT TESTED (canlı GPU yasak) | real_model_e2e_test.go canlı olmadan atlanıyor |

## Uygulananlar (§13: kök neden → test → düzeltme → test → regresyon)
1. **Proje türü tespiti + doğru test aracı.** Kök neden: runProjectTests, go.mod ve package.json yoksa her durumda Python unittest'e düşüyordu. Tekrar üreten testler (düzeltmeden önce FAIL): `TestP6RunnerDoesNotPickPythonForGoWithoutMod`, `TestP6RunnerDoesNotPickPythonForJSWithoutPackage`. Düzeltme: yeni `project_kind.go` `detectProjectKind` (önce manifest, sonra sınırlı ve symlink izlemeyen uzantı taraması), runner'da go-nomod/node-nopkg için açık mesaj. Test: `TestP6DetectProjectKind`. **DONE**
2. **Kod sembol arama.** Yeni `symbol_search.go` `findSymbols(root,name,limit)`: Go/Python/JS/TS/PHP tanımları (çağrı yerleri değil) file:line olarak döner. Symlink'leri, node_modules gibi atlanan ağaçları, gizli görünen yolları ve 1 MiB üstü dosyaları atlar, satırları redact eder. Testler: `TestP6FindSymbolDefinitions`, `TestP6FindSymbolsSkipsSymlinks`. **DONE (birim)**; model istemine/araç akışına bağlanmadı → **NOT DONE (entegrasyon)**.

## Testler (p6/term)
- go vet ./...: temiz. gofmt: yeni ve değişen dosyalar temiz.
- go test ./...: 400 PASS, 0 FAIL, 7 SKIP (canlı/ortam gerektirenler). EXIT=0
- go test -race ./...: hepsi ok.
- S2 regresyon: PASS=36 FAIL=0 BLOCKED=0 (results/20261108_005003).

## Kalan eksikler / NOT DONE
- Model hatası kurtarma (geçici 5xx veya zaman aşımında yeniden deneme): NOT DONE. Açık karar: yeniden deneme kredi rezervasyonunu çift harcatabilir; idempotent rezervasyon anahtarı ile mi yapılsın?
- Sembol aramanın ajan döngüsüne bağlanması: NOT DONE (prompt/araç protokolü değişikliği, PANEL DEV görüşü gerekiyor).
- Kullanıcı başına çalışma alanı izolasyonu (sunucu/çok kullanıcılı): NOT DONE, mimari karar gerekiyor.
- İçerik tabanlı dosya önceliklendirme: NOT DONE.
- §6'daki önceki hataların gerçek modelle tekrar testi: NOT TESTED (canlı GPU onayı gerekiyor).

## Gereken onaylar
Canlı model ve GPU ile e2e test için Demir'in onayı gerekiyor. Dağıtım yapılmadı.
