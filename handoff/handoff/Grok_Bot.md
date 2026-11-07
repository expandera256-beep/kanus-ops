# Grok Bot devir dosyası

Hazırlanma: 8 Kasım 2026, 01:45 civarı (İstanbul saati). Hazırlayan: Grok Bot (agent id 63d1fe7e-27cc-48d6-85f9-114907613902).
Bu dosyada hiçbir gizli değer yoktur. Yalnızca adlar ve yollar var.

## 1) Kimlik

- **Ad:** Grok Bot (Demir bu adı kendisi seçti, korunmalı)
- **Başlık:** Grok Bot
- **Açıklama:** A teammate with its own computer that takes real work off your plate.
- **Rol:** Demir'in ana genel amaçlı asistanı. Ağırlıklı iş: turkdini VPS (31.220.63.191) üzerindeki Node servisinin bellek/çökme sorunlarını incelemek, çözüm önermek, salt-okuma izleme yapmak. Ayrıca Hanedan projesi için yapılmış bitmiş bir yama var.

### Sistem talimatı taslağı

> Sen Grok Bot'sun, Demir'in sıcak ve kısa konuşan masaüstü asistanısın. Demir'le her zaman Türkçe konuş, uzun tire kullanma. Saatleri İstanbul saatiyle (UTC+3) ver.
> Ana sorumluluğun turkdini VPS'idir (31.220.63.191, Node servisi `/opt/turkdini/vps/server.mjs`, systemd). Kesin kural: Demir açıkça onaylamadan sunucuda hiçbir şeyi değiştirme, servisi yeniden başlatma; tradebot'lar dahil başka servislere dokunma. Staging testleri izole olmalı ve canlı sistemi asla yavaşlatmamalı; yavaşlatırsa hemen durdur.
> Sunucuyu salt-okuma ile izleyebilir, bulguları kısa ve sayılarla raporlayabilirsin. Aynı onay sorusunu tekrar tekrar sorma; yalnızca anlamlı yeni bulgu varsa haber ver.
> Değişiklik yapmadan önce mevcut durumu yedekle ve git'e commit'le. Yaptığın her değişikliği `/workspace/turkdini-degisiklikler.md` dosyasına kaydet.

## 2) Hafıza

Profile (hepsi 2026-10-24):
- Kullanıcı Türkçe iletişim kurar.
- Kullanıcı asistanın adının "Grok Bot" olmasını istiyor.
- Kullanıcının adı Demir.

Journal ve paylaşılan hafıza: kayıt yok. Aşağıdaki çalışma geçmişi sohbetten derlendi:
- 2026-10-24: Onboarding yapıldı. Gmail eklentisi (plugin id 45893410) kuruldu, bağlantı kartı gösterildi, Demir yetkilendirmedi.
- 2026-10-25: turkdini'de onaylı değişiklikler yapıldı: MemoryHigh/MemoryMax, heap 2816 MB, kuyruk zaman aşımları, denetim zamanlayıcısı, 01:30 UTC'de `/var/backups/turkdini/` altına SQLite yedeği (commit c55785f). Ayrıntı `/workspace/turkdini-degisiklikler.md` dosyasında.
- 2026-10-25/26: Çökme kök nedeni bulundu: tüm store'un `cloneJson` ile kopyalanması ve her yazmada bütün store'un serileştirilmesi. Store tek bir JSON (`state.sqlite`, aynası `data.json` 122 MB). Bunun yaklaşık 100 MB'ı türetilmiş `developmentEvaluation.cases` verisi ve üç kez tutuluyor. Önerilen yamalar: 01, 02, 04. Hiçbiri uygulanmadı.
- 2026-10-26: VPS'te ayrı portta yapılan staging testi 11 dakikada durduruldu, çünkü canlı sistemi yavaşlattı. Snapshot 122 MB'tan 37 MB'a indi ama RSS 1,35 GB'a çıktı. 05:15'teki tek seferlik rutin, sunucu aşırı yüklü olduğu için testi atladı ve kendini sildi. Başka makinede staging de engellendi: box'ta yalnızca yaklaşık 2,5 GB RAM var, swap yok ve Node 20 kurulu (VPS'te Node 22.22.3).
- 2026-10-30 22:16 İstanbul: Kayıt dışı bir değişiklik limitleri heap 2048 / MemoryHigh 2000M / MemoryMax 2400M'e düşürdü ve commit'lenmemiş bir `server.mjs` değişikliği bıraktı. SSH oturumları AWS IP'lerinden geliyordu, muhtemelen Demir'in kendi ajanlarından biri. Tanınmayan anahtar parmak izi `SHA256:zwwcGE5N…`. Bu değişiklikten önceki 42 saatte hiç çökme yoktu.
- 2026-11-02 ve sonrası: Çökme döngüsü başladı. 2 Kasım'dan beri yaklaşık 256 ABRT, NRestarts 275'e çıktı (5 Kasım 14:10 itibarıyla). Swap zaman zaman 4/4 GB doldu, `/api/health` zaman aşımına uğradı, nginx eski önbelleği sunuyor. Python süreçleri (muhtemelen tradebot'lar, tradebot-okx) 3 GB'tan fazla bellek kullanıyor.
- 2026-11-02 rapor: `/workspace/turkdini-otonomi-raporu.md`. Rahim kendi işini kendisi onaylıyor (L5). Rapor bir `/etc/turkdini/PAUSE` durdurma anahtarı ve D.1 bölümünde bellek ayarının geri alınmasını öneriyor.

## 3) Rutinler

Kayıtlı rutin yok. 26 Ekim 05:15'teki tek seferlik staging rutini çalıştıktan sonra kendini sildi. Sistemin kendi "Periodic proactive work" otomasyonu salt-okuma kontrolleri yapıyor, ama bu bir kullanıcı rutini değil.
Önerilen ama kurulmamış rutin: haftalık pazartesi 09:00 turkdini kontrolü (Demir'in onayını bekliyor).

## 4) Beceriler, bağlayıcılar, servisler

- Beceriler: onboarding-bot, onboarding-email-ask-v4, onboarding-connections-v4, routines, box-desktop, unblock-research
- Bağlayıcılar: Gmail eklentisi (kurulu, yetkisiz), cursor-github, cursor-origin, Hostinger (hosting/vps/dns/domains ve diğerleri; execute zaman aşımına uğruyordu), x, x-drafts, user-Shopify (bu botta kullanılmadı)
- Servisler: turkdini VPS (SSH, root, 31.220.63.191), systemd, nginx, SQLite, Node 22

## 5) Açık işler ve kalınan yer

- turkdini çökme döngüsü sürüyor. Son kontrol 5 Kasım 14:10: limitler hâlâ 2048 / 2000M / 2400M, son çökme 12:20. Sunucuda hiçbir şey değiştirilmedi.
- Onay gelirse yapılacak plan: (a) mevcut `override.conf` ve commit'lenmemiş `server.mjs` farkını `/root/turkdini-ops-backups/` altına yedekle ve commit'le; (b) 25 Ekim değerlerine dön: heap 2816 MB, MemoryHigh 3000M, MemoryMax 3300M; (c) çökme raporu bayrağını ekle; (d) servisi bir kez yeniden başlat ve doğrula; (e) `/workspace/turkdini-degisiklikler.md` dosyasına kaydet.
- Uzun vadeli kalıcı çözüm: 01, 02 ve 04 yamaları (03 isteğe bağlı). Önce izole staging testi gerekiyor. Sunucunun genel bellek bütçesi tradebot'larla birlikte ele alınmalı.
- Araç notu: `/workspace/tdr/ssh_run.py` bozuk (paramiko eksikti). Çalışan yardımcı `/workspace/wk1102/run.py`.

## 6) Bekleyen Demir kararları ve onaylar

1. Bellek limitlerini 25 Ekim değerlerine geri alıp turkdini'yi bir kez yeniden başlatma onayı. 2, 5 ve 6 Kasım'da birkaç kez soruldu, yanıt gelmedi. Tekrar sorarken yeni veri ekle, aynı soruyu boş yere yineleme.
2. PAUSE durdurma anahtarı ve/veya haftalık pazartesi 09:00 kontrol rutini.
3. Staging yöntemi için seçenek 1-4 seçimi.
4. Gmail yetkilendirmesi (isteğe bağlı).
5. 30 Ekim değişikliğini hangi ajanın yaptığını araştırma (seçenek olarak sunuldu).

## 7) Kritik dosya yolları ve sha256 değerleri

Box (`/workspace`):
- `turkdini-degisiklikler.md` d38386e6d788b751577fa70704410a407799a53ea413dc4577a0a6f2484597bd
- `turkdini-otonomi-raporu.md` 203eb11ad057536da412cd535c758682e39edd61addeb770d81a47719f0a0fdc
- `hanedan-tum-fazlar.patch` (Hanedan, 44 commit, testler 165/165) 59354d0c825df51b44cc97f2c57b4e76af5c08fdc97ae01f66485c511335fac4
- `turkdini-proposal/01-compact-evaluation-cases.patch` 2f8870e2060783dd5d0605a183a0b0a691581df6e1e5bc92675c5802d46cc25c
- `turkdini-proposal/02-rollback-drill-shallow-clone.patch` 7e7fcdca570afd9945c13625dd3eec26fa67e04f331d054d1232201df89e055a
- `turkdini-proposal/03-readstore-structuredclone-OPTIONAL.patch` 4c78fd6ad8d1bf6b0b4ec8ab7c56bfd6399352cb3f2dbb30be49847cdea03692
- `turkdini-proposal/04-datajson-checkpoint-hourly.patch` 23422a84f2e0e5fff858e5ceb9b979c9a53d5fe6315bf9098124f1fc6f258301
- `turkdini-proposal/README.md` e3054c40b876f01754ee16f4a8e583c18efe8a09bbd21e5378d25e0058524124
- `turkdini-proposal/orig/`, `new/`, `evidence/` (`staging-monitor.log`, `staging-precheck-20261026.log`)
- `wk1102/run.py` (SSH yardımcısı) d88c6bebcabdf4fec84a1d07c3ee1237bec87ef7e52cce2ddfc739503b8d3f62
- `wk1102/chk.sh` c3f02b617a34823ef5bf2b53563a31c68b839ae7b1217bf9305c07746d357931
- `wk1102/chk2.sh` 36450a5c6777f4b00b983409ae335e5806659588bb74437812ebd84f2c089fa8
- `wk1102/newbackup.sh` 82f3cff47b19d53459497c630b4c399cd9bd5714b4d796a4136430e2cee7015b

VPS (31.220.63.191):
- `/opt/turkdini/vps/server.mjs`, `storage.mjs`, `state.sqlite`, `data.json`
- systemd `override.conf` (turkdini servisi), `/var/backups/turkdini/`, `/root/turkdini-ops-backups/`
- Bilinen commit: c55785f (25 Ekim değişiklikleri)
- Önerilen: `/etc/turkdini/PAUSE` (henüz yok)

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)

- SSH özel anahtarı `~/.ssh/id_ed25519_otonomi`: turkdini VPS'e root SSH erişimi için (`wk1102/run.py` bunu kullanıyor). Box ortak olduğu için aynı box kullanılacaksa dosya zaten orada.
- `~/.ssh/id_ed25519_ticari`: Ticari botunun anahtarı, bu bot kullanmadı.
- Gmail OAuth bağlantısı: hiç yetkilendirilmedi.
- Hostinger bağlayıcı yetkilendirmesi: Hostinger araçları için.
- GitHub / Cursor Origin bağlantısı: cursor-github ve cursor-origin için.
