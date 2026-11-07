package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPanelRejectsForeignHostHeader(t *testing.T) {
	s:=newView("q8");s.root=t.TempDir();h:=guiMux(&s)
	bad:=httptest.NewRequest(http.MethodGet,"http://attacker.example/api/state",nil);w:=httptest.NewRecorder();h.ServeHTTP(w,bad)
	if w.Code!=http.StatusForbidden{t.Fatalf("foreign host status %d",w.Code)}
	good:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:4477/",nil);w=httptest.NewRecorder();h.ServeHTTP(w,good)
	if w.Code!=http.StatusOK{t.Fatalf("loopback host status %d",w.Code)}
}

func TestWorkspacePaginationCoversMoreThanOneHundredFiles(t *testing.T) {
	root:=t.TempDir()
	for i:=0;i<155;i++ {if err:=os.WriteFile(filepath.Join(root,fmt.Sprintf("file%03d.txt",i)),[]byte("x"),0600);err!=nil{t.Fatal(err)}}
	if err:=os.Mkdir(filepath.Join(root,"node_modules"),0700);err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(filepath.Join(root,"node_modules","skip.txt"),[]byte("x"),0600);err!=nil{t.Fatal(err)}
	var names []string
	for offset:=0;;offset+=80 { page,more,err:=workspacePage(root,offset,80);if err!=nil{t.Fatal(err)};names=append(names,page...);if !more{break} }
	if len(names)!=155||names[0]!="file000.txt"||names[154]!="file154.txt"{t.Fatalf("count=%d names=%v",len(names),names)}
}

func TestModelFileReadRejectsBinaryAndInvalidUTF8(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string][]byte{"binary.dat": {1, 0, 2}, "bad.txt": {0xff, 0xfe}} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil { t.Fatal(err) }
		s := newView("q8"); s.root = root
		if text, ok, complete := s.localTextWithLimit(name, 1024); ok || complete || !strings.Contains(text, "ikili") { t.Fatalf("%s: %q %v %v", name, text, ok, complete) }
	}
}

func TestGenerationCannotOverwriteUnreadExistingFile(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	root:=t.TempDir(); file:=filepath.Join(root,"existing.txt")
	if err:=os.WriteFile(file,[]byte("original\n"),0600);err!=nil{t.Fatal(err)}
	s:=newView("q8");s.root=root;s.key="test-key";s.modelCh=make(chan modelResult,1)
	s.sendModel=func(workerID,string,string)modelResult{return modelResult{ok:true,body:"DOSYA existing.txt\nchanged\nBİTTİ\n"}}
	s.draft="bu dosyayı yeniden oluştur";s.submit()
	data,err:=os.ReadFile(file);if err!=nil||string(data)!="original\n"{t.Fatalf("file changed: %q %v",data,err)}
}

func TestDirectWriteStopsAtEndMarker(t *testing.T) {
	root:=t.TempDir(); note,blocked:=applyReplyFilesCheckedVersioned(root,"DOSYA out.txt\nhello\nBİTTİ\nModelin açıklaması\n",map[string]bool{},nil)
	if len(blocked)>0||note==""{t.Fatalf("note=%q blocked=%v",note,blocked)}
	data,err:=os.ReadFile(filepath.Join(root,"out.txt"));if err!=nil||string(data)!="hello\n"{t.Fatalf("%q %v",data,err)}
}

func TestProjectRunnerReportsRealFailure(t *testing.T) {
	root:=t.TempDir()
	if _,_,ok:=findPython(root,root);!ok{t.Skip("Python yok")}
	if err:=os.WriteFile(filepath.Join(root,"test_failure.py"),[]byte("import unittest\nclass Check(unittest.TestCase):\n def test_fail(self): self.fail('expected failure')\n"),0600);err!=nil{t.Fatal(err)}
	report:=runProjectTests(root)
	if !strings.Contains(report,"Test başarısız")||!strings.Contains(report,"expected failure")||!strings.Contains(report,"süre"){t.Fatal(report)}
}
