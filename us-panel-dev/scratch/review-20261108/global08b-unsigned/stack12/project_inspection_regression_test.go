package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectInspectionSuppliesRealPaths(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 90; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("file%02d.py", i)), []byte("print(1)"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, task := range []string{"Projeyi incele", "Klasördeki dosyaları listele", "Review the project"} {
		if !taskNeedsPaths(task) {
			t.Fatalf("inspection not detected: %s", task)
		}
		wire := boundedChatWire(task, root, taskNeedsPaths(task), initialWireBudget)
		if !strings.Contains(wire, "file00.py") || !strings.Contains(wire, "file79.py") || !strings.Contains(wire, "Liste sınırlı olabilir") {
			t.Fatalf("missing paths for %q: %s", task, wire)
		}
		if strings.Contains(wire, "print(1)") {
			t.Fatal("file contents leaked into index")
		}
	}
}
