package privacy
import("testing";"time";"net/http/httptest")
func TestReviewV4RetentionExactBounds(t *testing.T){for _,c:=range []struct{raw string;days int}{ {"",90},{"1",1},{"3650",3650},{"3651",90},{"0",90},{"-1",90},{"213504",90},{"9223372036854775807",90},{"9999999999999999999999",90},{"1d",90},{" ",90} }{t.Run(c.raw,func(t *testing.T){t.Setenv("US_PRIVACY_RETENTION_DAYS",c.raw);got:=NewHandlerFromEnv(nil,nil).Keep;want:=time.Duration(c.days)*24*time.Hour;if got!=want{t.Fatalf("input=%q got=%v want=%v",c.raw,got,want)}})}}
func TestReviewV4OriginMatrix(t *testing.T){for _,c:=range []struct{o,allow,host string;want bool}{
 {"http://localhost.evil.com","http://localhost.evil.com","api.example",false},
 {"http://127.0.0.1.nip.io","http://127.0.0.1.nip.io","api.example",false},
 {"https://user@api.example","https://user@api.example","api.example",false},
 {"http://user@localhost:3000","http://user@localhost:3000","api.example",false},
 {"http://[::1]:3000","http://[::1]:3000","api.example",true},
 {"http://[::1]:3001","http://[::1]:3000","api.example",false},
 {"http://localhost:3000","","localhost:3000",false},
 {"http://localhost:3000","http://localhost:3000","api.example",true},
 {"https://api.example:443","","api.example:443",true},
 {"https://api.example:444","","api.example:443",false},
 {"https://api.example/x","https://api.example/x","api.example",false},
 {"ftp://api.example","ftp://api.example","api.example",false},
 }{t.Run(c.o,func(t *testing.T){t.Setenv("US_PRIVACY_ALLOWED_ORIGINS",c.allow);r:=httptest.NewRequest("POST","https://"+c.host+"/",nil);r.Header.Set("Origin",c.o);if got:=originOK(r);got!=c.want{t.Fatalf("got=%v want=%v",got,c.want)}})}}
