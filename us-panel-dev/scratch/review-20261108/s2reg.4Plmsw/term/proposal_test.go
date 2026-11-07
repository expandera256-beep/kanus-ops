package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProposalReviewBeforeApplyAndRestart(t *testing.T) {
	root:=t.TempDir(); file:=filepath.Join(root,"a.txt"); if e:=os.WriteFile(file,[]byte("old\n"),0600);e!=nil{t.Fatal(e)}
	p,e:=prepareProposal(root,"DOSYA a.txt\nnew\n",map[string]bool{"a.txt":true},map[string]string{"a.txt":hashContent([]byte("old\n"))});if e!=nil||p==nil{t.Fatalf("%v",e)}
	if e=saveProposal(p);e!=nil{t.Fatal(e)};if b,_:=os.ReadFile(file);string(b)!="old\n"{t.Fatal("changed before approval")}
	s:=newView("q8");s.root=root;s.proposal=loadProposal(root);if !strings.Contains(proposalPreview(s.proposal),"new"){t.Fatal("proposal not recovered")};s.applyProposal()
	if b,_:=os.ReadFile(file);string(b)!="new\n"{t.Fatalf("%q %s",b,s.notice)};if loadProposal(root)!=nil{t.Fatal("proposal remains")};if _,e=undoLastChange(root);e!=nil{t.Fatal(e)}
}
func TestProposalRejectsFileChangedAfterReview(t *testing.T) {
	root:=t.TempDir();file:=filepath.Join(root,"a.txt");os.WriteFile(file,[]byte("old\n"),0600);p,e:=prepareProposal(root,"DOSYA a.txt\nnew",map[string]bool{"a.txt":true},map[string]string{"a.txt":hashContent([]byte("old\n"))});if e!=nil{t.Fatal(e)};os.WriteFile(file,[]byte("user\n"),0600);s:=newView("q8");s.root=root;s.proposal=p;s.applyProposal();if b,_:=os.ReadFile(file);string(b)!="user\n"{t.Fatal("user changes overwritten")}
}
func TestProposalRejectsReadDirectiveInFileBody(t *testing.T) {
	root:=t.TempDir(); original:=[]byte("print('old')\n"); file:=filepath.Join(root,"main.py"); if err:=os.WriteFile(file,original,0600);err!=nil{t.Fatal(err)}
	bad:="DOSYA main.py\nprint('ok')\nOKU backend/package.json\n{\"name\":\"x\"}\n"
	if p,err:=prepareProposal(root,bad,map[string]bool{"main.py":true},map[string]string{"main.py":hashContent(original)});err==nil||p!=nil{t.Fatal("corrupted file accepted")}
	s:=newView("q8");s.root=root;s.proposal=&fileProposal{Root:root,Files:[]proposedFile{{Path:"main.py",New:"print('ok')\nOKU backend/package.json\n{}\n",Existed:true,OldHash:hashContent(original)}}};s.applyProposal()
	if b,_:=os.ReadFile(file);string(b)!=string(original){t.Fatal("persisted bad proposal was applied")}
}
func TestProposalStopsFileAtEndMarker(t *testing.T) {
	root:=t.TempDir();p,err:=prepareProposal(root,"DOSYA a.txt\nhello\nBİTTİ\nAçıklama\n",nil,nil)
	if err!=nil||p==nil||p.Files[0].New!="hello\n"{t.Fatalf("proposal=%+v error=%v",p,err)}
}
func TestRelevantProjectPathsPrioritizesTask(t *testing.T) {
	root:=t.TempDir();os.WriteFile(filepath.Join(root,"readme.md"),[]byte("x"),0600);os.WriteFile(filepath.Join(root,"billing.go"),[]byte("x"),0600);os.WriteFile(filepath.Join(root,"secret_billing.go"),[]byte("x"),0600);got:=relevantProjectPaths(root,"billing hatasını düzelt");if len(got)!=2||got[0]!="billing.go"{t.Fatal(got)}
}
