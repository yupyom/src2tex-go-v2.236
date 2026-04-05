package font

import (
"testing"
"strings"
)

func TestResolveCodeFontLine_Hackgen(t *testing.T) {
	line := ResolveCodeFontLine("hackgen")
	t.Logf("MonoFontLine: %s", line)
	if !strings.Contains(line, "Path=") {
		t.Errorf("expected Path= in MonoFontLine, got: %s", line)
	}
}

func TestResolveCodeFontCJKMonoLine_Hackgen(t *testing.T) {
	line := ResolveCodeFontCJKMonoLine("hackgen")
	t.Logf("CJKMonoFontLine: %q", line)
	if line == "" {
		t.Error("expected non-empty CJKMonoFontLine for unified font hackgen")
	}
	if !strings.Contains(line, "Path=") {
		t.Errorf("expected Path= in CJKMonoFontLine, got: %s", line)
	}
}

func TestResolveCodeFontCJKMainLine_Hackgen(t *testing.T) {
	line := ResolveCodeFontCJKMainLine("hackgen")
	t.Logf("CJKMainLine: %q", line)
	if line == "" {
		t.Error("expected non-empty CJKMainLine for unified font hackgen")
	}
}
