package licensesrv
import("context";"fmt";"net/http";"net/http/httptest";"testing";"time";"sync")
func TestPanelReadinessRejectsTrailingJSON(t *testing.T){
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprint(w,`{"data":[{"id":"probe-model"}]} BROKEN`)}));defer s.Close()
 p,e:=NewModelProxy(nil,s.URL,"","probe-model",time.Second,1,2);if e!=nil{t.Fatal(e)}
 ready,_:=p.modelReady(context.Background());if ready{t.Fatal("malformed trailing JSON cached as model ready")}
}
func TestPanelMetadataMissingDataInvalidates(t *testing.T){
 h,_:=setup(t);post(h,payload,sign("whsec",payload),"p");key:=h.Store.st.Orders["555"][0]
 for _,body:=range []string{`null`,`{}`,`{"data":null}`}{t.Run(body,func(t *testing.T){s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprint(w,body)}));defer s.Close();p,_:=NewModelProxy(h.Store,s.URL,"","probe-model",time.Second,1,2);p.readyUntil=time.Now().Add(time.Minute);r:=httptest.NewRequest("GET","/v1/models",nil);r.Header.Set("Authorization","Bearer "+key);w:=httptest.NewRecorder();p.Models(w,r);if time.Now().Before(p.readyUntil){t.Fatalf("missing model list kept readiness: code=%d body=%s",w.Code,w.Body.String())}})}
}
func TestPanelBcryptCancelledWaitersDoNotLeak(t *testing.T){
 a,_:=cost12Handler(t);hold:=fillBcryptSlots(a,cap(a.work));var wg sync.WaitGroup
 for i:=0;i<8;i++{wg.Add(1);go func(){defer wg.Done();ctx,c:=context.WithTimeout(context.Background(),20*time.Millisecond);defer c();w,_:=loginReq(a,ctx,"unknown@example.test","valid-length-wrong-password");if w.Code!=503{t.Errorf("waiter code=%d",w.Code)}}()};wg.Wait();if len(a.work)!=cap(a.work){t.Fatal("cancelled waiter drained another slot")};hold();if len(a.work)!=0{t.Fatal("slot leaked")}
}
