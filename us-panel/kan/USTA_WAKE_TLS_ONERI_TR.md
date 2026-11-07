# usta-wake /a100/* güvenlik önerisi (UYGULANMADI)
Tarih: 06.11.2026 09:40 TSİ. Hazırlayan: KAN. Durum: öneri, onay bekliyor.

## Mevcut durum (salt okuma)
- usta-wake (34.41.186.170) `kalfa-http` etiketli. `kalfa-allow-http` tcp:80 ve `kalfa-allow-worker` tcp:8788 kurallarının ikisi de 0.0.0.0/0'a açık.
- /a100/start ve /a100/status düz HTTP :80 üzerinden çalışıyor. Bearer token şifresiz gidiyor; uygulama yalnız 31.220.63.191'i kabul ediyor.
- :80 diğer usta-wake akışlarında da kullanılıyor (X-Usta-Worker). Bu yüzden :80'i yalnız VPS IP'lerine daraltmak başka istemcileri bozabilir.
- Ek gözlem: `default-allow-ssh` (22) ve `default-allow-rdp` (3389) tüm internete açık.

## Öneri (TLS + daraltılmış firewall birlikte)
1. usta-wake'e yeni bir dinleyici: :8443 TLS, yalnızca /a100/start ve /a100/status. Python `ssl` ile kurulur, nginx gerekmez. Sertifika self-signed (ECDSA P-256, 2 yıl, SAN IP:34.41.186.170). Sertifika VPS'te pinlenir: `curl --cacert /etc/us-model/usta-wake.pem` ya da `--pinnedpubkey sha256//...`. Let's Encrypt alan adı (DNS) gerektirdiği için tercih edilmedi.
2. Yeni etiket `usta-a100-api` ve yeni firewall kuralı `usta-a100-api-allow`: INGRESS tcp:8443, kaynak 31.220.63.191/32 ve 82.29.184.225/32, hedef bu etiket. Mevcut kurallara dokunulmaz.
3. Uygulama: allowlist'e 82.29.184.225 eklenir. /a100/* :80'de kapatılır (404), mevcut :80 akışları aynı kalır.
4. Token şifresiz gittiği için yeni bir token üretilir. Yenisi /etc/usta-wake/a100.token ve VPS'te /etc/us-model/wake.token dosyasına yazılır; eski token geçersiz olur.
5. US'a bildirilecek yeni uç: `https://34.41.186.170:8443/a100/{start,status}`, pin dosyası /etc/us-model/usta-wake.pem. 82.29.184.225 sunucusuna da token ve pin dosyası gerekiyorsa erişim yolu ayrıca belirlenmeli.

## Maliyet ve risk
- Ek ücret yok. Kesinti en fazla birkaç saniye (usta-wake servisi yeniden başlar).
- Geri alma: `/opt/usta-wake.py` yedekten döner, yeni kural ve etiket kaldırılır (silme işlemi için ayrıca onay gerekir) ya da kural disable edilir.

## Onay gerekenler
- Yeni firewall kuralı ve etiket (Demir'in kuralı: firewall değişikliği onaysız yapılmaz).
- Token yenileme ve US proxy adresinin değişmesi.
