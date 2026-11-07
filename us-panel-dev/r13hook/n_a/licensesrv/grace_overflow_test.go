package licensesrv

import (
	"testing"
	"time"
)

// PANEL DEV R3-A01: 16 uç durum (PANEL_API_REVIEW_R3FIX_20261106.md).
func TestCommitGraceFromEnvOverflowEdges(t *testing.T) {
	def, max := 10*time.Minute, 15*time.Minute
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"", def}, {" ", def}, {"0", 0}, {"-0", 0}, {"-1", def},
		{"900", max}, {"901", max}, {"+901", max}, {" 901 ", max},
		{"1s", def}, {"abc", def}, {"9999999999999999999999999", def},
		{"9223372036", max}, {"9223372037", max}, {"18446744074", max},
		{"9223372036854775807", max},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", c.in)
			if got := commitGraceFromEnv(); got != c.want {
				t.Fatalf("input=%q got=%v want=%v", c.in, got, c.want)
			}
		})
	}
}
