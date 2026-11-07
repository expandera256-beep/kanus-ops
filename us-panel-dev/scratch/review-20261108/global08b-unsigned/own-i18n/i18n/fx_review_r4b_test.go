package i18n
import("fmt";"math";"strings";"testing";"time")
func reviewFXDoc(date, eur, usd string) string{return fmt.Sprintf(`{"schema":"kanus.fx/1","base":"TRY","date":%q,"rates":{"EUR":%s,"USD":%s}}`,date,eur,usd)}
func TestReviewFXPolicy(t *testing.T){utc:=time.UTC;tsi:=time.FixedZone("TSİ",3*3600);now:=time.Date(2026,11,8,12,0,0,0,utc);good:=reviewFXDoc("2026-11-08","55","50");cs:=[]struct{name,raw string;now time.Time;bad bool}{
{"today",good,now,false},
{"fifth_day_exact",reviewFXDoc("2026-11-03","55","50"),time.Date(2026,11,8,0,0,0,0,utc),false},
{"fifth_calendar_day_tsi_end",reviewFXDoc("2026-11-03","55","50"),time.Date(2026,11,8,23,59,59,0,tsi),false},
{"same_local_day_before_utc",good,time.Date(2026,11,8,0,1,0,0,tsi),false},
{"sixth_day_utc",reviewFXDoc("2026-11-02","55","50"),time.Date(2026,11,8,0,0,0,0,utc),true},
{"sixth_day_tsi",reviewFXDoc("2026-11-02","55","50"),time.Date(2026,11,8,0,0,0,0,tsi),true},
{"sixth_day_plus_second",reviewFXDoc("2026-11-02","55","50"),time.Date(2026,11,8,0,0,1,0,utc),true},
{"tomorrow",reviewFXDoc("2026-11-09","55","50"),now,true},
{"future_36h",reviewFXDoc("2026-11-10","55","50"),now,true},
{"future_over_36h",reviewFXDoc("2026-11-10","55","50"),now.Add(-time.Second),true},
{"injected_clock_2030",reviewFXDoc("2030-01-01","55","50"),time.Date(2030,1,1,12,0,0,0,utc),false},
{"missing_eur",`{"schema":"kanus.fx/1","base":"TRY","date":"2026-11-08","rates":{"USD":50}}`,now,true},
{"wrong_second_key",`{"schema":"kanus.fx/1","base":"TRY","date":"2026-11-08","rates":{"USD":50,"GBP":55}}`,now,true},
{"wrong_schema",strings.Replace(good,"kanus.fx/1","kanus.fx/2",1),now,true},
{"wrong_base",strings.Replace(good,`"TRY"`,`"EUR"`,1),now,true},
{"bad_calendar",reviewFXDoc("2026-02-29","55","50"),now,true},
{"trailing_json",good+`{}`,now,true},
{"trailing_space",good+" \n",now,false},
{"size_4096",good+strings.Repeat(" ",4096-len(good)),now,false},
{"size_4097",good+strings.Repeat(" ",4097-len(good)),now,true},
{"cross_min",reviewFXDoc("2026-11-08","80","100"),now,false},
{"cross_max",reviewFXDoc("2026-11-08","160","100"),now,false},
{"cross_below",reviewFXDoc("2026-11-08","79.999","100"),now,true},
{"cross_above",reviewFXDoc("2026-11-08","160.001","100"),now,true},
}
for _,v:=range []string{"0","-0","-1","null","NaN","Inf","-Inf","1e309",`"NaN"`,`"55"`}{cs=append(cs,struct{name,raw string;now time.Time;bad bool}{"invalid_rate_"+v,reviewFXDoc("2026-11-08",v,"50"),now,true})}
for _,c:=range cs{t.Run(c.name,func(t *testing.T){s,e:=ParseFX([]byte(c.raw),c.now);t.Logf("now=%s date=%s parsed=%+v err=%v",c.now.Format(time.RFC3339),c.raw,s,e);if (e!=nil)!=c.bad{t.Fatalf("reject=%v want=%v",e!=nil,c.bad)}})}}
func TestReviewFXRounding(t *testing.T){b,e:=Load();if e!=nil{t.Fatal(e)};for _,c:=range []struct{name string;minor int64;rate,want string}{{"half_cent_29_rate2",29,"2","≈ 0.15 € · ≈ $0.15"},{"half_cent_21_rate6",21,"6","≈ 0.04 € · ≈ $0.04"}}{t.Run(c.name,func(t *testing.T){now:=time.Date(2026,11,8,12,0,0,0,time.UTC);s,e:=ParseFX([]byte(reviewFXDoc("2026-11-08",c.rate,c.rate)),now);if e!=nil{t.Fatal(e)};a,_:=b.For("en").ApproxFX(c.minor,s);t.Logf("minor=%d rate=%s got=%q want=%q",c.minor,c.rate,a,c.want);if a!=c.want{t.Fatal("decimal half-cent rounding")}})}}
func TestReviewFXApproxRobustness(t *testing.T){b,e:=Load();if e!=nil{t.Fatal(e)};d:=time.Date(2026,11,8,0,0,0,0,time.UTC);for name,s:=range map[string]FXSnapshot{"nan":{Date:d,EUR:math.NaN(),USD:math.NaN()},"inf":{Date:d,EUR:math.Inf(1),USD:math.Inf(1)}}{t.Run(name,func(t *testing.T){a,n:=b.For("en").ApproxFX(100,s);t.Logf("approx=%q note=%q",a,n);if a!=""||n!=""{t.Fatal("invalid nonfinite snapshot must hide FX")}})};t.Run("max_int64",func(t *testing.T){s,e:=ParseFX([]byte(reviewFXDoc("2026-11-08","1","1")),d);if e!=nil{t.Fatal(e)};a,_:=b.For("en").ApproxFX(math.MaxInt64,s);t.Logf("approx=%q",a);if strings.Contains(a,"-"){t.Fatal("nonnegative TRY amount rendered negative after float-to-int64 overflow")}})}
