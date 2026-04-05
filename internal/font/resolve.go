// Package font — resolve.go provides Path=-aware font resolution for LaTeX preambles.
//
// These functions generate \setmonofont / \setCJKmonofont / \setCJKmainfont commands
// with [Path=...] options when the font files are available in ~/.src2tex/fonts/.
// This allows fontspec to find fonts without requiring system-wide installation.
package font

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveFontDir returns the directory containing a font's files.
//
// If RegularFile is a bare filename (no "/"), the font is assumed to be under
// ~/.src2tex/fonts/{Name}/. If RegularFile contains a path separator, the
// directory portion is used directly (e.g. /Library/Fonts/).
func resolveFontDir(fd *FontDef) string {
	if strings.Contains(fd.RegularFile, "/") {
		return filepath.Dir(fd.RegularFile)
	}
	return filepath.Join(DefaultFontDir(), fd.Name)
}

// resolveFontBaseName returns the base filename of RegularFile.
func resolveFontBaseName(fd *FontDef) string {
	return filepath.Base(fd.RegularFile)
}

// resolveBoldBaseName returns the base filename of BoldFile.
func resolveBoldBaseName(fd *FontDef) string {
	return filepath.Base(fd.BoldFile)
}

// isFontFilePresent checks if a font's regular file exists at the resolved path.
func isFontFilePresent(fd *FontDef) bool {
	dir := resolveFontDir(fd)
	regFile := filepath.Join(dir, resolveFontBaseName(fd))
	_, err := os.Stat(regFile)
	return err == nil
}

// ResolveCodeFontLine returns the \setmonofont[...]{...} LaTeX command for a code font.
//
// Resolution order:
//  1. fontName is empty → try CMU Typewriter Text from TeX Live, then platform default.
//  2. fontName matches a known font (builtin or fonts.json) and is installed →
//     \setmonofont[Path=<dir>/, BoldFont=<bold>]{<regular>}
//  3. Fallback → \setmonofont{<fontName>} (assumes system font).
func ResolveCodeFontLine(fontName string) string {
	if fontName == "" {
		// No explicit font: try CMU Typewriter Text from TeX Live.
		if cmu := DetectCMUTypewriter(); cmu != nil {
			return fmt.Sprintf(`\setmonofont[Path=%s/, BoldFont=%s]{%s}`,
				cmu.Dir, cmu.BoldFile, cmu.RegularFile)
		}
		// Fallback to platform default.
		pd := GetPlatformDefaults()
		return `\setmonofont{` + pd.MonoFont + `}`
	}

	// Look up the font in builtin + fonts.json.
	fd := LookupFont(fontName)
	if fd != nil && isFontFilePresent(fd) {
		dir := resolveFontDir(fd)
		reg := resolveFontBaseName(fd)
		bold := resolveBoldBaseName(fd)
		if bold != "" && bold != "." {
			return fmt.Sprintf(`\setmonofont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
		}
		return fmt.Sprintf(`\setmonofont[Path=%s/]{%s}`, dir, reg)
	}

	// Unknown or not installed: bare name (system font assumption).
	return `\setmonofont{` + fontName + `}`
}

// ResolveCodeFontCJKMonoLine returns the \setCJKmonofont[Path=..., BoldFont=...]{...}
// command for unified code fonts. Returns "" if the font is not unified, not found,
// or not installed.
//
// This is only relevant for XeLaTeX (xeCJK provides \setCJKmonofont).
func ResolveCodeFontCJKMonoLine(fontName string) string {
	if fontName == "" {
		return ""
	}
	fd := LookupFont(fontName)
	if fd == nil || !fd.Unified || !isFontFilePresent(fd) {
		return ""
	}
	dir := resolveFontDir(fd)
	reg := resolveFontBaseName(fd)
	bold := resolveBoldBaseName(fd)
	if bold != "" && bold != "." {
		return fmt.Sprintf(`\setCJKmonofont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
	}
	return fmt.Sprintf(`\setCJKmonofont[Path=%s/]{%s}`, dir, reg)
}

// ResolveCodeFontCJKMainLine returns the \setCJKmainfont[Path=..., BoldFont=...]{...}
// command for unified code fonts when no separate -commentfont is specified.
// Returns "" if the font is not unified, not found, or not installed.
//
// When a unified font (e.g. HackGen) is used and no comment font is specified,
// the CJK main font should also be set to the code font so that CJK characters
// in comments render with the same font family.
func ResolveCodeFontCJKMainLine(fontName string) string {
	if fontName == "" {
		return ""
	}
	fd := LookupFont(fontName)
	if fd == nil || !fd.Unified || !isFontFilePresent(fd) {
		return ""
	}
	dir := resolveFontDir(fd)
	reg := resolveFontBaseName(fd)
	bold := resolveBoldBaseName(fd)
	if bold != "" && bold != "." {
		return fmt.Sprintf(`\setCJKmainfont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
	}
	return fmt.Sprintf(`\setCJKmainfont[Path=%s/]{%s}`, dir, reg)
}

// ResolveCommentFontLine returns the \setCJKmainfont[Path=..., ...]{...} command
// for the specified comment font name.
//
// Resolution order:
//  1. Look up as a comment font (BuiltinCommentFonts) → resolve from its directory.
//  2. Look up as a code font (BuiltinFonts / fonts.json) → resolve from its directory.
//  3. Fallback → \setCJKmainfont{name} (assumes system font).
func ResolveCommentFontLine(commentFont string) string {
	if commentFont == "" {
		return ""
	}

	// Already a full LaTeX command?
	if strings.HasPrefix(commentFont, `\setCJKmainfont`) {
		return strings.TrimRight(commentFont, "\n")
	}

	// Try as a comment font first.
	cfd := GetCommentFontDef(commentFont)
	if cfd != nil {
		if cfd.TexLive {
			dir, ok := findTexLiveFont(cfd.RegularFile)
			if ok {
				reg := cfd.FontSpec
				if reg == "" {
					reg = cfd.RegularFile
				}
				opts := fmt.Sprintf("Path=%s/", dir)
				if cfd.Extension != "" {
					opts += fmt.Sprintf(", Extension=%s", cfd.Extension)
				}
				if cfd.BoldFile != "" {
					opts += fmt.Sprintf(", BoldFont=%s", cfd.BoldFile)
				}
				return fmt.Sprintf(`\setCJKmainfont[%s]{%s}`, opts, reg)
			}
		} else {
			// Downloaded comment font: ~/.src2tex/fonts/comment-{name}/
			dir := filepath.Join(DefaultFontDir(), "comment-"+commentFont)
			regFile := cfd.RegularFile
			if _, err := os.Stat(filepath.Join(dir, regFile)); err == nil {
				reg := cfd.FontSpec
				if reg == "" {
					reg = regFile
				}
				opts := fmt.Sprintf("Path=%s/", dir)
				if cfd.Extension != "" {
					opts += fmt.Sprintf(", Extension=%s", cfd.Extension)
				}
				if cfd.BoldFile != "" {
					opts += fmt.Sprintf(", BoldFont=%s", cfd.BoldFile)
				}
				return fmt.Sprintf(`\setCJKmainfont[%s]{%s}`, opts, reg)
			}
		}
	}

	// Try as a code font (e.g. -commentfont hackgen).
	fd := LookupFont(commentFont)
	if fd != nil && isFontFilePresent(fd) {
		dir := resolveFontDir(fd)
		reg := resolveFontBaseName(fd)
		bold := resolveBoldBaseName(fd)
		if bold != "" && bold != "." {
			return fmt.Sprintf(`\setCJKmainfont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
		}
		return fmt.Sprintf(`\setCJKmainfont[Path=%s/]{%s}`, dir, reg)
	}

	// Unknown font: bare name fallback.
	return `\setCJKmainfont{` + commentFont + `}`
}

// ResolveCodeFontCJKSansLine returns the \setCJKsansfont[Path=..., BoldFont=...]{...}
// command for unified code fonts. Returns "" if the font is not unified, not found,
// or not installed.
//
// When a unified font (e.g. HackGen) is used as the code font, the CJK sans font
// should also use the same font so that \textsf{} in comments renders consistently.
func ResolveCodeFontCJKSansLine(fontName string) string {
	if fontName == "" {
		return ""
	}
	fd := LookupFont(fontName)
	if fd == nil || !fd.Unified || !isFontFilePresent(fd) {
		return ""
	}
	dir := resolveFontDir(fd)
	reg := resolveFontBaseName(fd)
	bold := resolveBoldBaseName(fd)
	if bold != "" && bold != "." {
		return fmt.Sprintf(`\setCJKsansfont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
	}
	return fmt.Sprintf(`\setCJKsansfont[Path=%s/]{%s}`, dir, reg)
}

// ResolveMainFontLine returns the \setmainfont[Path=..., ...]{...} command
// for the specified comment font name. This is the Latin pair of \setCJKmainfont.
//
// Resolution order:
//  1. Look up as a comment font (BuiltinCommentFonts) → resolve from its directory.
//  2. Look up as a code font (BuiltinFonts / fonts.json) → resolve from its directory.
//  3. Fallback → \setmainfont{name} (assumes system font).
func ResolveMainFontLine(commentFont string) string {
	if commentFont == "" {
		return ""
	}

	// Try as a comment font first.
	cfd := GetCommentFontDef(commentFont)
	if cfd != nil {
		if cfd.TexLive {
			dir, ok := findTexLiveFont(cfd.RegularFile)
			if ok {
				reg := cfd.FontSpec
				if reg == "" {
					reg = cfd.RegularFile
				}
				opts := fmt.Sprintf("Path=%s/", dir)
				if cfd.Extension != "" {
					opts += fmt.Sprintf(", Extension=%s", cfd.Extension)
				}
				if cfd.BoldFile != "" {
					opts += fmt.Sprintf(", BoldFont=%s", cfd.BoldFile)
				}
				return fmt.Sprintf(`\setmainfont[%s]{%s}`, opts, reg)
			}
		} else {
			dir := filepath.Join(DefaultFontDir(), "comment-"+commentFont)
			regFile := cfd.RegularFile
			if _, err := os.Stat(filepath.Join(dir, regFile)); err == nil {
				reg := cfd.FontSpec
				if reg == "" {
					reg = regFile
				}
				opts := fmt.Sprintf("Path=%s/", dir)
				if cfd.Extension != "" {
					opts += fmt.Sprintf(", Extension=%s", cfd.Extension)
				}
				if cfd.BoldFile != "" {
					opts += fmt.Sprintf(", BoldFont=%s", cfd.BoldFile)
				}
				return fmt.Sprintf(`\setmainfont[%s]{%s}`, opts, reg)
			}
		}
	}

	// Try as a code font (e.g. -commentfont hackgen).
	fd := LookupFont(commentFont)
	if fd != nil && isFontFilePresent(fd) {
		dir := resolveFontDir(fd)
		reg := resolveFontBaseName(fd)
		bold := resolveBoldBaseName(fd)
		if bold != "" && bold != "." {
			return fmt.Sprintf(`\setmainfont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
		}
		return fmt.Sprintf(`\setmainfont[Path=%s/]{%s}`, dir, reg)
	}

	// Unknown font: bare name fallback.
	return `\setmainfont{` + commentFont + `}`
}

// ResolveSansFontLine returns the \setsansfont[Path=..., BoldFont=...]{...}
// command for the code font. This is the Latin pair of \setCJKsansfont.
// Returns "" if the font is not specified, not found, or not installed.
func ResolveSansFontLine(fontName string) string {
	if fontName == "" {
		return ""
	}
	fd := LookupFont(fontName)
	if fd != nil && isFontFilePresent(fd) {
		dir := resolveFontDir(fd)
		reg := resolveFontBaseName(fd)
		bold := resolveBoldBaseName(fd)
		if bold != "" && bold != "." {
			return fmt.Sprintf(`\setsansfont[Path=%s/, BoldFont=%s]{%s}`, dir, bold, reg)
		}
		return fmt.Sprintf(`\setsansfont[Path=%s/]{%s}`, dir, reg)
	}
	// Unknown or not installed: bare name.
	return `\setsansfont{` + fontName + `}`
}
