package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSingleExplicitTargetRejectsWrongFileProposal(t *testing.T) {
	t.Setenv("US_DESK",filepath.Join(t.TempDir(),"desk.json"))
	root:=t.TempDir();if err:=os.WriteFile(filepath.Join(root,"sample.txt"),[]byte("original\n"),0600);err!=nil{t.Fatal(err)}
	s:=newView("q8");s.root=root;s.key="test-key";s.reviewBeforeApply=true;s.modelCh=make(chan modelResult,2)
	calls:=0;s.sendModel=func(workerID,string,string)modelResult{calls++;if calls==1{return modelResult{ok:true,body:"OKU sample.txt"}};return modelResult{ok:true,body:"DOSYA yeni_deneme.txt\noriginal\n"}}
	s.askModel("sample.txt dosyasını önce OKU ile incele, sonra yalnızca içeriğini değiştir.",intent{})
	if s.proposal!=nil{t.Fatal("unrequested path was accepted")}
	if _,err:=os.Stat(filepath.Join(root,"yeni_deneme.txt"));!os.IsNotExist(err){t.Fatal("unrequested file written")}
}
