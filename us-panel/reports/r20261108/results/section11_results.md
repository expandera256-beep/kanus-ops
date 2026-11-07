## §11 — MOCK upstream + gerçek yerel.11 hook
Kontrollü yapılandırma:capacity3/queue64/request5sn; varsayılan prod kapasitesi iddiası değildir.4ardışık batch; mock TTFT20ms,4içerik parçası,150parça/sn. Gerçek token/GPU throughput değil. Küçük örnekte p99 kapasite/SLO kanıtı sayılmaz.

| Kullanıcı | n | p50/p95/p99 ms | TTFT p50 ms | Başarı | Beklenen/gerçek kredi | İzolasyon |
|---:|---:|---|---:|---:|---|---|
| 1 | 4 | 46.88/65.58/68.22 | 22.81 | 100% | 4/4 | PASS |
| 5 | 20 | 50.75/98.51/99.77 | 25.09 | 100% | 20/20 | PASS |
| 10 | 40 | 97.24/190.73/203.73 | 72.45 | 100% | 40/40 | PASS |
| 25 | 100 | 237.86/431.04/456.79 | 214.47 | 100% | 100/100 | PASS |
| 50 | 200 | 477.12/856.50/910.86 | 450.99 | 100% | 200/200 | PASS |

- Ana yük364/364 tam yanıt; kredi364/364; hold0. Beşbağımsız sentetik lisans/benzersiz yanıt markerı ile1/5/10/25/50 düzeylerinde ek91istek: tüm yanıtlar doğru isteğe ait, lisansbaşına düşüm eşleşiyor, hold0. Bu dosya/tenant çalışma alanı izolasyonunun tamamını kanıtlamaz.
- GPU belleği **N/A/MOCK**; gerçek toplam fatura maliyeti **NOT MEASURED**,0diye uydurulmadı. Varsayımsal$0.002/krediyle yalnız ana364isteğin simülasyon toplamı$0.728. Yerel CPU maliyeti fiyatlandırılmadı. GPU API çağrısı0.
- Go model route kuyruk bekleme header’ını dışarı sunmuyor: N/A; Python fixture ölçülen queue dağılımları mock_results.* içinde. Login route yeni header sağlıyor ve aşağıda gerçekten ölçüldü.
- Kapasite1/queue2/deadline1sn, upstream TTFT1.5sn,50istek:47×429+2×504+1×HTTP200 SSEtimeout; tam başarı0, kredi0, hold0.429/504 Retry-After1 doğrulandı; HTTP200 tek başına başarı değil.
- **Güncel.11 strict-DONE:** normal1düşüm; truncate/missingDONE/malformed/SSEerror/HTTPerror/clientcancel0düşüm+hold0. Önceki missingDONE politika açığı bu yerel kaynakta artık gözlenmedi; product_policy_failures boş. Yeni ürün kodunu PERF uygulamadı; mevcut kaynağı test etti.

