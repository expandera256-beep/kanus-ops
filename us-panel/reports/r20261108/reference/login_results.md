# LOCAL SYNTHETIC — login / iki CPU

2026-11-06T09:40:55.065495+03:00 (TSİ)

| Sürüm | Kullanıcı | n | Tüm p50/p95 ms | Başarılı p50/p95 ms | Hata % | HTTP | Tüm RPS / başarılı RPS |
|---|---:|---:|---|---|---:|---|---|
| baseline_dot10_prior_binary | 1 | 50 | 210.0/220.2 | 210.0/220.2 | 0.0 | {'200': 50} | 4.71/4.71 |
| baseline_dot10_prior_binary | 5 | 50 | 407.8/469.1 | 414.3/473.1 | 20.0 | {'200': 40, '429': 10} | 11.30/9.04 |
| baseline_dot10_prior_binary | 10 | 50 | 66.6/511.4 | 427.0/525.9 | 60.0 | {'429': 30, '200': 20} | 22.12/8.85 |
| baseline_dot10_prior_binary | 20 | 60 | 46.1/422.6 | 420.0/445.0 | 80.0 | {'429': 48, '200': 12} | 45.97/9.19 |
| round11_default2_wait10s | 1 | 50 | 210.3/230.8 | 210.3/230.8 | 0.0 | {'200': 50} | 4.68/4.68 |
| round11_default2_wait10s | 5 | 50 | 507.0/1045.7 | 507.0/1045.7 | 0.0 | {'200': 50} | 5.28/5.28 |
| round11_default2_wait10s | 10 | 50 | 640.7/1087.4 | 640.7/1087.4 | 0.0 | {'200': 50} | 9.26/9.26 |
| round11_default2_wait10s | 20 | 60 | 1225.3/1321.8 | 1225.3/1321.8 | 0.0 | {'200': 60} | 9.21/9.21 |
| round11_overload_wait1s_DIAGNOSTIC_NOT_DEFAULT | 20 | 60 | 1034.5/1076.5 | 943.0/1077.1 | 16.7 | {'503': 10, '200': 50} | 10.35/8.63 |

taskset0,1 + GOMAXPROCS2; /proc process+thread affinities verified. bcrypt cost12. Four fixture accounts; four warmup logins excluded per level. Source/binary hashes and per-request metrics in JSON.

Baseline .10 candidate is exactly the prior measured binary; not a separately release/commit-attested build. Fast429 responses lower overall latency percentiles; compare success-only percentiles too. Request bursts are not a steady-state/open-loop SLO test.

Default round11 20-concurrent: zero503; Retry-After branch not exercised there. All default response times<10s, so observed queue durations bounded above by total response time<10s; exact queue time not exposed. Separate1s-wait diagnostic:503s have Retry-After5; not a measured default10s timeout test.
