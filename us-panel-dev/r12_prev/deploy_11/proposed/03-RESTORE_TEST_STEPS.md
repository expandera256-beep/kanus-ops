# Store restore test (UYGULANMADI — prosedür taslağı)

Servisi uzun süre kapatmadan doğrulama önerisi:

1. Günlük yedek scriptini bir kez elle çalıştır; `/var/backups/us-license-hook/licenses.json.<stamp>` oluşsun.
2. **Staging/test kopyası** (tercih): yedeği `/tmp/licenses.restore-test.json` olarak kopyala; `python3 -m json.tool` ile geçerli JSON olduğunu doğrula; dosya boyutunu canlı store ile karşılaştır.
3. Canlı restore (yalnızca onaylı bakım penceresi):
   - `systemctl stop us-license-hook` (kısa kesinti — bu adım operatör onayı ister)
   - canlı `licenses.json` → `licenses.json.pre-restore`
   - seçilen yedeği `US_STORE_PATH` konumuna `install -o uslic -g uslic -m 640`
   - `systemctl start us-license-hook`
   - `curl -s http://127.0.0.1:8787/healthz` → 200
4. Geri alma: `pre-restore` dosyasını geri koy → start.
5. Bu turda **hiçbir restore uygulanmadı**.
