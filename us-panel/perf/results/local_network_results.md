# MOCK_NETWORK — düşük hızlı health ölçümü

Kaynak zamanı: 2026-11-06T09:34:30.715812+03:00 (TSİ)

| HTTP | DNS ms | TCP connect ms | TLS ms | HTTP TTFB ms | Toplam ms |
|---|---:|---:|---:|---:|---:|
| 200 | 0.015 | 0.059 | N/A | 0.686 | 0.712 |

Egress: `{}`

- TCP connect is connect completion minus DNS, NOT pure RTT.
- TLS increment can include network/proxy overhead; no packet-level RTT inferred.
- Health HTTP TTFB is NOT LLM TTFT.
- Single box vantage does not measure US/Europe end-user latency.
