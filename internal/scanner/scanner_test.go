package scanner

import (
	"testing"

	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
)

func TestScan_GoLineComment(t *testing.T) {
	ld := lang.FindByFlag("go")
	src := []byte("x := 1 // set x\ny := 2\n")

	tokens := Scan(src, ld)

	want := []struct {
		kind TokenKind
		text string
	}{
		{TokenCode, "x := 1 "},
		{TokenLineComment, "// set x\n"},
		{TokenCode, "y := 2\n"},
	}

	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d\n%v", len(tokens), len(want), tokens)
	}
	for i, w := range want {
		if tokens[i].Kind != w.kind {
			t.Errorf("token[%d].Kind = %v, want %v", i, tokens[i].Kind, w.kind)
		}
		if string(tokens[i].Text) != w.text {
			t.Errorf("token[%d].Text = %q, want %q", i, tokens[i].Text, w.text)
		}
	}
}

func TestScan_ShellLineComment(t *testing.T) {
	ld := lang.FindByFlag("sh")
	src := []byte("echo hello # greet\necho bye\n")

	tokens := Scan(src, ld)

	want := []struct {
		kind TokenKind
		text string
	}{
		{TokenKeyword, "echo"},
		{TokenCode, " "},
		{TokenCode, "hello "},
		{TokenLineComment, "# greet\n"},
		{TokenKeyword, "echo"},
		{TokenCode, " "},
		{TokenCode, "bye\n"},
	}

	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(tokens), len(want), tokens)
	}
	for i, w := range want {
		if tokens[i].Kind != w.kind {
			t.Errorf("token[%d].Kind = %v, want %v", i, tokens[i].Kind, w.kind)
		}
		if string(tokens[i].Text) != w.text {
			t.Errorf("token[%d].Text = %q, want %q", i, tokens[i].Text, w.text)
		}
	}
}

func TestScan_CommentAtEOF(t *testing.T) {
	// Line comment with no trailing newline — common at end of file.
	ld := lang.FindByFlag("go")
	src := []byte("x := 1 // no newline")

	tokens := Scan(src, ld)

	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2", len(tokens))
	}
	if tokens[0].Kind != TokenCode || string(tokens[0].Text) != "x := 1 " {
		t.Errorf("token[0] = %+v", tokens[0])
	}
	if tokens[1].Kind != TokenLineComment || string(tokens[1].Text) != "// no newline" {
		t.Errorf("token[1] = %+v", tokens[1])
	}
}

func TestScan_NoComment(t *testing.T) {
	// Without any comment markers, no comment tokens must appear.
	// ("package" is a Go keyword, so tokens may include TokenKeyword,
	// but never TokenLineComment or TokenBlockComment.)
	ld := lang.FindByFlag("go")
	src := []byte("package main\n")

	tokens := Scan(src, ld)

	for _, tok := range tokens {
		if tok.Kind == TokenLineComment || tok.Kind == TokenBlockComment {
			t.Errorf("unexpected comment token %v %q", tok.Kind, tok.Text)
		}
	}
	// The full source text must be reconstructed from all tokens.
	var got []byte
	for _, tok := range tokens {
		got = append(got, tok.Text...)
	}
	if string(got) != string(src) {
		t.Errorf("reconstructed text = %q, want %q", got, src)
	}
}

func TestScan_CommentOnlyLine(t *testing.T) {
	ld := lang.FindByFlag("c")
	src := []byte("// entire line\nnext\n")

	tokens := Scan(src, ld)

	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2", len(tokens))
	}
	if tokens[0].Kind != TokenLineComment || string(tokens[0].Text) != "// entire line\n" {
		t.Errorf("token[0] = %+v", tokens[0])
	}
	if tokens[1].Kind != TokenCode || string(tokens[1].Text) != "next\n" {
		t.Errorf("token[1] = %+v", tokens[1])
	}
}

func TestScan_CommentInsideDoubleQuoteString(t *testing.T) {
	// "//" inside a string literal must NOT be treated as a line comment.
	// (Identifiers like "fmt" and "Printf" may be split into separate TokenCode
	// tokens, but no TokenLineComment must appear.)
	ld := lang.FindByFlag("go")
	src := []byte(`fmt.Printf("// not a comment")`)

	tokens := Scan(src, ld)

	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			t.Errorf("unexpected TokenLineComment %q inside string literal", tok.Text)
		}
	}
}

func TestScan_CommentInsideSingleQuoteString(t *testing.T) {
	// "#" inside a single-quoted string must NOT start a comment.
	// "echo" is a shell keyword so it gets its own TokenKeyword token.
	ld := lang.FindByFlag("sh")
	src := []byte("echo '# not a comment'\n")

	tokens := Scan(src, ld)

	// Ensure no TokenLineComment appears — that's the key invariant.
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			t.Errorf("unexpected TokenLineComment %q inside single-quoted string", tok.Text)
		}
	}
	// Reconstruct source to verify no bytes were dropped.
	var got []byte
	for _, tok := range tokens {
		got = append(got, tok.Text...)
	}
	if string(got) != string(src) {
		t.Errorf("reconstructed text = %q, want %q", got, src)
	}
}

func TestScan_CommentInsideRawString(t *testing.T) {
	// "//" inside a Go raw string literal must NOT start a comment.
	ld := lang.FindByFlag("go")
	src := []byte("x := `// not a comment`\n")

	tokens := Scan(src, ld)

	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1: %v", len(tokens), tokens)
	}
	if tokens[0].Kind != TokenCode {
		t.Errorf("token[0].Kind = %v, want TokenCode", tokens[0].Kind)
	}
}

func TestScan_EscapedQuoteInsideString(t *testing.T) {
	// Escaped \" inside a string must not close the string early.
	ld := lang.FindByFlag("go")
	src := []byte(`x := "say \"// hi\""` + "\n")

	tokens := Scan(src, ld)

	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1: %v", len(tokens), tokens)
	}
	if tokens[0].Kind != TokenCode {
		t.Errorf("token[0].Kind = %v, want TokenCode", tokens[0].Kind)
	}
}

func TestScan_StringThenComment(t *testing.T) {
	// String ends, then a real comment follows on the same line.
	ld := lang.FindByFlag("go")
	src := []byte(`x := "val" // real comment` + "\n")

	tokens := Scan(src, ld)

	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2: %v", len(tokens), tokens)
	}
	if tokens[0].Kind != TokenCode {
		t.Errorf("token[0].Kind = %v, want TokenCode", tokens[0].Kind)
	}
	if tokens[1].Kind != TokenLineComment {
		t.Errorf("token[1].Kind = %v, want TokenLineComment", tokens[1].Kind)
	}
}

func TestScan_KeywordEmitted(t *testing.T) {
	// "func" must be emitted as TokenKeyword in Go.
	ld := lang.FindByFlag("go")
	src := []byte("func main() {}\n")

	tokens := Scan(src, ld)

	// Find a TokenKeyword token with text "func".
	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenKeyword && string(tok.Text) == "func" {
			found = true
		}
	}
	if !found {
		t.Errorf("TokenKeyword 'func' not found in %v", tokens)
	}
}

func TestScan_NonKeywordIsCode(t *testing.T) {
	// "funcName" is not a keyword — must be TokenCode, not TokenKeyword.
	ld := lang.FindByFlag("go")
	src := []byte("funcName()\n")

	tokens := Scan(src, ld)

	for _, tok := range tokens {
		if tok.Kind == TokenKeyword {
			t.Errorf("unexpected TokenKeyword %q", tok.Text)
		}
	}
}

func TestScan_KeywordAdjacentToCode(t *testing.T) {
	// "if x > 0" — "if" is keyword, "x" is code, no merging of adjacent tokens.
	ld := lang.FindByFlag("go")
	src := []byte("if x > 0 {\n")

	tokens := Scan(src, ld)

	kinds := make(map[TokenKind][]string)
	for _, tok := range tokens {
		kinds[tok.Kind] = append(kinds[tok.Kind], string(tok.Text))
	}
	if len(kinds[TokenKeyword]) == 0 {
		t.Errorf("no TokenKeyword found; tokens=%v", tokens)
	}
	found := false
	for _, kw := range kinds[TokenKeyword] {
		if kw == "if" {
			found = true
		}
	}
	if !found {
		t.Errorf("'if' not among keywords %v", kinds[TokenKeyword])
	}
}

func TestScan_PythonDocstringMultiline(t *testing.T) {
	// Triple-quoted string at line start must be treated as block comment.
	ld := lang.FindByFlag("python")
	src := []byte(`"""
hello
"""
x = 1
`)

	tokens := Scan(src, ld)

	// First token must be TokenBlockComment covering the docstring.
	if len(tokens) == 0 {
		t.Fatal("got no tokens")
	}
	if tokens[0].Kind != TokenBlockComment {
		t.Errorf("token[0].Kind = %v, want TokenBlockComment", tokens[0].Kind)
	}
	want0 := "\"\"\"\nhello\n\"\"\""
	if string(tokens[0].Text) != want0 {
		t.Errorf("token[0].Text = %q, want %q", tokens[0].Text, want0)
	}
}

func TestScan_PythonDocstringSingleLine(t *testing.T) {
	// Single-line triple-quoted docstring at line start.
	ld := lang.FindByFlag("python")
	src := []byte("    '''Move n discs.'''\n")

	tokens := Scan(src, ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment {
			found = true
			want := "'''Move n discs.'''"
			if string(tok.Text) != want {
				t.Errorf("docstring token.Text = %q, want %q", tok.Text, want)
			}
		}
	}
	if !found {
		t.Errorf("no TokenBlockComment found; tokens=%v", tokens)
	}
}

func TestScan_PythonDocstringNotAtLineStart(t *testing.T) {
	// Triple-quote after non-whitespace must NOT start a docstring block comment.
	ld := lang.FindByFlag("python")
	src := []byte(`x = """text"""` + "\n")

	tokens := Scan(src, ld)

	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment {
			t.Errorf("unexpected TokenBlockComment %q in assignment context", tok.Text)
		}
	}
	// Verify source reconstruction.
	var got []byte
	for _, tok := range tokens {
		got = append(got, tok.Text...)
	}
	if string(got) != string(src) {
		t.Errorf("reconstructed text = %q, want %q", got, src)
	}
}

func TestScan_PythonHashCommentStillWorks(t *testing.T) {
	// Regular # comment must still be recognized in Python.
	ld := lang.FindByFlag("python")
	src := []byte("x = 1  # a comment\n")

	tokens := Scan(src, ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			found = true
		}
	}
	if !found {
		t.Errorf("no TokenLineComment found; tokens=%v", tokens)
	}
}

func TestScan_NoKeywordsWhenLangHasNone(t *testing.T) {
	// Makefile has no keywords defined — all identifiers must be TokenCode.
	ld := lang.FindByFlag("make")
	src := []byte("for i in 1 2 3; do echo $i; done\n")

	tokens := Scan(src, ld)

	for _, tok := range tokens {
		if tok.Kind == TokenKeyword {
			t.Errorf("unexpected TokenKeyword %q in shell (no keywords defined)", tok.Text)
		}
	}
}

func TestScan_XMLBlockComment(t *testing.T) {
	ld := lang.FindByFlag("xml")
	if ld == nil {
		t.Fatal("xml lang not found")
	}
	tokens := Scan([]byte("<?xml version=\"1.0\"?>\n<!-- XML comment -->\n<root/>"), ld)
	// <!-- XML comment --> が TokenBlockComment であること
	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "<!-- XML comment -->" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TokenBlockComment for <!-- XML comment -->, got tokens: %v", tokens)
	}
}

func TestScan_XMLMultilineComment(t *testing.T) {
	ld := lang.FindByFlag("xml")
	if ld == nil {
		t.Fatal("xml lang not found")
	}
	src := "<!--\n  multi\n  line\n-->\n<root/>"
	tokens := Scan([]byte(src), ld)
	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment {
			found = true
			expected := "<!--\n  multi\n  line\n-->"
			if string(tok.Text) != expected {
				t.Errorf("expected %q, got %q", expected, string(tok.Text))
			}
		}
	}
	if !found {
		t.Error("expected TokenBlockComment for multiline XML comment")
	}
}

func TestScan_CSSBlockComment(t *testing.T) {
	ld := lang.FindByFlag("css")
	if ld == nil {
		t.Fatal("css lang not found")
	}
	tokens := Scan([]byte("body {\n  /* CSS comment */\n  color: red;\n}"), ld)
	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* CSS comment */" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TokenBlockComment for /* CSS comment */")
	}
}

func TestScan_CSSNoLineComment(t *testing.T) {
	ld := lang.FindByFlag("css")
	if ld == nil {
		t.Fatal("css lang not found")
	}
	// CSS has no line comment; // should be treated as code
	tokens := Scan([]byte("a { content: \"//not-comment\"; }"), ld)
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			t.Error("CSS should not have line comments")
		}
	}
}

func TestHasPrefixCI(t *testing.T) {
	tests := []struct {
		src    string
		i      int
		prefix string
		want   bool
	}{
		{"<style>", 0, "<style", true},
		{"<STYLE>", 0, "<style", true},
		{"<Style type>", 0, "<style", true},
		{"<stylesheet>", 0, "<style", true}, // hasPrefixCI only checks the prefix itself
		{"<script>", 0, "<style", false},
		{"</style>", 0, "</style>", true},
		{"</STYLE>", 0, "</style>", true},
	}
	for _, tt := range tests {
		got := hasPrefixCI([]byte(tt.src), tt.i, tt.prefix)
		if got != tt.want {
			t.Errorf("hasPrefixCI(%q, %d, %q) = %v, want %v", tt.src, tt.i, tt.prefix, got, tt.want)
		}
	}
}

func TestScan_HTMLComment(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	tokens := Scan([]byte("<html>\n<!-- HTML comment -->\n</html>"), ld)
	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "<!-- HTML comment -->" {
			found = true
		}
	}
	if !found {
		t.Error("expected TokenBlockComment for <!-- HTML comment -->")
	}
}

func TestScan_HTMLTagBold(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	tokens := Scan([]byte("<div class=\"main\">text</div>"), ld)
	kwCount := 0
	for _, tok := range tokens {
		if tok.Kind == TokenKeyword && string(tok.Text) == "div" {
			kwCount++
		}
	}
	if kwCount != 2 {
		t.Errorf("expected 2 TokenKeyword for 'div', got %d", kwCount)
	}
}

func TestScan_HTMLStyleCSS(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<style>\n/* css comment */\nbody { color: red; }\n</style>"
	tokens := Scan([]byte(src), ld)

	foundStyleKW := false
	foundCSSComment := false
	for _, tok := range tokens {
		if tok.Kind == TokenKeyword && string(tok.Text) == "style" {
			foundStyleKW = true
		}
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* css comment */" {
			foundCSSComment = true
		}
	}
	if !foundStyleKW {
		t.Error("expected TokenKeyword for 'style'")
	}
	if !foundCSSComment {
		t.Error("expected TokenBlockComment for CSS /* */ inside <style>")
	}
}

func TestScan_HTMLScriptJS(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<script>\n// js line comment\nvar x = 1;\n/* js block comment */\n</script>"
	tokens := Scan([]byte(src), ld)

	foundLineComment := false
	foundBlockComment := false
	foundVar := false
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			foundLineComment = true
		}
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* js block comment */" {
			foundBlockComment = true
		}
		if tok.Kind == TokenKeyword && string(tok.Text) == "var" {
			foundVar = true
		}
	}
	if !foundLineComment {
		t.Error("expected TokenLineComment for JS // comment")
	}
	if !foundBlockComment {
		t.Error("expected TokenBlockComment for JS /* */ comment")
	}
	if !foundVar {
		t.Error("expected TokenKeyword for JS 'var'")
	}
}

func TestScan_HTMLCloseTagInJSString(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	// </script> inside a JS string should NOT pop the context
	src := "<script>\nvar s = \"</script>\";\n// still js\n</script>"
	tokens := Scan([]byte(src), ld)

	foundLineComment := false
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			foundLineComment = true
		}
	}
	if !foundLineComment {
		t.Error("</script> in JS string should not pop context; expected line comment for '// still js'")
	}
}

func TestScan_HTMLCaseInsensitive(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<STYLE>\n/* comment */\n</STYLE>"
	tokens := Scan([]byte(src), ld)

	foundComment := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* comment */" {
			foundComment = true
		}
	}
	if !foundComment {
		t.Error("expected CSS /* */ comment inside case-insensitive <STYLE>")
	}
}

func TestScan_HTMLNoFalseSubLang(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<strong>bold text</strong>"
	tokens := Scan([]byte(src), ld)
	foundStrong := false
	for _, tok := range tokens {
		if tok.Kind == TokenKeyword && string(tok.Text) == "strong" {
			foundStrong = true
		}
	}
	if !foundStrong {
		t.Error("expected TokenKeyword for 'strong' tag name (BoldTags)")
	}
}

func TestScan_PHPLineComment(t *testing.T) {
	// Standalone PHP: // comment must be recognized.
	ld := lang.FindByFlag("php")
	if ld == nil {
		t.Fatal("php lang not found")
	}
	src := []byte("<?php\n$x = 1; // set x\necho $x;\n")
	tokens := Scan(src, ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TokenLineComment for // in PHP, got tokens: %v", tokens)
	}
	// Source must reconstruct.
	var got []byte
	for _, tok := range tokens {
		got = append(got, tok.Text...)
	}
	if string(got) != string(src) {
		t.Errorf("reconstructed text = %q, want %q", got, src)
	}
}

func TestScan_PHPBlockComment(t *testing.T) {
	ld := lang.FindByFlag("php")
	if ld == nil {
		t.Fatal("php lang not found")
	}
	src := []byte("<?php\n/* block comment */\necho 'hi';\n")
	tokens := Scan(src, ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* block comment */" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TokenBlockComment for /* block comment */, got: %v", tokens)
	}
}

func TestScan_PHPKeyword(t *testing.T) {
	ld := lang.FindByFlag("php")
	if ld == nil {
		t.Fatal("php lang not found")
	}
	src := []byte("<?php\nfunction hello() { echo 'hi'; }\n")
	tokens := Scan(src, ld)

	foundFunction := false
	foundEcho := false
	for _, tok := range tokens {
		if tok.Kind == TokenKeyword {
			switch string(tok.Text) {
			case "function":
				foundFunction = true
			case "echo":
				foundEcho = true
			}
		}
	}
	if !foundFunction {
		t.Errorf("expected TokenKeyword 'function' in PHP")
	}
	if !foundEcho {
		t.Errorf("expected TokenKeyword 'echo' in PHP")
	}
}

func TestScan_HTMLEmbeddedPHP(t *testing.T) {
	// <?php ... ?> inside HTML must be scanned as PHP context.
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<html>\n<?php\n// php comment\necho 'hi';\n?>\n</html>"
	tokens := Scan([]byte(src), ld)

	foundLineComment := false
	foundEcho := false
	for _, tok := range tokens {
		if tok.Kind == TokenLineComment {
			foundLineComment = true
		}
		if tok.Kind == TokenKeyword && string(tok.Text) == "echo" {
			foundEcho = true
		}
	}
	if !foundLineComment {
		t.Errorf("expected TokenLineComment for PHP // inside HTML <?php ?>")
	}
	if !foundEcho {
		t.Errorf("expected TokenKeyword 'echo' inside HTML <?php ?>")
	}
}

func TestScan_HTMLEmbeddedPHPBlockComment(t *testing.T) {
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<div>\n<?php\n/* php block */\n?>\n</div>"
	tokens := Scan([]byte(src), ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "/* php block */" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TokenBlockComment '/* php block */' inside HTML <?php ?>, got: %v", tokens)
	}
}

func TestScan_HTMLPHPCloseRestoresHTML(t *testing.T) {
	// After ?>, HTML comment must be recognized again.
	ld := lang.FindByFlag("html")
	if ld == nil {
		t.Fatal("html lang not found")
	}
	src := "<?php echo 'hi'; ?>\n<!-- html comment -->"
	tokens := Scan([]byte(src), ld)

	found := false
	for _, tok := range tokens {
		if tok.Kind == TokenBlockComment && string(tok.Text) == "<!-- html comment -->" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HTML <!-- --> comment after ?>, got: %v", tokens)
	}
}

func TestMatchSubLangOpen(t *testing.T) {
	rules := []lang.SubLanguageRule{
		{OpenTag: "<style", CloseTag: "</style>", LangFlag: "css"},
		{OpenTag: "<script", CloseTag: "</script>", LangFlag: "js"},
	}
	tests := []struct {
		src  string
		want string // expected LangFlag or "" for nil
	}{
		{"<style>", "css"},
		{"<STYLE>", "css"},
		{"<style type=\"text/css\">", "css"},
		{"<stylesheet>", ""},         // 's' after "<style" — no match
		{"<script>", "js"},
		{"<Script src=\"app.js\">", "js"},
		{"<div>", ""},
		{"<strong>", ""},             // "<s" matches start of "<style" but full "<style" doesn't match
	}
	for _, tt := range tests {
		got := matchSubLangOpen([]byte(tt.src), 0, rules)
		gotFlag := ""
		if got != nil {
			gotFlag = got.LangFlag
		}
		if gotFlag != tt.want {
			t.Errorf("matchSubLangOpen(%q) = %q, want %q", tt.src, gotFlag, tt.want)
		}
	}
}
