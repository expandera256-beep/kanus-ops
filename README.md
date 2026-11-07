# KAN.US geliştirme ve operasyon arşivi

Demir’in KAN.US geliştirmeleri, raporları ve durum kayıtları için **özel (PRIVATE)** depo. Bu bir canlıya alma veya satış onayı değildir. Kaynak klasörler değiştirilmeden, 8 Kasım 2026 tarihli bir anlık kopya hazırlanmıştır (TSİ, Europe/Istanbul). Eşzamanlı geliştirmeler sürüyor; bu depo çalışma sırasında alınan arşivdir, atomik yayın paketi değildir.

## Klasörler

- `us-panel/`: kaynaklar, `prompt/`, `deploy/`, `reports/`, `qa/`, planlar ve operasyon kanıtları.
- `us-panel-dev/`: geliştirme kopyaları, `patches/`, incelemeler, durumlar ve `global-work/site/` site paketleri.
- `workspace-reports/`: çalışma alanının kökündeki ilgili ek KAN.US raporları.
- [EXCLUDED_ARTIFACTS.md](EXCLUDED_ARTIFACTS.md): yüklenmeyen derlemelerin/büyük dosyaların yol, boyut ve SHA256 listesi.
- [ARCHIVE_SECURITY.md](ARCHIVE_SECURITY.md): sır taraması ve bilinçli kapsam eksikleri.

## Son durum — 8 Kasım 2026 (TSİ)

- **.13 geliştirme ve düzeltme aşamasında**; .11 ve .12 FAIL. .13 için güvenlik PASS veya deploy onayı varsayılmamalı. Windows dosya katmanı ve kredi/GPU kontrollerinin yeniden doğrulanması bekleniyor.
- **api.kanus.org:** HOST A seçeneği; mevcut **82.29.184.225** üzerinde ayrı üretim kurulumu. Canlı secrets/planlar, ledger/rollback uyumu, SMTP teslim kabulü ve ilgili onaylar tamamlanmadan satışa hazır sayılmaz.
- **FX cron:** canlı; hafta içi **16:05 TSİ** TCMB kuru güncellemesi.
- **Panel → lisans sunucusu TLS:** pin yok, sistem güven deposuyla standart sertifika ve hostname doğrulaması. Bu karar ayrı GPU wake bağlantısının mevcut PEM doğrulamasını kaldırmaz.
- **Görsel motoru:** FLUX.2 klein 4B seçildi; video daha sonra. GPU açılışı, ölçüm ve kredi oranları ayrı onay gerektirir.

## Başlıca rapor ve planlar

- `us-panel/reports/r20261108/KANUS_CURRENT_STATE.md`
- `us-panel/reports/r20261108/KANUS_SECURITY_REPORT.md`
- `us-panel/reports/r20261108/SHOP_LIVE_PRESALE_CHECKLIST_20261108.md`
- `us-panel/reports/r20261108/KAN_IMAGE_ENGINE_PLAN.md`
- `us-panel/reports/r20261108/GLOBAL_FX_UPLOAD_20261108.md`
- `us-panel/deploy/hook-move/GOLIVE_PLAN_APIKANUS_20261108.md`
- `us-panel-dev/PANEL_R13_WINDOWS_REVIEW_20261108.md`
- `us-panel-dev/PANEL_DEV_STATUS.md`

Bazı kaynak belgeler eski veya birbiriyle çelişen kararları yansıtabilir; yukarıdaki özet arşivleme sırasındaki son kullanıcı kararlarını belirtir. Sır şüphesi taşıyan dosyalar yüklenmez; bir raporun bulunmaması güvenlik filtresinden kaynaklanabilir. Bu kopya sırlar ve ikili çıktılar olmadan tek başına yeniden derlenebilir dağıtım garantisi vermez.
