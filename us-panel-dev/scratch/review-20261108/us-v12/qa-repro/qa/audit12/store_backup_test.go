package licensesrv
import (
 "encoding/json"
 "fmt"
 "net/http/httptest"
 "os"
 "path/filepath"
 "testing"
 "time"
)
func TestQAAuditGoStoreBackupRestore(t *testing.T) {
 h,path:=setup(t)
 var o order
 if err:=json.Unmarshal([]byte(`{"id":900099,"email":"qa-backup@example.test","financial_status":"paid","line_items":[{"id":1,"product_id":42,"quantity":1}]}`),&o);err!=nil {t.Fatal("synthetic fixture invalid")}
 if _,err:=h.process("qa-local-only",o);err!=nil {t.Fatal("synthetic process failed")}
 h.Store.mu.Lock();keys:=append([]string(nil),h.Store.st.Orders["900099"]...);h.Store.mu.Unlock()
 if len(keys)!=1 {t.Fatal("expected one synthetic license")}
 now:=time.Now();rid,err:=h.Store.Reserve(keys[0],1,time.Minute,now);if err!=nil {t.Fatal("reserve failed")}
 if err=h.Store.Settle(rid,true);err!=nil {t.Fatal("settle failed")}
 before,ok:=h.Store.CreditStatus(keys[0],now);if !ok {t.Fatal("status failed")}
 if err=h.Store.Close();err!=nil {t.Fatal("close synthetic store failed")}
 raw,err:=os.ReadFile(path);if err!=nil {t.Fatal("snapshot read failed")}
 backup:=filepath.Join(t.TempDir(),"synthetic-backup.json");if os.WriteFile(backup,raw,0600)!=nil {t.Fatal("backup write failed")}
 restored:=filepath.Join(t.TempDir(),"restored.json");if os.WriteFile(restored,raw,0600)!=nil {t.Fatal("restore write failed")}
 st,err:=OpenStore(restored);if err!=nil {t.Fatal("restore store open failed")};defer st.Close()
 after,ok:=st.CreditStatus(keys[0],now)
 if !ok || before.Remaining!=after.Remaining || before.Reserved!=after.Reserved || before.Committed!=after.Committed || before.Valid!=after.Valid || before.Remaining!=499 {t.Fatal("restored quota/history mismatch")}
 if st2,err:=OpenStore(restored);err==nil {st2.Close();t.Fatal("restored store lock not enforced")}
 t.Log("synthetic JSON ledger backup+reopen: credits/history/validity and exclusive lock preserved; no customer data")
}

// Existing IP-only layer works as designed, but does not aggregate one account across IPs.
func TestQAAuditIPOnlyRateDesign(t *testing.T) {
 h:=&AccountHandler{rates:map[string]rateBucket{}}
 accepted:=0
 for i:=1;i<=11;i++ {
  r:=httptest.NewRequest("POST","/account/login",nil)
  r.RemoteAddr="127.0.0.1:4477";r.Header.Set("X-Real-IP",fmt.Sprintf("203.0.113.%d",i))
  if h.rate(r,"login",10) {accepted++}
 }
 t.Logf("IP-only layer: rotating_source_requests=11 accepted=%d per_ip_limit=10; account limiter must be separate",accepted)
 if accepted!=11 {t.Fatal("IP-only source behavior changed; reassess design evidence")}
}
