package licensesrv
import("testing";"time")
func TestReviewGraceParseEdges(t *testing.T){for _,tc:=range []struct{raw string;sec int64}{
 {"",600},{" ",600},{"0",0},{"-0",0},{"-1",600},{"900",900},{"901",900},{"+901",900},{" 901 ",900},{"1s",600},{"abc",600},{"9999999999999999999999999",600},{"9223372036",900},{"9223372037",900},{"18446744074",900},{"9223372036854775807",900},
 }{t.Run(tc.raw,func(t *testing.T){t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS",tc.raw);got:=commitGraceFromEnv();want:=time.Duration(tc.sec)*time.Second;t.Logf("input=%q got=%v want=%v",tc.raw,got,want);if got!=want{t.Fail()}})}}
