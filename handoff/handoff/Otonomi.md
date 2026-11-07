# Otonomi — Devir Dosyası (8 Kasım 2026)

## 1) Kimlik
- **Ad:** Otonomi
- **Başlık:** Otonomi
- **Açıklama:** Otonomi kavramını düzenli araştırıp Demir'in VPS'teki turkdini programını bu kavrama göre geliştiren asistan.
- **Rol:** Otonomi / ajan yönetişimi araştırması + turkdini'nin (turkdini.com, Rahim adlı YZ kimliğinin yönettiği Türk dinleri yayın platformu) kararlılık, denetim ve sürümleme iyileştirmeleri.
- **Sistem talimatı taslağı:**
  > Sen Otonomi'sin. Demir ile Türkçe, sıcak ve kısa konuş. Görevin: otonomi kavramını (otonomi seviyeleri, ajan yönetişimi, denetim izi, güvenli öz-düzeltme) düzenli araştırmak ve Demir'in VPS'indeki turkdini programını buna göre geliştirmek. Demir'in kuralı: turkdini'nin otonomi seviyesini DÜŞÜRECEK hiçbir değişiklik yapma (onay kapısı ekleme, humanReviewRequired'ı açma, kabuk/kaynak mutasyonunu kapatma, AUTONOMY_MAX/kill switch'i varsayılan etkin yapma, fullAutonomyEnabled'ı düşürme); bunları yalnızca öneri olarak raporla. Seviyeyi düşürmeyen işleri (yedek, kararlılık, kuyruk, denetim izi, sürümleme) önce yedek alarak doğrudan yapabilirsin. Sır değerlerini asla yazma. SSH'ta tek bağlantı kullan, 30 sn arayla az sayıda yeniden dene. Her değişiklik için geri alma adımı yaz.

## 2) Hafıza
- (2026-10-24, profile) VPS: root@31.220.63.191 (hostname srv1662550, Hostinger). Kutudan SSH, özel anahtar `~/.ssh/id_ed25519_otonomi` ile; kutuda ssh ikilisi yok, `/workspace/venv` içindeki paramiko kullanılıyor. Sunucudaki projeler: /opt/growth-os, /opt/turkdini, /opt/tradebot*. Demir turkdini'yi seçti.
- (2026-10-24, profile) Demir yetkisi: turkdini'de otonomi seviyesini düşürmeyen her eylem serbest; onay kapıları, kabuk/kaynak yetkisi kısıtlaması ve seviye tavanına dokunulmaz. (Önceki "yalnızca inceleme" talimatının yerine geçti.)
- (2026-10-24, olay) İlk inceleme: turkdini fiilen L5 tam otonomi (`rahim-governance.mjs`: humanReviewRequired:false, sourceMutationAllowed:true, arbitraryShellAllowed:true). Canlı kod /opt/turkdini/vps/server.mjs + ~20 rahim-*.mjs; servis turkdini.service (kullanıcı turkdini, port 3000); veri /var/lib/turkdini/data.json (122 MB) ve state.sqlite (413 MB); modeller yerel Ollama.
- (2026-10-25, olay) Uygulanan değişiklikler: yedek /root/turkdini-autonomy-backup-20261024-211212/; heap 2560→2816 MB, MemoryHigh 3000M, MemoryMax 3300M (override.conf); kuyruk bekleme dua 30 sn→5 dk, methiye 60 sn→10 dk (server.mjs); saatlik hash zincirli denetim izi /usr/local/sbin/turkdini-audit-chain.py → /var/lib/turkdini-audit/chain.jsonl; günlük 04:30 SQLite yedeği /usr/local/sbin/turkdini-sqlite-backup.sh → /var/backups/turkdini/ (son 7); 119 .bak → /opt/turkdini/archive-bak-20261025/; git init, ilk commit c55785f.
- (2026-10-25, olay) Heap OOM kök neden analizi (sunucuya uygulanmadı): /workspace/turkdini-proposal/README.md. Tek JSON blob, `developmentEvaluation.cases` tekrarları; önerilen yamalar 01, 02, 04.
- (2026-10-26, olay) 1. haftalık rapor: 23–25 Eki 26 çökme, swap dolu, Ollama 13 OOM-kill.
- (2026-11-02, journal) 2. haftalık rapor: 30 Eki 22:06–22:51 İstanbul, AWS IP'lerinden bizim olmayan root anahtarıyla (SHA256:zwwcGE5N…) giriş; override.conf bellek sınırları düşürülmüş (heap 2816→2048, MemoryMax 3300→2400M, MemoryHigh 3000→2000M) ve server.mjs'te commit'lenmemiş 9 satır. Çökmeler 26→151. data.json günlük yedeğe eklendi; eski betik /root/turkdini-ops-backups/20261102T061649Z/turkdini-sqlite-backup.sh.orig.

## 3) Rutinler
- **Ad:** Haftalık otonomi araştırması ve turkdini özeti
- **Zamanlama:** Her Pazartesi 09:00 Europe/Istanbul (cron `0 9 * * 1`, esnek ±15 dk)
- **Kayıtlı prompt:**
  > Demir'in isteği: otonomi kavramını düzenli öğren ve turkdini programını buna göre geliştir. Her çalışmada: (1) otonomi, ajan yönetişimi ve güvenli otonom sistemler üzerine yeni kaynakları kısaca tara; (2) VPS'e root@31.220.63.191 adresine, kutudaki /home/box/.ssh/id_ed25519_otonomi anahtarı ve /workspace/venv/bin/python içindeki paramiko ile TEK bağlantıyla bağlanıp (zaman aşımı olursa 30 sn arayla az sayıda yeniden dene) turkdini'nin son 7 günlük servis, çökme ve hata özetini SALT OKUNUR çıkar (sır değerlerini asla yazma); (3) /workspace/turkdini-otonomi-raporu.md planındaki ilerlemeyi ve 1-2 yeni öneriyi ekle; (4) Demir'e Türkçe kısa özet gönder. Demir'in kuralı: otonomi seviyesini DÜŞÜRECEK hiçbir değişiklik yapma (onay kapısı, kabuk/kaynak yetkisini kapatma, seviye tavanı gibi); bunlar yalnızca öneri olarak raporlanır. Seviyeyi düşürmeyen küçük iyileştirmeler için önce yedek al. Hepsini iş delege ederek yap.

## 4) Beceriler, bağlayıcılar, servisler
- Beceriler: routines, canvas (kullanılmadı)
- Bağlayıcılar: yok (Hostinger bağlayıcıları kurulu ama kullanılmadı)
- Servisler: SSH (paramiko), WebSearch/WebFetch, Hostinger VPS, Ollama (sunucuda)

## 5) Açık işler ve kalınan yer
- 8 Kasım 00:07 itibarıyla VPS'e SSH zaman aşımı veriyordu (sunucu yükü yüksek, aralıklı). Mevcut durum doğrulanamadı.
- Bellek sınırlarının 25 Ekim değerlerine geri alınması Demir'in cevabını bekliyor.
- Heap OOM kök çözümü (yamalar 01, 02, 04) hazır, uygulanmadı.
- Önerilen izleyici: config/kod değişikliği veya saatte >3 çökmede uyarı (henüz kurulmadı).
- Kuyruk backoff (5→10→20→60 dk, günlük tavan) yapılmadı.
- `rahim-continuous-growth.mjs:310` yeni .bak.uimicro-* dosyaları üretiyor, arşive taşınmalı.
- Methiye hata mesajında yanlış etiket ("Mantra-Matik").

## 6) Bekleyen Demir kararları
1. 30 Ekim'deki değişikliği (AWS IP'leri, anahtar SHA256:zwwcGE5N…) Demir ya da kendi aracı mı yaptı? Değilse yetkisiz erişim olabilir, root authorized_keys kontrol edilmeli.
2. Onaylarsa: override.conf ve server.mjs farkını yedekleyip git'e kaydet, heap 2816 / MemoryHigh 3000M / MemoryMax 3300M + `--report-on-fatalerror`, daemon-reload, turkdini restart.
3. Heap OOM yamaları 01, 02, 04'ün uygulanması (rollout adımları README'de).
4. Uyarı izleyicisinin kurulması.

## 7) Kritik dosya yolları ve sha256
Kutu (/workspace):
- turkdini-otonomi-raporu.md — 203eb11ad057536da412cd535c758682e39edd61addeb770d81a47719f0a0fdc
- turkdini-degisiklikler.md — d38386e6d788b751577fa70704410a407799a53ea413dc4577a0a6f2484597bd
- turkdini-proposal/README.md — e3054c40b876f01754ee16f4a8e583c18efe8a09bbd21e5378d25e0058524124
- turkdini-proposal/01-compact-evaluation-cases.patch — 2f8870e2060783dd5d0605a183a0b0a691581df6e1e5bc92675c5802d46cc25c
- turkdini-proposal/02-rollback-drill-shallow-clone.patch — 7e7fcdca570afd9945c13625dd3eec26fa67e04f331d054d1232201df89e055a
- turkdini-proposal/03-readstore-structuredclone-OPTIONAL.patch — 4c78fd6ad8d1bf6b0b4ec8ab7c56bfd6399352cb3f2dbb30be49847cdea03692
- turkdini-proposal/04-datajson-checkpoint-hourly.patch — 23422a84f2e0e5fff858e5ceb9b979c9a53d5fe6315bf9098124f1fc6f258301
- wk1102/ (2 Kasım ham çıktılar), turkdini-otonomi-raporu.md.bak-2026-10-26, .bak-2026-11-02
Sunucu:
- /opt/turkdini/vps/server.mjs, rahim-governance.mjs, rahim-autonomy-kernel.mjs; git commit c55785f
- /etc/systemd/system/turkdini.service.d/override.conf
- /var/lib/turkdini/data.json, state.sqlite
- /root/turkdini-autonomy-backup-20261024-211212/ ; /root/turkdini-ops-backups/20261102T061649Z/
- /usr/local/sbin/turkdini-audit-chain.py, turkdini-sqlite-backup.sh ; /var/lib/turkdini-audit/chain.jsonl ; /var/backups/turkdini/
- /opt/turkdini/archive-bak-20261025/MANIFEST.txt

## 8) Yeniden girilmesi gereken gizli değerler (yalnız adlar)
- **SSH özel anahtarı** `id_ed25519_otonomi` — VPS root erişimi. Yeni hesapta yeni anahtar çifti üretilip açık anahtar (`otonomi-box`) sunucudaki /root/.ssh/authorized_keys'e eklenmeli; eskisi kaldırılabilir.
- Sunucudaki admin.key ve .env dosyaları — turkdini'nin kendi sırları; sunucuda duruyor, taşınmaya gerek yok.
