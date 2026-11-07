# §7 — KANUS SaaS Project Manager (7 Kasım2026 promptu)

US Panel .11 ajanıyla sıfırdan çok dosyalı KANUS SaaS Project Manager üret. Backend API, frontend, PostgreSQL, kullanıcı girişi, rol tabanlı erişim, proje yönetimi, görev yönetimi, dosya yükleme, yönetim paneli, API dokümantasyonu, Docker, otomatik testler ve kurulum kılavuzu zorunlu.

Planla → dosyaları gerçekten yaz/uygula → doğru araçla derle → çalıştır → test et → hata kayıtlarını ajana geri ver → düzelt → yeniden test et. Sonunda çalışan uygulama, test kayıtları ve kaynak dosyaları bulunmalı. Yalnız model cevabı/kod metni veya mock fixture PASS, gerçek §7 kabulü değildir.

Kabul matrisi:
- Backend/frontend build ve boot; tarayıcıda frontend→API akışı.
- PostgreSQL migration; restart sonrası kalıcılık; rollback planı.
- Kullanıcı girişi; secretsiz log; eksik auth401.
- Member/admin RBAC; yetkisiz admin403; yönetim paneli fonksiyonları.
- Proje/görev create/list/update/delete; tenant ve eşzamanlı kullanıcı izolasyonu.
- Dosya yükleme boyut/tür sınırı, path traversal reddi, yetkili indirme ve tenant izolasyonu.
- OpenAPI ve gerçek route uyumu; Docker config/build/run; otomatik testler; yeniden üretilebilir kurulum kılavuzu.
- Hata→ajan düzeltme→yeniden build/test kanıtı; test sonuçları ve dosya manifesti.

Prompt §7 sayısal token/tur/süre/dosya bütçesi veya SLO belirtmiyor. Önceki2tur/100dosya/30dk/$3 önerileri prompt eşiği olarak gösterilmemeli; gerçek koşudan önce ayrı onaylı sınırlar belirlenmeli. .11deploy+GPU penceresi bekleniyor: gerçek model adımları BLOCKED. Bu belge model/GPU çağrısı veya deployment yetkisi değildir.

Güncel idle kabul kriteri: policy v5,600sn/10dk; Demir8Kasım2026 00:20TSİ kararı. Aktif/queued iş varken kapanma yok; son işten600sn sonra kapanış sağlayıcı API teyidiyle. Gerçek GPU testi bu mock hazırlığıyla doğrulanmış sayılmaz.
