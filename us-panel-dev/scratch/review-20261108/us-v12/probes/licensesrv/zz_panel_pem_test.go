package licensesrv
import("context";"crypto/sha256";"crypto/tls";"crypto/x509";"encoding/hex";"net/http";"path/filepath";"testing";"time")
func TestPanelIndependentPEMOnlyPins(t *testing.T){
 s,hits:=tlsGPU(t);sum:=sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo)
 for _,c:=range []struct{name,pin string;ok bool}{{"good-PEM",certPEMOf(s),true},{"wrong-PEM",unrelatedPEM(t),false},{"SPKI-correct-key",hex.EncodeToString(sum[:]),false},{"invalid","bad pin",false},{"empty","",false}}{t.Run(c.name,func(t *testing.T){g:=pinnedWaker(t,s,writePin(t,c.pin));before:=hits.Load();if got:=g.ModelReady(context.Background());got!=c.ok{t.Fatalf("ready=%v want=%v",got,c.ok)};if !c.ok&&hits.Load()!=before{t.Fatal("rejected pin reached application")};if c.ok{cl,e:=g.client(g.StatusURL);if e!=nil{t.Fatal(e)};cfg:=cl.Transport.(*http.Transport).TLSClientConfig;if cfg.InsecureSkipVerify||cfg.RootCAs==nil||cfg.VerifyConnection==nil{t.Fatal("PEM normal validation not enforced")}}})}
 t.Run("missing-file",func(t *testing.T){g:=pinnedWaker(t,s,filepath.Join(t.TempDir(),"missing"));if g.ModelReady(context.Background()){t.Fatal("missing pin failed open")}})
 t.Run("hostname-and-empty-peer",func(t *testing.T){cfg:=pemTLS("wrong.invalid",[]*x509.Certificate{s.Certificate()});cl:=&http.Client{Timeout:time.Second,Transport:&http.Transport{TLSClientConfig:cfg}};if _,e:=cl.Get(s.URL);e==nil{t.Fatal("PEM wrong hostname accepted")};if cfg.VerifyConnection(tls.ConnectionState{})==nil{t.Fatal("empty peer accepted")}})
}
func TestPanelIndependentMixedOriginPins(t *testing.T){
 t.Setenv("US_GPU_WAKE_PIN_FILE","");t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","0")
 for _,wake:=range []string{"http://127.0.0.1:1/a100/start","https://localhost:1/a100/start"}{status:="https://gpu.invalid/a100/status";g,e:=NewGPUWaker(wake,status,"unused",0);if e!=nil{t.Fatal(e)};if g.PinFile==""||!g.pinnedFor(status)||g.pinnedFor(wake){t.Fatal("mixed origin pin decision incorrect")};g.PinFile=filepath.Join(t.TempDir(),"missing");if _,e=g.client(status);e==nil{t.Fatal("remote missing pin failed open")}}
 t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA","1");g,e:=NewGPUWaker("https://gpu.invalid/a100/start","https://gpu.invalid/a100/status","unused",time.Minute);if e!=nil||g.PinFile!=""{t.Fatal("CA opt-in incorrect")};t.Setenv("US_GPU_WAKE_PIN_FILE",filepath.Join(t.TempDir(),"missing"));g,_=NewGPUWaker("https://gpu.invalid/a100/start","https://gpu.invalid/a100/status","unused",0);if _,e=g.client(g.StatusURL);e==nil{t.Fatal("explicit pin bypassed by CA opt-in")}
}
