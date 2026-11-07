package licensesrv
import("testing")
func TestReviewMailMatching(t *testing.T){for _,tc:=range []struct{name,to string;oid int64;email string;want bool}{
 {"case","USER@EXAMPLE.TEST",0,"user@example.test",true},
 {"space"," user@example.test ",0,"user@example.test",true},
 {"display","User <user@example.test>",0,"user@example.test",true},
 {"multi","user@example.test, other@example.test",0,"user@example.test",true},
 {"empty","",0,"",false},{"zero-order","other@example.test",0,"user@example.test",false},
 }{t.Run(tc.name,func(t *testing.T){got:=mailBelongsToAccount(&Mail{To:tc.to,OrderID:tc.oid},tc.email,map[int64]bool{0:true,42:true});if got!=tc.want{t.Fatalf("got %v want %v",got,tc.want)}})}}
