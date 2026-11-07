package licensesrv

import (
	"strings"
	"testing"
)

func TestLicenseMailUsesOrderNameAndKeepsIDIdempotency(t *testing.T) {
	h, path := setup(t)
	body := strings.Replace(payload, `"id":555`, `"id":555,"name":"#1001"`, 1)
	body = strings.Replace(body, `"product_id":42,"quantity":1`, `"product_id":42,"quantity":2`, 1)
	if c := post(h, body, sign("whsec", body), "paid-name"); c != 200 {
		t.Fatal(c)
	}
	mail := h.Store.Outbox()[0]
	if !strings.Contains(mail.Subject, "#1001") || !strings.Contains(mail.Body, "(#1001)") || strings.Contains(mail.Body, "(#555)") || strings.Contains(mail.Subject, "555") || mail.OrderID != 555 || mail.OrderName != "#1001" {
		t.Fatal("wrong display identity", mail.Subject, mail.OrderName)
	}
	changed := strings.Replace(body, "#1001", "#other", 1)
	post(h, changed, sign("whsec", changed), "another-id")
	if len(h.Store.Outbox()) != 1 || len(h.Store.st.Orders["555"]) != 2 {
		t.Fatal("name used as idempotency key")
	}
	h.Store.Close()
	again, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	h.Store = again
	if c := postTopic(h, "refunds/create", `{"id":900,"order_id":555,"refund_line_items":[{"line_item_id":1,"quantity":1}]}`, "partial-name"); c != 200 {
		t.Fatal(c)
	}
	mail = again.Outbox()[0]
	if !strings.Contains(mail.Subject, "#1001") || !strings.Contains(mail.Body, "(#1001)") {
		t.Fatal("refund refresh lost order name")
	}
	if strings.ContainsAny(orderMailSubject("#1001\r\nInjected"), "\r\n") {
		t.Fatal("subject control characters")
	}
}
