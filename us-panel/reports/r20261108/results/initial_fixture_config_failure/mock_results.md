# MOCK — yerel performans sonuçları

Gerçek model/GPU/uluslararası kullanıcı ölçümü değildir. Kredi maliyeti simülasyonu: $0.002/kredi. p99 küçük örneklerde kapasite SLO kanıtı değildir.

| Hedef | Kullanıcı | n | p50 ms | p95 ms | p99 ms | TTFT p50 ms | Kuyruk p95 ms | Hata % | Beklenen/gerçek kredi | Hold |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| python_credit_fixture | 1 | 4 | 43.202 | 49.262 | 50.076 | 21.1 | 0.005 | 0.0 | 4/4 | 0 |
| python_credit_fixture | 5 | 20 | 52.175 | 96.989 | 97.006 | 25.667 | 46.695 | 0.0 | 20/20 | 0 |
| python_credit_fixture | 10 | 40 | 95.329 | 182.128 | 191.596 | 69.391 | 135.728 | 0.0 | 40/40 | 0 |
| python_credit_fixture | 25 | 100 | 230.165 | 380.48 | 425.589 | 206.762 | 330.628 | 0.0 | 100/100 | 0 |
| python_credit_fixture | 50 | 200 | 460.873 | 822.846 | 868.83 | 435.615 | 763.707 | 0.0 | 200/200 | 0 |
| python_capacity1_queue2_timeout50ms | 50 | 50 | 50.173 | 72.811 | 165.868 | 217.343 | 0.001 | 98.0 | 1/1 | 0 |
| upstream_capacity1_queue1_timeout30ms | 10 | 10 | 7.26 | 125.269 | 181.287 | 172.001 | 17.057 | 90.0 | 1/None | None |
| seeded_error_rate25pct | 25 | 50 | 146.359 | 306.827 | 328.453 | 148.108 | 241.341 | 32.0 | 34/34 | 0 |
| unmodified_US_Panel_Go_hook | 1 | 4 | 0.422 | 0.732 | 0.769 | None | None | 100.0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | 5 | 20 | 1.064 | 3.702 | 4.684 | None | None | 100.0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | 10 | 40 | 1.713 | 6.577 | 6.672 | None | None | 100.0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | 25 | 100 | 6.332 | 7.982 | 8.259 | None | None | 100.0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | 50 | 200 | 10.618 | 16.12 | 16.442 | None | None | 100.0 | 0/0 | 0 |
| unmodified_US_Panel_capacity1_queue2_timeout1s | 50 | 50 | 28.257 | 35.101 | 35.641 | None | None | 100.0 | 0/0 | 0 |

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
| unmodified_US_Panel_Go_hook | normal | {'503': 1} | 0 | 1/0 | 0 |
| unmodified_US_Panel_Go_hook | truncate | {'503': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | missing_done | {'503': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | malformed | {'503': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | sse_error | {'503': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | http_error | {'503': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | client_cancel | {'503': 1} | 0 | 0/0 | 0 |

## Ham kanıt
mock_results.json her isteğin zamanlarını, HTTP durumunu, Retry-After ve SSE bütünlük bayraklarını içerir. Ürün kredi semantiği ayrıca doğrulanmalıdır. Go hook kuyruk başlığı sunmazsa süre N/A, istemci gecikmesinden kuyruk uydurulmaz.
