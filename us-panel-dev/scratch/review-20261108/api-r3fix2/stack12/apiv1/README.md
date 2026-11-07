# apiv1 — US .11 + UYUM üzerine ekleyici API yüzeyi

Bu paket `licensesrv/**` ve `model_proxy.go` dosyalarına dokunmaz.
`cmd/us-license-hook/api_mount.go` yalnızca mux’a `/v1/account/` alias ve
`API-Version` sarmalayıcı ekler; UYUM `privacy.Mount` çağrısından **sonra** gelir.

Tam hata zarfı / API key / credits usk_ entegrasyonu için bkz. `patches/api-r11/US_ONERILER.md`.
