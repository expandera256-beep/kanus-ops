# US önerileri — UYGULANMADI, DERLENMEDİ

Hedef work2/kiyas/term; yalnız diff oluşturuldu, gerçek US dosyasına dokunulmadı.

1. **01-backpressure-observability.diff**: gerçek yerel testte kuyruk dolu429 ve timeout504 yanıtlarında Retry-After yok; kuyruk süresi dışarı sunulmuyor. Sabit1sn yeniden deneme önerisi ve X-US-Queue-Wait-Ms ekler. 504 korunur; 503'e geçiş ayrı API kararıdır. Dağıtık istemcilerde jitter/exponential backoff gerekir. Büyük kuyruk önerilmez.
2. **02-strict-upstream-done-policy.diff**: mevcut geçit finish_reason=stop görülür görülmez krediyi kesinleştiriyor, upstream [DONE] gelmeden downstream [DONE] üretiyor. Fault injection eksik[DONE] senaryosu1 kredi tüketti. Bu diff strict OpenAI SSE politikası istenirse tamamlanmayı finish+upstream[DONE]'a bağlar; aksi mevcut davranış tasarım kararıdır. DONE olmadan EOF/hata/timeout rezervasyonu bırakmalı. Terminal stop tek başına kabul edilecekse bu patch uygulanmamalı ve ürün sözleşmesi belgelenmeli.

Gerekli regresyon: normal SSE1 düşüm; finish+EOF0; DONE-only0; malformed/error/truncate/cancel0; content+finish+DONE1; queue full/timeout0 ve hold0; Retry-After parse edilebilir; queue header ölçeği ve admission-before-reserve. İptal settlement'tan sonra gerçekleşirse ücret semantiği ayrıca belirlenmeli. Yalnız diff incelemesi yapıldı; patch test edilmiş sayılmaz.
