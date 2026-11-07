package licensesrv
import("context";"crypto/sha256";"crypto/tls";"crypto/x509";"encoding/base64";"encoding/hex";"encoding/pem";"net/http";"net/http/httptest";"os";"path/filepath";"strings";"testing";"time")
func TestPanelTLSProbes(t *testing.T){
 s,hits:=tlsGPU(t);sum:=sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo);good:=hex.EncodeToString(sum[:]);bad:=sha256.Sum256([]byte("wrong"))
 for _,c:=range []struct{name,pin string;allow bool}{{"SPKI-correct",good,true},{"SPKI-wrong",hex.EncodeToString(bad[:]),false},{"PEM-correct",string(pem.EncodeToMemory(&pem.Block{Type:"CERTIFICATE",Bytes:s.Certificate().Raw})),true},{"invalid-short","sha256/abcd",false},{"invalid-empty","",false},{"base64",base64.StdEncoding.EncodeToString(sum[:]),true},{"raw-base64",base64.RawStdEncoding.EncodeToString(sum[:]),true},{"prefixed-hex","sha256:"+good,true},{"mixed-PEM-SPKI",string(pem.EncodeToMemory(&pem.Block{Type:"CERTIFICATE",Bytes:s.Certificate().Raw}))+good,false}}{
 t.Run(c.name,func(t *testing.T){g:=pinnedWaker(t,s,writePin(t,c.pin));before:=hits.Load();ok:=g.ModelReady(context.Background());if ok!=c.allow{t.Fatalf("ready=%v want=%v",ok,c.allow)};if !c.allow&&hits.Load()!=before{t.Fatal("bad pin reached app")}})
 }
 t.Run("pin-file-missing",func(t *testing.T){g:=pinnedWaker(t,s,filepath.Join(t.TempDir(),"missing"));before:=hits.Load();if g.ModelReady(context.Background())||hits.Load()!=before{t.Fatal("missing pin failed open")}})
 t.Run("callback-leaf-empty-host-date",func(t *testing.T){
 other:=*s.Certificate();other.RawSubjectPublicKeyInfo=[]byte("other key")
 cfg:=pinnedTLS("127.0.0.1",[][32]byte{sum});for _,cs:=range []tls.ConnectionState{{},{PeerCertificates:[]*x509.Certificate{&other,s.Certificate()}},{PeerCertificates:[]*x509.Certificate{&other},DidResume:true}}{if cfg.VerifyConnection(cs)==nil{t.Fatal("empty/wrong leaf accepted")}}
 if pinnedTLS("wrong.invalid",[][32]byte{sum}).VerifyConnection(tls.ConnectionState{PeerCertificates:[]*x509.Certificate{s.Certificate()}})==nil{t.Fatal("hostname not checked")}
 expired:=*s.Certificate();expired.NotAfter=time.Now().Add(-time.Minute);if cfg.VerifyConnection(tls.ConnectionState{PeerCertificates:[]*x509.Certificate{&expired}})==nil{t.Fatal("expired accepted")}
 if pinnedTLS("127.0.0.1",nil).VerifyConnection(tls.ConnectionState{PeerCertificates:[]*x509.Certificate{s.Certificate()}})==nil{t.Fatal("nil pins accepted")}
 })
 t.Run("PEM-normal-hostname",func(t *testing.T){cfg:=pemTLS("wrong.invalid",[]*x509.Certificate{s.Certificate()});if cfg.InsecureSkipVerify||cfg.RootCAs==nil{t.Fatal("normal verification disabled")};cl:=&http.Client{Transport:&http.Transport{TLSClientConfig:cfg},Timeout:time.Second};if _,e:=cl.Get(s.URL);e==nil{t.Fatal("PEM wrong host accepted")}})
}
func TestPanelTLSResumption(t *testing.T){
 s:=httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){}));s.TLS=&tls.Config{MaxVersion:tls.VersionTLS12};s.StartTLS();defer s.Close()
 sum:=sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo);pins:=[][32]byte{sum};cfg:=pinnedTLS("127.0.0.1",pins);cfg.MaxVersion=tls.VersionTLS12;cfg.ClientSessionCache=tls.NewLRUClientSessionCache(2)
 callback:=cfg.VerifyConnection;calls,resumed:=0,false;cfg.VerifyConnection=func(cs tls.ConnectionState)error{calls++;resumed=resumed||cs.DidResume;return callback(cs)}
 for i:=0;i<2;i++{c,e:=tls.Dial("tcp",s.Listener.Addr().String(),cfg);if e!=nil{t.Fatal(e)};c.Close()}
 if calls!=2||!resumed{t.Fatalf("calls=%d resumed=%v",calls,resumed)}
 pins[0]=sha256.Sum256([]byte("revoked"));c,e:=tls.Dial("tcp",s.Listener.Addr().String(),cfg);if c!=nil{c.Close()};if e==nil||calls!=3{t.Fatalf("resumed session bypassed rotated pin: err=%v calls=%d",e,calls)}
}
func TestPanelHTTPAndEnv(t *testing.T){
 t.Setenv("US_GPU_WAKE_PIN_FILE","");t.Setenv("US_GPU_WAKE_HTTP_ALLOW","");t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","0")
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Write([]byte(`{"model_ready":true}`))}));defer s.Close();tok:=filepath.Join(t.TempDir(),"token");os.WriteFile(tok,[]byte("local-probe-token"),0600)
 g,e:=NewGPUWaker(s.URL+"/a100/start",s.URL+"/a100/status",tok,0);if e!=nil||!g.ModelReady(context.Background()){t.Fatal("loopback http refused",e)}
 for _,flag:=range []string{"0","1"}{t.Setenv("US_GPU_WAKE_ALLOW_INSECURE_HTTP",flag);if _,e:=NewGPUWaker("http://gpu.invalid/a100/start","http://gpu.invalid/a100/status",tok,0);e==nil{t.Fatal("legacy broad HTTP flag bypassed exact allowlist")}}
 t.Setenv("US_GPU_WAKE_HTTP_ALLOW","gpu.invalid:80");if _,e:=NewGPUWaker("http://gpu.invalid/a100/start","http://gpu.invalid/a100/status",tok,0);e!=nil{t.Fatal("exact allowlist rejected",e)}
 t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","1");g,e=NewGPUWaker("https://gpu.invalid/a100/start","https://gpu.invalid/a100/status",tok,0);if e!=nil||g.PinFile!=""{t.Fatal("CA opt-in failed")};t.Setenv("US_GPU_WAKE_PIN_FILE",filepath.Join(t.TempDir(),"missing"));g,_=NewGPUWaker("https://gpu.invalid/a100/start","https://gpu.invalid/a100/status",tok,0);if _,e=g.client(g.WakeURL);e==nil{t.Fatal("explicit missing pin ignored")}
}
func TestPanelMixedOriginsNeedPin(t *testing.T){
 t.Setenv("US_GPU_WAKE_PIN_FILE","");t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","0")
 for _,wake:=range []string{"http://127.0.0.1:1/a100/start","https://localhost:1/a100/start"}{g,e:=NewGPUWaker(wake,"https://gpu.invalid/a100/status","unused",0);if e!=nil{continue};if strings.TrimSpace(g.PinFile)==""{t.Errorf("non-loopback status lost mandatory pin with wake=%s",wake)}}
}
