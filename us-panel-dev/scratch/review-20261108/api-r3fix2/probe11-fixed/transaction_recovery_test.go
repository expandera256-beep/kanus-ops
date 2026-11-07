package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTransactionSurvivesProcessExit(t *testing.T) {
	if root:=os.Getenv("US_TEST_INTERRUPTED_ROOT");root!="" {
		p,err:=prepareProposal(root,"DOSYA a.txt\nnew\n",map[string]bool{"a.txt":true},map[string]string{"a.txt":hashContent([]byte("old\n"))})
		if err!=nil {os.Exit(71)}
		a:=filepath.Join(root,"a.txt")
		if beginTransaction(root,[]proposalBatchItem{{file:p.Files[0],full:a}})!=nil {os.Exit(72)}
		if writeModelFile(root,a,[]byte("new\n"))!=nil {os.Exit(73)}
		os.Exit(70)
	}
	root:=t.TempDir();a:=filepath.Join(root,"a.txt");if err:=os.WriteFile(a,[]byte("old\n"),0600);err!=nil{t.Fatal(err)}
	cmd:=exec.Command(os.Args[0],"-test.run=^TestTransactionSurvivesProcessExit$")
	cmd.Env=append(os.Environ(),"US_TEST_INTERRUPTED_ROOT="+root)
	if err:=cmd.Run();err==nil{t.Fatal("child unexpectedly exited successfully")}
	if !pendingTransaction(root){t.Fatal("journal did not survive process exit")}
	if err:=recoverProposalBatch(root);err!=nil{t.Fatal(err)}
	if b,_:=os.ReadFile(a);string(b)!="old\n"{t.Fatalf("wrong recovery: %q",b)}
}

func TestRecoverInterruptedProposal(t *testing.T) {
	root:=t.TempDir();a:=filepath.Join(root,"a.txt");if err:=os.WriteFile(a,[]byte("old\n"),0600);err!=nil{t.Fatal(err)}
	p,err:=prepareProposal(root,"DOSYA a.txt\nnew\nDOSYA b.txt\nsecond\n",map[string]bool{"a.txt":true},map[string]string{"a.txt":hashContent([]byte("old\n"))});if err!=nil{t.Fatal(err)}
	var items []proposalBatchItem;for _,f:=range p.Files{full,err:=checkProposalVersion(root,f);if err!=nil{t.Fatal(err)};items=append(items,proposalBatchItem{file:f,full:full})}
	if err=beginTransaction(root,items);err!=nil{t.Fatal(err)}
	if err=writeModelFile(root,a,[]byte("new\n"));err!=nil{t.Fatal(err)}
	if _,err=applyProposalBatch(root,p.Files,writeModelFile);err==nil{t.Fatal("pending journal ignored")}
	if err=recoverProposalBatch(root);err!=nil{t.Fatal(err)}
	if b,_:=os.ReadFile(a);string(b)!="old\n"{t.Fatalf("wrong recovery: %q",b)}
	if pendingTransaction(root){t.Fatal("journal remains")}
}

func TestRecoveryDoesNotOverwriteExternalEdits(t *testing.T) {
	root:=t.TempDir();a:=filepath.Join(root,"a.txt");if err:=os.WriteFile(a,[]byte("old\n"),0600);err!=nil{t.Fatal(err)}
	p,err:=prepareProposal(root,"DOSYA a.txt\nnew\n",map[string]bool{"a.txt":true},map[string]string{"a.txt":hashContent([]byte("old\n"))});if err!=nil{t.Fatal(err)}
	if err=beginTransaction(root,[]proposalBatchItem{{file:p.Files[0],full:a}});err!=nil{t.Fatal(err)}
	if err=os.WriteFile(a,[]byte("user edit"),0600);err!=nil{t.Fatal(err)}
	if err=recoverProposalBatch(root);err==nil{t.Fatal("external edit ignored")}
	if b,_:=os.ReadFile(a);string(b)!="user edit" {t.Fatalf("edit lost: %q",b)}
	if !pendingTransaction(root){t.Fatal("journal lost")}
}
