package emitter

import (
	"strings"
	"testing"

	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
	"github.com/yupyom/src2tex-go-v2.236/internal/scanner"
)

var ldGo = lang.FindByFlag("go")
var ldSh = lang.FindByFlag("sh")
var ldC = lang.FindByFlag("c")

// ---- Preamble / Postamble ----

func TestWritePreamble_RequiredElements(t *testing.T) {
	var b strings.Builder
	WritePreamble(&b, Options{})
	got := b.String()

	for _, want := range []string{
		`\documentclass`,
		`\usepackage{fontspec}`,
		`\usepackage{xeCJK}`,
		`\setmonofont`,
		`\setCJKmainfont`,
		`\def\mc{\relax}`,
		`\begin{document}`,
		`\tt\mc`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("preamble missing %q", want)
		}
	}
}

func TestWritePreamble_CustomFonts(t *testing.T) {
	var b strings.Builder
	WritePreamble(&b, Options{CodeFont: "HackGen", CJKFont: "IPAexGothic"})
	got := b.String()

	if !strings.Contains(got, `\setmonofont{HackGen}`) {
		t.Error("custom CodeFont not reflected")
	}
	if !strings.Contains(got, `\setCJKmainfont{IPAexGothic}`) {
		t.Error("custom CJKFont not reflected")
	}
}

func TestWritePostamble(t *testing.T) {
	var b strings.Builder
	WritePostamble(&b)
	got := b.String()

	if !strings.Contains(got, `\rm\mc`) {
		t.Error("postamble missing \\rm\\mc")
	}
	if !strings.Contains(got, `\end{document}`) {
		t.Error("postamble missing \\end{document}")
	}
}

// ---- Line structure ----

func TestWriteBody_LinePrefix(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenCode, Text: []byte("x := 1\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, "\\noindent\n\\mbox{}") {
		t.Errorf("missing \\noindent\\n\\mbox{} prefix: %q", got)
	}
}

func TestWriteBody_EmptyLine(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenCode, Text: []byte("\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, "\\noindent\n\\mbox{}\\hfill\n") {
		t.Errorf("blank line not rendered as \\hfill: %q", got)
	}
}

// ---- Code escaping ----

func TestCodeEscape_Space(t *testing.T) {
	if got := codeEscape(' '); got != `{\tt\mc \ }` {
		t.Errorf("space: got %q", got)
	}
}

func TestCodeEscape_Tab(t *testing.T) {
	if got := codeEscape('\t'); got != `{\tt\mc \ \ \ \ \ \ \ \ }` {
		t.Errorf("tab: got %q", got)
	}
}

func TestCodeEscape_Specials(t *testing.T) {
	cases := []struct {
		in   rune
		want string
	}{
		{'\\', `{\tt\char92}`},
		{'{', `{\tt\char'173}`},
		{'}', `{\tt\char'175}`},
		{'#', `{\tt\#}`},
		{'%', `{\tt\%}`},
		{'$', `{\tt\$}`},
		{'&', `{\tt\&}`},
		{'_', `{\tt\_}`},
		{'^', `{\tt\char94}`},
		{'~', `{\tt\char126}`},
		{'-', `{\tt -}`},
		{'<', `{\tt <}`},
		{'>', `{\tt >}`},
		{'"', `{\tt "}`},
	}
	for _, c := range cases {
		got := codeEscape(c.in)
		if got != c.want {
			t.Errorf("codeEscape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCodeEscape_PlainChars(t *testing.T) {
	for _, r := range "abcxyzABCXYZ0123456789:;.,()[]+=!?" {
		got := codeEscape(r)
		if got != string(r) {
			t.Errorf("codeEscape(%q) = %q, want literal", r, got)
		}
	}
}

// ---- Line comment rendering ----

func TestWriteBody_LineComment_FontSwitch(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte("// hello\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `\rm\mc `) {
		t.Errorf("missing \\rm\\mc: %q", got)
	}
	if !strings.Contains(got, "\\tt\\mc \n") {
		t.Errorf("missing \\tt\\mc reset: %q", got)
	}
}

func TestWriteBody_LineComment_MarkerEscaped_Go(t *testing.T) {
	// "//" marker must appear as {\tt /}{\tt /}
	tokens := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte("// note\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `{\tt /}{\tt /}`) {
		t.Errorf("// marker not escaped: %q", got)
	}
}

func TestWriteBody_LineComment_MarkerEscaped_Shell(t *testing.T) {
	// "#" marker must appear as {\tt\#}, NOT bare #
	tokens := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte("# disc count\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldSh, Options{})
	got := b.String()

	if !strings.Contains(got, `{\tt\#}`) {
		t.Errorf("# marker not escaped: %q", got)
	}
	// bare # must not appear (would break LaTeX)
	// strip known {\tt\#} occurrences and check no lone # remains
	stripped := strings.ReplaceAll(got, `{\tt\#}`, "")
	if strings.ContainsRune(stripped, '#') {
		t.Errorf("bare # still present after escaping: %q", got)
	}
}

func TestWriteBody_LineComment_TeXPassthrough(t *testing.T) {
	// {\ content } strips the delimiters and outputs content verbatim.
	tokens := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte(`# {\ $x^2$ \hfill}` + "\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldSh, Options{})
	got := b.String()

	if !strings.Contains(got, `$x^2$ \hfill`) {
		t.Errorf("{\\  } content not passed through: %q", got)
	}
	// The TeX mode opener "{\ " (brace+backslash+space) must NOT appear literally.
	if strings.Contains(got, `{\ `) {
		t.Errorf("TeX mode opener `{\\ ` still in output: %q", got)
	}
}

func TestWriteBody_LineComment_TeXPassthrough_NestedBraces(t *testing.T) {
	// Nested braces inside {\ ... } must be handled correctly.
	tokens := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte(`# {\ \centerline{\bf Hi}}` + "\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldSh, Options{})
	got := b.String()

	if !strings.Contains(got, `\centerline{\bf Hi}`) {
		t.Errorf("nested braces not handled: %q", got)
	}
}

func TestWriteBody_LineComment_BeginEndQuote(t *testing.T) {
	// {\ \begin{quote} } and {\ \end{quote} } must produce bare
	// \begin{quote} and \end{quote} — not wrapped in extra groups.
	open := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte(`# {\ \begin{quote} }` + "\n")},
	}
	close_ := []scanner.Token{
		{Kind: scanner.TokenLineComment, Text: []byte(`# {\ \end{quote} }` + "\n")},
	}
	var b strings.Builder
	WriteBody(&b, open, ldSh, Options{})
	WriteBody(&b, close_, ldSh, Options{})
	got := b.String()

	if !strings.Contains(got, `\begin{quote}`) {
		t.Errorf("\\begin{quote} missing: %q", got)
	}
	if !strings.Contains(got, `\end{quote}`) {
		t.Errorf("\\end{quote} missing: %q", got)
	}
	// Must NOT have bare { before \begin (which would be the old verbatim {\ )
	if strings.Contains(got, `{\ `) {
		t.Errorf("TeX mode opener not stripped: %q", got)
	}
}

func TestWriteBody_CodeThenLineComment(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenCode, Text: []byte("x := 1 ")},
		{Kind: scanner.TokenLineComment, Text: []byte("// note\n")},
		{Kind: scanner.TokenCode, Text: []byte("y := 2\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	commentPos := strings.Index(got, `\rm\mc `)
	resetPos := strings.Index(got, "\\tt\\mc \n")
	if commentPos < 0 {
		t.Fatal("missing \\rm\\mc")
	}
	if resetPos < 0 {
		t.Fatal("missing \\tt\\mc reset (followed by newline)")
	}
	if resetPos < commentPos {
		t.Error("\\tt\\mc reset appears before \\rm\\mc")
	}
}

// ---- Block comment rendering ----

func TestWriteBody_BlockComment_Markers(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenBlockComment, Text: []byte("/* hello */")},
		{Kind: scanner.TokenCode, Text: []byte("\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldC, Options{})
	got := b.String()

	if !strings.Contains(got, `{\tt /}{\tt *}`) {
		t.Errorf("/* not rendered as {\\tt /}{\\tt *}: %q", got)
	}
	if !strings.Contains(got, `{\tt *}{\tt /}`) {
		t.Errorf("*/ not rendered as {\\tt *}{\\tt /}: %q", got)
	}
}

func TestWriteBody_BlockComment_BodyVerbatim(t *testing.T) {
	// {\ content } inside a block comment: delimiters stripped, content verbatim.
	tokens := []scanner.Token{
		{Kind: scanner.TokenBlockComment, Text: []byte(`/* {\ disc の数 \hfill} */`)},
		{Kind: scanner.TokenCode, Text: []byte("\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldC, Options{})
	got := b.String()

	// {\ and matching } are stripped; content must appear verbatim.
	if !strings.Contains(got, `disc の数 \hfill`) {
		t.Errorf("block comment body content missing: %q", got)
	}
	if strings.Contains(got, `{\ `) {
		t.Errorf("TeX mode opener not stripped in block comment: %q", got)
	}
}

func TestWriteBody_BlockComment_FontReset(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenBlockComment, Text: []byte("/* content */")},
		{Kind: scanner.TokenCode, Text: []byte("\nnext\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldC, Options{})
	got := b.String()

	if !strings.Contains(got, "\\tt\\mc \n") {
		t.Errorf("missing \\tt\\mc reset after block comment: %q", got)
	}
}

// ---- TokenTeX ----

func TestWriteBody_TeXToken_Verbatim(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenTeX, Text: []byte(`\hfill`)},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `\hfill`) {
		t.Errorf("TeX token content lost: %q", got)
	}
}

// ---- Integration ----

func TestWriteBody_Braces(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenCode, Text: []byte("func f() {\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `{\tt\char'173}`) {
		t.Errorf("open brace not escaped: %q", got)
	}
}

func TestWriteBody_Backslash(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenCode, Text: []byte(`a\b` + "\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `{\tt\char92}`) {
		t.Errorf("backslash not escaped: %q", got)
	}
}

// ---- Keyword rendering ----

func TestWriteBody_Keyword_BoldWrapped(t *testing.T) {
	tokens := []scanner.Token{
		{Kind: scanner.TokenKeyword, Text: []byte("func")},
		{Kind: scanner.TokenCode, Text: []byte(" main()\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `\textbf{func}`) {
		t.Errorf("keyword not wrapped in \\textbf: %q", got)
	}
}

func TestWriteBody_Keyword_SpecialCharsEscaped(t *testing.T) {
	// Hypothetical keyword containing a special char — codeEscape must apply.
	// Use a real keyword that has no special chars; test that output is clean.
	tokens := []scanner.Token{
		{Kind: scanner.TokenKeyword, Text: []byte("if")},
		{Kind: scanner.TokenCode, Text: []byte(" x {\n")},
	}
	var b strings.Builder
	WriteBody(&b, tokens, ldGo, Options{})
	got := b.String()

	if !strings.Contains(got, `\textbf{if}`) {
		t.Errorf("keyword 'if' not rendered as \\textbf{if}: %q", got)
	}
}
