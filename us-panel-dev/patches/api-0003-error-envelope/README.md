# API-0003 — Standart hata zarfı

## Ne değişti
- `licensesrv/api_error.go`: `ErrorBody{error,code,message,request_id}`, `WriteAPIError` (dual-mode), `WriteAccountError`.
- Credits/model: `Accept: application/json` veya `X-Kanus-Error-Envelope: 1` veya `?error_envelope=1` → JSON zarf; aksi halde eski düz metin (`http.Error` şekli).
- Account hataları: her zaman JSON; legacy `"error"` string korunur, `code`/`message`/`request_id` eklenir; `X-Request-Id` başlığı.
- Kredi defteri semantiğine dokunulmadı.

## Geriye uyumluluk
- Varsayılan credits/model yanıtları hâlâ düz metin + newline.
- Account gövdesinde `"error":"account request rejected"` alanı duruyor.

## Test
- `go vet ./...` PASS
- `go test -race ./...` PASS (kiyas, hook, licensesrv)
- Yeni: `TestAPIErrorEnvelopeJSONAdditiveAndPlainFallback`, `TestAccountErrorKeepsLegacyErrorField`
