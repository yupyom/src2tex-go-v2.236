# Miscellaneous Notes — On the Design of src2tex-go

This document records the background behind the design of src2tex-go.

## How It Started

The story of the original src2tex goes back to my encounter with Prof. Amano 30 years ago. I had gone to ask him how to read his lecture notes, and that was the beginning. I was a Macintosh user at the time and had no experience with TeX, so it felt fresh and exciting.

Even then, I wanted to port it so I could run it on my Mac, and the professor kindly shared the source code with me — but I never managed to get it done. Years passed, and one day while building a small program, I started wondering how to document it, and the memory came back.

By then, Prof. Amano had long since retired, and I had lost his contact information. The old lab website was gone too. I eventually tracked down the source from an old Linux archive and started porting it from there.

## The Problem

The original program was written a long time ago — in the old Kernighan & Ritchie style of C — and its preamble was written for the TeX conventions of that era, with EUC-JP encoding assumed. Using it in a modern environment took some effort.

From what I remember, the professor used Solaris and Linux back then. C wasn't impossible to compile, but setting up the environment was sometimes a hassle (though far easier than it used to be).

So I decided to port it to Go for better portability, and to separate the LaTeX preamble into external configuration files.

## Why Not Use a Parser Generator like goyacc?

Go has goyacc, as well as ANTLR and tree-sitter bindings. But src2tex-go uses a hand-written scanner and emitter. There are several reasons, but the most fundamental one is: **the task does not match what parser generators are good at**.

### The Goal Is Classification, Not Parsing

Parser generators like goyacc, ANTLR, and tree-sitter are designed to **build an AST (abstract syntax tree)**. In other words, they understand `if (x > 0) { return x; }` structurally — recognizing the condition, the block, the return statement.

But src2tex doesn't need a syntax tree. It just needs to **classify the input into five categories**:

| TokenKind | Purpose |
|-----------|---------|
| `TokenCode` | Output as-is in monospace font |
| `TokenKeyword` | Output in bold |
| `TokenLineComment` | Output in serif font + TeX passthrough |
| `TokenBlockComment` | Same as above |
| `TokenTeX` | Output verbatim |

Statement structure, operator precedence, scopes — none of that is needed. This is a **subset of lexical analysis**, not even full parsing.

### The Cost of Handling over 25 Languages with One Grammar

If we used goyacc, we would need a `.y` grammar file for each language. Writing grammars for 25 languages would be an **enormous amount of work**. And all src2tex actually needs from each language is: the comment start/end tokens, the keyword list, and the string literal delimiters.

The table-driven design (`LangDef` struct) lets a single scanner handle all languages simply by **swapping the data**. Using goyacc here would make things harder, not easier.

### The State Machine Is Simple

The scanner is essentially a **finite automaton** — `stateCode → stateLineComment → stateCode`, and so on. About 500 lines of code handle string literal skipping, the context stack for composite languages (HTML → CSS/JS/PHP switching), and nestable block comments (Pascal's `{ }`).

Bringing in goyacc wouldn't simplify any of this. It would only add **abstraction overhead**.

### Similarity to the Original

The original src2tex from the 1990s also used a flag-based hand-written state machine. In porting to Go, I modernized the approach from "flag management" to an "explicit state machine," but I chose to **keep the fundamental processing model intact**. Introducing a parser generator would have made it harder to debug subtle differences in the output.

## Key Points as a Transformation Tool

src2tex is essentially a **transformation pipeline** — source code → LaTeX. It belongs to the same category as HTML rendering engines and Markdown → HTML converters.

Here are the main design points from that perspective.

### Two-Stage Pipeline (Scanner → Emitter)

```
src → [Scanner] → []Token → [Emitter] → LaTeX (.tex)
```

This is the same design as markdown-it's `parse() → render()`. What matters is that **the token list acts as the interface**. The scanner decides what is a comment; the emitter focuses only on how to convert comments into LaTeX.

This separation means that adding an HTML emitter in the future, for example, would require no changes to the scanner at all.

### TeX Special Character Escaping

Just as `<` → `&lt;` is the most important escaping step in HTML rendering, **TeX special character escaping** is the heart of src2tex's conversion.

```
// Code context (monospace font)
' ' → {\tt\mc \ }
'#' → {\tt\#}
'{' → {\tt\char'173}

// Comment context (serif font)
'_' → \_
'$' → \$
```

**The same character is escaped differently depending on context (code vs. comment).** This is why token classification is necessary, and why simple regex substitution falls short.

### Line-Oriented Rendering

TeX has a **weak notion of lines**. To enforce source code's line structure in TeX, src2tex uses a distinctive output pattern:

```tex
\noindent
\mbox{}{\tt\mc \ }{\tt\mc \ }{\tt f}{\tt o}{\tt r}  % indented code line
```

The `\noindent` + `\mbox{}` combination treats each line as an independent paragraph, with `\hfill` representing blank lines. **Bridging the impedance mismatch between TeX's paragraph model and source code's line model** was the biggest design challenge.

### Font Mode Switching

The reason src2tex output looks good is the clear separation: **code in monospace (`\tt`), comments in serif (`\rm`)**. Even the comment markers (`//`, `#`) are rendered in monospace with `{\tt /}`. This fine-grained font control greatly improves the readability of the final PDF.

### TeX Passthrough

```c
/* {\centerline{\epsfbox{figure.eps}}} */
```

The mechanism that outputs `{\ ... }` in comments directly as TeX goes beyond a simple syntax converter — it allows **math, figures, and arbitrary TeX commands to be embedded in source code documentation**. This is the most original design decision in the original src2tex, and it has been faithfully preserved in the Go port.

Note: In the original version, not only `{\ }` markers but also `$...$` in comments were implicitly treated as TeX passthrough. This worked by pre-scanning the entire comment buffer at the start of each comment; if a `$` was found, the whole comment was switched to TeX mode (`RMFlag=1`) — a coarse approach.

In v2.236, this pre-scan was not carried over. Instead, only the `$...$` range is passed through precisely. Characters outside math in a comment are still escaped by `commentTextEscape`, while only the content inside `$...$` is output as raw TeX math. This is a more predictable design that fits naturally with the token-based architecture.

## Summary

Looking back at all of this, the design ends up closely following the spirit of the original.

| Aspect | Design decision |
|--------|----------------|
| Parser choice | Lexical classification is sufficient → hand-written state machine + table-driven |
| Multi-language support | 25 languages unified via data (`LangDef`), not grammar rules |
| Core of conversion | Context-dependent TeX escaping |
| Biggest challenge | Bridging TeX's paragraph model and source code's line model |
| Unique strength | TeX passthrough inside comments for rich expression |

In short: by recognizing that the task is **"classify surface-level lexical patterns and convert them to context-appropriate TeX output"** — not "understand syntax" — the lightweight scanner + emitter architecture turned out to be the right answer, rather than a parser generator.

I hope this is useful to someone.

## Extras

### Tower of Hanoi

The `testdata/input` directory includes Tower of Hanoi programs in 20 different languages, based on the C version that came with the original src2tex. They all actually run. The compilers and tools you need will vary by environment, but feel free to update `CompileTask.yml` to match your setup and give them a try. They solve the puzzle in $2^8 - 1 = 255$ moves.
