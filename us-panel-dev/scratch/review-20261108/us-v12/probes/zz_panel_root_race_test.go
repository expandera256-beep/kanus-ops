//go:build linux
package main
import("os";"path/filepath";"sync/atomic";"testing";"time";"golang.org/x/sys/unix")
func TestPanelIndependentRootSwapWriteDelete(t *testing.T){
 for _,op:=range []string{"write","delete"}{t.Run(op,func(t *testing.T){
 p:=t.TempDir();root:=filepath.Join(p,"root");outside:=filepath.Join(p,"outside");link:=filepath.Join(p,"link")
 os.Mkdir(root,0700);os.Mkdir(outside,0700);os.Symlink(outside,link)
 os.WriteFile(filepath.Join(root,"victim"),[]byte("inside"),0600);os.WriteFile(filepath.Join(outside,"victim"),[]byte("outside"),0600)
 ft,err:=newFileTools(root);if err!=nil{t.Fatal(err)}
 var stop atomic.Bool;done:=make(chan struct{});go func(){defer close(done);for !stop.Load(){_ = unix.Renameat2(unix.AT_FDCWD,root,unix.AT_FDCWD,link,unix.RENAME_EXCHANGE)}}()
 attempts,success:=0,0;escaped:=false
 for end:=time.Now().Add(8*time.Second);time.Now().Before(end);{
 attempts++;var e error;if op=="write"{e=ft.Write("written",[]byte("probe"))}else{e=ft.Delete("victim")};if e==nil{success++}
 if op=="write"{_,e=os.Stat(filepath.Join(outside,"written"));escaped=e==nil}else{_,e=os.Stat(filepath.Join(outside,"victim"));escaped=os.IsNotExist(e)}
 if escaped{break}
 }
 stop.Store(true);<-done;t.Logf("attempts=%d success=%d outsideChanged=%v",attempts,success,escaped)
 if escaped{t.Fatal("root-swap escaped pinned workspace boundary")}
 })}
}
func TestPanelChildSwapWriteDelete(t *testing.T){
 root,out:=t.TempDir(),t.TempDir();os.Mkdir(filepath.Join(root,"realdir"),0700);ft,_:=newFileTools(root);os.WriteFile(filepath.Join(out,"victim"),[]byte("outside"),0600)
 stop:=startDirSwap(root,out);attempts,success:=0,0
 for end:=time.Now().Add(3*time.Second);time.Now().Before(end);{attempts++;if ft.Write("d/x",[]byte("x"))==nil{success++};_ = ft.Delete("d/victim")}
 stop();ents,_:=os.ReadDir(out);t.Logf("attempts=%d successfulWrites=%d outsideNames=%v",attempts,success,outsideNames(out));if len(ents)!=1||ents[0].Name()!="victim"{t.Fatal("child symlink escape")}
}
