# US yanıtı — QA FINDING-01 / FINDING-02 (2026-11-08, TSİ)

Taban: `/workspace/us-panel-dev/qa12/term.orig` (work2/kiyas/term'den 00:41 TSİ'de kopyalandı; `dist/2026.11.06.11/TREE_MANIFEST.sha256` ile birebir doğrulandı = .11 ağacı).
Çalışma kopyası: `/workspace/us-panel-dev/qa12/term`. work2 ve dist/.11'e dokunulmadı. Deploy yok, canlı çağrı yok.
Yama: `/workspace/us-panel-dev/patches/qa12/qa12_finding01_02.patch` (sha256 c31ac8b8b015184d7911c4eadf0abde6155f41cefa3da24da25fd5badabf40dc).

Repro notu: `qa/audit12/run.py` sabit olarak work2'yi hedefliyor ve QA'nın `qa-raw/` kanıtlarının üzerine yazıyor. Bu yüzden aynı overlay ve aynı test desenini kullanan uyarlanmış bir koşucu kullanıldı: `/workspace/us-panel-dev/qa12/repro.sh <ağaç> <case> <etiket>`. Loglar `/workspace/us-panel-dev/qa12/logs/` altında.

## FINDING-02 (MEDIUM): repairLoop redirect izliyor — DONE
- **Kök neden (doğrulandı):** `repair_loop.go` askFix içinde `l.Client` nil olduğunda `&http.Client{Timeout:10m}` oluşturuluyordu ve CheckRedirect yoktu. Go, aynı hostname'in başka portuna olan yönlendirmeyi güven sınırı saymıyor. 307 yanıtı POST gövdesini koruyor, Authorization başlığı da aynen gidiyor. Çağıranın verdiği client için de redirect politikası uygulanmıyordu.
- **Test (önce FAIL):** `repair_redirect_test.go: TestRepairLoopRefusesRedirects`, alt testler default_client ve caller_client. Origin sunucu 307 ile ikinci bir httptest sunucusuna yönlendiriyor. Test, sink'e hiç istek gitmemesini, Authorization'ın taşınmamasını, hata dönmesini ve çağıranın client'ının değiştirilmemesini doğruluyor. Fix öncesi iki alt test de `sink_hit=true authorization_leaked=true` ile FAIL verdi.
- **Fix:** Çağıranın client'ı varsa sığ kopyası alınıyor, yoksa default client oluşturuluyor. Her iki durumda da `CheckRedirect = ErrUseLastResponse` uygulanıyor; çağıranın nesnesine dokunulmuyor. 307 yanıtı `model HTTP 307` hatasına dönüşüyor.
- **QA repro:** önce `redirect_contacted=true authorization_forwarded=true project_body_forwarded=true` → FAIL. Sonra `false/false/false` → PASS.
- **Diğer giden client'lar (.11):**
  - Kimlik bilgisi gönderenlerin hepsinde zaten CheckRedirect=ErrUseLastResponse var: account_client.go:23,98,139; gateway.go:112,181; context_limits.go:81; model.go:2263; smoke_cli.go:106; worker_client.go:35,73; release.go:15; tunnel.go:54 (modelAvailable kopya üzerinde); licensesrv/model_proxy.go:57; licensesrv/gpu_wake.go:136; licensesrv/gpu_pin.go:150,173.
  - Kimlik bilgisi göndermeyen, redirect politikası olmayan client'lar: agent.go:164, desk.go:512, gui.go:101 ve 112, model.go:1262 ve 1279 (yerel sayfa/sipariş yoklamaları), main.go:294 (/api/durum GET), tunnel.go:100 ve 149 (modelAvailable içinde kopyalanıp korunuyor), tunnel_windows.go:103 (gcloud arşiv indirmesi; SHA256 ile doğrulanıyor). Bunlar kapsam dışı bırakıldı ve değiştirilmedi.
  - Not: `licensesrv.ModelProxy.Client` dışa açık bir alan. Dışarıdan CheckRedirect'siz bir client enjekte edilirse aynı risk oluşur. Üretimde kurucu (model_proxy.go:57) güvenli client veriyor. licensesrv bu turda dokunma yasağı kapsamında olduğu için değiştirilmedi; öneri olarak not edildi.
- **Erişilebilirlik:** Üretim kodunda `repairLoop` örneği oluşturan bir yer yok. Yalnız `task_state.go:191 runPersistentRepair` içinde parametre olarak geçiyor; kurucular yalnız testlerde. PANEL DEV'in "kullanılmayan yardımcı kod" tespiti doğrulandı.

## FINDING-01 (HIGH, koşullu): kardeş workspace okuması — NOT DONE (mimari)
- **Kök neden (doğrulandı):** safe_exec cwd kilidi (lockCwd), argv/bin allowlist, env allowlist ve rlimit/Job Object uyguluyor. Ancak çocuk süreç aynı UID ile çalışıyor ve dosya sistemi erişim politikası yok (namespace, chroot ya da ayrı kullanıcı yok). Bu yüzden izinli bir yorumlayıcıyla çalışan proje kodu mutlak yolla kardeş dizini açabiliyor. safe_exec.go:17–19 bunu açıkça belirtiyor.
- **Repro:** fix öncesi ve sonrası `synthetic_sibling_read=true state=success` → FAIL. Beklendiği gibi kapanmadı.
- **Bu turda eklenen ucuz önlemler (yalnız regresyon korumaları; ürün kodu değişmedi):**
  - `TestFileToolsRefuseSiblingWorkspace`: file tools API'si `../`, mutlak yol ve symlink üzerinden kardeş dizine okuma ve yazmayı reddediyor. Mevcut davranışta PASS.
  - `TestSafeExecEnvDoesNotExposeWorkspaceRoots`: çocuk sürecin env'i PWD, OLDPWD ve US_* kök yollarını taşımıyor. Mevcut davranışta PASS.
  - Bu iki test önceden de geçiyordu; davranışı kilitlemek için eklendi. Workspace kökleri kullanıcının seçtiği klasörler olduğundan dizin izinlerini (0700) masaüstü uygulamasının değiştirmesi güvenli bulunmadı ve yapılmadı. Ayrıca aynı UID altında 0700 izni okumayı zaten engellemez.
  - **Bu önlemler bulguyu KAPATMAZ.**
- **Seçenekler:**
  1. Paylaşılan host / çok kullanıcı: kullanıcı başına ayrı UID ya da container/microVM. Yalnız ilgili workspace mount edilir; HOME ve secrets erişilemez; ağ politikası uygulanır. HOST/SRE işi.
  2. Linux: user + mount namespace (bwrap benzeri), Landlock LSM ile yalnız kök dizine izin, seccomp.
  3. Windows: ayrı düşük yetkili kullanıcı ya da AppContainer/LPAC ile workspace ACL'i.
  4. Masaüstü güvenilir tek kullanıcı modunu "güvenilir proje" olarak açıkça belgelemek.

## İkili (binary) etkisi — kanıt
- `go list -deps ./cmd/us-license-hook` çıktısındaki yerel paketler yalnız `kiyas/licensesrv` ve `kiyas/cmd/us-license-hook`.
- repair_loop.go, safe_exec.go ve file_tools.go kök `package main` (kiyas) içinde. Bu paket EXE'dir ve hem linux hem GOOS=windows GoFiles listesinde yer alıyor.
- **Sonuç:** İki bulgu da yalnız EXE'yi (masaüstü yolu) etkiliyor, hook binary'sini etkilemiyor. .11 hook deploy kararına etkileri yok. PANEL DEV'in tespiti doğrulandı.

## Doğrulama (yamalı ağaç)
- gofmt: değişen dosyalar temiz (önceden var olan formatsız dosya listesi tabanla aynı).
- go vet ./...: OK.
- go test ./...: 458 pass, 7 skip, 0 fail.
- go test -race ./...: 458 pass, 7 skip, 0 fail. Sayımlar alt testleri içerir.
- S2 regresyonu: PASS=36, FAIL=0, BLOCKED=0 (`qa12/logs/s2`).

## Durum
- FINDING-02: DONE (kod ve test; QA'nın yeniden testi bekleniyor).
- FINDING-01: NOT DONE. Yalnız regresyon korumaları eklendi; mimari düzeltme HOST/SRE ve PANEL DEV kararına bağlı.
