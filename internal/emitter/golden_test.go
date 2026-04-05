package emitter

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
	"github.com/yupyom/src2tex-go-v2.236/internal/scanner"
)

var updateGolden = flag.Bool("update", false, "overwrite golden files with current output")

// runGolden converts inputFile using the given lang flag and compares the
// output to goldenFile.  inputSrcFile is passed as Options.SourceFile
// (the footer label — same as what shortPath() returns in main.go).
func runGolden(t *testing.T, inputFile, goldenFile, langFlag, inputSrcFile string) {
	t.Helper()

	src, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("read input %s: %v", inputFile, err)
	}

	ld := lang.FindByFlag(langFlag)
	if ld == nil {
		t.Fatalf("unknown lang flag %q", langFlag)
	}

	tokens := scanner.Scan(src, ld)

	var b strings.Builder
	WritePreamble(&b, Options{SourceFile: inputSrcFile})
	WriteBody(&b, tokens, ld, Options{})
	WritePostamble(&b)
	got := b.String()

	if *updateGolden {
		if err := os.WriteFile(goldenFile, []byte(got), 0644); err != nil {
			t.Fatalf("write golden %s: %v", goldenFile, err)
		}
		return
	}

	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenFile, err)
	}

	if got != string(want) {
		// Show first differing line for diagnosis.
		gotLines := strings.Split(got, "\n")
		wantLines := strings.Split(string(want), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Errorf("first diff at line %d:\n  got:  %q\n  want: %q", i+1, gotLines[i], wantLines[i])
				return
			}
		}
		t.Errorf("output length differs: got %d lines, want %d lines", len(gotLines), len(wantLines))
	}
}

const (
	inputDir  = "../../testdata/input/"
	goldenDir = "../../testdata/golden/"
)

func TestGolden_HanoiGo(t *testing.T) {
	runGolden(t, inputDir+"hanoi.go", goldenDir+"hanoi.go.tex", "go", "input/hanoi.go")
}

func TestGolden_HanoiSh(t *testing.T) {
	runGolden(t, inputDir+"hanoi.sh", goldenDir+"hanoi.sh.tex", "sh", "input/hanoi.sh")
}

func TestGolden_HanoiPas(t *testing.T) {
	runGolden(t, inputDir+"hanoi.pas", goldenDir+"hanoi.pas.tex", "pascal", "input/hanoi.pas")
}

func TestGolden_PopgenRed(t *testing.T) {
	runGolden(t, inputDir+"popgen.red", goldenDir+"popgen.red.tex", "reduce", "input/popgen.red")
}
