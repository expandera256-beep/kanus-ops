package main

import (
 "context"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "sync/atomic"
 "testing"
 "time"
)

func TestQAAuditCommandInjection(t *testing.T) {
 root:=t.TempDir()
 for _,raw:=range []string{"go test; echo AUDIT", "go test && echo AUDIT", "sh -c echo", "python3 -c print(1)"} {
  if _,err:=terminalArgv(raw,root);err==nil {t.Fatal("unsafe terminal argv accepted")}
 }
}
func TestQAAuditPathTraversal(t *testing.T) {
 root:=t.TempDir(); sibling:=t.TempDir();os.WriteFile(filepath.Join(sibling,"synthetic.txt"),[]byte("synthetic-only"),0600)
 ft,err:=newFileTools(root);if err!=nil {t.Fatal("fixture init failed")}
 for _,p:=range []string{"../synthetic.txt",filepath.Join(sibling,"synthetic.txt"),`..\synthetic.txt`} {
  if _,e:=ft.Read(p);e==nil {t.Fatal("fileTools accepted outside path")}
 }
 if err:=os.Symlink(sibling,filepath.Join(root,"outside"));err==nil {
  if _,e:=ft.Read("outside/synthetic.txt");e==nil {t.Fatal("fileTools accepted symlink escape")}
 }
}
func TestQAAuditChildEnvironment(t *testing.T) {
 sentinel:="AUDIT_SYNTHETIC_ENV_ONLY_20261108"
 t.Setenv("SHOPIFY_WEBHOOK_SECRET",sentinel)
 t.Setenv("GOOGLE_APPLICATION_CREDENTIALS",sentinel)
 if strings.Contains(strings.Join(safeExecEnv(),"\n"),sentinel) {t.Fatal("sensitive environment inherited")}
}
// Negative regression: expected denial for a multi-user untrusted-code boundary.
// All files are newly generated synthetic files in t.TempDir. Never reads HOME/secrets.
func TestQAAuditExecSiblingIsolation(t *testing.T) {
 parent:=t.TempDir();root:=filepath.Join(parent,"user-a");sibling:=filepath.Join(parent,"user-b")
 os.Mkdir(root,0700);os.Mkdir(sibling,0700)
 sentinel:="AUDIT_SYNTHETIC_OTHER_WORKSPACE_ONLY"
 outside:=filepath.Join(sibling,"synthetic.txt");os.WriteFile(outside,[]byte(sentinel),0600)
 script:=filepath.Join(root,"main.py")
 os.WriteFile(script,[]byte("import pathlib,sys\nprint(pathlib.Path(sys.argv[1]).read_text())\n"),0600)
 out,state,_,err:=safeRun(context.Background(),root,root,[]string{"python3",script,outside},5*time.Second,nil)
 leaked:=err==nil && strings.Contains(out,sentinel)
 t.Logf("synthetic_sibling_read=%t state=%s",leaked,state)
 if leaked {t.Fatal("untrusted project can read sibling workspace: cwd/resource limits are not filesystem sandbox")}
}
// Local redirect only; never contacts metadata services or real credential endpoints.
func TestQAAuditRepairRedirectIsolation(t *testing.T) {
 var contacted,authForwarded,bodyForwarded atomic.Bool
 sentinel:="AUDIT_SYNTHETIC_REPAIR_KEY_ONLY"
 sink:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  contacted.Store(true);authForwarded.Store(r.Header.Get("Authorization")=="Bearer "+sentinel)
  bodyForwarded.Store(r.Method=="POST" && r.ContentLength>0)
  w.Header().Set("Content-Type","application/json")
  w.Write([]byte(`{"choices":[{"message":{"content":"{\"files\":[]}"}}]}`))
 }));defer sink.Close()
 origin:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){http.Redirect(w,r,sink.URL+"/sink",http.StatusTemporaryRedirect)}));defer origin.Close()
 root:=t.TempDir();os.WriteFile(filepath.Join(root,"main.go"),[]byte("package main\nfunc main(){}\n"),0600)
 ft,err:=newFileTools(root);if err!=nil {t.Fatal("fixture init failed")}
 loop:=&repairLoop{Root:root,Endpoint:origin.URL,Key:sentinel,Model:"qa-mock"}
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel()
 _,_=loop.askFix(ctx,ft,repairStep{State:"failed",Stderr:"synthetic build failure"})
 t.Logf("redirect_contacted=%t authorization_forwarded=%t project_body_forwarded=%t",contacted.Load(),authForwarded.Load(),bodyForwarded.Load())
 if contacted.Load() {t.Fatal("repair default HTTP client followed cross-port redirect; expected refusal")}
}
