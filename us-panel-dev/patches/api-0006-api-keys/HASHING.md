# API key hashing (tuzsuz SHA-256)

Anahtar gövdesi `usk_` + 192-bit CSPRNG (24 rastgele bayt → hex). Saklama: `SHA-256(plaintext)` hex.
Tuz/HMAC-pepper yok: sır zaten yüksek entropili; rainbow table pratik değil.
İleride `US_API_KEY_PEPPER` ile HMAC-SHA256’ya geçiş önerilir (US_ONERILER); bu turda davranış değişmedi.
