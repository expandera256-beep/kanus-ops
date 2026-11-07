package licensesrv
import("crypto/sha256";"crypto/tls";"io";"net/http";"net/http/httptest";"testing";"time")
func TestPanelTLS13Resumption(t *testing.T){
 s:=httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Write([]byte("OK"))}));s.TLS=&tls.Config{MinVersion:tls.VersionTLS13};s.StartTLS();defer s.Close()
 sum:=sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo);pins:=[][32]byte{sum};cfg:=pinnedTLS("127.0.0.1",pins);cfg.MinVersion=tls.VersionTLS13;cfg.ClientSessionCache=tls.NewLRUClientSessionCache(2)
 callback:=cfg.VerifyConnection;calls,resumed:=0,false;cfg.VerifyConnection=func(cs tls.ConnectionState)error{calls++;resumed=resumed||cs.DidResume;return callback(cs)}
 tr:=&http.Transport{TLSClientConfig:cfg,DisableKeepAlives:true};defer tr.CloseIdleConnections();cl:=&http.Client{Transport:tr,Timeout:time.Second}
 for i:=0;i<2;i++{r,e:=cl.Get(s.URL);if e!=nil{t.Fatal(e)};io.Copy(io.Discard,r.Body);r.Body.Close()}
 if calls!=2||!resumed{t.Fatalf("calls=%d resumed=%v",calls,resumed)}
 pins[0]=sha256.Sum256([]byte("revoked"));r,e:=cl.Get(s.URL);if r!=nil{r.Body.Close()};if e==nil||calls!=3{t.Fatalf("resumed session bypassed changed pin: err=%v calls=%d",e,calls)}
 t.Logf("TLS1.3 callback=%d resumed=%v rotated-pin-rejected=true",calls,resumed)
}
