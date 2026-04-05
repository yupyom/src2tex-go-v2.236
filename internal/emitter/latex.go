package emitter

import (
	"fmt"
	"io"
	"strings"

	"github.com/yupyom/src2tex-go-v2.236/internal/font"
	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
	"github.com/yupyom/src2tex-go-v2.236/internal/scanner"
)

const version = "2.236"

// Options controls preamble generation.
type Options struct {
	SourceFile           string // used in footer; may be empty
	CodeFont             string // \setmonofont value; default from font.GetPlatformDefaults()
	CJKFont              string // \setCJKmainfont / \setCJKsansfont value; default from font.GetPlatformDefaults()
	CommentFont          string // font name for comment text; if empty, uses CJKFont
	ExplicitCodeFont     bool   // true if -font was explicitly specified by the user
	ExplicitCommentFont  bool   // true if -commentfont was explicitly specified by the user
	HeaderContent        string // if non-empty, replace \fancyhead[R]{...} with this content
	FooterContent        string // if non-empty, replace \fancyfoot[R]{...} with this content
	LineNumbers          bool   // if true, emit line numbers in the left margin
	TabWidth             int    // tab expansion width in spaces (default: 8)
	Engine               string // TeX engine name ("xelatex", "pdflatex", etc.; empty = xelatex)
	PaperSize            string // paper size ("a4", "b5", "letter"; empty = a4)
}

// cjkMonoFont returns the font name for \setCJKmonofont.
// When the user has set CJKFont explicitly, use it for mono too.
// Otherwise use the platform's CJKMonoFont (which may differ from CJKSansFont on Windows).
func cjkMonoFont(opts Options, pd font.PlatformDefaults) string {
	if opts.CJKFont != "" {
		return opts.CJKFont
	}
	return pd.CJKMonoFont
}

func headerLine(opts Options) string {
	if opts.HeaderContent != "" {
		return strings.TrimRight(opts.HeaderContent, "\n")
	}
	return `\fancyhead[R]{\rm src2tex version ` + version + `}`
}

func footerLine(opts Options) string {
	if opts.FooterContent != "" {
		return strings.TrimRight(opts.FooterContent, "\n")
	}
	return `\fancyfoot[R]{\rm ` + texifyPath(opts.SourceFile) + `\qquad page \thepage}`
}

// WritePreamble emits a TeX preamble to w.
// When opts.Engine is empty or "xelatex" and no user template exists,
// it uses the legacy hardcoded XeLaTeX preamble for backward compatibility.
// When opts.Engine is set, it loads the engine template from
// ~/.src2tex/engines/<engine>/ or the embedded defaults.
func WritePreamble(w io.Writer, opts Options) {
	engine := opts.Engine
	if engine == "" {
		engine = "xelatex"
	}

	// Build template data.
	pd := font.GetPlatformDefaults()

	cjkFont := opts.CJKFont
	if cjkFont == "" {
		cjkFont = pd.CJKSansFont
	}

	// --- MonoFontLine: Path=-aware resolution via font.ResolveCodeFontLine ---
	monoFontLine := font.ResolveCodeFontLine(opts.CodeFont)

	// --- CJKMonoFontLine: Path=-aware for unified code fonts ---
	cjkMonoFontLine := font.ResolveCodeFontCJKMonoLine(opts.CodeFont)

	// --- SansFontLine / CJKSansFontLine: Path=-aware for code fonts ---
	// Latin \setsansfont is only set when -font is explicitly specified,
	// to avoid overriding the engine's default Latin sans-serif font.
	var sansFontLine string
	if opts.ExplicitCodeFont {
		sansFontLine = font.ResolveSansFontLine(opts.CodeFont)
	}
	cjkSansFontLine := font.ResolveCodeFontCJKSansLine(opts.CodeFont)

	// --- MainFontLine / CJKMainFont: comment font, engine-specific ---
	// Latin \setmainfont is only set when -commentfont is explicitly specified,
	// to avoid overriding the engine's default Latin serif font.
	var mainFontLine string
	if opts.ExplicitCommentFont {
		mainFontLine = font.ResolveMainFontLine(opts.CommentFont)
	}
	var setCJKmain string
	if opts.CommentFont != "" {
		// Comment font specified: resolve with Path= support.
		setCJKmain = font.ResolveCommentFontLine(opts.CommentFont)
	} else if cjkMainFromCode := font.ResolveCodeFontCJKMainLine(opts.CodeFont); cjkMainFromCode != "" {
		// Unified code font, no comment font → use code font for CJKmainfont too.
		setCJKmain = cjkMainFromCode
	} else {
		// Default: use CJK sans font.
		setCJKmain = `\setCJKmainfont{` + cjkFont + `}`
	}

	// For LuaLaTeX, convert xeCJK commands to luatexja-fontspec equivalents.
	// luatexja-fontspec uses \setmainjfont, \setsansjfont, \setmonojfont
	// instead of xeCJK's \setCJKmainfont, \setCJKsansfont, \setCJKmonofont.
	// See: https://mirrors.ibiblio.org/CTAN/macros/luatex/generic/luatexja/doc/luatexja-ja.pdf
	if engine == "lualatex" {
		setCJKmain = strings.Replace(setCJKmain, `\setCJKmainfont`, `\setmainjfont`, 1)
		cjkSansFontLine = strings.Replace(cjkSansFontLine, `\setCJKsansfont`, `\setsansjfont`, 1)
		cjkMonoFontLine = strings.Replace(cjkMonoFontLine, `\setCJKmonofont`, `\setmonojfont`, 1)
	}

	paperSize := NormalizePaperSize(opts.PaperSize)
	if paperSize == "" {
		paperSize = "a4paper"
	}

	data := PreambleData{
		PaperSize:       paperSize,
		Margin:          MarginForPaper(paperSize),
		MainFontLine:    mainFontLine,
		SansFontLine:    sansFontLine,
		MonoFontLine:    monoFontLine,
		CJKMainFont:     setCJKmain,
		CJKSansFont:     cjkFont,
		CJKSansFontLine: cjkSansFontLine,
		CJKMonoFont:     cjkMonoFont(opts, pd),
		CJKMonoFontLine: cjkMonoFontLine,
		Header:          headerLine(opts),
		Footer:          footerLine(opts),
	}

	// Try to load engine template.
	cfg, tmpl, err := LoadEngine(engine)
	if err != nil {
		fmt.Fprintf(w, "%% src2tex: engine %q not found, using fallback: %v\n", engine, err)
		writePreambleLegacy(w, opts, data)
		return
	}

	// Engines without fontspec (e.g. upLaTeX, pdfLaTeX) can't use \setmonofont.
	if !cfg.HasFontspec() {
		data.MonoFontLine = ""
		data.CJKMonoFontLine = ""
	}

	if err := tmpl.Execute(w, data); err != nil {
		fmt.Fprintf(w, "%% src2tex: template error: %v\n", err)
		writePreambleLegacy(w, opts, data)
	}
}


// writePreambleLegacy is the original hardcoded XeLaTeX preamble, used as
// fallback when template loading fails.
func writePreambleLegacy(w io.Writer, opts Options, data PreambleData) {
	lines := []string{
		`\documentclass[` + data.PaperSize + `,10pt]{article}`,
		`\usepackage{fontspec}`,
		`\usepackage{xeCJK}`,
		`\usepackage[` + data.PaperSize + `,margin=` + data.Margin + `]{geometry}`,
		`\usepackage{graphicx}`,
		`\usepackage{fancyhdr}`,
		data.MonoFontLine,
		data.CJKMainFont,
		`\setCJKsansfont{` + data.CJKSansFont + `}`,
		`\setCJKmonofont{` + data.CJKMonoFont + `}`,
		`\XeTeXlinebreaklocale ""`,
		`\xeCJKsetup{CJKglue={},CJKecglue={}}`,
		`\newdimen\charwd`,
		`{\tt\global\setbox0=\hbox{x}\global\charwd=\wd0}`,
		`\frenchspacing`,
		`\pagestyle{fancy}`,
		`\renewcommand{\headrulewidth}{0pt}`,
		`\fancyhf{}`,
		data.Header,
		data.Footer,
		`% plain TeX compatibility`,
		`\makeatletter`,
		`\providecommand{\eqalign}[1]{\vcenter{\openup1\jot\m@th`,
		`  \ialign{\strut\hfil$\displaystyle{##}$&$\displaystyle{{}##}$\hfil\crcr#1\crcr}}}`,
		`\makeatother`,
		`\ifx\sevenrm\undefined`,
		`  \font\sevenrm=cmr7 scaled \magstep0`,
		`\fi`,
		`\def\mc{\relax}`,
		`\def\gt{\relax}`,
		`\def\sc{\scshape}`,
		`\begin{document}`,
		`\tt\mc `,
		``,
	}
	io.WriteString(w, strings.Join(lines, "\n"))
}

// WritePostamble emits the closing boilerplate to w.
func WritePostamble(w io.Writer) {
	io.WriteString(w, "\n\\rm\\mc\n\n\\end{document}\n")
}

// WriteBody converts tokens to LaTeX body text and writes to w.
//
// ld is needed to know the comment marker strings (e.g. "//" or "#").
//
// Token rendering rules:
//   - TokenCode:         \tt\mc font, special chars escaped per codeEscape
//   - TokenLineComment:  \rm\mc; marker in {\tt X} form; content via texPassthrough
//   - TokenBlockComment: \rm\mc; markers in {\tt X} form; body via texPassthrough
//   - TokenTeX:          verbatim, no escaping
//
// Line structure:
//   - Each source line   → \noindent\n\mbox{} + content
//   - Empty source line  → \noindent\n\mbox{}\hfill
//   - After any comment  → \tt\mc on its own line + blank separator line
func WriteBody(w io.Writer, tokens []scanner.Token, ld *lang.LangDef, opts Options) {
	tabWidth := opts.TabWidth
	if tabWidth <= 0 {
		tabWidth = 8
	}
	bw := &bodyWriter{w: w, ld: ld, atLineStart: true, lineNumbers: opts.LineNumbers, tabWidth: tabWidth}
	for _, tok := range tokens {
		switch tok.Kind {
		case scanner.TokenCode:
			bw.writeCodeToken(tok.Text)
		case scanner.TokenKeyword:
			bw.writeKeywordToken(tok.Text)
		case scanner.TokenLineComment:
			bw.writeLineCommentToken(tok.Text)
		case scanner.TokenBlockComment:
			bw.writeBlockCommentToken(tok.Text)
		case scanner.TokenTeX:
			bw.writeTeXToken(tok.Text)
		}
	}
	if bw.needTTReset {
		io.WriteString(w, "\n\\tt\\mc \n")
	}
}

// bodyWriter holds the stateful rendering context.
type bodyWriter struct {
	w            io.Writer
	ld           *lang.LangDef
	atLineStart  bool // true at document start and after every '\n'
	needTTReset  bool // true after a comment; emit \tt\mc before next \noindent
	rawTeXDepth  int  // nesting depth of {\ } / {\cmd } blocks in RawTeX comments
	lineNumbers  bool // whether to emit line numbers
	lineNo       int  // current source line number (1-based)
	tabWidth     int  // tab expansion width in spaces
}

// beginLine emits \tt\mc reset (if pending) and a blank line, then \noindent\n\mbox{}.
// Every \noindent must be preceded by a blank line so LaTeX treats each source
// line as a separate paragraph.
func (bw *bodyWriter) beginLine() {
	if bw.needTTReset {
		io.WriteString(bw.w, "\\tt\\mc \n\n")
		bw.needTTReset = false
	} else {
		io.WriteString(bw.w, "\n")
	}
	io.WriteString(bw.w, "\\noindent\n\\mbox{}")
	if bw.lineNumbers {
		bw.lineNo++
		fmt.Fprintf(bw.w, "\\rlap{\\kern-2.5em{\\sevenrm %d}}", bw.lineNo)
	}
	bw.atLineStart = false
}

// writeKeywordToken renders a TokenKeyword as \textbf{keyword} in \tt\mc context.
// Keywords never contain newlines, so atLineStart handling is the same as
// an ordinary non-newline code character.
func (bw *bodyWriter) writeKeywordToken(text []byte) {
	if bw.atLineStart {
		bw.beginLine()
	}
	io.WriteString(bw.w, `\textbf{`)
	for _, r := range string(text) {
		io.WriteString(bw.w, codeEscape(r))
	}
	io.WriteString(bw.w, `}`)
}

// writeCodeToken renders a TokenCode fragment.
func (bw *bodyWriter) writeCodeToken(text []byte) {
	for _, r := range string(text) {
		if bw.atLineStart {
			if r == '\n' {
				bw.beginLine()
				io.WriteString(bw.w, "\\hfill\n")
				bw.atLineStart = true
				continue
			}
			bw.beginLine()
		}
		if r == '\n' {
			io.WriteString(bw.w, "\n")
			bw.atLineStart = true
		} else if r == '\t' {
			// Tab expansion using configured width
			for i := 0; i < bw.tabWidth; i++ {
				io.WriteString(bw.w, codeEscape(' '))
			}
		} else {
			io.WriteString(bw.w, codeEscape(r))
		}
	}
}

// writeLineCommentToken renders a TokenLineComment fragment.
//
// The comment marker (e.g. "//" or "#") is rendered in {\tt X} form.
// The content after the marker is processed via texPassthrough so that
// {\ ... } regions are stripped of their delimiters and emitted as raw TeX,
// while text outside {\ ... } is output verbatim in \rm\mc context.
//
// RawTeX languages (REDUCE, MATLAB) are handled by writeRawTeXComment.
func (bw *bodyWriter) writeLineCommentToken(text []byte) {
	if bw.ld.Comment.RawTeX {
		bw.writeRawTeXComment(text)
		return
	}
	if bw.atLineStart {
		bw.beginLine()
	}
	io.WriteString(bw.w, "\\rm\\mc ")

	marker := bw.ld.Comment.LineComment
	for _, r := range marker {
		io.WriteString(bw.w, commentMarkerChar(r))
	}

	afterMarker := text[len(marker):]
	content := []rune(string(afterMarker))
	i := 0
	inMath := false
	mathDelim := ""
	for i < len(content) {
		r := content[i]
		if r == '\n' {
			io.WriteString(bw.w, "\n")
			bw.atLineStart = true
			bw.needTTReset = true
			return
		}
		if inMath {
			if mathDelim == "$$" && r == '$' && i+1 < len(content) && content[i+1] == '$' {
				io.WriteString(bw.w, "$$")
				inMath = false
				i += 2
				continue
			} else if mathDelim == "$" && r == '$' {
				io.WriteString(bw.w, "$")
				inMath = false
				i++
				continue
			}
			io.WriteString(bw.w, string(r))
			i++
			continue
		}
		// Detect $$ or $ math mode: pass through to closing delimiter without escaping.
		if r == '$' {
			if i+1 < len(content) && content[i+1] == '$' {
				inMath = true
				mathDelim = "$$"
				io.WriteString(bw.w, "$$")
				i += 2
				continue
			}
			inMath = true
			mathDelim = "$"
			io.WriteString(bw.w, "$")
			i++
			continue
		}
		if r == '{' && i+1 < len(content) && content[i+1] == '\\' {
			// {\cmd...}: TeX mode marker — strip the opening { and closing },
			// output content (from \ onward) verbatim.
			end := findMatchingBrace(content, i)
			io.WriteString(bw.w, string(content[i+1:end]))
			i = end + 1
		} else {
			io.WriteString(bw.w, commentTextEscape(r))
			i++
		}
	}
	// EOF without '\n'.
	bw.needTTReset = true
}

// writeRawTeXComment handles TokenLineComment for RawTeX languages (REDUCE, MATLAB).
//
// At rawTeXDepth == 0 (outside a TeX block):
//   - emits the % marker as {\tt\%}
//   - processes content with normal commentTextEscape
//   - detects {\ or {\cmd opening a multi-line TeX block (no matching } on same line)
//   - detects {\ or {\cmd with matching } on same line as single-line passthrough
//
// At rawTeXDepth > 0 (inside a multi-line TeX block):
//   - suppresses \noindent\mbox{} and the % marker
//   - emits content as raw TeX, tracking { / } for depth
//   - when the outermost } is found, emits \rm\mc and returns to depth 0
func (bw *bodyWriter) writeRawTeXComment(text []byte) {
	marker := bw.ld.Comment.LineComment

	if bw.rawTeXDepth > 0 {
		// Inside a multi-line TeX block: raw output, track brace depth.
		// Skip just the marker; keep any space as-is (LaTeX formatting).
		afterMarker := text[len(marker):]
		content := []rune(string(afterMarker))
		i := 0
		for i < len(content) {
			r := content[i]
			if r == '\n' {
				io.WriteString(bw.w, "\n")
				bw.atLineStart = true
				bw.needTTReset = true
				return
			}
			if r == '{' {
				bw.rawTeXDepth++
				io.WriteString(bw.w, "{")
			} else if r == '}' {
				bw.rawTeXDepth--
				if bw.rawTeXDepth == 0 {
					// Outermost closing brace: end of TeX block, switch to rm\mc.
					io.WriteString(bw.w, "\\rm\\mc ")
					// Consume remainder of line (typically just whitespace + \n).
					i++
					for i < len(content) {
						if content[i] == '\n' {
							io.WriteString(bw.w, "\n")
							bw.atLineStart = true
							bw.needTTReset = true
							return
						}
						i++
					}
					bw.needTTReset = true
					return
				}
				io.WriteString(bw.w, "}")
			} else {
				io.WriteString(bw.w, string(r))
			}
			i++
		}
		bw.needTTReset = true
		return
	}

	// rawTeXDepth == 0: show % marker, use normal comment escaping.
	if bw.atLineStart {
		bw.beginLine()
	}
	io.WriteString(bw.w, "\\rm\\mc ")
	for _, r := range marker {
		io.WriteString(bw.w, commentMarkerChar(r))
	}

	afterMarker := text[len(marker):]
	content := []rune(string(afterMarker))
	i := 0
	for i < len(content) {
		r := content[i]
		if r == '\n' {
			io.WriteString(bw.w, "\n")
			bw.atLineStart = true
			bw.needTTReset = true
			return
		}
		if r == '{' && i+1 < len(content) && content[i+1] == '\\' {
			end := findMatchingBrace(content, i)
			if end < len(content) {
				// Single-line {\ ... } passthrough: strip { and }, emit from \ onward.
				io.WriteString(bw.w, string(content[i+1:end]))
				i = end + 1
			} else {
				// No matching } on this line: start of a multi-line TeX block.
				bw.rawTeXDepth++
				// Emit content from \ to end of line.
				rest := content[i+1:]
				for _, rc := range rest {
					if rc == '\n' {
						io.WriteString(bw.w, "\n")
						bw.atLineStart = true
						bw.needTTReset = true
						return
					}
					io.WriteString(bw.w, string(rc))
				}
				bw.needTTReset = true
				return
			}
		} else {
			io.WriteString(bw.w, commentTextEscape(r))
			i++
		}
	}
	bw.needTTReset = true
}

// writeBlockCommentToken renders a TokenBlockComment fragment.
//
// The token text is the full /* ... */ region (both markers included).
// For docstring tokens (BlockOpen == ""), the delimiter (""" or ''') is
// detected from the token text itself.
// Body content is processed via the same {\ ... } passthrough logic.
func (bw *bodyWriter) writeBlockCommentToken(text []byte) {
	if bw.atLineStart {
		bw.beginLine()
	}

	// Detect open/close markers from the token text itself, not from bw.ld.
	// This is essential for composite languages (e.g. HTML with embedded CSS/JS):
	// a CSS /* */ comment inside <style> must be rendered with /* */ markers,
	// not with the HTML <!-- --> markers from the top-level LangDef.
	open, close_ := detectBlockCommentMarkers(text)

	// \rm\mc {\tt /}{\tt *}\
	io.WriteString(bw.w, "\\rm\\mc ")
	for _, r := range open {
		io.WriteString(bw.w, commentMarkerChar(r))
	}
	io.WriteString(bw.w, `\ `) // space after opener

	// Body: everything between the markers, with {\ ... } passthrough.
	// Text outside {\ ... } is in \rm\mc context and needs TeX-special escaping,
	// unless the language uses RawTeX comments (e.g. REDUCE, MATLAB).
	escFn := commentTextEscapeFor(bw.ld)
	body := []rune(string(text[len(open) : len(text)-len(close_)]))
	i := 0
	leadingWS := false // true only after a newline inside the body
	inMath := false
	mathDelim := ""

	// Pascal { } comments: if the body (after trimming leading whitespace) starts
	// with '\', treat the entire body as TeX passthrough (raw LaTeX output).
	// This handles patterns like { \underline{\textsf{...}} } where the body
	// contains TeX commands that would otherwise conflict with comment escaping.
	if open == "{" && close_ == "}" {
		trimmed := strings.TrimLeft(string(body), " \t")
		if len(trimmed) > 0 && trimmed[0] == '\\' {
			io.WriteString(bw.w, string(body))
			if bw.atLineStart {
				bw.beginLine()
			}
			io.WriteString(bw.w, `\rm\mc \ `)
			for _, r := range close_ {
				io.WriteString(bw.w, commentMarkerChar(r))
			}
			bw.needTTReset = true
			return
		}
	}

	// Handle /** pattern: if body starts with *, render as monospace marker.
	// This ensures the third character of /** appears in \tt font like /* and */.
	if len(body) > 0 && body[0] == '*' {
		io.WriteString(bw.w, commentMarkerChar('*'))
		i = 1
	}

	for i < len(body) {
		r := body[i]

		// Inside $...$ or $$...$$: output raw without escaping.
		// Newlines are emitted as-is (no \noindent\mbox{} machinery).
		if inMath {
			if r == '\n' {
				io.WriteString(bw.w, "\n")
				i++
				continue
			}
			if mathDelim == "$$" && r == '$' && i+1 < len(body) && body[i+1] == '$' {
				io.WriteString(bw.w, "$$")
				inMath = false
				i += 2
				continue
			} else if mathDelim == "$" && r == '$' {
				io.WriteString(bw.w, "$")
				inMath = false
				i++
				continue
			}
			io.WriteString(bw.w, string(r))
			i++
			continue
		}

		if r == '\n' {
			if bw.atLineStart {
				// Consecutive newline = empty line within block comment.
				bw.beginLine()
				io.WriteString(bw.w, "\\hfill\n")
				bw.atLineStart = true
			} else {
				// End of a content line within block comment.
				io.WriteString(bw.w, "\n")
				bw.atLineStart = true
				bw.needTTReset = true
			}
			leadingWS = true
			i++
			continue
		}
		// If at start of a new line within the block comment, emit
		// \noindent\mbox{} and re-enter \rm\mc comment font.
		if bw.atLineStart {
			bw.beginLine()
			io.WriteString(bw.w, "\\rm\\mc ")
		}
		// Emit leading whitespace as monospaced spaces so that continuation
		// lines align with the opening /* which uses {\tt\mc \ }.
		if leadingWS && (r == ' ' || r == '\t') {
			io.WriteString(bw.w, `{\tt\mc \ }`)
			i++
			continue
		}
		// Render * immediately after leading whitespace as monospace to match
		// the /* and */ comment markers. Handles " * content" continuation lines.
		if leadingWS && r == '*' {
			io.WriteString(bw.w, commentMarkerChar(r))
			leadingWS = false
			i++
			continue
		}
		leadingWS = false
		// Detect $$ or $ math mode: pass through to closing delimiter without escaping.
		// Not applicable to RawTeX languages where content is already raw LaTeX.
		if !bw.ld.Comment.RawTeX && r == '$' {
			if i+1 < len(body) && body[i+1] == '$' {
				inMath = true
				mathDelim = "$$"
				io.WriteString(bw.w, "$$")
				i += 2
				continue
			}
			inMath = true
			mathDelim = "$"
			io.WriteString(bw.w, "$")
			i++
			continue
		}
		if !bw.ld.Comment.RawTeX && r == '{' && i+1 < len(body) && body[i+1] == '\\' {
			// {\cmd...}: TeX mode marker — strip { and matching }, output from \ onward.
			// Skipped for RawTeX languages: all block-comment content is already raw LaTeX.
			end := findMatchingBrace(body, i)
			passthroughContent := string(body[i+1 : end])
			if containsParagraphLevelCmd(passthroughContent) && !strings.Contains(passthroughContent, "\n") {
				// Wrap paragraph-level commands (e.g. \centerline) in \parbox
				// to contain their paragraph-break side effects. Without this,
				// \centerline would push the closing */ marker to a new line.
				// Use 0.85\textwidth to leave room for the comment markers.
				io.WriteString(bw.w, "\\parbox[t]{0.85\\textwidth}{")
				io.WriteString(bw.w, passthroughContent)
				io.WriteString(bw.w, "}")
			} else {
				io.WriteString(bw.w, passthroughContent)
			}
			i = end + 1
		} else {
			io.WriteString(bw.w, escFn(r))
			i++
		}
	}

	// If the body ended with \n, the closing delimiter goes on its own line.
	if bw.atLineStart {
		bw.beginLine()
	}
	// \rm\mc \ {\tt *}{\tt /}
	io.WriteString(bw.w, `\rm\mc \ `)
	for _, r := range close_ {
		io.WriteString(bw.w, commentMarkerChar(r))
	}

	bw.needTTReset = true
	// The '\n' after */ comes from the next TokenCode token.
}

// detectBlockCommentMarkers detects the open/close markers from the token text.
// This allows the emitter to work correctly for composite languages where the
// BlockComment token may come from a sub-language with different markers than
// the top-level LangDef (e.g. /* */ inside HTML's <style>).
func detectBlockCommentMarkers(text []byte) (open, close_ string) {
	s := string(text)

	// Known block comment marker pairs, ordered by specificity.
	type markerPair struct {
		open, close string
	}
	pairs := []markerPair{
		{"<!--", "-->"},  // HTML/XML
		{"/*", "*/"},     // C-family, CSS
		{"(*", "*)"},     // Pascal alt block
		{"{", "}"},       // Pascal
		{`"""`, `"""`},   // Python docstring (double)
		{`'''`, `'''`},   // Python docstring (single)
	}

	for _, p := range pairs {
		if len(s) >= len(p.open)+len(p.close) &&
			s[:len(p.open)] == p.open &&
			s[len(s)-len(p.close):] == p.close {
			return p.open, p.close
		}
	}

	// Fallback: no markers detected — treat entire text as body.
	return "", ""
}

// containsParagraphLevelCmd reports whether s contains a TeX command that
// creates a paragraph break (e.g. \centerline). Such commands disrupt the
// inline flow and push subsequent content to a new visual line.
// Note: \begin{}/\end{} are not included because they typically appear in
// matched pairs and don't disrupt the closing marker on the same line.
func containsParagraphLevelCmd(s string) bool {
	cmds := []string{
		`\centerline`,
	}
	for _, cmd := range cmds {
		if strings.Contains(s, cmd) {
			return true
		}
	}
	return false
}

// writeTeXToken renders a TokenTeX fragment verbatim.
func (bw *bodyWriter) writeTeXToken(text []byte) {
	for _, r := range string(text) {
		if bw.atLineStart {
			if r == '\n' {
				bw.beginLine()
				io.WriteString(bw.w, "\\hfill\n")
				bw.atLineStart = true
				continue
			}
			bw.beginLine()
		}
		if r == '\n' {
			io.WriteString(bw.w, "\n")
			bw.atLineStart = true
		} else {
			io.WriteString(bw.w, string(r))
		}
	}
}

// findMatchingBrace returns the index of the } that closes the { at pos,
// tracking nested brace depth. Returns len(runes) if no match is found.
func findMatchingBrace(runes []rune, pos int) int {
	depth := 1
	i := pos + 1
	for i < len(runes) {
		switch runes[i] {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth == 0 {
			return i
		}
		i++
	}
	return len(runes)
}

// codeEscape converts a single rune to its LaTeX representation in \tt\mc context.
func codeEscape(r rune) string {
	switch r {
	case ' ':
		return `{\tt\mc \ }`
	case '\t':
		return `{\tt\mc \ \ \ \ \ \ \ \ }` // one tab = 8 spaces
	case '\\':
		return `{\tt\char92}`
	case '{':
		return `{\tt\char'173}`
	case '}':
		return `{\tt\char'175}`
	case '#':
		return `{\tt\#}`
	case '%':
		return `{\tt\%}`
	case '$':
		return `{\tt\$}`
	case '&':
		return `{\tt\&}`
	case '_':
		return `{\tt\_}`
	case '^':
		return `{\tt\char94}`
	case '~':
		return `{\tt\char126}`
	case '-':
		return `{\tt -}`
	case '<':
		return `{\tt <}`
	case '>':
		return `{\tt >}`
	case '"':
		return `{\tt "}`
	default:
		return string(r)
	}
}

// commentTextEscapeFor returns commentTextEscape, or identity when the language
// uses RawTeX comments (e.g. REDUCE, MATLAB), where content is raw LaTeX and
// must not have $ & ^ etc. escaped.
func commentTextEscapeFor(ld *lang.LangDef) func(rune) string {
	if ld.Comment.RawTeX {
		return func(r rune) string { return string(r) }
	}
	return commentTextEscape
}

// commentTextEscape escapes the minimum set of TeX special characters in plain
// comment text (content outside {\ ... } regions, rendered in \rm\mc context).
//
// Characters that are NOT escaped: \  {  }
//   These are valid TeX constructs (commands, groups) and must pass through so
//   that {\hrulefill}, {\tt X}, etc. work correctly without {\ ... } wrappers.
func commentTextEscape(r rune) string {
	switch r {
	case '_':
		return `\_`
	case '^':
		return `\^{}`
	case '$':
		return `\$`
	case '%':
		return `\%`
	case '&':
		return `\&`
	case '~':
		return `\~{}`
	case '#':
		return `\#`
	default:
		return string(r)
	}
}

// commentMarkerChar renders a single rune of a comment marker (e.g. '/', '*', '#')
// as {\tt X} so it looks like code within \rm\mc context.
func commentMarkerChar(r rune) string {
	switch r {
	case '/':
		return `{\tt /}`
	case '*':
		return `{\tt *}`
	case '#':
		return `{\tt\#}`
	case ';':
		return `{\tt ;}`
	case '%':
		return `{\tt\%}`
	case '{':
		return `{\tt\char'173}`
	case '}':
		return `{\tt\char'175}`
	default:
		return `{\tt ` + string(r) + `}`
	}
}

// texifyPath converts a file path string for use in LaTeX (e.g. the footer),
// replacing '/' with {\tt /} and escaping '_' as {\tt\_} so the path renders
// correctly in \rm\mc context.
func texifyPath(p string) string {
	p = strings.ReplaceAll(p, "/", `{\tt /}`)
	p = strings.ReplaceAll(p, "_", `{\tt\_}`)
	return p
}
