# release_us_fx2 — (release_us_fx düzeltmesi, PANEL DEV R4B FX01/FX02) yaklaşık döviz gösterimi (YEREL, YÜKLENMEDİ)
Demir kararı: "TL yanında yaklaşık döviz de gösterilsin". Kaynak TCMB today.xml ForexSelling (EUR, USD). Tarayıcı yalnız /fx.json okur; TCMB'ye istek atmaz.
Bileşen: `data-kanus-try="<TL tutarı>"` taşıyan her öğeyi "₺… (≈ € … · ≈ $ …)" + 4 dilde not ("Yaklaşık tutar; ödeme TL ile alınır. Kur: TCMB döviz satış, <tarih>.") olarak gösterir. fx.json yok/bozuk/geçersiz tarihli (gelecek ya da TSİ takvimiyle bugün-6 ve öncesi) ise yalnız TL gösterilir.
Fiyat verisi: kanus.org SPA'sında şu an render edilen fiyat YOK (payments rotası /us'a yönlenir; fiyatlar shop.kanus.org/Shopify'da). Bu yüzden bileşen hazır ama sitede şu an görünür bir etkisi yok; uydurma fiyat eklenmedi.
Ödeme ve ledger koduna dokunulmadı; ödeme TL.
Günlük güncelleme: aşağıya bakın.
Yeniden üretim: `python3 site/us/fx_snapshot.py /tmp/fx.json && python3 site/us/build_us_fx.py /tmp/fx.json`

## Günlük güncelleme
Demir 00:48 TSİ: KAN.US hPanel'de hafta içi 16:05 cron kuruyor. Kurulum: `cron/INSTALL_CRON.md` (Seçenek A: `cron/upload_public_html/_fx/` → public_html/_fx/).

## fx2 farkları
- Tazelik: TCMB tarihi Türkiye (UTC+3) takvim günü; bugün-5..bugün kabul, gelecek ve bugün-6 red (panel i18n.FXFresh, fx_snapshot.py, fx_cron.php ile aynı).
- Yuvarlama: kuruş tam sayı, BigInt ile kesin bölme, yarım sentte sıfırdan uzağa; 10^15 kuruş üstü gizlenir.
- Canlı i18nfx-5a375f29 eski kurallara sahip (1 gün gelecek ve 6. gün kabul, float yuvarlama); sitede `data-kanus-try` öğesi olmadığından görünür etki yok.
