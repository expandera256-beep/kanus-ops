package main
import("os";"path/filepath";"testing")
// Overlay-only test of the actual panel file and terminal boundaries.
func TestQASec8PanelFileIsolation(t *testing.T) {
 parent:=t.TempDir();a:=filepath.Join(parent,"A");b:=filepath.Join(parent,"B")
 if e:=os.MkdirAll(a,0700);e!=nil {t.Fatal(e)};if e:=os.MkdirAll(b,0700);e!=nil {t.Fatal(e)}
 sentinel:=[]byte("B-private-sentinel");os.WriteFile(filepath.Join(b,"private.txt"),sentinel,0600)
 for _,path:=range []string{"../B/private.txt",filepath.Join(b,"private.txt")} {
  if _,_,e:=sourceSnapshot(a,path);e==nil {t.Fatal("cross-workspace read accepted",path)}
 }
 if e:=os.Symlink(b,filepath.Join(a,"B-link"));e!=nil {t.Fatal(e)}
 if _,_,e:=sourceSnapshot(a,"B-link/private.txt");e==nil {t.Fatal("symlink escape accepted")}
 text:="print('injected')\n"
 for _,path:=range []string{"../B/injected.py",filepath.Join(b,"injected.py"),"B-link/injected.py"} {
  envelope:=&modelActionEnvelope{Version:1,Operations:[]modelOperation{{Op:"create_file",Path:path,Content:&text}}}
  if _,e:=prepareActionProposal(a,envelope,nil,nil,nil);e==nil {t.Fatal("cross-workspace write proposal accepted",path)}
 }
 got,_:=os.ReadFile(filepath.Join(b,"private.txt"));if string(got)!=string(sentinel) {t.Fatal("B sentinel changed")}
 if _,e:=os.Stat(filepath.Join(b,"injected.py"));!os.IsNotExist(e) {t.Fatal("B injection file exists")}
 t.Log("PASS actual panel sourceSnapshot + prepareActionProposal deny traversal, absolute path and symlink across A/B workspaces")
}
func TestQASec8PanelCommandInjection(t *testing.T) {
 root:=t.TempDir()
 for _,raw:=range []string{`["sh","-c","touch qa-injected"]`,`["python3","../private.py"]`,`["go","test",";touch qa-injected"]`,`["node","--eval","1"]`} {
  if _,e:=terminalArgv(raw,root);e==nil {t.Fatal("unsafe terminal argv accepted",raw)}
 }
 if _,e:=os.Stat(filepath.Join(root,"qa-injected"));!os.IsNotExist(e) {t.Fatal("injection marker exists")}
 t.Log("PASS actual terminal argv validator denies shell, traversal, shell metacharacters and eval; no marker created")
}
