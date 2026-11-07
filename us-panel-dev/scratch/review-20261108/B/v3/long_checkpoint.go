package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const taskDir = ".us-tasks"
type longCheckpoint struct { Root string; Text string; TaskHash string; ChunkBytes int; PromptLimit int; Worker workerID; Model string; Gateway string; Summaries []string }
func taskPath(root string) string { return filepath.Join(root,taskDir,"long.json") }
func hasLongCheckpoint(root string) bool { dir:=filepath.Join(root,taskDir);d,e:=os.Lstat(dir);if e!=nil||!d.IsDir()||d.Mode()&os.ModeSymlink!=0{return false};st,e:=os.Lstat(taskPath(root));return e==nil&&st.Mode().IsRegular()&&st.Size()<=16<<20 }
func loadLongCheckpoint(root string) *longCheckpoint {
	dir:=filepath.Join(root,taskDir);d,e:=os.Lstat(dir);if e!=nil||!d.IsDir()||d.Mode()&os.ModeSymlink!=0{return nil};path:=taskPath(root);st,e:=os.Lstat(path);if e!=nil||!st.Mode().IsRegular()||st.Size()>16<<20{return nil};b,e:=os.ReadFile(path);if e!=nil{return nil};var cp longCheckpoint;if json.Unmarshal(b,&cp)!=nil||cp.Root!=root||cp.TaskHash!=hashContent([]byte(cp.Text))||len(cp.Text)>longInputLimit{return nil};return &cp
}
func saveLongCheckpoint(cp *longCheckpoint) error {
	dir:=filepath.Join(cp.Root,taskDir);d,e:=os.Lstat(dir);if e==nil&&(!d.IsDir()||d.Mode()&os.ModeSymlink!=0){return errors.New("görev klasörü geçersiz")};if e!=nil&&!os.IsNotExist(e){return e};if e=os.Mkdir(dir,0700);e!=nil&&!os.IsExist(e){return e};b,e:=json.Marshal(cp);if e!=nil{return e};f,e:=os.CreateTemp(dir,"checkpoint-*");if e!=nil{return e};name:=f.Name();defer os.Remove(name);if e=f.Chmod(0600);e==nil{_,e=f.Write(b)};if c:=f.Close();e==nil{e=c};if e!=nil{return e};return os.Rename(name,taskPath(cp.Root))
}
func removeLongCheckpoint(root string) error { e:=os.Remove(taskPath(root));if os.IsNotExist(e){return nil};return e }
