# MOCK — yerel performans sonuçları

Gerçek model/GPU/uluslararası kullanıcı ölçümü değildir. Kredi maliyeti simülasyonu: $0.002/kredi. p99 küçük örneklerde kapasite SLO kanıtı değildir.

| Hedef | Kullanıcı | n | p50 ms | p95 ms | p99 ms | TTFT p50 ms | Kuyruk p95 ms | Hata % | Beklenen/gerçek kredi | Hold |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| python_credit_fixture | 1 | 4 | 43.216 | 44.148 | 44.256 | 20.941 | 0.008 | 0.0 | 4/4 | 0 |
| python_credit_fixture | 5 | 20 | 45.705 | 88.907 | 88.919 | 22.315 | 43.799 | 0.0 | 20/20 | 0 |
| python_credit_fixture | 10 | 40 | 89.854 | 175.817 | 177.943 | 66.21 | 131.746 | 0.0 | 40/40 | 0 |
| python_credit_fixture | 25 | 100 | 224.334 | 360.582 | 400.14 | 200.702 | 313.445 | 0.0 | 100/100 | 0 |
| python_credit_fixture | 50 | 200 | 411.403 | 728.914 | 772.036 | 386.699 | 676.118 | 0.0 | 200/200 | 0 |
| python_capacity1_queue2_timeout50ms | 50 | 50 | 7.662 | 35.11 | 121.871 | 161.798 | 0.001 | 98.0 | 1/1 | 0 |
| upstream_capacity1_queue1_timeout30ms | 10 | 10 | 2.273 | 110.783 | 162.179 | 153.095 | 16.747 | 90.0 | 1/None | None |
| seeded_error_rate25pct | 25 | 50 | 137.546 | 269.29 | 270.307 | 114.643 | 222.492 | 32.0 | 34/34 | 0 |
| unmodified_US_Panel_Go_hook | 1 | 4 | 46.876 | 65.583 | 68.223 | 22.81 | None | 0.0 | 4/4 | 0 |
| unmodified_US_Panel_Go_hook | 5 | 20 | 50.749 | 98.514 | 99.769 | 25.086 | None | 0.0 | 20/20 | 0 |
| unmodified_US_Panel_Go_hook | 10 | 40 | 97.243 | 190.731 | 203.731 | 72.455 | None | 0.0 | 40/40 | 0 |
| unmodified_US_Panel_Go_hook | 25 | 100 | 237.858 | 431.044 | 456.785 | 214.466 | None | 0.0 | 100/100 | 0 |
| unmodified_US_Panel_Go_hook | 50 | 200 | 477.116 | 856.501 | 910.857 | 450.987 | None | 0.0 | 200/200 | 0 |
| unmodified_US_Panel_capacity1_queue2_timeout1s | 50 | 50 | 9.59 | 566.776 | 1010.17 | None | None | 100.0 | 0/0 | 0 |

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
| unmodified_US_Panel_Go_hook | missing_done | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | malformed | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | sse_error | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | http_error | {'200': 1} | 0 | 0/0 | 0 |
| unmodified_US_Panel_Go_hook | client_cancel | {'200': 1} | 0 | 0/0 | 0 |

## Ham kanıt
mock_results.json her isteğin zamanlarını, HTTP durumunu, Retry-After ve SSE bütünlük bayraklarını içerir. Ürün kredi semantiği ayrıca doğrulanmalıdır. Go hook kuyruk başlığı sunmazsa süre N/A, istemci gecikmesinden kuyruk uydurulmaz.
