package i18n
import("testing";"os";"path/filepath")
func TestReviewSymlinkRoot(t *testing.T){d:=t.TempDir();a:=filepath.Join(d,"a");os.Mkdir(a,0700);l:=filepath.Join(d,"link");os.Symlink(a,l);p:=reviewPref(l);if e:=p.Set("en");e!=nil{t.Fatal(e)};if p.Get()!="en"{t.Fatal("not persisted")}}
func TestReviewRootRebind(t *testing.T){d:=t.TempDir();a:=filepath.Join(d,"a");b:=filepath.Join(d,"b");os.Mkdir(a,0700);os.Mkdir(b,0700);l:=filepath.Join(d,"link");os.Symlink(a,l);p:=reviewPref(l);if e:=p.Set("en");e!=nil{t.Fatal(e)};os.Remove(l);os.Symlink(b,l);if e:=p.Set("tr");e!=nil{t.Fatal(e)};if _,e:=os.Stat(filepath.Join(b,"lang.json"));!os.IsNotExist(e){t.Fatal("write escaped pinned root")}}

func reviewPref(r string)*PrefStore{return &PrefStore{Root:func()string{return r},Name:"lang.json"}}
