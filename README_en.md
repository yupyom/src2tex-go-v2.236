# src2tex-go <small>v2.236</small>

A CLI tool that converts program source code into LaTeX documents for XeLaTeX / LuaLaTeX / upLaTeX / pdfLaTeX. You can write TeX math formulas inside comments, so your working source code becomes a ready-to-print document. For example, you can add explanations to sample code and immediately produce handout materials for students.

This is a complete rewrite in Go, based on version 2.12 of `src2tex` — a C program originally developed by Prof. Kazuo Amano (Mathematics, Josai University) in 1992.

The version number follows the decimal expansion of √5, approaching the true value with each release. The current version is **2.236**.

The previous version aimed to stay close to the original behavior. This version updates the architecture for modern LaTeX, targeting current TeX engines. See ARCHITECTURE_en.md and NOTE_en.md for details.

## Requirements

| Tool | Purpose |
|------|---------|
| [XeLaTeX](https://tug.org/xetex/) etc.$^1$ (TeX Live / MacTeX) | PDF generation |
| [Ghostscript](https://www.ghostscript.com/) (`gs`) | EPS → PDF conversion |

<small>1: LuaLaTeX, upLaTeX, and pdfTeX (for ASCII-only documents) are also supported. Plain TeX support was available in the [previous compatible version](https://github.com/yupyom/src2tex-go), but has been dropped here since the author no longer uses plain TeX.</small>

## Installation

Requires Go 1.21 or later.

```bash
go install github.com/yupyom/src2tex-go-v2.236/cmd/src2tex@latest
```

Or clone and build:

```bash
git clone <repo-url>
cd src2tex-go-v2.236
go build -o src2tex ./cmd/src2tex/
```

## Basic Usage

```bash
# Convert a Go file (produces hanoi.go.tex)
src2tex hanoi.go

# Specify language explicitly
src2tex -lang reduce popgen.red

# stdin → stdout
cat hanoi.go | src2tex -lang go

# Generate PDF (run from testdata/input/)
cd testdata/input
xelatex -halt-on-error -interaction=nonstopmode hanoi.go.tex
```

The output filename is `<input-filename>.tex` — the original extension is kept and `.tex` is appended.

### Supported Languages

| Category | Language | Extensions | `-lang` value | Bold keywords | Docstring |
|----------|----------|-----------|---------------|:---:|:---:|
| C-family | C | `.c`, `.h` | `c` | ✅ | — |
| | Go | `.go` | `go` | ✅ | — |
| | Java | `.java` | `java` | ✅ | — |
| | C++ | `.cpp`, `.cc`, `.cxx`, `.hpp` | `cpp` | ✅ | — |
| | C# | `.cs` | `csharp` | ✅ | — |
| | Dart | `.dart` | `dart` | ✅ | — |
| | JavaScript | `.js`, `.mjs` | `js` | ✅ | — |
| | TypeScript | `.ts`, `.tsx` | `ts` | ✅ | — |
| | Rust | `.rs` | `rust` | ✅ | — |
| | Kotlin | `.kt`, `.kts` | `kotlin` | ✅ | — |
| | Swift | `.swift` | `swift` | ✅ | — |
| Hash-style | Shell | `.sh`, `.bash` | `sh` | ✅ | — |
| | Python | `.py` | `python` | ✅ | ✅ (`"""` / `'''`) |
| | Ruby | `.rb` | `ruby` | ✅ | — |
| | Perl | `.pl`, `.pm` | `perl` | ✅ | — |
| | Makefile$^2$ | `.mk`, `Makefile`* | `make` | — | — |
| | Tcl | `.tcl` | `tcl` | ✅ | — |
| Percent-style | REDUCE | `.red` | `reduce` | ✅ | — |
| | MATLAB/Octave | `.m` | `matlab` | — | — |
| Semicolon-style | Lisp/Scheme | `.lisp`, `.scm`, `.el` | `lisp` | ✅ | — |
| Pascal-style | Pascal | `.pas`, `.p` | `pascal` | ✅ | — |
| XML-family | XML | `.xml`, `.xsl`, `.xslt`, `.svg`, `.xhtml` | `xml` | — | — |
| CSS | CSS | `.css` | `css` | — | — |
| HTML-family | HTML | `.html`, `.htm`, `.php` | `html` | ✅ (tag names) | — |

<small>2: `Makefile`, `makefile`, and `GNUmakefile` are detected by filename even without an extension.</small>


### Command-Line Options

| Option | Description |
|--------|-------------|
| `-lang <name>` | Specify language explicitly (when extension is ambiguous) |
| `-o <file>` | Output file path (default: `<input>.tex`) |
| `-font <name>` | Monospace font for code (sets `\setmonofont`; default: `CMU Typewriter Text`, or `Courier New` if TeX Live is not detected) |
| `-commentfont <name>` | CJK serif font for comments (default: auto-detected) |
| `-fontdir <path>` | Font install directory (default: `~/.src2tex/fonts/`) |
| `-listfonts` | List available fonts and exit |
| `-header <file>` | Replace `\fancyhead[R]{...}` with file contents |
| `-footer <file>` | Replace `\fancyfoot[R]{...}` with file contents |
| `-linenumbers` | Show line numbers in the left margin |
| `-tab <n>` | Tab width in spaces (default: 8) |
| `-engine <name>` | TeX engine: `xelatex`, `lualatex`, `uplatex`, `pdflatex`, `tectonic` (default: `xelatex`) |
| `-paper <size>` | Paper size: `a4`, `b5`, `letter` (default: `a4`) |


## Preamble Templates

Running `src2tex engine init` extracts the preamble templates to your user directory at `~/.src2tex/engines/`. You can customize them to fit your project.

### Supported Engines

| Engine | `-engine` value | CJK | fontspec | Notes |
|--------|----------------|:---:|:--------:|-------|
| XeLaTeX | `xelatex` (default) | ✅ | ✅ | fontspec + xeCJK |
| Tectonic | `tectonic` | ✅ | ✅ | XeTeX-compatible, fontspec support |
| LuaLaTeX | `lualatex` | ✅ | ✅ | luatexja-fontspec |
| upLaTeX | `uplatex` | ✅ | ❌ | jsarticle + otf, dvipdfmx pipeline |
| pdfLaTeX | `pdflatex` | ✅ | ❌ | bxjsarticle + bxcjkjatype (`ja=standard`) |

### Examples

```bash
# Export templates to ~/.src2tex/engines/
src2tex engine init

# List available engines
src2tex engine list

# Convert an ASCII-only source with pdfLaTeX
src2tex -engine pdflatex -paper letter hello.c

# Convert a Japanese source with LuaLaTeX
src2tex -engine lualatex report.go

# Use upLaTeX + dvipdfmx (manual .dvi → .pdf step required)
src2tex -engine uplatex report.go
uplatex report.go.tex && dvipdfmx report.go.dvi

# Convert with B5 paper size
src2tex -paper b5 report.go
```

Templates are at `~/.src2tex/engines/<engine>/preamble.tmpl`, written in Go `text/template` syntax with `<% %>` delimiters.

### Template Protection (the `auto` flag)

When you run `engine init`, each engine's `engine.json` gets `"auto": true` set automatically. This flag controls whether `engine init` overwrites your customizations.

| `auto` value in `engine.json` | Behavior on `engine init` |
|---|---|
| `true` | **Overwritten** with the built-in template (kept up to date) |
| `false` or absent | **Skipped** (your edits are preserved) |

To customize a built-in engine template, remove `"auto": true` (or set it to `false`) from `engine.json`, then edit `preamble.tmpl`.

### Creating Custom Templates

You can copy an existing engine config to create your own:

```bash
# Create a custom template based on LuaLaTeX
cp -r ~/.src2tex/engines/lualatex ~/.src2tex/engines/my-lualatex

# Edit engine.json (remove "auto", change the name)
vi ~/.src2tex/engines/my-lualatex/engine.json

# Customize preamble.tmpl freely
vi ~/.src2tex/engines/my-lualatex/preamble.tmpl

# Use your custom engine
src2tex -engine my-lualatex report.go
```

Custom engines appear in `engine list` with a `[custom]` label:
```
Available engines:
  lualatex     Unicode TeX engine with luatexja (CJK support)
  xelatex      Unicode TeX engine with fontspec + xeCJK (default)
  my-lualatex  My customized LuaLaTeX [custom]
```

> **Note**: Do **not** add `"auto": true` to user-created custom engines. If you do, `engine init` may overwrite them with a built-in template.

### Verifying Engine Output

To check that the generated LaTeX compiles correctly in your environment, install [Task](https://taskfile.dev/) and run:

```bash
# Verify all engines
task verify:all

# Verify individual engines
task verify:xelatex
task verify:pdflatex
task verify:lualatex
task verify:uplatex
task verify:tectonic
```

## Font Management

With fontspec-compatible engines, you can customize fonts for both code and comments. The recommended fonts are monospace fonts that align full-width and half-width characters, which works well for space-aligned source code. Built-in utility commands make it easy to install recommended fonts, but you can also use any font you have installed yourself.

```bash
# List available fonts
src2tex font list

# Install a font (downloaded from GitHub)
src2tex font install hackgen
src2tex font install all

# Install a comment font
src2tex font install-comment haranoaji
src2tex font install-comment all

# Convert using installed fonts
src2tex -font hackgen -commentfont haranoaji hanoi.go
```

Fonts are installed to `~/.src2tex/fonts/`. Use `-fontdir` to change this.

### Font Config File (`~/.src2tex/fonts.json`)

`~/.src2tex/fonts.json` is created automatically when you install fonts or run `src2tex font init`. It lists all available fonts and their paths.

```bash
# Generate/update fonts.json
src2tex font init
```

#### File Structure

```json
{
  "_comment": "src2tex font configuration. Entries with \"auto\": true are system-managed and regenerated on init. Remove \"auto\" or set it to false to prevent overwriting.",
  "code_fonts": [
    {
      "name": "hackgen",
      "display_name": "HackGen",
      "license": "SIL OFL",
      "unified": true,
      "regular_file": "HackGen-Regular.ttf",
      "bold_file": "HackGen-Bold.ttf",
      "font_dir": "/Users/xxx/.src2tex/fonts/hackgen",
      "installed": true,
      "auto": true,
      "description": "Hack + GenJyuuGothic. A popular programming font."
    }
  ],
  "comment_fonts": [
    {
      "name": "haranoaji",
      "display_name": "Harano Aji Mincho",
      "license": "SIL OFL",
      "regular_file": "HaranoAjiMincho-Regular.otf",
      "bold_file": "HaranoAjiMincho-Bold.otf",
      "font_dir": "/usr/local/texlive/2026/texmf-dist/fonts/opentype/public/haranoaji",
      "texlive": true,
      "installed": true,
      "auto": true,
      "description": "High-quality Mincho font bundled with TeX Live. No download needed."
    }
  ]
}
```

#### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | **Required.** The name used with `-font` / `-commentfont` |
| `display_name` | string | Display name shown in `font list` |
| `license` | string | License type (informational) |
| `unified` | bool | If `true`, the font covers both Latin and CJK glyphs; `\setCJKmonofont` uses the same font |
| `regular_file` | string | **Required.** Regular-weight font filename |
| `bold_file` | string | Bold-weight font filename. If absent, no bold variant is set |
| `font_dir` | string | Absolute path to the directory containing the font files. If absent, defaults to `~/.src2tex/fonts/{name}/` |
| `texlive` | bool | If `true`, the font is bundled with TeX Live (no download needed) |
| `installed` | bool | Auto-detected install status. No need to edit manually |
| `auto` | bool | **Important.** If `true`, this is a system-managed entry and will be regenerated from built-in definitions on `font init`. If `false` or absent, user edits are preserved |
| `description` | string | Short description (informational) |

#### How the `auto` Flag Works

| `auto` value | Behavior on `font init` |
|---|---|
| `true` | Regenerated from built-in definition (`font_dir`, `installed`, etc. refreshed) |
| `false` or absent | **Not overwritten**, even if the name matches a built-in font |

To customize a built-in font entry, remove `"auto": true` (or set it to `false`), then edit `font_dir` or other fields as needed.

#### Path Resolution

The font path is resolved from `font_dir` and `regular_file` combined:

1. **If `font_dir` is set** → `{font_dir}/{regular_file}` is used as the font path
2. **If `font_dir` is absent** → `~/.src2tex/fonts/{name}/{regular_file}` is used
3. **If `regular_file` contains a path separator (`/`)** → its `dirname` part is used as the font directory (takes precedence over `font_dir`)

Example of the generated LaTeX command:
```latex
\setmonofont[Path=/Users/xxx/.src2tex/fonts/hackgen/, BoldFont=HackGen-Bold.ttf]{HackGen-Regular.ttf}
```

#### Adding a Custom Font

You can use any font installed on your system or in any directory. Add a new entry to the `code_fonts` array:

```json
{
  "name": "myricam",
  "display_name": "MyricaM",
  "license": "SIL OFL",
  "unified": true,
  "regular_file": "MyricaM-Regular.ttf",
  "bold_file": "MyricaM-Bold.ttf",
  "font_dir": "/Library/Fonts",
  "installed": true
}
```

After adding it, use `-font myricam` on the command line. Running `src2tex font init` again will not overwrite entries that have no `"auto"` key (or have `"auto": false`).

> **Note**: Do **not** add `"auto": true` to entries you create manually. If you do, the next `font init` may overwrite them with built-in definitions.

> **Note**: Font selection only works with fontspec-compatible engines (XeLaTeX, LuaLaTeX, Tectonic). The `-font` option is ignored for upLaTeX and pdfLaTeX.

### Proxy Settings for Font Downloads

If you are behind a corporate proxy, set these environment variables:

```bash
# Linux / macOS
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
```

```powershell
# Windows (PowerShell)
$env:HTTP_PROXY = "http://proxy.example.com:8080"
$env:HTTPS_PROXY = "http://proxy.example.com:8080"
```

```bat
:: Windows (cmd)
set HTTP_PROXY=http://proxy.example.com:8080
set HTTPS_PROXY=http://proxy.example.com:8080
```


## TeX Comment Syntax

Any block starting with `{\` inside a comment is output directly as a LaTeX command.

### Inline Math

Embed TeX commands at the end of a comment:

```c
/* radius $r = \sqrt{x^2 + y^2}$ */
```

```python
# integral $\int_0^1 f(x)\,dx$
```

```pascal
{ Euclid's algorithm {\ \hfill --- greatest common divisor} }
```

### `{\ }` Syntax (Inline TeX Block)

Writing `{\ ... }` inside a comment outputs the block as raw LaTeX:

```python
# {\ some text \hfill}
# {\ \hrulefill}
# {\ \begin{quote} indented text \end{quote} }
```

### `{\null ... }` Syntax (Multi-line TeX Block)

A block starting with `{\null` can span multiple lines:

```python
# {\null
# \begin{eqnarray}
#   f(x) &=& x^2 + 1 \\
#   g(x) &=& \sqrt{x}
# \end{eqnarray}
# }
```

In Pascal, `{ }` is also the block comment syntax, so `{\ ... }` works as a passthrough directly.

### Python Docstrings (Triple Quotes)

Python's `"""..."""` (or `'''...'''`) is treated as a block comment when it appears at the start of a line. You can use `{\ ... }` syntax inside:

```python
"""
{\ \centerline{\bf Module Overview} }
This module solves the Towers of Hanoi.
"""

def hanoi(n, src, dst, tmp):
    '''Move n discs.'''
```

Triple quotes that are not at the start of a line are treated as string literals, not comments.

## Custom Headers and Footers

```bash
# Replace the header
printf '\\fancyhead[L]{\\rm My Project}\n' > myheader.tex
src2tex -header myheader.tex hanoi.go

# Replace the footer
printf '\\fancyfoot[C]{\\thepage}\n' > myfooter.tex
src2tex -footer myfooter.tex hanoi.go
```

Write `\fancyhead` / `\fancyfoot` commands directly in the file.
`\usepackage{fancyhdr}` and `\pagestyle{fancy}` are already included in the preamble automatically.


## License

Distributed under the MIT License, as-is, with no warranty. Use freely.

The original `src2tex` version 2.12 was created by Kazuo AMANO and Shinichi NOMOTO, and was previously released as free software.
