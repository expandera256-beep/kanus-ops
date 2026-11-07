package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Bound the scan and omit ignored, hidden and symlinked files. Ranking uses
// path names only; contents are read separately through the normal safeguards.
func relevantProjectPaths(root, task string) []string {
	stop:=map[string]bool{"projeyi":true,"proje":true,"tamamla":true,"onar":true,"düzelt":true,"degistir":true,"dosya":true,"için":true,"with":true,"the":true,"and":true}
	var words []string
	for _,w:=range strings.FieldsFunc(strings.ToLower(task),func(r rune)bool{return !unicode.IsLetter(r)&&!unicode.IsDigit(r)}) { if len([]rune(w))>=3&&!stop[w] { words=append(words,w) } }
	type candidate struct { path string; score int }; var all []candidate; scanned:=0
	_ = filepath.WalkDir(root,func(path string,e os.DirEntry,err error)error { if err!=nil{return nil};if path==root{return nil};if e.Type()&os.ModeSymlink!=0{return nil};if e.IsDir(){if skipTreeDir(e.Name())||depth(root,path)>workspaceDepth{return filepath.SkipDir};return nil};if skipTreeDir(e.Name()){return nil};scanned++;if scanned>5000{return filepath.SkipAll}; rel:=relPath(root,path); low:=strings.ToLower(rel);for _,secret:=range []string{"gizli","secret","credential",".env",".pem",".key"}{if strings.Contains(low,secret){return nil}};score:=0;for _,w:=range words{if strings.Contains(low,w){score+=10}};switch strings.ToLower(filepath.Base(path)){case "readme.md","go.mod","package.json","pyproject.toml","main.py","main.go":score+=3};all=append(all,candidate{rel,score});return nil })
	sort.SliceStable(all,func(i,j int)bool{if all[i].score!=all[j].score{return all[i].score>all[j].score};return all[i].path<all[j].path}); if len(all)>100 {all=all[:100]};out:=make([]string,0,len(all));for _,c:=range all{out=append(out,c.path)};return out
}
