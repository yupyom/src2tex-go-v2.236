package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yupyom/src2tex-go-v2.236/internal/emitter"
	"github.com/yupyom/src2tex-go-v2.236/internal/font"
	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
	"github.com/yupyom/src2tex-go-v2.236/internal/postprocess"
	"github.com/yupyom/src2tex-go-v2.236/internal/scanner"
)

func main() {
	// Subcommand: src2tex font ...
	if len(os.Args) >= 2 && os.Args[1] == "font" {
		runFontCmd(os.Args[2:])
		return
	}

	// Subcommand: src2tex engine ...
	if len(os.Args) >= 2 && os.Args[1] == "engine" {
		runEngineCmd(os.Args[2:])
		return
	}

	langFlag := flag.String("lang", "", "language override (c, go, sh, ...)")
	outFlag := flag.String("o", "", "output file (default: <input>.tex, or stdout for stdin)")
	fontFlag := flag.String("font", "", "monospace font for \\setmonofont (default: Courier New)")
	commentFontFlag := flag.String("commentfont", "", "CJK mincho font for comment text (default: auto-detect)")
	fontDirFlag := flag.String("fontdir", "", "directory for downloaded fonts (default: ~/.src2tex/fonts/)")
	listFontsFlag := flag.Bool("listfonts", false, "list available fonts and exit")
	headerFlag := flag.String("header", "", "file whose contents replace \\fancyhead[R]{...}")
	footerFlag := flag.String("footer", "", "file whose contents replace \\fancyfoot[R]{...}")
	lineNumbersFlag := flag.Bool("linenumbers", false, "emit source line numbers in the left margin")
	tabFlag := flag.Int("tab", 8, "tab width in spaces")
	engineFlag := flag.String("engine", "xelatex", "TeX engine (xelatex, lualatex, uplatex, pdflatex, tectonic)")
	paperFlag := flag.String("paper", "a4", "paper size (a4, b5, letter)")
	flag.Usage = usage
	flag.Parse()

	fontDir := *fontDirFlag

	if *listFontsFlag {
		font.PrintFontList(fontDir)
		return
	}

	commentFont := resolveCommentFont(*commentFontFlag, fontDir)

	headerContent, err := readOptionalFile(*headerFlag, "header")
	if err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
		os.Exit(1)
	}
	footerContent, err := readOptionalFile(*footerFlag, "footer")
	if err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		// stdin → stdout: -lang flag is the only hint, use flag-first resolution.
		ld := resolveByFlag(*langFlag)
		if err := run(os.Stdin, os.Stdout, ld, *fontFlag, commentFont, "", headerContent, footerContent, *lineNumbersFlag, *tabFlag, *engineFlag, *paperFlag, *fontFlag != "", *commentFontFlag != ""); err != nil {
			fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
			os.Exit(1)
		}
		return
	}

	inputPath := flag.Arg(0)

	// Language resolution:
	//   -lang explicitly set → FindByFlag first (e.g. -lang php → PHP LangDef)
	//   auto-detected from extension → FindByExt first (e.g. .php → HTML LangDef)
	//   no extension → try FindByFileName (e.g. "Makefile")
	var ld *lang.LangDef
	if *langFlag != "" {
		ld = resolveByFlag(*langFlag)
	} else {
		ext := extWithoutDot(inputPath)
		if ext != "" {
			ld = resolveByExt(ext)
		} else {
			// No extension: try matching by exact filename (e.g. "Makefile").
			ld = lang.FindByFileName(filepath.Base(inputPath))
		}
	}

	outPath := *outFlag
	if outPath == "" {
		outPath = addExt(inputPath, ".tex")
	}

	in, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	if err := run(in, w, ld, *fontFlag, commentFont, shortPath(inputPath), headerContent, footerContent, *lineNumbersFlag, *tabFlag, *engineFlag, *paperFlag, *fontFlag != "", *commentFontFlag != ""); err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: %v\n", err)
		os.Exit(1)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: flush: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "src2tex: %s -> %s\n", inputPath, outPath)

	// Post-process: convert \special{epsfile=...} and EPS→PDF for referenced images
	if err := postprocess.ProcessTexFile(outPath); err != nil {
		fmt.Fprintf(os.Stderr, "src2tex: postprocess: %v\n", err)
		// Non-fatal: the .tex file was written successfully
	}
}

// runFontCmd handles the "font" subcommand: font list | font install <name> | font init.
func runFontCmd(args []string) {
	fset := flag.NewFlagSet("font", flag.ExitOnError)
	fontDirFlag := fset.String("fontdir", "", "directory for downloaded fonts (default: ~/.src2tex/fonts/)")
	fset.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: src2tex font <subcommand> [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Subcommands:")
		fmt.Fprintln(os.Stderr, "  list                    list available fonts")
		fmt.Fprintln(os.Stderr, "  init                    generate/update ~/.src2tex/fonts.json")
		fmt.Fprintln(os.Stderr, "  install <name>          download and install a code font")
		fmt.Fprintln(os.Stderr, "  install-comment <name>  download and install a comment (mincho) font")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fset.PrintDefaults()
	}

	if len(args) == 0 {
		fset.Usage()
		os.Exit(1)
	}

	// Parse flags after the subcommand word
	subcmd := args[0]
	_ = fset.Parse(args[1:])
	fontDir := *fontDirFlag

	switch subcmd {
	case "list":
		font.PrintFontList(fontDir)

	case "init":
		if err := font.SaveFontsConfig(fontDir); err != nil {
			fmt.Fprintf(os.Stderr, "src2tex: font init: %v\n", err)
			os.Exit(1)
		}

	case "install":
		name := fset.Arg(0)
		if name == "" {
			fmt.Fprintln(os.Stderr, "Usage: src2tex font install <name|all>")
			fmt.Fprintln(os.Stderr, "Run 'src2tex font list' to see available fonts.")
			os.Exit(1)
		}
		if err := font.Install(name, fontDir); err != nil {
			fmt.Fprintf(os.Stderr, "src2tex: font install: %v\n", err)
			os.Exit(1)
		}

	case "install-comment":
		name := fset.Arg(0)
		if name == "" {
			fmt.Fprintln(os.Stderr, "Usage: src2tex font install-comment <name|all>")
			fmt.Fprintln(os.Stderr, "Run 'src2tex font list' to see available comment fonts.")
			os.Exit(1)
		}
		if err := font.InstallCommentFont(name, fontDir); err != nil {
			fmt.Fprintf(os.Stderr, "src2tex: font install-comment: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "src2tex: unknown font subcommand %q\n", subcmd)
		fset.Usage()
		os.Exit(1)
	}
}

// runEngineCmd handles the "engine" subcommand: engine init | engine list.
func runEngineCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: src2tex engine <subcommand>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Subcommands:")
		fmt.Fprintln(os.Stderr, "  init    extract built-in engine templates to ~/.src2tex/engines/")
		fmt.Fprintln(os.Stderr, "  list    list available engines")
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		if err := emitter.InitEngines(); err != nil {
			fmt.Fprintf(os.Stderr, "src2tex: engine init: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "src2tex: engine templates extracted to ~/.src2tex/engines/")
	case "list":
		engines := emitter.ListEngines()
		fmt.Println("Available engines:")
		for _, name := range engines {
			cfg, _, err := emitter.LoadEngine(name)
			if err != nil {
				fmt.Printf("  %-12s (error: %v)\n", name, err)
			} else if emitter.IsCustomEngine(name) {
				fmt.Printf("  %-12s %s [custom]\n", name, cfg.Description)
			} else {
				fmt.Printf("  %-12s %s\n", name, cfg.Description)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "src2tex: unknown engine subcommand %q\n", args[0])
		os.Exit(1)
	}
}

// resolveCommentFont returns the comment font name to use.
// If commentFontName is empty, auto-detection is used.
// Returns the font name (not a LaTeX command); LaTeX command generation is
// handled by emitter.WritePreamble via font.ResolveCommentFontLine.
func resolveCommentFont(commentFontName, fontDir string) string {
	name := commentFontName
	if name == "" {
		name = font.AutoDetectCommentFont(fontDir)
	}
	return name
}

// run executes the Scanner → Emitter pipeline.
func run(in io.Reader, out io.Writer, ld *lang.LangDef, fontHint, commentFont, srcFile, headerContent, footerContent string, lineNumbers bool, tabWidth int, engine, paper string, explicitFont, explicitCommentFont bool) error {
	if ld == nil {
		return fmt.Errorf("unknown language — use -lang to specify (c, go, sh, ...)")
	}

	src, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	tokens := scanner.Scan(src, ld)

	opts := emitter.Options{
		SourceFile:          srcFile,
		CodeFont:            fontHint,
		CommentFont:         commentFont,
		ExplicitCodeFont:    explicitFont,
		ExplicitCommentFont: explicitCommentFont,
		HeaderContent:       headerContent,
		FooterContent:       footerContent,
		LineNumbers:         lineNumbers,
		TabWidth:            tabWidth,
		Engine:              engine,
		PaperSize:           paper,
	}
	var buf bytes.Buffer
	emitter.WritePreamble(&buf, opts)
	emitter.WriteBody(&buf, tokens, ld, opts)
	emitter.WritePostamble(&buf)

	processed := postProcessQuoteBlocks(buf.String())
	_, err = fmt.Fprint(out, processed)
	return err
}

// postProcessQuoteBlocks fixes \begin{quote}...\end{quote} blocks inside hash
// comments so that the # marker stays at the left margin while only the text
// is indented with \hspace{\leftmargini}.
//
// In v2.236 output, TeX-passthrough comment lines look like:
//
//	\mbox{}\rm\mc {\tt\#} \ \begin{quote}
//	\mbox{}\rm\mc {\tt\#} \ text content
//	\mbox{}\rm\mc {\tt\#} \ \end{quote}
//
// We remove \begin{quote} / \end{quote} from those lines and prepend
// \hspace{\leftmargini} to text lines inside the block instead.
var (
	// commentMarkerPat matches comment markers: {\\tt\\#}, {\\tt ;}, {\\tt\\%}
	commentMarkerPat = `\{\\tt(?:\\#| ;|\\%)\}`

	quoteStartRe  = regexp.MustCompile(`(` + commentMarkerPat + `[^\n]*)\\ \\begin\{quote\}`)
	quoteEndRe    = regexp.MustCompile(`(` + commentMarkerPat + `[^\n]*)\\ \\end\{quote\}`)
	commentTextRe = regexp.MustCompile(`(` + commentMarkerPat + ` \\ )`)
)

func postProcessQuoteBlocks(texContent string) string {
	lines := strings.Split(texContent, "\n")
	inQuote := false

	for i, line := range lines {
		switch {
		case quoteStartRe.MatchString(line):
			lines[i] = quoteStartRe.ReplaceAllString(line, "$1")
			inQuote = true
		case quoteEndRe.MatchString(line):
			lines[i] = quoteEndRe.ReplaceAllString(line, "$1")
			inQuote = false
		case inQuote && commentTextRe.MatchString(line):
			lines[i] = commentTextRe.ReplaceAllString(line, `$1\hspace{\leftmargini}`)
		}
	}

	return strings.Join(lines, "\n")
}

// readOptionalFile reads the file at path and returns its contents as a string.
// If path is empty, it returns ("", nil). If the file does not exist or cannot
// be read, it returns an error with the flag name for context.
func readOptionalFile(path, flagName string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("-%s %s: %w", flagName, path, err)
	}
	return string(data), nil
}

// resolveByFlag looks up by flag name first, then by extension.
// Use when the user explicitly provides a -lang value.
func resolveByFlag(hint string) *lang.LangDef {
	if hint == "" {
		return nil
	}
	if ld := lang.FindByFlag(hint); ld != nil {
		return ld
	}
	return lang.FindByExt(hint)
}

// resolveByExt looks up by file extension first, then by flag name.
// Use when auto-detecting language from a file's extension.
func resolveByExt(ext string) *lang.LangDef {
	if ext == "" {
		return nil
	}
	if ld := lang.FindByExt(ext); ld != nil {
		return ld
	}
	return lang.FindByFlag(ext)
}

// addExt returns path with newExt appended (e.g. "foo.sh" -> "foo.sh.tex").
func addExt(path, newExt string) string {
	return path + newExt
}

// shortPath returns the last two path elements (parent/file) of p.
func shortPath(p string) string {
	dir, file := filepath.Split(filepath.Clean(p))
	parent := filepath.Base(filepath.Clean(dir))
	if parent == "." || parent == "/" {
		return file
	}
	return parent + "/" + file
}

// extWithoutDot returns the file extension of path without the leading dot.
func extWithoutDot(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: src2tex [options] [file]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nSupported -lang values:\n")
	fmt.Fprintf(os.Stderr, "  C family:    c, go, java, js, ts, rust, kotlin, swift, cpp, csharp, dart\n")
	fmt.Fprintf(os.Stderr, "  Hash:        sh, python, ruby, perl, make, tcl\n")
	fmt.Fprintf(os.Stderr, "  Percent:     reduce, matlab\n")
	fmt.Fprintf(os.Stderr, "  Semicolon:   lisp\n")
	fmt.Fprintf(os.Stderr, "  Pascal:      pascal\n")
	fmt.Fprintf(os.Stderr, "  Markup:      xml, css, html\n")
	fmt.Fprintf(os.Stderr, "\nAuto-detected file names: Makefile, GNUmakefile\n")
	fmt.Fprintf(os.Stderr, "\nFont subcommand:\n")
	fmt.Fprintf(os.Stderr, "  src2tex font list\n")
	fmt.Fprintf(os.Stderr, "  src2tex font install <name|all>\n")
	fmt.Fprintf(os.Stderr, "  src2tex font install-comment <name|all>\n")
	fmt.Fprintf(os.Stderr, "\nEngine subcommand:\n")
	fmt.Fprintf(os.Stderr, "  src2tex engine list\n")
	fmt.Fprintf(os.Stderr, "  src2tex engine init\n")
	fmt.Fprintf(os.Stderr, "\nSupported -engine values: xelatex, lualatex, uplatex, pdflatex, tectonic\n")
	fmt.Fprintf(os.Stderr, "\nIf no file is given, reads from stdin and writes to stdout.\n")
}
