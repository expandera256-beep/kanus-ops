package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLongCheckpointCorruptRecordIsPreserved(t *testing.T) {
	s:=newView("q8");s.root=t.TempDir();s.reviewBeforeApply=true
	dir:=filepath.Join(s.root,taskDir);if err:=os.Mkdir(dir,0700);err!=nil{t.Fatal(err)}
	path:=taskPath(s.root);original:=[]byte("{broken checkpoint");if err:=os.WriteFile(path,original,0600);err!=nil{t.Fatal(err)}
	calls:=0
	_,r:=s.summarizeLongTextWithLimit(strings.Repeat("görev ",1000),func(string,int)modelResult{calls++;return modelResult{ok:true,body:"not"}},promptWireBudget)
	if r.ok || calls!=0 || !strings.Contains(r.body,"korunarak") { t.Fatalf("%+v calls=%d",r,calls) }
	got,err:=os.ReadFile(path);if err!=nil || string(got)!=string(original){t.Fatalf("checkpoint overwritten: %q %v",got,err)}
}

func TestLongCheckpointResumesOnlyFinishedChunks(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "")
	s:=newView("q8");s.root=t.TempDir();s.reviewBeforeApply=true
	text:=strings.Repeat("deneme paragrafı ",700)
	firstCalls:=0
	_,first:=s.summarizeLongTextWithLimit(text,func(_ string,_ int)modelResult{firstCalls++;if firstCalls==2{return modelResult{body:"bağlantı kesildi"}};return modelResult{ok:true,body:"kayıtlı özet"}},promptWireBudget)
	if first.ok||firstCalls!=2{t.Fatalf("expected interrupted run, calls=%d",firstCalls)}
	cp:=loadLongCheckpoint(s.root);if cp==nil||len(cp.Summaries)!=1{t.Fatalf("checkpoint missing: %+v",cp)}
	mapCalls:=0
	_,second:=s.summarizeLongTextWithLimit(text,func(wire string,_ int)modelResult{if strings.Contains(wire,"--- PARÇA ---") { mapCalls++ };return modelResult{ok:true,body:"kısa not"}},promptWireBudget)
	if !second.ok{t.Fatal(second.body)}
	want:=len(splitLongText(text,longChunkBytes))-1;if mapCalls!=want{t.Fatalf("mapCalls=%d want=%d",mapCalls,want)}
}

func TestLongCheckpointRejectsDifferentTask(t *testing.T) {
	s:=newView("q8");s.root=t.TempDir();s.reviewBeforeApply=true
	a:=strings.Repeat("ilk görev ",1000);b:=strings.Repeat("başka görev ",1000)
	_,_=s.summarizeLongTextWithLimit(a,func(string,int)modelResult{return modelResult{body:"kesildi"}},promptWireBudget)
	calls:=0;_,r:=s.summarizeLongTextWithLimit(b,func(string,int)modelResult{calls++;return modelResult{ok:true,body:"not"}},promptWireBudget)
	if r.ok||calls!=0||!strings.Contains(r.body,"uyumsuz"){t.Fatalf("%+v calls=%d",r,calls)}
}
