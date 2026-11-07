# BUILD_INFO — i18n_windows_amd64.test.exe sha256 166d869a5dad7e21985a5b910f894c4283af1799244653fce3f0c97104f974fe

- **Kaynak ağaç:** `/workspace/us-panel-dev/global-work/v10dev` = `pinned/v10` (181 dosya) + sırayla:
  1. global_i18n_panel_core.patch `e4f10a339c78818ceddbc49e2f687b4ee80afa86a7fce9d407c22bd8536892c7`
  2. global_i18n_02_gui_strings.patch `314fab48b009f06e221ce140758724105b9f5b972c69d0c15c7ad2c9db8d3669`
  3. global_i18n_03_pref.patch `ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e` (donmuş)
  4. global_i18n_04_errors.patch `1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991`
  5. global_i18n_06_de_ru.patch `83aaca5d4a3aeb8cf7d4ed0e3854d0c3408d1553bb139ddb2eee22f8b5d013c6`
  (05/07 Python, binary'ye girmez.) Ağacın bu zincirle birebir eş olduğu Tur 9'da `diff -rq` ile doğrulandı.
- **Komut (cwd = v10dev):** `GOOS=windows GOARCH=amd64 go test -c -o ../ci/i18n_windows_amd64.test.exe ./i18n`
- **Go:** go1.27.3 linux/amd64 (`go version -m`: go1.27.3). x/sys v0.30.0 (h1:QjkSwP/36a20jFYWkSue1YwXzLmsV5Gfq7Eiy72C1uc=).
- **Hedef/ortam (binary buildinfo):** GOOS=windows, GOARCH=amd64, GOAMD64=v1, CGO_ENABLED=0 (çapraz derlemede otomatik), -buildmode=exe, -compiler=gc; -race YOK; -trimpath YOK; GOFLAGS boş; mod (devel), VCS bilgisi gömülü değil.
- **Derleme zamanı:** dosya mtime 08.11.2026 00:38:34 TSİ (Tur 9 shell çıktısıyla uyumlu; ayrı build logu yok).
- **Yeniden derleme (08.11.2026 ~00:59 TSİ, aynı komut, /tmp'ye):** sha256 `166d869a5dad7e21985a5b910f894c4283af1799244653fce3f0c97104f974fe` → **BİREBİR AYNI** (bu makine/yol/Go sürümünde tekrar üretilebilir; farklı yol veya Go sürümünde -trimpath olmadığı için farklı çıkabilir).
- Windows r4 (23 PASS / 0 FAIL / 1 SKIP) bu binary ile koşuldu; hash PC'de doğrulama ve RESULT exit=2 / ERRORLEVEL 2: **KAN.US beyanı, logda yok**.
