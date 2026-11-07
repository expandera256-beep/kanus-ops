package licensesrv

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Sender interface{ Send(m Mail) error }

// SMTPSender: Hostinger için implicit TLS (smtp.hostinger.com:465).
// Kimlik bilgileri yalnız ortamdan: US_SMTP_HOST, US_SMTP_PORT, US_SMTP_USER,
// US_SMTP_PASS, US_SMTP_FROM.
type SMTPSender struct{ Host, Port, User, Pass, From string }

func (c SMTPSender) Send(m Mail) error {
	if !SMTPRecipientAllowed(m.To) {
		return errors.New("recipient blocked by allowlist")
	}
	if c.Host == "" || c.User == "" || c.Pass == "" || c.From == "" {
		return errors.New("SMTP yapılandırması eksik")
	}
	addr := net.JoinHostPort(c.Host, c.Port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 20 * time.Second}, "tcp", addr, &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	cl, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer cl.Close()
	if err := cl.Auth(smtp.PlainAuth("", c.User, c.Pass, c.Host)); err != nil {
		return err
	}
	if err := cl.Mail(c.From); err != nil {
		return err
	}
	if err := cl.Rcpt(m.To); err != nil {
		return err
	}
	wc, err := cl.Data()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("From: KAN.US <%s>\r\nTo: %s\r\nSubject: =?UTF-8?B?%s?=\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nDate: %s\r\nMessage-ID: <%s@kanus.org>\r\n\r\n%s",
		c.From, m.To, b64(m.Subject), time.Now().UTC().Format(time.RFC1123Z), m.ID, m.Body)
	if _, err := wc.Write([]byte(msg)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return cl.Quit()
}

// DryRunSender gerçek e-posta göndermez; yalnız kaydeder.
type DryRunSender struct{ Log func(string) }

func (d DryRunSender) Send(m Mail) error {
	if d.Log != nil {
		d.Log("[dry-run] e-posta " + maskEmail(m.To) + ": " + m.Subject)
	}
	return errors.New("dry-run: gönderilmedi")
}

const maxAttempts = 8

// FlushOutbox bekleyen e-postaları gönderir; hatada üstel geri çekilme uygular.
func (s *Store) FlushOutbox(sender Sender, now time.Time) (sent, failed int) {
	s.mu.Lock()
	var due []*Mail
	for _, m := range s.st.Outbox {
		if m.SentAt.IsZero() && !m.Dead && m.Skipped == "" && !now.Before(m.NextTry) {
			due = append(due, m)
		}
	}
	s.mu.Unlock()
	for _, m := range due {
		s.mu.Lock()
		if m.Dead || m.Skipped != "" || !m.SentAt.IsZero() {
			s.mu.Unlock()
			continue
		}
		if !SMTPRecipientAllowed(m.To) {
			_ = s.transactionLocked(func() error { m.Skipped = "recipient not on SMTP allowlist"; return nil })
			s.mu.Unlock()
			continue
		}
		mail := *m // do not read a mutable queue item while cancellation edits it
		if m.OrderID != 0 {
			// Delivery mail: the outbox keeps only masked keys; the full
			// keys are rendered here from the license store and are never
			// written back.
			keys := s.activeOrderKeysLocked(m.OrderID)
			if len(keys) == 0 {
				_ = s.transactionLocked(func() error {
					m.Dead, m.LastErr, m.Body = true, "sipariş iptal/iade", scrubLicenseKeys(m.Body)
					return nil
				})
				s.mu.Unlock()
				continue
			}
			mail.Body = orderMailBody(m.OrderName, keys)
		}
		s.mu.Unlock()
		err := sender.Send(mail)
		s.mu.Lock()
		if err == nil {
			m.SentAt = now
			m.LastErr = ""
			m.Body = scrubLicenseKeys(m.Body)
			sent++
		} else {
			m.Attempts++
			m.LastErr = strings.TrimSpace(err.Error())
			failed++
			m.NextTry = now.Add(time.Duration(1<<min(m.Attempts, 10)) * time.Minute)
			if m.Attempts >= maxAttempts {
				m.Dead = true
				m.Body = scrubLicenseKeys(m.Body)
			}
		}
		_ = s.saveLocked()
		s.mu.Unlock()
	}
	return
}

// Unset means legacy unrestricted; set-empty denies all. Guard is evaluated
// before sender invocation, and again by SMTPSender before any connection.
func SMTPRecipientAllowed(to string) bool {
	if strings.ContainsAny(to, "\r\n") {
		return false
	}
	raw, set := os.LookupEnv("US_SMTP_ALLOWLIST")
	if !set {
		return true
	}
	to = strings.ToLower(strings.TrimSpace(to))
	for _, v := range strings.Split(raw, ",") {
		if strings.TrimSpace(v) != "" && strings.EqualFold(strings.TrimSpace(v), to) {
			return true
		}
	}
	return false
}
