# FX08 bağımsız bulgular (kalıcı ara not)
08 yalnız i18n/fx.go ve fx_test.go; snapshot SHA6ac3588b… doğrulandı. v10 tam race425/0/7 PASS; v11 devam ediyor.
MEDIUM FX-01 fx.go:46-52: tarih UTC parse, alt eşik -36h, üst eşik5gün+24h. 6gün önce UTC00 ve TSİ00 ile gelecek12h/36h kabul:4 negatif testFAIL. Takvim günü/TCMB zaman dilimi açıkça belirlenmeli; bugünün tarihi yerel00:01'de yanlış reddedilmemeli ve5.takvim gününün sonu korunmalı.
LOW FX-02 fx.go:84-94: 29kuruş/2 oranı0.145 birim→0.14 yerine yarım yukarı0.15 beklenir;21/6→0.03 yerine0.04. MaxInt64 ve parse-geçerli1/1 negatif gösterim üretir. Ayrıca elle yapılmış NaN/Inf snapshot gizlenmiyor (ParseFX zaten bunları reddediyor; yalnız public helper defansif notu).
Kendi39 yaprakvaka30PASS/9FAIL;4tarih,2rounding,3sağlamlık. Native snapshot/gerçekTCMB verisi değil, testfixturedır.
Windows r4 harici23/0/1 hamlog;Scaling +6/+2 (aktarılan+0 yanlış). CI binary166d… hash doğru; yeni counter/test sembolleri var. Logdosyalarında actual/expected hashstdout, RESULT exit ve build-source manifesti yok. Runner1skipte2PARTIALbekler;gerçekexitgözlenmedi. KendiWindowskoşusuNOTTESTED.
