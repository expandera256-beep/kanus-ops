package privacy
import("testing";"sync";"time";"path/filepath")
func TestReviewSameAccountConcurrentFields(t *testing.T){p,e:=OpenPrefStore(filepath.Join(t.TempDir(),"p.json"));if e!=nil{t.Fatal(e)};v:=true;for i:=0;i<100;i++{p.Put(Preferences{AccountID:"a"},time.Now());var w sync.WaitGroup;w.Add(2);go func(){defer w.Done();p.Update("a",&v,nil,nil,"",time.Now())}();go func(){defer w.Done();p.Update("a",nil,&v,nil,"",time.Now())}();w.Wait();r:=p.Get("a");if !r.MarketingConsent||!r.ProductUpdates{t.Fatalf("lost field at iteration %d: %+v",i,r)}}}
