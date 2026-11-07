package licensesrv
import("context";"fmt";"net/http";"net/http/httptest";"os";"path/filepath";"sync";"sync/atomic";"testing";"time")
func TestPanelConcurrentHTTPWakeDebounce(t *testing.T){
 h,_:=setup(t);post(h,payload,sign("whsec",payload),"p");key:=h.Store.st.Orders["555"][0];var starts,status atomic.Int32
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path=="/a100/start"{starts.Add(1);w.WriteHeader(500);return};status.Add(1);fmt.Fprint(w,`{"model_ready":false}`)}));defer s.Close();t.Setenv("US_GPU_WAKE_PIN_FILE","");tok:=filepath.Join(t.TempDir(),"token");os.WriteFile(tok,[]byte("test-token"),0600);g,_:=NewGPUWaker(s.URL+"/a100/start",s.URL+"/a100/status",tok,0)
 p,_:=NewModelProxy(h.Store,"http://127.0.0.1:1","","test-model",time.Second,1,2);p.Waker=g;var wg sync.WaitGroup
 for i:=0;i<50;i++{wg.Add(1);go func(){defer wg.Done();w:=proxyRequest(p,key,proxyInput,context.Background());if w.Code!=503||w.Header().Get("Retry-After")!="30"{t.Errorf("bad response %d %v",w.Code,w.Header())}}()};wg.Wait();panelDrainWake(g)
 t.Logf("50 concurrent HTTP requests: starts=%d status=%d",starts.Load(),status.Load());if starts.Load()!=1||status.Load()!=50{t.Fatal("unexpected debounce/probe count")}
}
