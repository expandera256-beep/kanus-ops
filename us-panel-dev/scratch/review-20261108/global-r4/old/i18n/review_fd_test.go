package i18n
import("testing";"os";"path/filepath";"fmt";"runtime")
func TestReviewRootFDGrowth(t *testing.T){d:=t.TempDir();r:=d;p:=&PrefStore{Root:func()string{return r}};before,_:=os.ReadDir("/proc/self/fd");for i:=0;i<64;i++{r=filepath.Join(d,fmt.Sprint(i));if e:=p.Set("en");e!=nil{t.Fatal(e)}};runtime.GC();after,_:=os.ReadDir("/proc/self/fd");t.Logf("fds before=%d after=%d roots=%d",len(before),len(after),len(p.roots));if len(after)-len(before)>4{t.Fatal("root handles accumulate without Close/eviction")}}
