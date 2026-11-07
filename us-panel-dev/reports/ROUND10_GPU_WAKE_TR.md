# Round 10 — ds-a100 GPU uyandırma (yerel, deploy YOK)

Canlı GPU/VPS çağrısı yapılmadı; tüm testler httptest sahte sunucularıyla. /workspace/secrets okunmadı.

## Değişiklikler (patches/r10_gpu_wake.patch, taban: .9 ağacı r10_gpu_base)
- licensesrv/gpu_wake.go (yeni): GPUWaker — URL koruması (yalnız tam /a100/start, /a100/status; "health" ve /v1/ reddedilir, kullanıcı:şifre reddedilir), ısınma süresi en az 45 dk, singleflight Trigger, durum yoklaması 5→30 sn geri çekilme, başarısızlıkta 30 sn bekleme, token her çağrıda dosyadan okunur (saklanmaz, loglanmaz, hata metnine girmez), writeWarming: 503 + Retry-After 30 + X-US-GPU-State + JSON.
- licensesrv/model_proxy.go: `Waker` alanı; kontrol kuyruk ve kredi rezervasyonundan ÖNCE → ısınmada rezervasyon/ücret yok. SSE, eşzamanlılık 1 / kuyruk 32 değişmedi. Hesap/lisans semantiği değişmedi.
- cmd/us-license-hook/main.go: US_MODEL_UPSTREAM varsayılan http://127.0.0.1:18000; US_GPU_WAKE_URL, US_GPU_STATUS_URL, US_GPU_WAKE_TOKEN_FILE (/etc/us-model/wake.token), US_GPU_WARMUP_TIMEOUT (45m). İki URL de boşsa özellik kapalı.
- EXE: gpu_warming.go + model.go → 503/X-US-GPU-State yanıtı "GPU ısınıyor… Yaklaşık N sn sonra tekrar deneyin. Kredi düşülmedi." gösterilir. EXE değiştiği için **.10** derlendi (release 2026.11.06.10, US_DERLE.bat _10.exe).
- deploy: env.example ve README_TR güncellendi.

## Testler
202, 200, ready false→true, timeout, auth hatası, redaction (TestGPUWakeAuthErrorAndTimeoutAndRedaction), singleflight, /health koruması + 45 dk taban, ısınmada rezervasyon yok + hazır olunca SSE (TestGPUWakeProxyWarmingNoChargeThenSSE), idle stop sonrası yeniden başlatma, başarısız ısınma ücretsiz, EXE mesajı (TestGPUWarmingMessageShown). race -count=3 PASS.
check.sh: gofmt, vet, windows vet, race (PASS=386 FAIL=0 SKIP=7), Windows build OK. Regresyon koşucusu .10 ağacında 36/36 PASS.

## Artefaktlar (sha256)
- dist/US_Ticari_Aday_2026_11_06_10.exe 96aefc86a8f57bbe53227481e8985cac12a98f3090b63430e40dcda83d4414dc
- dist/us-license-hook_linux 1d4bdc9392da19ae8625cc754d714cafab0e241d7eead8128dd6fb02607c0ce2
- deploy/us-license-hook_hostinger.tgz cf1363348604d072fd8866878c0b535e52ff420924fa3d5b1753a636d8446b7c
- Önceki tgz/binary/env/README yedeği: us-panel-dev/r10_prev/

## Notlar / öneri
- Erişilebilirlik TCP dial ile ölçülür; port açık ama model yüklenmemişse ReadyGrace (3 dk) içinde istek upstream'e gider.
- usta-wake sohbet tamponlama: önerilen düzeltme, ısınma yanıtının SSE yerine düz JSON 503 olarak hemen dönmesi (uygulandı) ve istemcinin Retry-After ile yeniden denemesi; sunucu tarafında istek tamponlanmaz.
- Windows'a özgü kontroller: TEST EDİLMEDİ.
