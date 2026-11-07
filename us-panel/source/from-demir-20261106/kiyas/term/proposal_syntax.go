package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Parse proposed Python text before committing any file. Isolated mode and a
// clean working directory prevent project sitecustomize/import side effects.
// No project code is executed. A missing interpreter means unverified syntax;
// project tests remain a separate, explicit action.
func checkProposedPython(files []proposedFile) error {
	var docs []map[string]string
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Path), ".py") {
			docs = append(docs, map[string]string{"path": f.Path, "content": f.New})
		}
	}
	if len(docs) == 0 {
		return nil
	}
	var bin string
	for _, name := range []string{"python3", "python", "py"} {
		if path, err := exec.LookPath(name); err == nil {
			bin = path
			break
		}
	}
	if bin == "" {
		return nil
	}
	script := `import ast,json,sys
for doc in json.load(sys.stdin):
 try: ast.parse(doc["content"],filename=doc["path"])
 except SyntaxError as error:
  print("%s:%s:%s: %s: %s"%(doc["path"],error.lineno,error.offset,type(error).__name__,error.msg))
  sys.exit(2)
`
	raw, _ := json.Marshal(docs)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := append(pythonPrefix(bin), "-I", "-S", "-c", script)
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = bytes.NewReader(raw)
	cmd.Env = projectTestEnv()
	cmd.Dir = os.TempDir()
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("Python sözdizimi denetimi zaman aşımı; hiçbir dosya yazılmadı")
		}
		detail := trimLog(redactLine(strings.TrimSpace(string(output))), 500)
		if detail == "" {
			detail = "yorumlayıcı denetimi tamamlayamadı"
		}
		return fmt.Errorf("Python sözdizimi geçmedi; hiçbir dosya yazılmadı: %s", detail)
	}
	return nil
}
