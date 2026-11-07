package i18n
import("testing";"time")
func TestOwnBoundaries(t *testing.T){z:=time.FixedZone("TSI",10800);d:=time.Date(2026,11,3,0,0,0,0,z);for _,c:=range []struct{name string;now time.Time;want bool}{{"5days",d.Add(120*time.Hour),true},{"5days1sec",d.Add(120*time.Hour+time.Second),true},{"before-midnight",d.Add(144*time.Hour-time.Second),true},{"midnight-sixth",d.Add(144*time.Hour),false},{"future-at-midnight",d.Add(-time.Second),false}}{t.Run(c.name,func(t *testing.T){if FXFresh(d,c.now)!=c.want{t.Fatal(c.now)}})}
 for _,r:=range []string{"-1","0","0.00000000001"}{t.Run("rate"+r,func(t *testing.T){if _,e:=ParseFX([]byte(reviewFXDoc("2026-11-03",r,r)),d);e==nil{t.Fatal("accepted invalid rate")}})}
 b,_:=Load();for _,n:=range []int64{FXMaxTRYMinor-1,FXMaxTRYMinor,FXMaxTRYMinor+1,-1}{t.Run(time.Duration(n).String(),func(t *testing.T){a,_:=b.For("en").ApproxFX(n,FXSnapshot{Date:d,EUR:2,USD:2});if (a!="")!=(n>=0&&n<=FXMaxTRYMinor){t.Fatal(n,a)}})}
 t.Run("negative-half-cent-hidden",func(t *testing.T){a,_:=b.For("en").ApproxFX(-1,FXSnapshot{Date:d,EUR:2,USD:2});if a!=""{t.Fatal(a)}})
}
