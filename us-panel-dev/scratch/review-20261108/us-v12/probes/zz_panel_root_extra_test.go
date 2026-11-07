//go:build linux
package main
import("os";"path/filepath";"testing";"golang.org/x/sys/unix")
func TestPanelIndependentPinnedMkdirRename(t *testing.T){
 p:=t.TempDir();root,out,link:=filepath.Join(p,"root"),filepath.Join(p,"outside"),filepath.Join(p,"link");os.Mkdir(root,0700);os.Mkdir(out,0700);os.Symlink(out,link);ft,e:=newFileTools(root);if e!=nil{t.Fatal(e)};defer ft.root.Close()
 // Swap exactly after temp fsync but before renameat. The parent path now points out.
 atomicWriteTestHook=func(string)error{return unix.Renameat2(unix.AT_FDCWD,root,unix.AT_FDCWD,link,unix.RENAME_EXCHANGE)};defer func(){atomicWriteTestHook=nil}()
 if e=ft.Write("new/deep/x",[]byte("inside"));e!=nil{t.Fatal(e)}
 if _,e=os.Stat(filepath.Join(out,"new"));!os.IsNotExist(e){t.Fatal("mkdir escaped",e)}
 if got,e:=os.ReadFile(filepath.Join(link,"new/deep/x"));e!=nil||string(got)!="inside"{t.Fatal("rename did not stay at pinned root",e)}
}
