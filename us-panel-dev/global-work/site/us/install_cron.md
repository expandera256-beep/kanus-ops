# fx_cron kurulumu (KAN.US; Demir 00:48 TSİ onayı)

## Önerilen: Seçenek A — public_html/_fx/
Neden: KAN.US'in Hostinger connector'ı yalnız public_html altına yükleyebiliyor; A connector ile yüklenip doğrulanabilir. İki kat koruma: `_fx/.htaccess` web erişimini reddeder (Apache 2.4 `Require all denied`, eski/LiteSpeed `Deny from all`), .htaccess atlansa bile script `PHP_SAPI !== 'cli'` ise 403 döner ve hiçbir şey yazmaz/kilitlemez.

Yüklenecekler (`upload_public_html/` içeriği → sunucu `public_html/`):
- `_fx/.htaccess`
- `_fx/fx_cron.php`

hPanel → Gelişmiş → Cron Jobs → Özel: dakika `5`, saat `16`, gün `*`, ay `*`, haftanın günü `1-5`, komut:
`/usr/bin/php /home/u346856246/domains/kanus.org/public_html/_fx/fx_cron.php`
(hPanel PHP yolunu farklı gösterirse onun önerdiği `php` yolunu kullanın; sunucu saat dilimi TSİ değilse saati ona göre çevirin.)

Yükleme sonrası kontrol (salt okuma):
- `curl -sI https://kanus.org/_fx/fx_cron.php` → 403 (veya 404), asla 200 değil
- `curl -sI https://kanus.org/_fx/fx_cron.lock` → 403
- Cron ilk çalıştıktan sonra `curl -s https://kanus.org/fx.json` → `tcmb_tarih` güncel

Çıkış kodları: 0 yazıldı · 2 geçersiz/eski/ağ hatası (fx.json'a dokunulmaz) · 3 başka kopya çalışıyor. Çıktıda sır yok.

## Seçenek B — public_html dışı (~/domains/kanus.org/fx/)
hPanel Dosya Yöneticisi ile elle konabilir; ancak connector bu yola ne yükleyebilir ne de listeleyebilir → **doğrulanamaz**. Script şu an `_fx/` içinden çalışmayı şart koşar; B seçilirse küçük bir yol değişikliği gerekir. Önerilmez.
