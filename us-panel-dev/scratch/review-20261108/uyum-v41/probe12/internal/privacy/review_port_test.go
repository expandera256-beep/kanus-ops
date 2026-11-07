package privacy
import("testing";"net/http/httptest")
func TestReviewInvalidOriginPort(t *testing.T){for _,o:=range []string{"http://localhost:65536","http://[::1]:65536","https://api.example:65536","http://localhost:"}{t.Run(o,func(t *testing.T){t.Setenv("US_PRIVACY_ALLOWED_ORIGINS",o);r:=httptest.NewRequest("POST","https://api.example/",nil);r.Header.Set("Origin",o);if originOK(r){t.Fatal("invalid/empty port accepted when allowlisted")}})}}
