//go:build unix
package privacy
import("testing";"os";"syscall";"path/filepath";"time")
func TestReviewOpenHelperAcceptsFIFO(t *testing.T){p:=filepath.Join(t.TempDir(),"fifo");if e:=syscall.Mkfifo(p,0600);e!=nil{t.Fatal(e)};done:=make(chan error,1);go func(){f,e:=openPrefsNoFollow(p);if e==nil{defer f.Close();fi,_:=f.Stat();if fi.Mode()&os.ModeNamedPipe!=0{e=nil}};done<-e}();select{case <-done:t.Log("returned without writer");return;case <-time.After(100*time.Millisecond):};w,e:=os.OpenFile(p,os.O_WRONLY,0);if e!=nil{t.Fatal(e)};w.Close();e=<-done;if e==nil{t.Fatal("helper blocked and accepted FIFO; no opened-handle type validation")}}
