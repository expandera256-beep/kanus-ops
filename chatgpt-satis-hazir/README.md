# ChatGPT satışa-hazır çalışma kiti

Bu klasör, KAN.US projesinin 8 Kasım 2026 (TSİ) itibarıyla **sırlar olmadan** derlenmiş çalışma setidir.
Amaç: ChatGPT’ye bağlam verip satışa hazır sürüme giden eksikleri kapatmak.

**Bu paket satışa hazır değildir.** FAIL / SKIP / NOT TESTED / BLOCKED maddeleri PASS sayılmaz.

## İçinde ne var

- `CHATGPT_PROMPT_SATIS_HAZIR.md` — ChatGPT’ye yapıştırılacak ana prompt
- `docs/SATIS_HAZIR_CHECKLIST_20261108.md` — satış kontrol listesi (genel durum NO-GO)
- `docs/US_V1_IMAGE_SOZLESMESI_20261108.md` — `/v1/image` sözleşme r2 (PLAN ONLY)
- `docs/KAN_IMAGE_ENGINE_PLAN.md` ve `docs/KAN_IMAGE_ENGINE_STAGE.md` — görsel motor planı ve evre
- `docs/EXE_DERLEME_RAPORU.md` — yerel Windows EXE sınırları
- `image-engine/` — CPU supervisor / adaptör kaynakları (ağırlık yok)
- `imageapi/` — ayrı derlenebilir `/v1/image` Go paketi (licensesrv bağlı değil)

## Bilinçli olarak yok

- Webhook sırları, lisans sırları, SSH anahtarları, `.env` değerleri, kart/ödeme bilgileri
- FLUX ağırlıkları (~16 GB) ve GCS kimlik bilgileri
- Müşteri Windows EXE ikilisi
- Çalışan VPS/GCP hesap oturumları

Sırları bu repoya veya ChatGPT’ye yapıştırma.
