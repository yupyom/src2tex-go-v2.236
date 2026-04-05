// Package font manages CJK fonts for src2tex.
// It supports TeX Live bundled fonts and fonts downloaded to ~/.src2tex/fonts/.
package font

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FontDef defines a monospace (code) CJK font.
type FontDef struct {
	Name            string
	DisplayName     string
	License         string
	Unified         bool    // true = Latin+CJK unified (one font covers both)
	RegularFile     string  // e.g. "HackGen-Regular.ttf"
	BoldFile        string  // e.g. "HackGen-Bold.ttf"
	MainRegularFile string  // for non-unified: mincho/serif regular
	MainBoldFile    string  // for non-unified: mincho/serif bold
	Scale           float64 // CJK scale factor (non-unified only)
	GithubRepo      string  // "owner/repo" for GitHub Releases API
	ZipAsset        string  // substring to match ZIP asset filename
	Description     string
}

// CommentFontDef defines a mincho/serif font for comment text.
type CommentFontDef struct {
	Name        string
	DisplayName string
	License     string
	TexLive     bool   // bundled with TeX Live (no download needed)
	RegularFile string // e.g. "HaranoAjiMincho-Regular.otf"
	BoldFile    string // may be empty
	Extension   string // ".otf" or ".ttf"
	FontSpec    string // fontspec name without extension
	GithubRepo  string // "owner/repo"
	ZipAsset    string // substring to match ZIP asset
	Description string
}

// FontInfo is the struct returned by ListAvailable.
type FontInfo struct {
	Name        string
	DisplayName string
	Path        string // directory containing the font files; empty for TeX Live fonts
	TexLive     bool   // true if bundled with TeX Live
	Comment     bool   // true if this is a mincho/comment font
}

// BuiltinFonts lists the supported code (monospace) fonts.
var BuiltinFonts = []FontDef{
	{
		Name:            "ipaex",
		DisplayName:     "IPAex Gothic",
		License:         "IPA",
		Unified:         false,
		RegularFile:     "ipaexg.ttf",
		BoldFile:        "ipaexg.ttf",
		MainRegularFile: "ipaexm.ttf",
		MainBoldFile:    "ipaexg.ttf",
		Scale:           1.05,
		Description:     "TeX Live 同梱の日本語フォント。ダウンロード不要。",
	},
	{
		Name:        "hackgen",
		DisplayName: "HackGen",
		License:     "SIL OFL",
		Unified:     true,
		RegularFile: "HackGen-Regular.ttf",
		BoldFile:    "HackGen-Bold.ttf",
		GithubRepo:  "yuru7/HackGen",
		ZipAsset:    "HackGen_v",
		Description: "Hack + 源柔ゴシック。プログラミング向けの人気フォント。",
	},
	{
		Name:        "udev",
		DisplayName: "UDEV Gothic",
		License:     "SIL OFL",
		Unified:     true,
		RegularFile: "UDEVGothic-Regular.ttf",
		BoldFile:    "UDEVGothic-Bold.ttf",
		GithubRepo:  "yuru7/udev-gothic",
		ZipAsset:    "UDEVGothic_v",
		Description: "JetBrains Mono + BIZ UDゴシック。UD（ユニバーサルデザイン）対応。",
	},
	{
		Name:        "firple",
		DisplayName: "Firple",
		License:     "SIL OFL",
		Unified:     true,
		RegularFile: "Firple-Regular.ttf",
		BoldFile:    "Firple-Bold.ttf",
		GithubRepo:  "negset/Firple",
		ZipAsset:    "Firple.zip",
		Description: "Fira Code + IBM Plex Sans JP。リガチャ対応。",
	},
}

// BuiltinCommentFonts lists the supported comment (mincho/serif) fonts.
var BuiltinCommentFonts = []CommentFontDef{
	{
		Name:        "haranoaji",
		DisplayName: "原ノ味明朝",
		License:     "SIL OFL",
		TexLive:     true,
		RegularFile: "HaranoAjiMincho-Regular.otf",
		BoldFile:    "HaranoAjiMincho-Bold.otf",
		Extension:   ".otf",
		FontSpec:    "HaranoAjiMincho-Regular",
		Description: "TeX Live 同梱の高品質明朝体。ダウンロード不要。",
	},
	{
		Name:        "ipaexm",
		DisplayName: "IPAex 明朝",
		License:     "IPA",
		TexLive:     true,
		RegularFile: "ipaexm.ttf",
		Extension:   ".ttf",
		FontSpec:    "ipaexm",
		Description: "TeX Live 同梱の明朝体。ダウンロード不要。",
	},
	{
		Name:        "noto-serif",
		DisplayName: "Noto Serif JP",
		License:     "SIL OFL",
		TexLive:     false,
		RegularFile: "NotoSerifJP-Regular.otf",
		BoldFile:    "NotoSerifJP-Bold.otf",
		Extension:   ".otf",
		FontSpec:    "NotoSerifJP-Regular",
		GithubRepo:  "notofonts/noto-cjk",
		ZipAsset:    "NotoSerifJP",
		Description: "Google Noto明朝体。高品質で幅広い字形をカバー。",
	},
}

// DefaultFontDir returns the default directory for downloaded fonts (~/.src2tex/fonts/).
func DefaultFontDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".src2tex", "fonts")
}

// FindFont searches TeX Live font directories for a font file by base name (without extension).
// It tries both .otf and .ttf suffixes. Returns the directory containing the file and true on success.
func FindFont(name string) (dir string, found bool) {
	for _, d := range texLiveFontDirs() {
		for _, ext := range []string{".otf", ".ttf"} {
			if _, err := os.Stat(filepath.Join(d, name+ext)); err == nil {
				return d, true
			}
		}
	}
	return "", false
}

// ListAvailable returns all fonts that are currently available on the system —
// either bundled with TeX Live or downloaded to fontDir.
func ListAvailable(fontDir string) []FontInfo {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	var result []FontInfo

	// Code fonts
	for _, fd := range BuiltinFonts {
		if fd.Name == "ipaex" {
			if isTexLiveAvailable() {
				result = append(result, FontInfo{
					Name:        fd.Name,
					DisplayName: fd.DisplayName,
					TexLive:     true,
				})
			}
			continue
		}
		fontPath := filepath.Join(fontDir, fd.Name)
		if _, err := os.Stat(filepath.Join(fontPath, fd.RegularFile)); err == nil {
			result = append(result, FontInfo{
				Name:        fd.Name,
				DisplayName: fd.DisplayName,
				Path:        fontPath,
			})
		}
	}

	// Comment fonts
	for _, cfd := range BuiltinCommentFonts {
		if cfd.TexLive {
			// Check if it's actually present in TeX Live
			dir, ok := findTexLiveFont(cfd.RegularFile)
			if ok {
				result = append(result, FontInfo{
					Name:        cfd.Name,
					DisplayName: cfd.DisplayName,
					Path:        dir,
					TexLive:     true,
					Comment:     true,
				})
			}
			continue
		}
		fontPath := filepath.Join(fontDir, "comment-"+cfd.Name)
		if _, err := os.Stat(filepath.Join(fontPath, cfd.RegularFile)); err == nil {
			result = append(result, FontInfo{
				Name:        cfd.Name,
				DisplayName: cfd.DisplayName,
				Path:        fontPath,
				Comment:     true,
			})
		}
	}

	return result
}

// GetFontDef returns the FontDef for name, or nil.
func GetFontDef(name string) *FontDef {
	for i := range BuiltinFonts {
		if BuiltinFonts[i].Name == name {
			return &BuiltinFonts[i]
		}
	}
	return nil
}

// GetCommentFontDef returns the CommentFontDef for name, or nil.
func GetCommentFontDef(name string) *CommentFontDef {
	for i := range BuiltinCommentFonts {
		if BuiltinCommentFonts[i].Name == name {
			return &BuiltinCommentFonts[i]
		}
	}
	return nil
}

// IsFontInstalled reports whether the code font is available (TeX Live or downloaded).
func IsFontInstalled(name, fontDir string) bool {
	fd := GetFontDef(name)
	if fd == nil {
		return false
	}
	if fd.Name == "ipaex" {
		return isTexLiveAvailable()
	}
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	_, err := os.Stat(filepath.Join(fontDir, fd.Name, fd.RegularFile))
	return err == nil
}

// IsCommentFontInstalled reports whether the comment font is available.
func IsCommentFontInstalled(name, fontDir string) bool {
	cfd := GetCommentFontDef(name)
	if cfd == nil {
		return false
	}
	if cfd.TexLive {
		_, ok := findTexLiveFont(cfd.RegularFile)
		return ok
	}
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	_, err := os.Stat(filepath.Join(fontDir, "comment-"+name, cfd.RegularFile))
	return err == nil
}

// CodeFontPath returns the directory containing the downloaded code font.
// Returns empty string for TeX Live fonts (ipaex) or if not found.
func CodeFontPath(name, fontDir string) string {
	if name == "ipaex" {
		return ""
	}
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	return filepath.Join(fontDir, name)
}

// CommentFontPath returns the directory containing the downloaded comment font.
func CommentFontPath(name, fontDir string) string {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	return filepath.Join(fontDir, "comment-"+name)
}

// AutoDetectCommentFont returns the best available comment (mincho) font name.
// Priority: haranoaji (TeX Live) > ipaexm (TeX Live) > downloaded comment fonts.
// Returns "" if none found (caller should fall back to code font for \setCJKmainfont).
func AutoDetectCommentFont(fontDir string) string {
	for _, cfd := range BuiltinCommentFonts {
		if cfd.TexLive {
			if _, ok := findTexLiveFont(cfd.RegularFile); ok {
				return cfd.Name
			}
		}
	}
	for _, cfd := range BuiltinCommentFonts {
		if !cfd.TexLive && IsCommentFontInstalled(cfd.Name, fontDir) {
			return cfd.Name
		}
	}
	return ""
}

// ResolveCommentFontXeLaTeX returns the \setCJKmainfont{...} LaTeX command for
// the given comment font name. fontDir is the downloaded-font directory.
// The returned string includes a trailing newline.
func ResolveCommentFontXeLaTeX(fontName, fontDir string) string {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}

	// Known comment font?
	cfd := GetCommentFontDef(fontName)
	if cfd != nil {
		if cfd.TexLive {
			dir, ok := findTexLiveFont(cfd.RegularFile)
			if ok {
				baseName := strings.TrimSuffix(cfd.RegularFile, cfd.Extension)
				if cfd.BoldFile != "" {
					return fmt.Sprintf(`\setCJKmainfont[Path=%s/, Extension=%s, BoldFont=%s]{%s}`+"\n",
						dir, cfd.Extension, cfd.BoldFile, baseName)
				}
				return fmt.Sprintf(`\setCJKmainfont[Path=%s/, Extension=%s]{%s}`+"\n",
					dir, cfd.Extension, baseName)
			}
		} else if IsCommentFontInstalled(fontName, fontDir) {
			dir := CommentFontPath(fontName, fontDir) + "/"
			baseName := strings.TrimSuffix(cfd.RegularFile, cfd.Extension)
			if cfd.BoldFile != "" {
				return fmt.Sprintf(`\setCJKmainfont[Path=%s, Extension=%s, BoldFont=%s]{%s}`+"\n",
					dir, cfd.Extension, cfd.BoldFile, baseName)
			}
			return fmt.Sprintf(`\setCJKmainfont[Path=%s, Extension=%s]{%s}`+"\n",
				dir, cfd.Extension, baseName)
		}
	}

	// Try raw name in TeX Live dirs
	for _, d := range texLiveFontDirs() {
		for _, ext := range []string{".otf", ".ttf"} {
			if _, err := os.Stat(filepath.Join(d, fontName+ext)); err == nil {
				return fmt.Sprintf(`\setCJKmainfont[Path=%s/, Extension=%s]{%s}`+"\n",
					d, ext, fontName)
			}
		}
	}

	// Fall back: bare name (system font or fontspec can locate it)
	return fmt.Sprintf(`\setCJKmainfont{%s}`+"\n", fontName)
}

// PrintFontList prints a human-readable font list to w.
func PrintFontList(fontDir string) {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}
	fmt.Fprintf(os.Stderr, "\nCode fonts (-font):\n\n")
	fmt.Fprintf(os.Stderr, "  %s %s %-10s %s\n", padRight("NAME", 12), padRight("DISPLAY NAME", 20), "LICENSE", "DESCRIPTION")
	fmt.Fprintf(os.Stderr, "  %s %s %-10s %s\n", padRight("----", 12), padRight("------------", 20), "-------", "-----------")
	for _, fd := range BuiltinFonts {
		status := statusLabel(IsFontInstalled(fd.Name, fontDir), fd.Name == "ipaex")
		fmt.Fprintf(os.Stderr, "  %s %s %-10s %s%s\n",
			padRight(fd.Name, 12), padRight(fd.DisplayName, 20), fd.License, fd.Description, status)
	}

	fmt.Fprintf(os.Stderr, "\nComment fonts (-commentfont, 明朝体):\n\n")
	fmt.Fprintf(os.Stderr, "  %s %s %-10s %s\n", padRight("NAME", 12), padRight("DISPLAY NAME", 22), "LICENSE", "DESCRIPTION")
	fmt.Fprintf(os.Stderr, "  %s %s %-10s %s\n", padRight("----", 12), padRight("------------", 22), "-------", "-----------")
	for _, cfd := range BuiltinCommentFonts {
		status := statusLabel(IsCommentFontInstalled(cfd.Name, fontDir), cfd.TexLive)
		fmt.Fprintf(os.Stderr, "  %s %s %-10s %s%s\n",
			padRight(cfd.Name, 12), padRight(cfd.DisplayName, 22), cfd.License, cfd.Description, status)
	}
	fmt.Fprintln(os.Stderr)
}

// padRight pads a string with spaces to the given display width.
// CJK characters (U+2E80–U+9FFF, U+F900–U+FAFF, U+FE30–U+FE4F,
// U+FF00–U+FFEF, U+20000–U+2FA1F) are counted as 2 columns wide.
func padRight(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	pad := width - w
	return s + strings.Repeat(" ", pad)
}

// displayWidth returns the terminal display width of a string,
// counting CJK/fullwidth characters as 2 columns.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isCJKOrFullwidth(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

// isCJKOrFullwidth returns true if the rune occupies 2 terminal columns.
func isCJKOrFullwidth(r rune) bool {
	return (r >= 0x2E80 && r <= 0x9FFF) || // CJK Radicals, Kangxi, CJK Unified Ideographs
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		(r >= 0xFE30 && r <= 0xFE4F) || // CJK Compatibility Forms
		(r >= 0xFF01 && r <= 0xFF60) || // Fullwidth ASCII
		(r >= 0xFFE0 && r <= 0xFFEF) || // Fullwidth symbols
		(r >= 0x20000 && r <= 0x2FA1F) || // CJK Extension B–F
		(r >= 0x3000 && r <= 0x303F) || // CJK Symbols and Punctuation
		(r >= 0x3040 && r <= 0x30FF) || // Hiragana, Katakana
		(r >= 0x31F0 && r <= 0x31FF) || // Katakana Phonetic Extensions
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul Syllables
}

// statusLabel returns " [TeX Live]" or " [installed]" or "".
func statusLabel(installed, texlive bool) string {
	if texlive {
		if installed {
			return " [TeX Live]"
		}
		return " [TeX Live — not found]"
	}
	if installed {
		return " [installed]"
	}
	return ""
}

// isTexLiveAvailable reports whether any TeX Live installation is detected.
func isTexLiveAvailable() bool {
	for _, d := range texLiveFontDirs() {
		if _, err := os.Stat(d); err == nil {
			return true
		}
	}
	return false
}

// findTexLiveFont searches TeX Live font directories for the given filename.
// Returns the directory containing the file.
func findTexLiveFont(filename string) (string, bool) {
	for _, d := range texLiveFontDirs() {
		if _, err := os.Stat(filepath.Join(d, filename)); err == nil {
			return d, true
		}
	}
	return "", false
}

// texLiveFontDirs returns the TeX Live font subdirectories to search, in priority order.
// Results are platform-specific.
func texLiveFontDirs() []string {
	var dirs []string
	for _, base := range texLiveBaseDirs() {
		// Specific well-known subdirs first (faster lookup)
		dirs = append(dirs,
			filepath.Join(base, "texmf-dist/fonts/opentype/public/haranoaji"),
			filepath.Join(base, "texmf-dist/fonts/truetype/public/ipaex"),
			filepath.Join(base, "texmf-dist/fonts/opentype/public/cm-unicode"),
			// Broader search dirs for other fonts
			filepath.Join(base, "texmf-dist/fonts/opentype/public"),
			filepath.Join(base, "texmf-dist/fonts/truetype/public"),
		)
	}
	return dirs
}

// CMUTypewriterResult holds the detection result for CMU Typewriter Text.
type CMUTypewriterResult struct {
	Dir         string // directory containing the .otf files
	RegularFile string // e.g. "cmuntt.otf"
	BoldFile    string // e.g. "cmuntb.otf"
}

// DetectCMUTypewriter searches TeX Live font directories for CMU Typewriter Text.
// Returns nil if not found.
func DetectCMUTypewriter() *CMUTypewriterResult {
	for _, base := range texLiveBaseDirs() {
		dir := filepath.Join(base, "texmf-dist/fonts/opentype/public/cm-unicode")
		regular := filepath.Join(dir, "cmuntt.otf")
		bold := filepath.Join(dir, "cmuntb.otf")
		if _, err := os.Stat(regular); err == nil {
			if _, err := os.Stat(bold); err == nil {
				return &CMUTypewriterResult{
					Dir:         dir,
					RegularFile: "cmuntt.otf",
					BoldFile:    "cmuntb.otf",
				}
			}
		}
	}
	return nil
}

// texLiveBaseDirs returns the TeX Live base installation directories for the current OS.
func texLiveBaseDirs() []string {
	years := []string{"2026", "2025", "2024", "2023"}
	var bases []string

	switch runtime.GOOS {
	case "windows":
		drives := []string{`C:\`, `D:\`}
		for _, drv := range drives {
			for _, y := range years {
				bases = append(bases, filepath.Join(drv, "texlive", y))
			}
		}
		// User-local TeX Live
		if home, err := os.UserHomeDir(); err == nil {
			for _, y := range years {
				bases = append(bases, filepath.Join(home, "texlive", y))
			}
		}
		// MiKTeX (user install)
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			bases = append(bases, filepath.Join(localAppData, "Programs", "MiKTeX"))
		}
	case "linux":
		for _, y := range years {
			bases = append(bases, "/usr/local/texlive/"+y)
		}
		// Distro-packaged TeX Live (Debian/Ubuntu etc.)
		bases = append(bases, "/usr/share/texlive")
		// Also check /usr/local without year (some minimal installs)
		bases = append(bases, "/usr/local/texlive/texmf-local")
	default: // darwin and others
		for _, y := range years {
			bases = append(bases, "/usr/local/texlive/"+y)
		}
	}
	return bases
}
