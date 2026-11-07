# LOCAL SYNTHETIC — login, not GPU

Timestamp: 2026-11-08T00:23:00.995380+03:00 (TSİ)

| Variant | Users | n | p50 ms | p95 ms | Error % | HTTP | Requests/s | Success/s |
|---|---:|---:|---:|---:|---:|---|---:|---:|
| baseline_dot10_prior_binary | 1 | 50 | 344.867 | 482.637 | 0.0 | {'200': 50} | 2.914 | 2.914 |
| baseline_dot10_prior_binary | 5 | 50 | 873.172 | 1282.925 | 20.0 | {'200': 40, '429': 10} | 5.031 | 4.025 |
| baseline_dot10_prior_binary | 10 | 50 | 68.775 | 857.096 | 60.0 | {'200': 20, '429': 30} | 14.592 | 5.837 |
| baseline_dot10_prior_binary | 20 | 60 | 56.016 | 784.522 | 80.0 | {'200': 12, '429': 48} | 30.251 | 6.050 |
| round11_default2_wait10s | 1 | 50 | 226.671 | 280.938 | 0.0 | {'200': 50} | 4.283 | 4.283 |
| round11_default2_wait10s | 5 | 50 | 612.044 | 1195.376 | 0.0 | {'200': 50} | 5.048 | 5.048 |
| round11_default2_wait10s | 10 | 50 | 658.54 | 1103.099 | 0.0 | {'200': 50} | 8.965 | 8.965 |
| round11_default2_wait10s | 20 | 60 | 1266.715 | 2188.818 | 0.0 | {'200': 60} | 7.998 | 7.998 |
| round11_overload_wait1s_DIAGNOSTIC_NOT_DEFAULT | 20 | 60 | 1055.398 | 1077.252 | 10.0 | {'200': 54, '503': 6} | 10.338 | 9.304 |

No passwords/hashes/sessions recorded. Unique logical IPs isolate bcrypt from IP rate limiting; real remote traffic stays loopback. Affinity and source/binary hashes in JSON. Queue waiting comes from X-US-Queue-Wait-Ms where exposed; baseline N/A.

## .11 measured queue header

| Users | p50/p95/max ms |
|---:|---|
| 1 | 0.0/0.0/0.0 |
| 5 | 245.5/794.6/1014.0 |
| 10 | 421.5/861.0/989.0 |
| 20 | 1021.0/1965.15/2089.0 |
