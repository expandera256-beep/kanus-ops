# API — grace taşma düzeltmesi + .11 rebase incelemesi + QA F04 (08.11.2026, TSİ)

KANUS_API_REPORT.md için bu turun özeti. `us-panel/` ağacına ve US'in `rebased-on-11/` dosyalarına yazılmadı. Deploy yok. Sır ya da env değeri yazdırılmadı. Bütün sonuçlar gerçek koşulardan; loglar `logs/` altında.

## Durum tablosu
| Madde | Durum | Kanıt |
|---|---|---|
| 1 R3-A01 taşma (r10_base) | **PASS**: düzeltmeden önce 13/3 FAIL (EXIT=1), sonra 16/16 PASS | logs/r10_FAIL_before.txt, logs/r10_PASS_after.txt |
| 1a r10 vet / race / gofmt | **PASS**: VET_EXIT=0, RACE_EXIT=0 (`./...`), gofmt licensesrv 0 satır | logs/r10_vet.txt, r10_race.txt, r10_gofmt_licensesrv.txt, r10_ttl_tests_after.txt (*) |
| 1a yamalar yeniden üretildi | **PASS**: B sırası 0001→0003→0005→0006→0008→0006b→US-proposal 7/7 EXIT=0, ağaç r10_base HEAD ile birebir | logs/r10_chain_apply.txt, logs/sha256_r10.txt |
| 1b .11 + US yığını | **PASS**: .11 manifest OK; 0004/0012, UYUM v4 16/16 ve additive EXIT=0; ağaç US'in FULL_STACK sonucuyla birebir | logs/on11_apply_us_stack.txt |
| 1b taşma düzeltmesi .11'de | **PASS**: önce 13/3 FAIL (EXIT=1), sonra 16/16 PASS | logs/on11_FAIL_before.txt, on11_PASS_after.txt |
| 1b vet / race / gofmt (.11 yığını + düzeltme) | **PASS**: VET_EXIT=0; `go test -race ./...` 5 paketin hepsi ok, RACE_EXIT=0; değişen dosyalarda gofmt 0 satır; ağaç genelinde 30 dosya, .11 tabanıyla aynı (önceden var) | logs/on11_vet.txt, on11_race.txt, on11_gofmt_changed.txt, on11_gofmt_tree.txt |
| 1c US rebase semantiği | **OK** (1 SORUN: zayıf test, yeni testle kapatıldı) | aşağıda |
| 1d gofmt api_keys_test.go:243 | **PASS**: 0006b yamasında `(24*time.Hour)` düzeltildi; zincir uygulanınca gofmt temiz | logs/r10_chain_apply.txt, logs/old_0006b.patch.bak |
| 2 QA FINDING-04 | **Yanıtlandı** (koddan; canlı repro koşulmadı), yerel birim testi 3/3 PASS | logs/qa_f04_account_limiter_unit.txt |

(*) r10 tam `-race ./...` koşusu taşma düzeltmesinden sonra yapıldı. Kilit testi (`ttl_grace_lock_test.go`) ondan sonra eklendi; onun için licensesrv'de `-race -run 'TestTTL|TestCommitGrace'` 7/7 PASS. .11 tam race koşusu her iki testi de içeriyor.

## Dosyalar
- `fix-grace-overflow-on-11.diff`: US'in .11 yığınının (FULL_STACK) üzerine gelen küçük ek diff. `store.go` içinde yalnız commitGraceFromEnv değişiyor, yanına 2 yeni test dosyası geliyor. `git apply --check` US FULL_STACK üzerinde OK. sha256 `df1ae528483634405cb6b21f3c3a0dd3481db9d2f808b2e10f7991d994c07c28`
- `r10-grace-overflow-fix.diff`: aynı değişiklik r10_base'e karşı. sha256 `84a42e17b55f94f76dbf25122d62a13d307e93bea495f5ad1145a4ec04b951d3`
- Yeniden üretilen dosyalar (tam liste logs/sha256_r10.txt'de):
  - `api-0012-ttl-commit-anchor/0001-API-0012-clamp-grace-seconds-as-int64-before-Duratio.patch`: `8ba71f1e907e9152fa62d4ebe41bdad422c9179292a2c66e1e4b5cae51e87f7b`
  - `api-0012-ttl-commit-anchor/0002-API-0012-test-reserved-stays-locked-during-grace-rel.patch`: `1d2aac59a2f07c296c2c3715d5cba30248bd87e40575a30b8d79a8d2fea84018`
  - `us-proposals/api-0004-ledger-hardening/store-and-tests.diff`, `us-proposals/api-0012-ttl-commit-anchor/store-and-tests.diff` ve `api-0012-ttl-commit-anchor/0001-US-proposal-store-0004-0012.diff` (üçü byte-eş): `ee9aee17b68cc82ffbfe944e78380f525e3dfdb4f5abad8feaab901b4bc604cf`
  - `api-0006-api-keys/0001-API-0006b-…patch` (gofmt): `75dec7b4eaf50277d6979a6ddc43a1d61c6d6dc78ff255b07a4df607982f03b0`
- Eski sürümlerin yedekleri: logs/old_*.bak

## Düzeltme (R3-A01)
Değer `strconv.ParseInt(raw,10,64)` ile int64 saniye olarak okunuyor. Kurallar:
- Parse hatası ya da negatif değer → 600
- 900'den büyük değer → 900
- Ancak bu kontrollerden sonra `time.Duration(n)*time.Second` hesaplanıyor.

0/-0 → 0 davranışı (eski TTL'e dönüş) korundu; PANEL'in beklentisi de bu. Tablo testi PANEL'in 16 durumunu birebir içeriyor (`grace_overflow_test.go`).

## 1c — US rebase incelemesi
- **OK**: store.go hunk #3. `s.commitGrace = commitGraceFromEnv()`, LicenseOwners init'inden sonra ve `good = true`'dan hemen önce duruyor (OpenStore'un başarılı yolu). Accounts ve LicenseOwners nil-init blokları korunmuş. Önerimle US'in diff'i arasındaki tek fark bu bağlam satırı (UsageEvents → LicenseOwners) ve "new file mode" başlıkları. Eklenen/silinen satırlar birebir aynı, kayıp ya da fazla hunk yok (logs/us_rebase_vs_api_proposal_content.diff).
- **OK**: ledger guard'ları. Reserve, settleFor ve expireLocked'daki 4 çağrıda applyReservedDelta var, Credits<0 geri alma yolu da korunmuş (store.go:263-359).
- **OK**: grace varsayılanı 600, üst sınır 900. Ancak US'in kopyasında R3-A01 taşması da duruyordu (on11_FAIL_before); bu ek diff onu kapatıyor.
- **OK**: main.go. `"kiyas/apiv1"` importu :15'te. `apiv1.MountAccountAlias(mux, accounts)` :132'de, yani UYUM v4 `privacy.Mount(...)` bloğundan (:126-131) sonra ve `/healthz`'den (:134) önce. UYUM Mount'una dokunulmamış.
- **OK**: additive diff. Değişen satırlar orijinal r11 additive yamasıyla birebir aynı; tek fark format-patch imza satırı (logs/us_additive_vs_original_changed_lines.diff).
- **OK**: 0012 testleri .11 tabanında geçiyor: AcceptsCommitWithinGrace, RejectsAfterGraceAndNoDoubleDebit, LegacyZeroGrace, IdempotentAndConcurrent, DefaultCapAndInvalid ve 3 ledger_hardening testi (logs/on11_0012_ledger_tests.txt).
- **SORUN (giderildi)**: "grace sırasında reserved kilitli" durumu hiçbir testte gerçekten doğrulanmıyordu. RejectsAfterGrace içindeki `if l.Reserved != 1 {}` bloğu boş. Bunun için `TestTTLGraceKeepsReservedLockedThenReleases` eklendi: grace içindeki sweep'ten sonra reserved=2 olmalı ve kredi düşmemeli; grace sonrasında release ve geç commit reddi bekleniyor; final credits=498. Test PASS. Mutasyonla da doğrulandı: commitGrace=0 yapılınca test FAIL veriyor ("got reserved=1"), log: logs/on11_grace_lock_MUTATION_grace0.txt.

## 2 — QA FINDING-04: hesap limiter'ının kapsamı
Kaynak: .11 = `r12_prev/term_11` (dist .11 manifest'i doğrulandı). r10 = `work/api/r10_base`.

| Konu | Koddan bulgu |
|---|---|
| Kapsam (hesap katmanı) | Yalnız `/account/login`. blocked: accounts.go:369-374, fail: :467, reset: :484. register, redeem ve parola sıfırlama hesap limiter'ında **yok**. |
| IP katmanı | rate() accounts.go:257-298, gruplar :345-357. register 5, login 10, reset (request+confirm) 5, verify 10. `/account/redeem`, balance, requests ve workers için **limit yok** (group=""); redeem oturum token'ı istiyor (:571). |
| `/v1/account/*` | Aynı handler'a `/v1` öneki atılarak gidiyor: .11 yığınında apiv1/version.go:18-25, r10'da main.go:105 StripV1AccountPrefix. Kovalar ortak, ayrı bir bütçe yok. Saf .11'de (additive yama olmadan) `/v1/account` hiç yok. |
| Anahtar | normalizedEmail(): `ToLower(TrimSpace)` + `mail.ParseAddress` tam eşleşme, en fazla 254 karakter (accounts.go:48-52). Gmail nokta/+etiket normalize edilmiyor; `a.b+x@` ayrı bir kova. IP katmanının anahtarı `group:ip`; X-Real-IP'ye yalnız loopback'ten gelen isteklerde güveniliyor (:262-266). |
| Pencere/limit | Varsayılan 10 başarısız deneme / 900 sn. Env ile değişebiliyor: `US_LOGIN_ACCOUNT_LIMIT` 1-1000, `US_LOGIN_ACCOUNT_WINDOW_SECONDS` 10-86400 (login_limiter.go:30-37). |
| Pencere türü | **Sabit pencere**, ilk başarısızlıktan başlıyor (login_limiter.go:61-76, start bir kez set ediliyor). Kilitliyken gelen denemeler sayılmıyor ve süreyi uzatmıyor. |
| Başarılı giriş | Sayaç sıfırlanıyor (:78-82, accounts.go:484). |
| 11. deneme | 10 başarısızlıktan sonra `fails>=limit` → **429**, `Retry-After` = pencerede kalan saniye (en az 1) (login_limiter.go:40-58, accounts.go:370-372). Kontrol bcrypt'ten önce yapılıyor. Doğru parola da kilit süresince 429 alıyor. IP katmanının 429'unda Retry-After **yok** (accounts.go:358-360). |
| Sayılmayanlar | Parola <12 ya da >72 karakterse veya e-posta geçersizse 400 dönüyor ve deneme sayılmıyor (:363-366). Kısa parolalı 15×400 istek bu sayacı test etmiyor. |
| Bellek | Süreç içinde bir map; restart'ta sıfırlanıyor. Çoklu instance'ta her instance kendi sayacını tutuyor (limit fiilen N×10). En fazla 10000 kova var; dolunca önce süresi dolanlar, sonra en eskisi siliniyor (:84-99). 10000 farklı e-postayla spray yapan bir saldırgan hedefin kovasını sildirip sayacı sıfırlatabilir (DÜŞÜK). IP katmanında 4096 kova dolunca yeni IP'ler 15 dk boyunca 429 alıyor (fail-closed, :276-280). |
| Enumeration | Kayıtlı olmayan e-postalar da aynı şekilde sayılıyor (dummy bcrypt + fail, :463-468). Limiter hesabın var olup olmadığını ele vermiyor. Yan etkisi: herkes, bilinen herhangi bir e-postayı 15 dk kilitleyebilir (hedefli kilitleme DoS'u). |
| r10_base | Hesap limiter'ı **yok**, yalnız IP katmanı var (accounts.go:250-292; gruplar :324-334, 15 dk :265, 4096 :269). |
| Yerel test | TestAccountLimiterRotatingIPs, SuccessResetsAndWindow ve MemoryBound **3/3 PASS** (.11 yığını). Canlı repro koşulmadı. |

**QA'ya kısa yanıt:** .11'de hesap limiter'ı yalnız `/account/login` ve onun `/v1/account/login` alias'ı için çalışıyor. Anahtar küçük harfe çevrilip boşlukları kırpılmış e-posta. Varsayılan sabit pencerede 900 sn içinde 10 başarısız deneme sayılıyor; başarılı giriş sayacı sıfırlıyor. 11. deneme bcrypt'ten önce 429 ve kalan süreyi gösteren Retry-After ile reddediliyor. Kayıtlı olmayan e-postalar da sayıldığı için enumeration yok, ama başkasının hesabını kilitleme (DoS) mümkün. register (5), reset (5) ve verify (10) yalnız IP katmanında (15 dk, Retry-After yok); redeem'de limit yok, oturum istiyor. Sayaç bellek içi: restart'ta sıfırlanıyor, instance'lar arasında paylaşılmıyor. r10'da hesap katmanı hiç yok. Canlı binary'de etkin olup olmadığı doğrulanmadı (F04 değerlendirmesi değişmedi: DESIGN GAP, kaynakta giderilmiş, canlıda doğrulanmadı).

## Test edilmeyenler
- .12 tabanı
- GLOBAL ile birlikte uygulama
- Windows çalışma zamanı
- Redocly
- Canlı ortam
