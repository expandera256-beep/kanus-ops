# release_us_fx — yaklaşık döviz gösterimi (YEREL, YÜKLENMEDİ)
Demir kararı: "TL yanında yaklaşık döviz de gösterilsin". Kaynak TCMB today.xml ForexSelling (EUR, USD). Tarayıcı yalnız /fx.json okur; TCMB'ye istek atmaz.
Bileşen: `data-kanus-try="<TL tutarı>"` taşıyan her öğeyi "₺… (≈ € … · ≈ $ …)" + 4 dilde not ("Yaklaşık tutar; ödeme TL ile alınır. Kur: TCMB döviz satış, <tarih>.") olarak gösterir. fx.json yok/bozuk/eski (>5 gün) ise yalnız TL gösterilir.
Fiyat verisi: kanus.org SPA'sında şu an render edilen fiyat YOK (payments rotası /us'a yönlenir; fiyatlar shop.kanus.org/Shopify'da). Bu yüzden bileşen hazır ama sitede şu an görünür bir etkisi yok; uydurma fiyat eklenmedi.
Ödeme ve ledger koduna dokunulmadı; ödeme TL.
Günlük güncelleme: aşağıya bakın.
Yeniden üretim: `python3 site/us/fx_snapshot.py /tmp/fx.json && python3 site/us/build_us_fx.py /tmp/fx.json`

## Günlük güncelleme
Demir 00:48 TSİ: KAN.US hPanel'de hafta içi 16:05 cron kuruyor. Kurulum: `cron/INSTALL_CRON.md` (Seçenek A: `cron/upload_public_html/_fx/` → public_html/_fx/).
