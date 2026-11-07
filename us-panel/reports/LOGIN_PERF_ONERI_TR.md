# Login/register performans önerisi (YALNIZ ÖNERİ — licensesrv değiştirilmedi)
06.11.2026 TSİ

## Durum
api-test VPS'te 2 vCPU var ve load 26–27. bcrypt-12 ile login 20–28 sn sürüyor. EXE'nin 20 sn istemci süresi dolunca nginx 499 yazdı, `--check-account` de exit 4 ile bitti. EXE tarafı .9'da düzeltildi: login/register için 90 sn süre ve ağ hatasında tek tekrar.

## (d) Ölçüm (US box, Go x/crypto/bcrypt v0.33, 5 tekrarın ortalaması)
| cost | tek karşılaştırma | 8 eşzamanlı, 8 CPU | 8 eşzamanlı, 2 CPU (taskset) |
|---|---|---|---|
| 10 | 53 ms | 56 ms | 213 ms |
| 11 | 105 ms | 124 ms | 423 ms |
| 12 | 209 ms | 213 ms | 839 ms |

**Önemli çıkarım:** boş 2 CPU'da bcrypt-12 yaklaşık 0,2 sn sürer. 20–28 sn'lik gecikme bcrypt'in kendi maliyeti değil; yaklaşık 100 kat CPU kıtlığı var. Load 26 başka bir süreçten (kaçak iş, swap veya I/O bekleme) geliyor olmalı. Önce `top`/`ps`/`vmstat` ile kaynağı bulunmalı. Bu kontrol yapılmadı; canlı sisteme dokunulmadı.

## (a) bcrypt semaforu ve kuyruk zaman aşımı
- En fazla 1–2 eşzamanlı hash/karşılaştırma işi olsun. Slot için en çok 10 sn beklensin; dolarsa `503` + `Retry-After: 5` + `{"status":"busy"}` dönsün. Aşırı yükte tüm istekler yavaşlamaz, açık bir ret yanıtı gelir. EXE 503'ü tekrar denemez, kullanıcıya gösterir.
```go
var bcryptSlots = make(chan struct{}, 2)
func withBcrypt(ctx context.Context, f func() error) error {
    t := time.NewTimer(10 * time.Second); defer t.Stop()
    select {
    case bcryptSlots <- struct{}{}: defer func() { <-bcryptSlots }(); return f()
    case <-t.C: return errBusy          // handler: 503 + Retry-After
    case <-ctx.Done(): return ctx.Err() // istemci koptu (499): CPU harcama
    }
}
```

## (b) Yeni hash'lerde cost 11
`bcrypt.CompareHashAndPassword` cost değerini hash'in kendisinden okur. Bu yüzden mevcut cost-12 hash'ler doğrulanmaya devam eder. Yeni kayıtlar cost 11 ile oluşturulur ve süreleri yarıya iner. İsteğe bağlı olarak login başarılı olunca `bcrypt.Cost(h) > 11` ise hash yeniden üretilip atomik store'a yazılabilir. Bu fazladan bir yazma ve bir hash maliyeti getirir; yük azalınca açılmalı.

## (c) Bilinmeyen kullanıcı için sahte karşılaştırma
Kullanıcı yoksa da aynı cost'ta sabit bir sahte hash ile `CompareHashAndPassword` çalıştırılsın ve aynı hata dönsün. Böylece yanıt süresinden e-posta adresinin kayıtlı olup olmadığı anlaşılamaz. Sahte işlem de semafordan geçmeli.
```go
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("us-dummy-not-a-password"), 11)
if !found { _ = bcrypt.CompareHashAndPassword(dummyHash, pw); return errBadLogin }
```

## (e) Riskler ve öneri
- Semafor kapasitesi çok düşük olursa 503'ler artar. Değer ölçümle ayarlanmalı (2 vCPU için 2). Login rate limit ile birlikte tutulmalı.
- Cost 11 kaba kuvvet saldırısını yaklaşık 2 kat ucuzlatır ama güncel kabul edilebilir alt sınırın üstünde kalır. Cost 10'a inilmemeli.
- Yeniden hash yazma işlemi store kilidi altında çalışmalı.
- **Öneri:** 1) VPS'teki load 26'nın kaynağını bulup giderin; asıl sorun bu. 2) Semafor (2) ve 503/Retry-After ekleyin; istemci koptuğunda işi iptal edin. 3) Sahte hash ile karşılaştırmayı ekleyin. 4) Yeni hash'lerde cost 11'e geçin, yeniden hash'i sonraya bırakın. Hepsi licensesrv değişikliğidir: KAN.US'un incelemesi ve Demir'in onayından sonra test hook'una uygulanmalı.
