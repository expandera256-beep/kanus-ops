package licensesrv

import (
	"strconv"
	"strings"
)

type refundPayload struct {
	ID      int64 `json:"id"`
	OrderID int64 `json:"order_id"`
	Lines   []struct {
		LineID   int64 `json:"line_item_id"`
		Quantity int   `json:"quantity"`
	} `json:"refund_line_items"`
}

func (h *Handler) cancel(webhookID string, orderID int64) error {
	s := h.Store
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transactionLocked(func() error {
		oid := strconv.FormatInt(orderID, 10)
		// A durable tombstone prevents late orders/paid deliveries granting credits.
		s.st.Cancelled[oid] = true
		for _, key := range s.st.Orders[oid] {
			s.revokeLocked(key)
		}
		h.refreshPendingMailLocked(orderID)
		if webhookID != "" {
			s.st.Webhooks[webhookID] = h.now()
		}
		return nil
	})
}

func (h *Handler) refund(webhookID string, f refundPayload) error {
	s := h.Store
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transactionLocked(func() error {
		// Refund ID, not just delivery ID, makes redelivery with a fresh header safe.
		fid := strconv.FormatInt(f.ID, 10)
		if _, seen := s.st.Refunds[fid]; !seen {
			lines := map[int64]int{}
			for _, line := range f.Lines {
				lines[line.LineID] = min(20, lines[line.LineID]+min(20, line.Quantity))
			}
			s.st.Refunds[fid] = Refund{OrderID: f.OrderID, Lines: lines}
			h.applyRefundsLocked(f.OrderID)
			h.refreshPendingMailLocked(f.OrderID)
		}
		if webhookID != "" {
			s.st.Webhooks[webhookID] = h.now()
		}
		return nil
	})
}

// One purchased unit creates one license. A line-item quantity refund revokes
// that many unit licenses (including spent units); credits never go negative.
// Financial-only refunds deliberately do not guess a credit conversion rate.
func (h *Handler) applyRefundsLocked(orderID int64) {
	quantities := map[int64]int{}
	for _, f := range h.Store.st.Refunds {
		if f.OrderID == orderID {
			for line, q := range f.Lines {
				quantities[line] = min(20, quantities[line]+min(20, q))
			}
		}
	}
	for _, key := range h.Store.st.Orders[strconv.FormatInt(orderID, 10)] {
		l := h.Store.st.Licenses[key]
		if l != nil && quantities[l.LineID] > 0 {
			h.Store.revokeLocked(key)
			quantities[l.LineID]--
		}
	}
}

func (h *Handler) activeMailKeysLocked(orderID int64) []string {
	return h.Store.activeOrderKeysLocked(orderID)
}

// activeOrderKeysLocked: the order's non-revoked keys, read from the license
// store at send time so full keys never have to be persisted in the outbox.
func (s *Store) activeOrderKeysLocked(orderID int64) []string {
	var keys []string
	for _, key := range s.st.Orders[strconv.FormatInt(orderID, 10)] {
		if l := s.st.Licenses[key]; l != nil && !l.Revoked {
			keys = append(keys, key)
		}
	}
	return keys
}

// Shopify's human order.name is display-only; order.id stays the ledger key.
func orderDisplayName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	runes := []rune(name)
	if len(runes) > 100 {
		name = string(runes[:100])
	}
	if name == "" {
		return "adı belirtilmedi"
	}
	return name
}
func orderMailSubject(name string) string {
	return "KAN.US US lisans anahtarınız — " + orderDisplayName(name)
}
func orderMailBody(name string, keys []string) string {
	return "Merhaba,\r\n\r\nSiparişiniz (" + orderDisplayName(name) + ") için US lisans anahtar(lar)ınız:\r\n\r\n" + strings.Join(keys, "\r\n") + "\r\n\r\nAnahtarı US panelinde Lisans alanına yapıştırın.\r\n\r\nKAN.US destek: support@kanus.org\r\n"
}
func (h *Handler) refreshPendingMailLocked(orderID int64) {
	oid := strconv.FormatInt(orderID, 10)
	keys := h.activeMailKeysLocked(orderID)
	for _, m := range h.Store.st.Outbox {
		matches := m.OrderID == orderID
		if m.OrderID == 0 {
			for _, key := range h.Store.st.Orders[oid] {
				if strings.Contains(m.Body, key) {
					matches = true
					break
				}
			}
		}
		if !matches || !m.SentAt.IsZero() {
			continue
		}
		if len(keys) == 0 {
			m.Dead = true
			m.LastErr = "sipariş iptal/iade"
			m.Body = scrubLicenseKeys(m.Body)
		} else {
			m.Subject = orderMailSubject(m.OrderName)
			m.Body = orderMailBody(m.OrderName, maskLicenseKeys(keys))
		}
	}
}
