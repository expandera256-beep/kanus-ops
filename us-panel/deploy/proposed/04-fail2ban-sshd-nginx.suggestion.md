# fail2ban önerisi (UYGULANMADI)

## Paket
```
apt-get install -y fail2ban
systemctl enable --now fail2ban
```
`enable --now` fail2ban servisini başlatır (yeni servis). Mevcut nginx/hook restart **gerekmez**.

## `/etc/fail2ban/jail.local` taslak
```ini
[DEFAULT]
bantime = 1h
findtime = 10m
maxretry = 5
backend = systemd

[sshd]
enabled = true
port = ssh
mode = aggressive

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
# rate-limit 429/503 loglarından ban (limit_req uygulandıktan sonra anlamlı)
```

Doğrulama: `fail2ban-client status` / `fail2ban-client status sshd`
Geri alma: `systemctl disable --now fail2ban` veya jail'leri `enabled = false` yapıp `fail2ban-client reload`.

Not: UFW zaten aktifse fail2ban iptables/nftables kuralları ekler; çakışma için önce `ufw status verbose` ile mevcut politikayı belgele.
