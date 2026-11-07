# MOCK — yerel performans sonuçları

Gerçek model/GPU/uluslararası kullanıcı ölçümü değildir. Kredi maliyeti simülasyonu: $0.002/kredi. p99 küçük örneklerde kapasite SLO kanıtı değildir.

| Hedef | Kullanıcı | n | p50 ms | p95 ms | p99 ms | TTFT p50 ms | Kuyruk p95 ms | Hata % | Beklenen/gerçek kredi | Hold |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| python_credit_fixture | 1 | 4 | 43.074 | 43.612 | 43.672 | 20.931 | 0.006 | 0.0 | 4/4 | 0 |
| python_credit_fixture | 5 | 20 | 45.083 | 87.916 | 87.935 | 21.652 | 43.704 | 0.0 | 20/20 | 0 |
| python_credit_fixture | 10 | 40 | 90.158 | 174.786 | 176.972 | 65.829 | 130.835 | 0.0 | 40/40 | 0 |
| python_credit_fixture | 25 | 100 | 221.837 | 353.471 | 395.054 | 198.336 | 307.119 | 0.0 | 100/100 | 0 |
| python_credit_fixture | 50 | 200 | 401.344 | 709.114 | 751.776 | 378.071 | 660.31 | 0.0 | 200/200 | 0 |
| python_capacity1_queue2_timeout50ms | 50 | 50 | 6.935 | 34.228 | 120.792 | 160.717 | 0.001 | 98.0 | 1/1 | 0 |
| upstream_capacity1_queue1_timeout30ms | 10 | 10 | 1.474 | 110.345 | 161.57 | 152.547 | 16.891 | 90.0 | 1/None | None |
| seeded_error_rate25pct | 25 | 50 | 135.776 | 267.013 | 267.361 | 113.32 | 220.994 | 32.0 | 34/34 | 0 |
| unmodified_US_Panel_Go_hook | 1 | 4 | 46.332 | 46.907 | 46.983 | 22.569 | None | 0.0 | 4/4 | 0 |
| unmodified_US_Panel_Go_hook | 5 | 20 | 50.278 | 97.075 | 97.746 | 24.871 | None | 0.0 | 20/20 | 0 |
| unmodified_US_Panel_Go_hook | 10 | 40 | 99.382 | 190.73 | 194.172 | 75.242 | None | 0.0 | 40/40 | 0 |
| unmodified_US_Panel_Go_hook | 25 | 100 | 238.918 | 403.83 | 434.372 | 215.388 | None | 0.0 | 100/100 | 0 |
| unmodified_US_Panel_Go_hook | 50 | 200 | 443.521 | 831.249 | 889.433 | 414.935 | None | 0.0 | 200/200 | 0 |
| unmodified_US_Panel_capacity1_queue2_timeout1s | 50 | 50 | 39.066 | 588.452 | 1038.432 | None | None | 100.0 | 0/0 | 0 |

## Yanıt bütünlüğü ve iptal

| Hedef | Mod | HTTP | Tam | Beklenen/gerçek düşüm | Hold |
|---|---|---|---:|---:|---:|
| python_credit_fixture | normal | {'200': 1} | 1 | 1/1 | 0 |
| python_credit_fixture | truncate | {'200': 1} | 0 | 0/0 | 0 |
| python_credit_fixture | missing_done | {'200': 1} | 0 | 0/0 | 0 |
| python_credit_fixture | malformed | {'200': 1} | 0 | 0/0 | 0 |
| python_credit_fixture | sse_error | {'200': 1} | 0 | 0/0 | 0 |
| python_credit_fixture | http_error | {'503': 1} | 0 | 0/0 | 0 |
| python_credit_fixture | client_cancel | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | normal | {'200': 1} | 1 | 1/1 | 0 |
| unmodified_US_Panel_Go_hook | truncate | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | missing_done | {'200': 1} | 1 | 0/1 | 0 |
| unmodified_US_Panel_Go_hook | malformed | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | sse_error | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | http_error | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | client_cancel | {'200': 1} | 0 | 0/0 | 0 |

## Ham kanıt
mock_results.json her isteğin zamanlarını, HTTP durumunu, Retry-After ve SSE bütünlük bayraklarını içerir. Ürün kredi semantiği ayrıca doğrulanmalıdır. Go hook kuyruk başlığı sunmazsa süre N/A, istemci gecikmesinden kuyruk uydurulmaz.
