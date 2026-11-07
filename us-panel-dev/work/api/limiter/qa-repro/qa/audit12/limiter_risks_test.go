package licensesrv
import (
 "fmt"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)
// Safety regression: this intentionally fails while an unauthenticated actor can block a valid password.
func TestQAAuditLimiterAccountLockDoS(t *testing.T) {
 h,_:=setup(t);a:=accountTestHandler(t,h.Store);accountRegister(t,a,"qa-victim@example.test","")
 for i:=0;i<10;i++ {w:=loginFromIP(a,fmt.Sprintf("203.0.113.%d",i+1),"qa-victim@example.test","wrong-password-at-least12");if w.Code!=401 {t.Fatal("fixture failed before lock")}}
 w:=loginFromIP(a,"198.51.100.8","qa-victim@example.test",accountPassword)
 t.Logf("failed_attempts=10 correct_password_http=%d retry_after_present=%t",w.Code,w.Header().Get("Retry-After")!="")
 if w.Code!=200 {t.Fatal("unauthenticated failures locked the legitimate correct password: account-lock DoS remains OPEN")}
}
func TestQAAuditLimiterFixedWindowAndVolatileState(t *testing.T) {
 l:=newAccountLimiter();now:=time.Now();email:="unknown-qa@example.test"
 for i:=0;i<10;i++ {l.fail(email,now)}
 locked,_:=l.blocked(email,now);if !locked {t.Fatal("fixture not locked")}
 l.blocked(email,now.Add(899*time.Second));if locked,_:=l.blocked(email,now.Add(901*time.Second));locked {t.Fatal("window extended by blocked request")}
 l.fail(email,now);fresh:=newAccountLimiter();if blocked,_:=fresh.blocked(email,now);blocked {t.Fatal("independent instance unexpectedly shares state")}
 t.Log("unregistered_email_counted=true fixed_window_900s=true blocked_checks_do_not_extend=true fresh_instance_or_restart_state_empty=true")
}
func TestQAAuditLimiterEvictionSpray(t *testing.T) {
 l:=newAccountLimiter();now:=time.Now();target:="target-qa@example.test"
 for i:=0;i<10;i++ {l.fail(target,now)}
 if b,_:=l.blocked(target,now);!b {t.Fatal("fixture not locked")}
 for i:=0;i<10000;i++ {l.fail(fmt.Sprintf("spray%d@example.test",i),now.Add(time.Second))}
 b,_:=l.blocked(target,now.Add(time.Second));t.Logf("default_max_keys=%d synthetic_other_emails=10000 target_counter_evicted=%t",l.maxKeys,!b)
 if b {t.Fatal("eviction behavior changed; reassess finding")}
}
func TestQAAuditLimiterIPCapacityAndRetryAfter(t *testing.T) {
 h,_:=setup(t);a:=accountTestHandler(t,h.Store);now:=time.Now()
 for i:=0;i<4096;i++ {a.rates[fmt.Sprintf("login:synthetic-ip-%d",i)]=rateBucket{Start:now,Count:1}}
 w:=loginFromIP(a,"203.0.113.98","unknown-qa@example.test","wrong-password-at-least12")
 t.Logf("ip_buckets=4096 new_ip_http=%d retry_after_seconds=%s",w.Code,w.Header().Get("Retry-After"))
 if w.Code!=429 || w.Header().Get("Retry-After")!="5" {t.Fatal("IP capacity behavior changed; reassess finding")}
}
func TestQAAuditLimiterRedeemHasNoRateBucket(t *testing.T) {
 h,_:=setup(t);a:=accountTestHandler(t,h.Store);accountRegister(t,a,"redeem-qa@example.test","");token:=accountLogin(t,a,"redeem-qa@example.test")
 before:=len(a.rates)
 for i:=0;i<12;i++ {
  r:=httptest.NewRequest("POST","/account/redeem",strings.NewReader(`{"license":"synthetic-invalid"}`));r.RemoteAddr="203.0.113.70:1234";r.Header.Set("Authorization","Bearer "+token)
  w:=httptest.NewRecorder();a.ServeHTTP(w,r);if w.Code==429 || w.Code==200 {t.Fatal("redeem fixture unexpectedly rate-limited or mutated")}
 }
 if len(a.rates)!=before {t.Fatal("redeem created a rate bucket; reassess scope")}
 t.Log("authenticated_redeem_invalid_attempts=12 rate_429_count=0 rate_bucket_delta=0 synthetic_no_mutation=true")
}
