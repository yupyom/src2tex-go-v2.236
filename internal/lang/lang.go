package lang

// CommentStyle describes how a language delimits comments.
type CommentStyle struct {
	LineComment         string // e.g. "//" or "#" or "%" — empty if none
	BlockOpen           string // e.g. "/*" — empty if none
	BlockClose          string // e.g. "*/" — empty if none
	AltBlockOpen        string // alternative block open, e.g. "(*" for Pascal
	AltBlockClose       string // alternative block close, e.g. "*)" for Pascal
	BlockNestable       bool   // true if block comments can be nested (e.g. Pascal { })
	RawTeX              bool   // true if comment content is raw LaTeX (no TeX special char escaping)
	DocstringDelimiter  string // e.g. `"""` — triple-quote treated as block comment when at line start
}

// SubLanguageRule defines an embedded language region within a composite language.
// For example, HTML contains CSS inside <style> and JS inside <script>.
type SubLanguageRule struct {
	OpenTag             string // opening tag prefix, case-insensitive (e.g. "<style")
	CloseTag            string // closing tag, case-insensitive (e.g. "</style>")
	LangFlag            string // Flag value of the sub-language's LangDef (e.g. "css")
	ImmediateActivation bool   // if true, context switches right after OpenTag (no '>' scan needed, e.g. "<?php")
}

// LangDef is a table-driven language definition.
// Add a new language by appending an entry to Languages — no procedural code needed.
type LangDef struct {
	Name         string   // display name, e.g. "C", "Go", "Shell"
	Exts         []string // file extensions without dot, e.g. ["c", "h"]
	Flag         string   // -lang flag value, e.g. "c", "go", "sh"
	Comment      CommentStyle
	Keywords     []string          // reserved words rendered in bold; nil means no bolding
	SubLanguages []SubLanguageRule // embedded language regions (nil for simple languages)
	BoldTags     bool              // if true, <tagname> tag names are emitted as TokenKeyword
	FileNames    []string          // exact file names to match (e.g. "Makefile", "GNUmakefile")
}

// cStyle is the comment style shared by C-family languages (// and /* */).
var cStyle = CommentStyle{LineComment: "//", BlockOpen: "/*", BlockClose: "*/"}

// hashStyle is the comment style for hash-comment languages (#).
var hashStyle = CommentStyle{LineComment: "#"}

// percentStyle is the comment style for percent-comment languages (%).
var percentStyle = CommentStyle{LineComment: "%"}

// semicolonStyle is the comment style for semicolon-comment languages (;).
var semicolonStyle = CommentStyle{LineComment: ";"}

// xmlCommentStyle is the comment style for XML/HTML (<!-- -->).
var xmlCommentStyle = CommentStyle{BlockOpen: "<!--", BlockClose: "-->"}

// cssCommentStyle is the comment style for CSS (/* */ with no line comment).
var cssCommentStyle = CommentStyle{BlockOpen: "/*", BlockClose: "*/"}

// phpCommentStyle is the comment style for PHP (// and /* */).
var phpCommentStyle = CommentStyle{LineComment: "//", BlockOpen: "/*", BlockClose: "*/"}

var goKeywords = []string{
	"break", "case", "chan", "const", "continue",
	"default", "defer", "else", "fallthrough", "for",
	"func", "go", "goto", "if", "import",
	"interface", "map", "package", "range", "return",
	"select", "struct", "switch", "type", "var",
	// predeclared identifiers
	"make", "nil", "true", "false",
}

var cKeywords = []string{
	"if", "else", "while", "for", "do",
	"switch", "case", "default", "break", "continue",
	"return", "goto",
	"void", "int", "char", "float", "double",
	"long", "short", "unsigned", "signed",
	"struct", "typedef", "enum", "union",
	"const", "static", "extern", "auto", "register",
}

var pascalKeywords = []string{
	"program", "begin", "end", "var", "const", "type",
	"procedure", "function",
	"if", "then", "else", "while", "do",
	"for", "to", "downto", "repeat", "until",
	"array", "of", "record", "set", "with",
	"integer", "real", "boolean", "char", "string",
	"true", "false", "nil",
	"and", "or", "not", "div", "mod",
	"in", "is",
}

var javaKeywords = []string{
	"class", "interface", "extends", "implements", "new", "this", "super",
	"public", "private", "protected", "static", "final", "abstract",
	"void", "int", "boolean", "String", "return", "if", "else", "for", "while",
	"do", "switch", "case", "default", "break", "continue", "try", "catch",
	"finally", "throw", "throws", "import", "package", "null", "true", "false",
}

var jsKeywords = []string{
	"function", "var", "let", "const", "if", "else", "for", "while", "do",
	"switch", "case", "default", "break", "continue", "return", "new", "this",
	"class", "extends", "import", "export", "from", "async", "await",
	"try", "catch", "finally", "throw", "typeof", "instanceof",
	"null", "undefined", "true", "false",
}

var tsKeywords = []string{
	"function", "var", "let", "const", "if", "else", "for", "while", "do",
	"switch", "case", "default", "break", "continue", "return", "new", "this",
	"class", "extends", "import", "export", "from", "async", "await",
	"try", "catch", "finally", "throw", "typeof", "instanceof",
	"null", "undefined", "true", "false",
	"type", "interface", "enum", "namespace", "declare", "readonly", "abstract",
	"implements", "keyof", "as",
}

var rustKeywords = []string{
	"fn", "let", "mut", "const", "if", "else", "match", "for", "while", "loop",
	"return", "break", "continue", "struct", "enum", "impl", "trait",
	"pub", "use", "mod", "crate", "self", "super", "where", "type", "as",
	"async", "await", "move", "unsafe", "dyn", "ref", "true", "false",
}

var kotlinKeywords = []string{
	"fun", "val", "var", "if", "else", "when", "for", "while", "do",
	"return", "break", "continue", "class", "object", "interface",
	"data", "sealed", "open", "abstract", "override", "private", "public",
	"internal", "protected", "companion", "import", "package",
	"is", "as", "in", "null", "true", "false", "this", "super", "throw", "try", "catch",
}

var swiftKeywords = []string{
	"func", "var", "let", "if", "else", "for", "while", "repeat", "switch", "case",
	"default", "break", "continue", "return", "class", "struct", "enum",
	"protocol", "extension", "import", "public", "private", "internal",
	"open", "static", "self", "super", "nil", "true", "false",
	"guard", "defer", "throw", "try", "catch", "do", "where", "as", "is", "in",
}

var pythonKeywords = []string{
	"def", "class", "if", "elif", "else", "for", "while", "break", "continue",
	"return", "pass", "import", "from", "as", "try", "except", "finally",
	"raise", "with", "yield", "lambda", "global", "nonlocal",
	"and", "or", "not", "is", "in", "True", "False", "None",
}

var rubyKeywords = []string{
	"def", "class", "module", "if", "elsif", "else", "unless", "case", "when",
	"for", "while", "until", "do", "begin", "end", "rescue", "ensure", "raise",
	"return", "break", "next", "yield", "require", "include", "extend",
	"attr_reader", "attr_writer", "attr_accessor",
	"nil", "true", "false", "self", "super", "and", "or", "not",
}

var perlKeywords = []string{
	"sub", "my", "our", "local", "if", "elsif", "else", "unless", "while", "until",
	"for", "foreach", "do", "return", "last", "next", "redo",
	"use", "require", "package", "die", "warn", "print", "chomp",
	"and", "or", "not", "eq", "ne", "lt", "gt", "le", "ge",
}

var shellKeywords = []string{
	"if", "then", "elif", "else", "fi", "for", "while", "until", "do", "done",
	"case", "esac", "in", "function", "return", "exit", "local", "export",
	"echo", "read", "set", "unset", "test", "true", "false",
}

var tclKeywords = []string{
	"proc", "if", "elseif", "else", "while", "for", "foreach", "switch",
	"return", "break", "continue", "set", "puts", "expr", "catch", "try",
	"namespace", "package", "variable", "global", "upvar",
}

var lispKeywords = []string{
	"define", "lambda", "if", "cond", "else", "let", "letrec",
	"begin", "set!", "car", "cdr", "cons", "null?", "list", "map", "apply",
	"defun", "defmacro", "quote", "and", "or", "not",
}

var reduceKeywords = []string{
	"procedure", "begin", "end", "if", "then", "else",
	"while", "do", "for", "return", "on", "off",
	"or", "and", "not", "neq",
}

var phpKeywords = []string{
	"echo", "print", "if", "else", "elseif", "while", "for", "foreach",
	"do", "switch", "case", "default", "break", "continue", "return",
	"function", "class", "new", "extends", "implements", "interface", "trait",
	"public", "private", "protected", "static", "abstract", "final",
	"namespace", "use", "require", "include", "require_once", "include_once",
	"null", "true", "false", "array",
}

var cppKeywords = []string{
	// C keywords
	"if", "else", "while", "for", "do",
	"switch", "case", "default", "break", "continue",
	"return", "goto",
	"void", "int", "char", "float", "double",
	"long", "short", "unsigned", "signed",
	"struct", "typedef", "enum", "union",
	"const", "static", "extern", "auto", "register",
	// C++ additions
	"class", "public", "private", "protected", "virtual", "override",
	"namespace", "using", "template", "typename",
	"new", "delete", "this", "throw", "try", "catch",
	"bool", "true", "false", "nullptr",
	"inline", "explicit", "constexpr", "noexcept",
	"operator", "friend", "mutable",
	"dynamic_cast", "static_cast", "reinterpret_cast", "const_cast",
	"std", "cout", "cin", "endl", "string", "vector",
	"include",
}

var csharpKeywords = []string{
	"using", "namespace", "class", "struct", "interface", "enum",
	"public", "private", "protected", "internal", "static", "abstract", "sealed", "override", "virtual",
	"void", "int", "bool", "string", "double", "float", "char", "long", "byte", "decimal", "object", "var",
	"if", "else", "for", "foreach", "while", "do", "switch", "case", "default",
	"break", "continue", "return", "throw", "try", "catch", "finally",
	"new", "this", "base", "null", "true", "false",
	"async", "await", "get", "set", "value",
	"readonly", "const", "ref", "out", "in", "params",
}

var dartKeywords = []string{
	"import", "library", "part", "class", "extends", "implements", "with", "mixin",
	"abstract", "factory", "const", "final", "var", "void", "dynamic",
	"int", "double", "bool", "String", "List", "Map", "Set",
	"if", "else", "for", "while", "do", "switch", "case", "default",
	"break", "continue", "return", "throw", "try", "catch", "finally", "rethrow",
	"new", "this", "super", "null", "true", "false",
	"async", "await", "yield", "sync",
	"static", "late", "required", "typedef", "enum",
	"is", "as", "in",
}

// Languages is the master table of supported languages.
var Languages = []LangDef{
	// ---- C family (// and /* */) ----
	{Name: "C", Exts: []string{"c", "h"}, Flag: "c", Comment: cStyle, Keywords: cKeywords},
	{Name: "Go", Exts: []string{"go"}, Flag: "go", Comment: cStyle, Keywords: goKeywords},
	{Name: "Java", Exts: []string{"java"}, Flag: "java", Comment: cStyle, Keywords: javaKeywords},
	{Name: "JavaScript", Exts: []string{"js", "mjs"}, Flag: "js", Comment: cStyle, Keywords: jsKeywords},
	{Name: "TypeScript", Exts: []string{"ts", "tsx"}, Flag: "ts", Comment: cStyle, Keywords: tsKeywords},
	{Name: "Rust", Exts: []string{"rs"}, Flag: "rust", Comment: cStyle, Keywords: rustKeywords},
	{Name: "Kotlin", Exts: []string{"kt", "kts"}, Flag: "kotlin", Comment: cStyle, Keywords: kotlinKeywords},
	{Name: "Swift", Exts: []string{"swift"}, Flag: "swift", Comment: cStyle, Keywords: swiftKeywords},
	{Name: "C++", Exts: []string{"cpp", "cc", "cxx", "hpp"}, Flag: "cpp", Comment: cStyle, Keywords: cppKeywords},
	{Name: "C#", Exts: []string{"cs"}, Flag: "csharp", Comment: cStyle, Keywords: csharpKeywords},
	{Name: "Dart", Exts: []string{"dart"}, Flag: "dart", Comment: cStyle, Keywords: dartKeywords},

	// ---- Hash family (#) ----
	{Name: "Shell", Exts: []string{"sh", "bash"}, Flag: "sh", Comment: hashStyle, Keywords: shellKeywords},
	{Name: "Python", Exts: []string{"py"}, Flag: "python", Comment: CommentStyle{LineComment: "#", DocstringDelimiter: `"""`}, Keywords: pythonKeywords},
	{Name: "Ruby", Exts: []string{"rb"}, Flag: "ruby", Comment: hashStyle, Keywords: rubyKeywords},
	{Name: "Perl", Exts: []string{"pl", "pm"}, Flag: "perl", Comment: hashStyle, Keywords: perlKeywords},
	{Name: "Makefile", Exts: []string{"mk"}, Flag: "make", Comment: hashStyle,
		FileNames: []string{"Makefile", "makefile", "GNUmakefile"}},
	{Name: "Tcl", Exts: []string{"tcl"}, Flag: "tcl", Comment: hashStyle, Keywords: tclKeywords},

	// ---- Percent family (%) ----
	// RawTeX: true because % comments in REDUCE/MATLAB are typically raw LaTeX
	// (math environments, \eqalign, $$...$$, etc.) without {\ } wrappers.
	{Name: "REDUCE", Exts: []string{"red"}, Flag: "reduce", Comment: CommentStyle{LineComment: "%", RawTeX: true}, Keywords: reduceKeywords},
	{Name: "MATLAB", Exts: []string{"m"}, Flag: "matlab", Comment: CommentStyle{LineComment: "%", RawTeX: true}},

	// ---- Semicolon family (;) ----
	{Name: "Lisp/Scheme", Exts: []string{"lisp", "scm", "el"}, Flag: "lisp", Comment: semicolonStyle, Keywords: lispKeywords},

	// ---- Pascal family (block comments only) ----
	{
		Name: "Pascal",
		Exts: []string{"pas", "p"},
		Flag: "pascal",
		Comment: CommentStyle{
			BlockOpen:     "{",
			BlockClose:    "}",
			AltBlockOpen:  "(*",
			AltBlockClose: "*)",
			BlockNestable: true,
		},
		Keywords: pascalKeywords,
	},

	// ---- XML family (<!-- -->) ----
	{Name: "XML", Exts: []string{"xml", "xsl", "xslt", "svg", "xhtml"}, Flag: "xml",
		Comment: xmlCommentStyle, BoldTags: true},

	// ---- CSS (/* */ only, no line comments, no keywords) ----
	{Name: "CSS", Exts: []string{"css"}, Flag: "css", Comment: cssCommentStyle},

	// ---- HTML (<!-- -->, with embedded CSS, JS, and PHP) ----
	{
		Name:     "HTML",
		Exts:     []string{"html", "htm", "php"},
		Flag:     "html",
		Comment:  xmlCommentStyle,
		BoldTags: true,
		SubLanguages: []SubLanguageRule{
			{OpenTag: "<style", CloseTag: "</style>", LangFlag: "css"},
			{OpenTag: "<script", CloseTag: "</script>", LangFlag: "js"},
			{OpenTag: "<?php", CloseTag: "?>", LangFlag: "php", ImmediateActivation: true},
		},
	},

	// ---- PHP (// and /* */, standalone or embedded in HTML) ----
	{Name: "PHP", Exts: nil, Flag: "php", Comment: phpCommentStyle, Keywords: phpKeywords},
}

// FindByExt returns the LangDef whose Exts list contains ext (without dot).
// Returns nil if no match.
func FindByExt(ext string) *LangDef {
	for i := range Languages {
		for _, e := range Languages[i].Exts {
			if e == ext {
				return &Languages[i]
			}
		}
	}
	return nil
}

// FindByFlag returns the LangDef matching the -lang flag value.
// Returns nil if no match.
func FindByFlag(flag string) *LangDef {
	for i := range Languages {
		if Languages[i].Flag == flag {
			return &Languages[i]
		}
	}
	return nil
}

// FindByFileName returns the LangDef whose FileNames list contains the given
// base file name (e.g. "Makefile"). Returns nil if no match.
func FindByFileName(name string) *LangDef {
	for i := range Languages {
		for _, fn := range Languages[i].FileNames {
			if fn == name {
				return &Languages[i]
			}
		}
	}
	return nil
}
