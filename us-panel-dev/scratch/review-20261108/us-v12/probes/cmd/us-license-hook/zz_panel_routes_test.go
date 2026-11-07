package main
import("net/http";"net/http/httptest";"path/filepath";"strings";"sync/atomic";"testing";"kiyas/licensesrv")
func TestPanelR12HookReadOnlyUnauthenticatedNeverWake(t *testing.T){
 var starts,status,up atomic.Int32
 gpu:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if strings.HasSuffix(r.URL.Path,"start"){starts.Add(1)}else{status.Add(1)};w.Write([]byte(`{"model_ready":false}`))}));defer gpu.Close()
 app:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){up.Add(1);w.Write([]byte(`{"data":[]}`))}));defer app.Close()
 t.Setenv("US_GPU_WAKE_DISABLE","0");t.Setenv("US_GPU_WAKE_URL",gpu.URL+"/a100/start");t.Setenv("US_GPU_STATUS_URL",gpu.URL+"/a100/status");t.Setenv("US_MODEL_UPSTREAM",app.URL);t.Setenv("US_GPU_WAKE_PIN_FILE","");t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","0");t.Setenv("US_SESSION_SECRET","panel-synthetic-local-only");t.Setenv("US_SESSION_TTL_SECONDS","86400")
 st,err:=licensesrv.OpenStore(filepath.Join(t.TempDir(),"store"));if err!=nil{t.Fatal(err)};defer st.Close();mux:=newMux(licensesrv.Config{},st)
 cases:=[][2]string{{"GET","/healthz"},{"OPTIONS","/healthz"},{"GET","/readyz"},{"OPTIONS","/v1/chat"},{"OPTIONS","/v1/chat/completions"},{"GET","/v1/models"},{"OPTIONS","/v1/models"},{"POST","/v1/chat"},{"POST","/v1/chat/completions"},{"POST","/v1/models"}}
 for _,c:=range cases{r:=httptest.NewRequest(c[0],c[1],strings.NewReader(`{"messages":[]}`));w:=httptest.NewRecorder();mux.ServeHTTP(w,r);t.Logf("%s %s status=%d",c[0],c[1],w.Code);if starts.Load()!=0||status.Load()!=0||up.Load()!=0{t.Fatalf("wake=%d status=%d upstream=%d",starts.Load(),status.Load(),up.Load())}}
 t.Setenv("US_GPU_WAKE_DISABLE","1");p,err:=configuredProxy(st);if err!=nil||p.Waker!=nil{t.Fatalf("disable failed %v",err)};t.Logf("all %d requests wake=0 status=0 upstream=0; disable Waker=nil",len(cases))
}
