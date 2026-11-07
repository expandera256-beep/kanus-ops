# A: protokol onayı bekleyen bölüm

Durum: NOT DONE. Bu dosya bir uygulama veya onay değildir.

.12 chat istemcisi X-US-Request-ID gönderir ama hook bunu okumaz; reserve uç noktası ise istemci anahtarı kabul etmez. Aynı başlıklı iki başarılı chat 2 kredi düşürür. Reserve cevabı kaybolup yeniden gönderilirse iki hold oluşur. Bu nedenle model/repair/reserve için kör retry eklenmedi.

Öneri (uygulanmadı):
1. Chat'te mevcut X-US-Request-ID başlığını kimliği doğrulanmış hesap/lisans + canonical request hash ile bağla. Aynı kimlikte farklı gövde 409, concurrent retry aynı singleflight işini bekler. Başlık isim olarak zaten var fakat idempotency semantiği yeni sözleşmedir.
2. /v1/credits/reserve'e de aynı başlığı yeni protokol alanı olarak kabul et (alternatif: body request_id; iki ayrı alan açma). İlk reserve öncesi kimlik istemcide oluşturulur ve task/checkpoint'te saklanır; reserve ACK kaybolunca aynı kimlikle sorgulanır/tekrarlanır.
3. Ledger'da request->reservation + request hash + pending/committed/released terminal durumunu kalıcı/atomic tut. Session yenilemesinden etkilenmeyecek hesap/lisans kimliğiyle scope et. Yanıt kaybolması için sınırlandırılmış replay mekanizması gerekir; terminal isteği yeni ücretli modele dönüştürme. Başarı gövdesi saklanmayacaksa 409/ayrı durum sözleşmesi ve UX açıkça kararlaştırılmalı.
4. Failed attempt release ile logical request terminal durumunu ayır. Aynı logical request'teki transient denemeler aynı hold'u paylaşmalı; hiçbir 4xx otomatik retry edilmemeli. Tamamlanan yanıt için aynı commit tek kez uygulanmalı.
5. Orphan TTL sweep startup ve periyodik olarak, aktif timeout'tan uzun TTL ile, transactionLocked içinde yapılmalı. Süreç yeniden başladıktan sonra aktif olmayan hold'lar timeout sonrası release edilir; committed kayıt asla release'e çevrilmez. Ledger tombstone saklama süresi/replay sınırı API ile kesinleşmeli.
6. EXE ve repair-loop en fazla 3 deneme, jitter'lı 100/300 ms bounded backoff; sadece 5xx, timeout ve reset. Header ve body aynı kalır, task devamında aynı mantıksal model isteği için request id korunur. Farklı içerik üreten continuation/repair turu ayrı ücretli iştir; id paylaşılmaz.

Yeni protokol alanı ve replay davranışı onayı gelmeden bu kısım durduruldu. Yerel güvenli değişiklik sadece zaten idempotent /settle isteğinin aynı id ve aynı commit kararıyla 3 denemesidir.

Tekrarlama: go test -tags post12repro -run '^TestPost12Repro' ./licensesrv
Bu iki kabul testi bilerek kırmızıdır; default build'e dahil değildir.
