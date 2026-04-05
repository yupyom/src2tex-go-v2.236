package scanner

import (
	"strings"

	"github.com/yupyom/src2tex-go-v2.236/internal/lang"
)

// TokenKind classifies a scanned token.
type TokenKind int

const (
	TokenCode         TokenKind = iota // ordinary source code
	TokenLineComment                   // from line-comment start to end of line (inclusive of '\n')
	TokenBlockComment                  // /* ... */ including both markers
	TokenTeX                           // {\ ... } TeX-mode content inside a comment — emitted verbatim
	TokenKeyword                       // reserved word rendered in bold
)

// Token is a contiguous slice of the input with its classification.
type Token struct {
	Kind TokenKind
	Text []byte
}

// state is the internal scanner state.
type state int

const (
	stateCode         state = iota
	stateLineComment        // consuming a line comment (until '\n' or EOF)
	stateBlockComment       // consuming a block comment (until BlockClose or EOF)
	stateDocstring          // consuming a triple-quoted docstring (""" or ''') as block comment
	stateStringDouble       // inside "..." — skip escaped \" until closing "
	stateStringSingle       // inside '...' — skip escaped \' until closing '
	stateStringRaw          // inside `...` raw string (Go) — no escapes, ends at `
	stateIdent              // accumulating an ASCII identifier for keyword lookup
	stateSubLangTag         // consuming attributes inside <style ...> or <script ...> until '>'
)

// Scan tokenises src according to ld's comment rules and keyword table.
// It returns one Token per contiguous region of the same kind.
func Scan(src []byte, ld *lang.LangDef) []Token {
	var tokens []Token
	cur := stateCode
	start := 0
	i := 0
	n := len(src)

	lineOpen := ld.Comment.LineComment
	blockOpen := ld.Comment.BlockOpen
	blockClose := ld.Comment.BlockClose
	altBlockOpen := ld.Comment.AltBlockOpen
	altBlockClose := ld.Comment.AltBlockClose
	altBlockActive := false // true when inside an AltBlockOpen..AltBlockClose comment
	blockNestable := ld.Comment.BlockNestable
	blockDepth := 0
	docstringDelim := ld.Comment.DocstringDelimiter
	docstringClose := "" // set when entering stateDocstring

	// Build keyword set for O(1) lookup.
	kwSet := make(map[string]bool, len(ld.Keywords))
	for _, kw := range ld.Keywords {
		kwSet[kw] = true
	}

	flush := func(end int, kind TokenKind) {
		if end > start {
			tokens = append(tokens, Token{Kind: kind, Text: src[start:end]})
		}
		start = end
	}

	// --- Sub-language context stack ---
	type scanCtx struct {
		lineOpen       string
		blockOpen      string
		blockClose     string
		altBlockOpen   string
		altBlockClose  string
		blockNestable  bool
		kwSet          map[string]bool
		boldTags       bool
		closeTag       string // lowercase close tag active when this context was pushed
		docstringDelim string
	}
	var ctxStack []scanCtx
	var activeCloseTag string          // current sub-language close tag (lowercase), "" if top-level
	var matchedSubRule *lang.SubLanguageRule // set when entering stateSubLangTag

	boldTags := ld.BoldTags
	subLangRules := ld.SubLanguages

	// pushContext saves current scanner variables, switches to sub-language.
	pushContext := func(rule *lang.SubLanguageRule) {
		ctxStack = append(ctxStack, scanCtx{
			lineOpen:       lineOpen,
			blockOpen:      blockOpen,
			blockClose:     blockClose,
			altBlockOpen:   altBlockOpen,
			altBlockClose:  altBlockClose,
			blockNestable:  blockNestable,
			kwSet:          kwSet,
			boldTags:       boldTags,
			closeTag:       activeCloseTag,
			docstringDelim: docstringDelim,
		})
		subLD := lang.FindByFlag(rule.LangFlag)
		if subLD == nil {
			return // unknown sub-language; stay in current context
		}
		lineOpen = subLD.Comment.LineComment
		blockOpen = subLD.Comment.BlockOpen
		blockClose = subLD.Comment.BlockClose
		altBlockOpen = subLD.Comment.AltBlockOpen
		altBlockClose = subLD.Comment.AltBlockClose
		blockNestable = subLD.Comment.BlockNestable
		docstringDelim = subLD.Comment.DocstringDelimiter
		boldTags = subLD.BoldTags
		subLangRules = subLD.SubLanguages

		kwSet = make(map[string]bool, len(subLD.Keywords))
		for _, kw := range subLD.Keywords {
			kwSet[kw] = true
		}

		activeCloseTag = strings.ToLower(rule.CloseTag)
	}

	// popContext restores scanner variables from the stack.
	popContext := func() {
		if len(ctxStack) == 0 {
			return
		}
		top := ctxStack[len(ctxStack)-1]
		ctxStack = ctxStack[:len(ctxStack)-1]
		lineOpen = top.lineOpen
		blockOpen = top.blockOpen
		blockClose = top.blockClose
		altBlockOpen = top.altBlockOpen
		altBlockClose = top.altBlockClose
		blockNestable = top.blockNestable
		kwSet = top.kwSet
		boldTags = top.boldTags
		activeCloseTag = top.closeTag
		docstringDelim = top.docstringDelim
		subLangRules = ld.SubLanguages
	}

	for i < n {
		switch cur {
		case stateCode:
			switch {
			// --- 1. Sub-language CloseTag check (highest priority) ---
			case activeCloseTag != "" && hasPrefixCI(src, i, activeCloseTag):
				flush(i, TokenCode)
				ct := activeCloseTag
				if len(ct) > 2 && ct[0] == '<' && ct[1] == '/' {
					// Emit "</" as TokenCode
					i += 2
					flush(i, TokenCode)
					// Extract and emit tag name as TokenKeyword
					tagStart := i
					for i < n && isTagNameChar(src[i]) {
						i++
					}
					if i > tagStart {
						flush(i, TokenKeyword)
					}
					// Consume ">" as TokenCode
					for i < n && src[i] != '>' {
						i++
					}
					if i < n {
						i++ // skip '>'
					}
					flush(i, TokenCode)
				} else {
					// Non-HTML close tag (e.g. "?>" for PHP)
					i += len(activeCloseTag)
					flush(i, TokenCode)
				}
				popContext()

			// --- 2. Block comment (<!-- --> for HTML, /* */ for CSS/JS) ---
			case blockOpen != "" && hasPrefix(src, i, blockOpen):
				flush(i, TokenCode)
				cur = stateBlockComment
				i += len(blockOpen) // skip past the opener; it will be part of the block token
				blockDepth = 1

			// --- 2b. Alternative block comment (e.g. (* *) for Pascal) ---
			case altBlockOpen != "" && hasPrefix(src, i, altBlockOpen):
				flush(i, TokenCode)
				cur = stateBlockComment
				i += len(altBlockOpen)
				blockDepth = 1
				altBlockActive = true

			// --- 3. Line comment ---
			case lineOpen != "" && hasPrefix(src, i, lineOpen):
				flush(i, TokenCode)
				cur = stateLineComment
				// do NOT advance i; stateLineComment consumes the marker too

			// --- 4. BoldTags: parse <tagname> and optionally trigger SubLanguage ---
			case boldTags && src[i] == '<':
				tagOpenPos := i // '<' の位置を保存
				j := i + 1
				isCloseTag := false
				if j < n && src[j] == '/' {
					j++
					isCloseTag = true
				}
				if j < n && isTagNameStart(src[j]) {
					// It's a tag. Flush code before '<'.
					flush(i, TokenCode)
					// Scan tag name
					tagNameStart := j
					for j < n && isTagNameChar(src[j]) {
						j++
					}
					// Emit "<" or "</" as TokenCode
					flush(tagNameStart, TokenCode)
					// Emit tag name as TokenKeyword
					flush(j, TokenKeyword)
					i = j

					if !isCloseTag {
						// Check for sub-language match
						rule := matchSubLangOpen(src, tagOpenPos, subLangRules)
						if rule != nil {
							matchedSubRule = rule
							cur = stateSubLangTag
						}
						// If no match, continue scanning (attributes are code)
					}
					// Close tags here are regular close tags (not activeCloseTag,
					// which was checked in priority 1)
				} else {
					// '<' not followed by a regular tag name (e.g. "<!--", "<?php").
					// Still check for sub-language rules like <?php.
					rule := matchSubLangOpen(src, tagOpenPos, subLangRules)
					if rule != nil {
						flush(tagOpenPos, TokenCode)
						i = tagOpenPos + len(rule.OpenTag)
						flush(i, TokenCode)
						matchedSubRule = rule
						cur = stateSubLangTag
					} else {
						i++
					}
				}

			// --- 5. SubLanguage OpenTag for non-boldTags contexts (fallback) ---
			case !boldTags && len(subLangRules) > 0 && matchSubLangOpen(src, i, subLangRules) != nil:
				rule := matchSubLangOpen(src, i, subLangRules)
				flush(i, TokenCode)
				matchedSubRule = rule
				cur = stateSubLangTag

			// --- 6. Docstring ---
			case docstringDelim != "" && hasPrefix(src, i, `"""`) && isAtLineStart(src, i):
				flush(i, TokenCode)
				docstringClose = `"""`
				i += 3
				cur = stateDocstring

			case docstringDelim != "" && hasPrefix(src, i, `'''`) && isAtLineStart(src, i):
				flush(i, TokenCode)
				docstringClose = `'''`
				i += 3
				cur = stateDocstring

			// --- 7. String literals ---
			case src[i] == '"':
				i++
				cur = stateStringDouble

			case src[i] == '\'':
				i++
				cur = stateStringSingle

			case src[i] == '`':
				i++
				cur = stateStringRaw

			// --- 8. Keywords ---
			case len(kwSet) > 0 && isIdentStart(src[i]):
				// Flush preceding code, then start accumulating an identifier.
				flush(i, TokenCode)
				cur = stateIdent
				i++

			default:
				i++
			}

		case stateLineComment:
			// Consume until '\n' (inclusive) or EOF.
			if src[i] == '\n' {
				i++
				flush(i, TokenLineComment)
				cur = stateCode
			} else {
				i++
			}

		case stateBlockComment:
			// Consume until BlockClose (or AltBlockClose) inclusive, or EOF.
			// For nestable block comments (e.g. Pascal { }),
			// track brace depth so inner {…} groups don't close the comment.
			// AltBlock comments (e.g. Pascal (* *)) are never nestable.
			if altBlockActive {
				if hasPrefix(src, i, altBlockClose) {
					i += len(altBlockClose)
					flush(i, TokenBlockComment)
					cur = stateCode
					altBlockActive = false
				} else {
					i++
				}
			} else if blockNestable && blockOpen != "" && hasPrefix(src, i, blockOpen) {
				blockDepth++
				i += len(blockOpen)
			} else if hasPrefix(src, i, blockClose) {
				if blockNestable {
					blockDepth--
					i += len(blockClose)
					if blockDepth == 0 {
						flush(i, TokenBlockComment)
						cur = stateCode
					}
				} else {
					i += len(blockClose)
					flush(i, TokenBlockComment)
					cur = stateCode
				}
			} else {
				i++
			}

		case stateDocstring:
			// Consume until the matching closing triple-quote (inclusive) or EOF.
			if hasPrefix(src, i, docstringClose) {
				i += len(docstringClose)
				flush(i, TokenBlockComment)
				cur = stateCode
			} else {
				i++
			}

		case stateStringDouble:
			// Consume until unescaped closing ".
			if src[i] == '\\' && i+1 < n {
				i += 2 // skip escape sequence (e.g. \", \\)
			} else if src[i] == '"' {
				i++
				cur = stateCode
			} else {
				i++
			}

		case stateStringSingle:
			// Consume until unescaped closing '.
			if src[i] == '\\' && i+1 < n {
				i += 2 // skip escape sequence (e.g. \', \\)
			} else if src[i] == '\'' {
				i++
				cur = stateCode
			} else {
				i++
			}

		case stateStringRaw:
			// Raw string (Go backtick): no escape processing, ends at `.
			if src[i] == '`' {
				i++
				cur = stateCode
			} else {
				i++
			}

		case stateIdent:
			// Accumulate identifier characters. On the first non-identifier byte,
			// classify the word.
			// - Keyword: flush [start:i] as TokenKeyword.
			// - Non-keyword: do NOT flush; return to stateCode leaving start unchanged
			//   so this identifier merges with the surrounding TokenCode region.
			if isIdentCont(src[i]) {
				i++
			} else {
				if kwSet[string(src[start:i])] {
					flush(i, TokenKeyword)
				}
				cur = stateCode
			}

		case stateSubLangTag:
			// Consume everything inside the opening tag until '>'.
			// E.g., <style type="text/css"> → consume ' type="text/css">'
			// For ImmediateActivation rules (e.g. <?php), context switches right away.
			if matchedSubRule != nil && matchedSubRule.ImmediateActivation {
				flush(i, TokenCode)
				pushContext(matchedSubRule)
				matchedSubRule = nil
				cur = stateCode
			} else if src[i] == '>' {
				i++
				flush(i, TokenCode)
				if matchedSubRule != nil {
					pushContext(matchedSubRule)
					matchedSubRule = nil
				}
				cur = stateCode
			} else {
				i++
			}
		}
	}

	// Flush whatever remains.
	if start < n {
		if cur == stateIdent && kwSet[string(src[start:n])] {
			tokens = append(tokens, Token{Kind: TokenKeyword, Text: src[start:n]})
		} else {
			flush(n, cur.tokenKind())
		}
	}

	return tokens
}

// tokenKind maps a terminal state to its TokenKind for EOF flushes.
func (s state) tokenKind() TokenKind {
	switch s {
	case stateLineComment:
		return TokenLineComment
	case stateBlockComment, stateDocstring:
		return TokenBlockComment
	default:
		return TokenCode
	}
}

// isIdentStart reports whether b can begin an ASCII identifier ([A-Za-z_]).
func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// isIdentCont reports whether b can continue an ASCII identifier ([A-Za-z0-9_]).
func isIdentCont(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

// isAtLineStart reports whether position i in src is at the start of a line,
// meaning only spaces and tabs appear between the most recent newline and i.
func isAtLineStart(src []byte, i int) bool {
	for j := i - 1; j >= 0; j-- {
		if src[j] == '\n' {
			return true
		}
		if src[j] != ' ' && src[j] != '\t' {
			return false
		}
	}
	return true // beginning of file
}

// hasPrefix reports whether src[i:] starts with prefix.
func hasPrefix(src []byte, i int, prefix string) bool {
	p := []byte(prefix)
	if i+len(p) > len(src) {
		return false
	}
	for k, b := range p {
		if src[i+k] != b {
			return false
		}
	}
	return true
}

// hasPrefixCI reports whether src[i:] starts with prefix, case-insensitively.
func hasPrefixCI(src []byte, i int, prefix string) bool {
	if i+len(prefix) > len(src) {
		return false
	}
	for k := 0; k < len(prefix); k++ {
		a, b := src[i+k], prefix[k]
		if a == b {
			continue
		}
		// lowercase both
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}

// isTagNameStart reports whether b can begin an HTML/XML tag name ([A-Za-z]).
func isTagNameStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// isTagNameChar reports whether b can continue an HTML/XML tag name ([A-Za-z0-9:_-]).
func isTagNameChar(b byte) bool {
	return isTagNameStart(b) || (b >= '0' && b <= '9') || b == '-' || b == ':' || b == '_'
}

// matchSubLangOpen checks if src[i:] matches any SubLanguageRule's OpenTag.
// The match is case-insensitive and requires that the character after the OpenTag
// is '>', whitespace, or EOF (to prevent "<stylesheet>" matching "<style").
// Returns the matched rule or nil.
func matchSubLangOpen(src []byte, i int, rules []lang.SubLanguageRule) *lang.SubLanguageRule {
	for idx := range rules {
		rule := &rules[idx]
		if !hasPrefixCI(src, i, rule.OpenTag) {
			continue
		}
		end := i + len(rule.OpenTag)
		if end >= len(src) {
			return rule // EOF right after tag — treat as match
		}
		ch := src[end]
		if ch == '>' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '/' {
			return rule
		}
	}
	return nil
}
