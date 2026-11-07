package main

import (
	"os"
	"strings"
	"testing"
)

// legacyBoundedChatWire12 is a verbatim copy of boundedChatWire from .12
// (model.go). With US_AGENT_SYMBOL_CONTEXT off the wire must equal it.
func legacyBoundedChatWire12(task, root string, withPaths bool, limit int) string {
	base := chatWireMode(task, root, false)
	if !withPaths {
		return base
	}
	ending := "\nListe sınırlı olabilir; diğer sayfaları JSON list_files ile iste. Mevcut dosyaları değiştirmeden önce içeriklerini read_file ile oku. Örnek veya yer tutucu yollar üretme."
	var out strings.Builder
	out.WriteString(base)
	out.WriteString("\nÇalışma alanındaki yollar:\n")
	for _, path := range relevantProjectPaths(root, task) {
		if out.Len()+len(path)+1+len(ending) > limit {
			break
		}
		out.WriteString(path)
		out.WriteByte('\n')
	}
	out.WriteString(ending)
	return out.String()
}

func p6SymbolFixture(t *testing.T) string {
	root := t.TempDir()
	p6write(t, root, "go.mod", "module x\n")
	p6write(t, root, "calc/total.go", "package calc\nfunc CalcTotal(x int) int { key := \"AKIAIOSFODNN7EXAMPLE\"; _ = key; return x }\n")
	p6write(t, root, "web/cart.js", "export function apply_discount(a) { return a }\nclass CartView {}\n")
	for i := 0; i < 40; i++ {
		p6write(t, root, "pkg/f"+itoa(i)+"/CalcTotal_"+itoa(i)+".go", "package f\nfunc CalcTotal() {}\n")
	}
	return root
}

var p6Tasks = []string{
	"mevcut projeyi onar: `CalcTotal` yanlış toplam veriyor, apply_discount ve CartView de bozuk",
	"proje dosyalarını incele ve listele",
	"merhaba",
	"mevcut CalcTotal OtherThing ThirdOne FourthOne fonksiyonlarını düzelt",
}

func TestP6SymbolContextOffIsByteIdenticalTo12(t *testing.T) {
	root := p6SymbolFixture(t)
	for _, v := range []string{"", "0", "false", "off"} {
		if v == "" {
			os.Unsetenv("US_AGENT_SYMBOL_CONTEXT")
		} else {
			t.Setenv("US_AGENT_SYMBOL_CONTEXT", v)
		}
		for _, task := range p6Tasks {
			for _, wp := range []bool{true, false} {
				for _, lim := range []int{initialWireBudget, promptWireBudget, 300} {
					if got, want := boundedChatWire(task, root, wp, lim), legacyBoundedChatWire12(task, root, wp, lim); got != want {
						t.Fatalf("flag=%q task=%q differs from .12", v, task)
					}
				}
			}
		}
	}
}

func TestP6SymbolContextOnBounded(t *testing.T) {
	root := p6SymbolFixture(t)
	t.Setenv("US_AGENT_SYMBOL_CONTEXT", "1")
	task := p6Tasks[0]
	wire := boundedChatWire(task, root, true, initialWireBudget)
	if len(wire) > initialWireBudget {
		t.Fatalf("wire %d > limit", len(wire))
	}
	i := strings.Index(wire, symbolContextHeader)
	if i < 0 {
		t.Fatalf("no symbol block: %s", wire)
	}
	block := symbolContextBlock(root, task, initialWireBudget)
	if len(block) == 0 || len(block) > 600 || len(block) > initialWireBudget/4 || !strings.Contains(wire, block) {
		t.Fatalf("block %d bytes", len(block))
	}
	for _, bad := range []string{"AKIA", "return", "{", "key"} {
		if strings.Contains(block, bad) {
			t.Fatalf("line text in block: %q\n%s", bad, block)
		}
	}
	if !strings.Contains(block, "calc/total.go:2 func CalcTotal") {
		t.Fatalf("missing hit:\n%s", block)
	}
	counts := map[string]int{}
	for _, ln := range strings.Split(strings.TrimSpace(strings.TrimPrefix(block, symbolContextHeader)), "\n") {
		f := strings.Fields(ln)
		if len(f) != 3 {
			t.Fatalf("bad line %q", ln)
		}
		counts[f[2]]++
	}
	if len(counts) > 3 {
		t.Fatalf("more than 3 names: %v", counts)
	}
	for n, c := range counts {
		if c > 5 {
			t.Fatalf("%s has %d hits", n, c)
		}
	}
	if got := symbolTaskNames(p6Tasks[3]); len(got) > 3 {
		t.Fatalf("names %v", got)
	}
	if small := symbolContextBlock(root, task, 200); len(small) > 50 {
		t.Fatalf("block exceeds 25%% of small budget: %d", len(small))
	}
	if b := symbolContextBlock(root, "merhaba proje", initialWireBudget); b != "" {
		t.Fatalf("no identifiers should give empty block: %q", b)
	}
}

func TestP6SymbolTaskNames(t *testing.T) {
	got := symbolTaskNames("`foo` ve CalcTotal, apply_discount, Proje, the, ab, `../x` CartView fazla")
	want := []string{"foo", "CalcTotal", "apply_discount"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v", got)
	}
}
