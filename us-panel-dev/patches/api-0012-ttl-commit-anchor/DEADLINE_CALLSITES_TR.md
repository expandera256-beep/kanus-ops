# Rezerv deadline'ını kendi başına hesaplayan yerler (08.11.2026, TSİ)

Amaç: US .13'teki ortak deadline fonksiyonu (ttl + grace, en fazla 900 sn) için tam liste.
Ağaçlar:
- **.12** = `/workspace/us-panel/work2/kiyas/term` (salt okundu)
- **STACK** = `.12 + UYUM v4 (16/16) + api-0004/0012 combined + df1ae528 fix + apiv1 additive` (`/workspace/us-panel-dev/work/api/on12/verify`)

accounts.go, model_meter.go, model_proxy.go ve credits.go iki ağaçta aynı; satır numaraları ikisinde de geçerli. store.go'nun satırları farklı, ikisi de verildi.
Kod değişikliği yapılmadı. US'le iş bölümü gereği kapı diff'i, 4 vakalık test ve ortak fonksiyon diff'i bu teslimde YOK.

## A. Rezervin oluşturulduğu yer (TTL → Expires)
| Dosya:satır | Ne hesaplıyor |
|---|---|
| store.go:246 (.12) / :291 (STACK) — `Reserve` | `Expires = now + ttl`; ttl çağırandan gelir |
| credits.go:19 + :59 | `/v1/credits/reserve` sabit `ReservationTTL = 45 dk` ile `Reserve` çağırır |
| model_meter.go:56 — `reserveModel` | Model rezervi: `Expires = now + budget + 5 dk` (budget = p.Timeout, model_proxy.go:157). **Ayrı bir TTL formülü** |

## B. Expire / release kararı (deadline'ı geçti mi?)
| Dosya:satır | Ne hesaplıyor |
|---|---|
| store.go:304 (.12) — `expireLocked` | `now >= r.Expires` ise release eder (grace yok) |
| store.go:356–357 (STACK) — `expireLocked` | `now >= reservationCommitDeadline(r)` = Expires+grace ise release eder |
| store.go:256–260 (STACK) — `reservationCommitDeadline` | **Tek merkezi formül (API 0012):** grace ≤ 0 ise Expires, değilse Expires+commitGrace |
| store.go:238–249 (STACK) — `commitGraceFromEnv` | grace: int64 saniye, varsayılan 600, en fazla 900, geçersiz/negatifse 600, 0 ise 0 |
| store.go:151 (STACK) — `OpenStore` | `s.commitGrace = commitGraceFromEnv()` (grace yalnız açılışta okunuyor) |
| model_meter.go:18 — `reserveModel` | `expireLocked(now)` çağırır, kararı store'a bırakır (STACK'te grace'e uyar) |
| store.go:274 (.12) / :280 (STACK) — `Reserve` | `expireLocked(now)` çağırır |

## C. Commit kabul sınırı (settle)
| Dosya:satır | Ne hesaplıyor |
|---|---|
| store.go:277 (.12) — `settleFor` | commit yalnız `now < r.Expires` ise kabul edilir |
| store.go:322 (STACK) — `settleFor` | commit yalnız `now < Expires+grace` ise kabul edilir (deadline fonksiyonu) |
| **model_meter.go:79 — `settleAmount`** | **Model proxy commit'i hâlâ ham `now < r.Expires`'a bakıyor, grace YOK.** STACK'te tutarsız: store rezervi grace boyunca kilitli tutuyor ama proxy aynı aralıkta commit edemiyor (ErrInvalid). Sonuç olarak rezerv sonradan expireLocked ile bedava release ediliyor. Fazla çekim olmuyor, gelir kaybı ya da tutarsızlık var. Ortak fonksiyona bağlanmalı |
| store.go:282 (.12) / :327 (STACK) | Lisans bitişi (`l.ExpiresAt`); rezerv deadline'ı değil, bilgi için |

## D. Kullanılabilir kredi / kapı (reserved'i hesaplayan yerler)
| Dosya:satır | Ne hesaplıyor |
|---|---|
| **accounts.go:763–776 (özellikle :769–770) — `usableCreditLocked`** | `now >= r.Expires` olan rezervi Reserved'dan düşer. STACK'te grace yok sayılıyor, PANEL R3FIX2 HIGH bu (kapı geçiyor, reserve ErrNoCredit, wake model_proxy.go:135) |
| accounts.go:747 ve :757 — `modelSpendable` | Kararı `usableCreditLocked`'tan alır (oturum ve lisans yolu); model_proxy.go:119 çağırır |
| accounts.go:117 | Hesap seçimi `l.Credits - l.Reserved > 0`: ham Reserved, deadline yok. Muhafazakâr, grace ile uyumlu |
| model_meter.go:46 ve :52 | `l.Credits - l.Reserved` ham. `expireLocked` sonrası olduğu için deadline'a dolaylı uyar |
| store.go:241 (.12) / :285 (STACK) — `Reserve` | Aynı ham `Credits-Reserved` kontrolü, `expireLocked` sonrası |

## E. API yanıtında reserved/available/expires dönen yerler
| Dosya:satır | Ne dönüyor |
|---|---|
| credits.go:132 — credit status | `reserved = l.Reserved`, `available = Credits - Reserved` (ham, sweep yok: "read neither sweeps", :96). Grace'teki ve süresi geçmiş ama henüz sweep edilmemiş rezervi de rezerve gösterir. Kendi deadline'ını hesaplamıyor ama kapıyla farklı sayı verebilir |
| accounts.go:166–191 — `balanceLocked` | Hesap bakiyesi: `available = Credits - Reserved` ham, `ReservedSeconds`/`Reserved` ham toplam. Deadline hesabı yok |
| accounts.go:704 — account credits yanıtı | `reserved`, `available`, `expiresAt`. Buradaki `expiresAt` **lisans** bitişi (b.ExpiresAt), rezerv deadline'ı değil |
| credits.go:114 `ExpiresAt` | Lisans bitişi; rezerv deadline'ı değil |
| — | **Hiçbir API yanıtı rezervin `expires`/deadline'ını dönmüyor.** `/v1/credits/reserve` yalnız id dönüyor |

## F. Benim yamalarım (api-0004, api-0012, fix df1ae528, api-0006/0008, apiv1, SDK, OpenAPI)
| Yer | Bulgu |
|---|---|
| api-0012 / df1ae528 (`store.go` STACK :151, :238–260, :322, :356–357) | Deadline yalnız `reservationCommitDeadline` içinde hesaplanıyor; settleFor ve expireLocked bunu kullanıyor. Başka kopya yok. US ortak fonksiyonu getirdiğinde bu fonksiyon onunla değiştirilmeli ya da ona delege etmeli |
| api-0004 (`applyReservedDelta`, store.go STACK :263–271) | Yalnız Reserved'in negatife düşmesini engelliyor; zaman/deadline hesabı yok |
| testler: `ttl_commit_anchor_test.go`, `grace_overflow_test.go`, `ttl_grace_lock_test.go` | Testler `Expires+grace` beklentisini kendi içinde kuruyor (ör. `now.Add(ttl+grace)`). Fonksiyon adı değişirse yalnız test referansları güncellenir |
| api-0006 (`0001-API-0006-…patch:58–76`, accounts credits yanıtı) | .12 accounts.go:704 ile aynı satır; ham `reserved`/`available`, lisans `expiresAt`. Rezerv deadline'ı yok. `GraceUntil`/`defaultGracePeriod` (patch :108, :179, :402; 0006b :39–66) **API anahtarı rotasyon grace'i**, kredi grace'iyle ilgisi yok |
| api-0006 (patch :574) | API anahtarıyla lisans çözümü: yalnız `l.ExpiresAt` |
| api-0008 kota | Günlük kota ve Retry-After; rezerv/deadline hesabı yok |
| apiv1 (`version.go`, `errors.go`) | Yalnız /v1/account alias ve envelope; deadline yok |
| SDK Python `client.py:186–195`, TS `client.ts:152–167` | reserve/settle'ı çağırıyor; TTL/deadline hesabı yok, retry etmiyor |
| OpenAPI api-0002 :648–693, api-0009 :897–942 | `reserved`, `reserved_seconds`, lisans `expiresAt` alanları; rezerv deadline alanı yok. `graceUntil` (:985) API anahtarı içindir |

## Özet: ortak fonksiyona bağlanması gereken yerler
1. accounts.go:769 `usableCreditLocked`: HIGH (US .13'te).
2. **model_meter.go:79 `settleAmount`: commit sınırı ham Expires. PANEL raporunda yok, yeni bulgu.**
3. store.go STACK :322 (settleFor) ve :357 (expireLocked): şu an `reservationCommitDeadline` kullanıyor; ortak fonksiyona delege edilebilir.
4. TTL üretimi iki ayrı formül: store.go `Reserve(ttl)` (credits 45 dk) ve model_meter.go:56 (`budget + 5 dk`). Grace ikisine de aynı biçimde ekleniyor; bu kabul ediliyorsa bilinçli olarak not edilmeli.
5. credits.go:132 / accounts.go:166–191: sweep etmeyen ham okuma. Kapıyla aynı sayıyı göstermesi istenirse aynı fonksiyonla hesaplanmalı (karar US'te).
