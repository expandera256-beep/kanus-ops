# COMPLETED — conditional remote phases

No stop condition.

| Phase | Service health | Remote seconds | Levels |
|---|---|---:|---|
| phase1 | HTTP200 | 38.812 | CLEAN |
| phase2 | HTTP200 | 21.422 | CLEAN |

| Phase | Users | n | Successful p50/p95/p99 ms | Success% | Expected429% | 5xx% |
|---|---:|---:|---|---:|---:|---:|
| phase1 | 1 | 20 | 437.754/488.575/509.286 | 100.0 | 0.0 | 0.0 |
| phase1 | 5 | 20 | 443.053/464.048/499.239 | 100.0 | 0.0 | 0.0 |
| phase1 | 10 | 20 | 441.42/447.621/448.032 | 100.0 | 0.0 | 0.0 |
| phase1 | 20 | 20 | 466.848/488.474/488.494 | 100.0 | 0.0 | 0.0 |
| phase2 | 1 | 5 | 695.783/720.947/723.542 | 100.0 | 0.0 | 0.0 |
| phase2 | 5 | 5 | 904.467/911.285/911.799 | 80.0 | 20.0 | 0.0 |
| phase2 | 10 | 10 | 898.912/906.934/907.186 | 40.0 | 60.0 | 0.0 |
| phase2 | 20 | 20 | 944.785/948.311/948.446 | 20.0 | 80.0 | 0.0 |

Dual public DoH: [{"provider": "dns.google", "addresses": ["82.29.184.225"], "pass": true, "http_status": 200, "DNS_status": 0, "TTLs": [300]}, {"provider": "cloudflare-dns.com", "addresses": ["82.29.184.225"], "pass": true, "http_status": 200, "DNS_status": 0, "TTLs": [300]}]
TLS certificates: validated with default CA trust and hostname api-test.kanus.org; normal SNI/Host, no IP pinning.
Local198.18.0.1 is box proxy DNS; TCP timings are NOT origin end-to-end RTT.
TLS stage includes local SSL context/CA setup.
Target HTTPS requests: 154; DoH HTTPS requests: 2

DNS: {'local_addresses': ['198.18.0.1'], 'required_public_A': '82.29.184.225', 'elapsed_ms': 315.3631180002776, 'verification': 'dual_DoH', 'pass': True}
Total remote time: 60.23395441100001seconds
Balance before/after: 10/10
Identity: HOST declaration7d72f449…56f0; not an independently measured binary hash.
