package licensesrv
import("context";"fmt";"net/http";"net/http/httptest";"os";"path/filepath";"strings";"sync";"sync/atomic";"testing";"time")
func TestPanelR12EntitlementChangesBeforeWake(t *testing.T){
 for _,change:=range []string{"credit-consumed","license-revoked"}{t.Run(change,func(t *testing.T){
 h,_:=setup(t);post(h,payload,sign("whsec",payload),"p");key:=h.Store.st.Orders["555"][0];h.Store.st.Licenses[key].Credits=1
 var starts atomic.Int32;entered,release:=make(chan struct{}),make(chan struct{});var once sync.Once
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path=="/a100/start"{starts.Add(1);w.WriteHeader(500);return};once.Do(func(){close(entered)});<-release;fmt.Fprint(w,`{"model_ready":false}`)}));defer s.Close()
 t.Setenv("US_GPU_WAKE_PIN_FILE","");tok:=filepath.Join(t.TempDir(),"token");os.WriteFile(tok,[]byte("synthetic-token"),0600);g,_:=NewGPUWaker(s.URL+"/a100/start",s.URL+"/a100/status",tok,0);p,_:=NewModelProxy(h.Store,"http://127.0.0.1:1","","test-model",time.Second,1,2);p.Waker=g
 done:=make(chan struct{});go func(){defer close(done);proxyRequest(p,key,proxyInput,context.Background())}();<-entered
 if change=="credit-consumed"{id,_,_,e:=h.Store.reserveModel(key,time.Second,time.Now());if e!=nil{t.Fatal(e)};if e=h.Store.settleAmount(key,id,1,time.Now());e!=nil{t.Fatal(e)}}else{h.Store.mu.Lock();h.Store.st.Licenses[key].Revoked=true;h.Store.mu.Unlock()}
 close(release);<-done;panelDrainWake(g);t.Logf("after %s wake=%d",change,starts.Load());if starts.Load()!=0{t.Fatal("stale eligibility snapshot woke GPU after entitlement disappeared")}
 })}
}
func TestPanelR12CancelledLeaderDoesNotWake(t *testing.T){
 h,_:=setup(t);post(h,payload,sign("whsec",payload),"p");key:=h.Store.st.Orders["555"][0];g,starts,_:=panelWakeFixture(t);p,_:=NewModelProxy(h.Store,"http://127.0.0.1:1","","test-model",time.Second,1,2);p.Waker=g
 ctx,cancel:=context.WithCancel(context.Background());cancel();w:=proxyRequest(p,key,proxyInput,ctx);panelDrainWake(g);t.Logf("cancelled leader code=%d wake=%d",w.Code,starts.Load());if starts.Load()!=0{t.Fatal("already-cancelled request woke GPU")}
}
func TestPanelR12CancelledFollowerDoesNotWake(t *testing.T){
 h,_:=setup(t);post(h,payload,sign("whsec",payload),"p");key:=h.Store.st.Orders["555"][0];var starts atomic.Int32
 entered,release:=make(chan struct{}),make(chan struct{});var once sync.Once
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path=="/a100/start"{starts.Add(1);w.WriteHeader(500);return};once.Do(func(){close(entered)});<-release;fmt.Fprint(w,`{"model_ready":true}`)}));defer s.Close()
 tok:=filepath.Join(t.TempDir(),"token");os.WriteFile(tok,[]byte("synthetic-token"),0600);t.Setenv("US_GPU_WAKE_PIN_FILE","");g,_:=NewGPUWaker(s.URL+"/a100/start",s.URL+"/a100/status",tok,0)
 upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprint(w,`{"data":[{"id":"test-model"}]}`)}));defer upstream.Close();p,_:=NewModelProxy(h.Store,upstream.URL,"","test-model",time.Second,1,2);p.Waker=g
 leader:=make(chan struct{});go func(){defer close(leader);p.modelReady(context.Background())}();<-entered
 ctx,cancel:=context.WithCancel(context.Background());cancel();w:=proxyRequest(p,key,proxyInput,ctx);panelDrainWake(g);close(release);<-leader;t.Logf("cancelled follower code=%d wake=%d",w.Code,starts.Load());if starts.Load()!=0{t.Fatal("cancelled follower treated cancellation as GPU-down and woke ready GPU")}
}
func TestPanelR12ReserveSettleNoOverspend(t *testing.T){
 var chats atomic.Int32;p,st,key:=proxyFixture(t,func(w http.ResponseWriter,r *http.Request){chats.Add(1);w.Header().Set("Content-Type","text/event-stream");fmt.Fprint(w,goodEvents)})
 st.mu.Lock();st.st.Licenses[key].Credits=1;st.mu.Unlock();p.requests=make(chan struct{},32)
 var ok,denied atomic.Int32;var wg sync.WaitGroup
 for i:=0;i<20;i++{wg.Add(1);go func(){defer wg.Done();w:=proxyRequest(p,key,proxyInput,context.Background());if w.Code==200&&strings.Contains(w.Body.String(),"[DONE]"){ok.Add(1)}else if w.Code==402{denied.Add(1)}else{t.Errorf("unexpected %d %s",w.Code,w.Body.String())}}()};wg.Wait();l,_:=st.License(key)
 t.Logf("ok=%d denied=%d chats=%d credits=%d reserved=%d",ok.Load(),denied.Load(),chats.Load(),l.Credits,l.Reserved);if ok.Load()!=1||denied.Load()!=19||chats.Load()!=1||l.Credits!=0||l.Reserved!=0{t.Fatal("reservation/settlement inconsistency")}
}
