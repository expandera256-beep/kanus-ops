# API-0010 — Python SDK (kanus-sdk)

OpenAPI `/v1` ince istemci: credits, usage, SSE chat; hata zarfı; 429+Retry-After; Idempotency-Key; Auth (license/session/api_key). Anahtar `repr` içinde kırpılır; testler mock HTTP’e karşı.

```bash
cd patches/api-0010-sdk-python
PYTHONPATH=src pytest -q
```
