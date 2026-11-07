# §11 — gerçek US Panel ajanı proje görevi (henüz modele gönderilmedi)

Çok kiracılı, çok kullanıcılı SaaS Proje Yönetim Sistemi oluştur. Backend: Python FastAPI; frontend: TypeScript/React; PostgreSQL; migration; Docker Compose; CI ve README. Kimlik doğrulama, admin/member RBAC, tenant izolasyonu, proje/görev REST CRUD, giriş doğrulama ve hata yanıtları ekle. Şifreler güvenli hash; parola/token kaynakta/logda yok. SQL parametreli; tenant kimliği istemciden güvenilir kabul edilmez.

İstenen dosyalar: backend/app/*, backend/tests/*, backend/migrations/*, frontend/src/*, frontend testleri, bağımlılık lock dosyaları, Dockerfile'lar, compose.yaml, CI yaml, .env.example (yalnız boş/değişken isimleri), README ve OpenAPI açıklaması.

Plan → dosyaları gerçekten yaz → build → test → hata analizi → düzelt → tekrar test. Gerçek stdout/stderr kanıtı sun; yapılmayan işi tamamlandı deme. Doğrulayıcı hatalarını ikinci turda kullan. İlk turdan sonra en çok bir düzeltme turu; toplam iki model turu. En çok100 dosya, toplam30dk aktif GPU/üretim-doğrulama, compute bütçesi $3 (öneri, onay bekliyor). Toplam input/output token bütçesini mevcut model kapasitesi ve fiyatlandırmayla koşu öncesi sabitle; sınırsız devam etme. Sınır aşımında dur ve kalan işi raporla.

Kabul: frontend/backend build; compose config; sağlıklı PostgreSQL; migration ve restart sonrası kalıcılık; kayıt/giriş/çıkış; admin/member red/allow; görev CRUD; iki tenant arasında liste/detail/update/delete izolasyonu; geçersiz girdiler; auth eksik401 ve yetki eksik403; otomatik testler; CI; yeniden üretilebilir README. UI'ı gerçek tarayıcı ve talep edilen EXE sürümüyle ayrıca doğrula. finish_reason=length görev tamamlandı kanıtı değildir.

10 dakika (600sn) idle, policy v5: Demir8Kasım2026 00:20TSİ kararı; önceki15dk kararının yerini alır. Aktif veya kuyruktaki görev kapanmayı engellemeli; son işten sonra600sn timer; sağlayıcı API'sinden gerçek stop doğrulanmalı. Bu dosya GPU başlatma/ödeme/dağıtım yetkisi vermez.
