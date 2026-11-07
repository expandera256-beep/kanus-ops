# API-0002 OpenAPI düzeltme (PANEL inceleme)

**Tarih:** 2026-11-06 TSİ · Kaynak inceleme: `PANEL_API_REVIEW_20261106.md`

## ORTA — settle auth belgesi
- Global `BearerAuth` kaldırıldı; yerine `LicenseBearer` (`us1…`) ve `SessionBearer` (`usa1…`).
- `/v1/credits/reserve` ve `/v1/credits/settle`: yalnız `LicenseBearer` + açıklamada oturum kabul edilmediği.
- `/v1/credits/status`: `LicenseBearer` **veya** `SessionBearer`.
- Model uçları: her iki şema (kod `authorizedCredential`).
- Account korumalı uçlar: yalnız `SessionBearer`.

## DÜŞÜK — ChatRequest / version / lint
- `ChatRequest`: `required: [model, messages, max_tokens]`; `max_tokens` 1..8192; `messages` 1..256; `n`/`best_of`/`max_completion_tokens` reddi notu.
- `info.version`: `2026.11.06.10` (.10 taban).
- Tüm operation'lara `operationId`; eksik `4XX` yanıtları eklendi (Redocly uyarı azaltma).
