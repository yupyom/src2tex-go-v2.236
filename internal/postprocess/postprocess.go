// Package postprocess transforms legacy dvips \special commands in generated
// .tex files into modern \includegraphics equivalents, and triggers EPS→PDF
// conversion for any referenced EPS files.
//
// This bridges the gap between the original src2tex (which used jtex/dvi2ps
// with \special{epsfile=...}) and modern TeX engines that expect PDF images.
package postprocess

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yupyom/src2tex-go-v2.236/internal/epsconv"
)

// specialRe matches \special{epsfile=<filename> ...optional params...}
var specialRe = regexp.MustCompile(`\\special\{epsfile=([^\s}]+)([^}]*)\}`)

// ProcessTexFile performs all post-processing on a generated .tex file:
//
//  1. Convert \special{epsfile=...} to \includegraphics{...}
//  2. Clean up \vskip artifacts from dvips→includegraphics conversion
//  3. Convert referenced EPS files to PDF using Ghostscript
func ProcessTexFile(texPath string) error {
	data, err := os.ReadFile(texPath)
	if err != nil {
		return fmt.Errorf("postprocess: %w", err)
	}

	content := string(data)
	modified := false

	// Step 1: Convert \special{epsfile=...} to \includegraphics{...}
	if specialRe.MatchString(content) {
		content = convertSpecials(content)
		content = cleanupVskip(content)
		content = wrapSideBySide(content)
		modified = true
	}

	// Write back if modified
	if modified {
		if err := os.WriteFile(texPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("postprocess: write: %w", err)
		}
	}

	// Step 2: Convert referenced EPS files to PDF
	n, errs := epsconv.ConvertReferencedEps(texPath)
	if n > 0 {
		fmt.Fprintf(os.Stderr, "src2tex: converted %d EPS file(s) to PDF\n", n)
	}
	if len(errs) > 0 {
		// Report errors but don't fail — the user might have PDFs already
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "src2tex: warning: %v\n", e)
		}
	}

	return nil
}

// convertSpecials replaces \special{epsfile=<file> hscale=<h> vscale=<v> hoffset=<x>}
// with \includegraphics[scale=<s>]{<file>} (without .eps extension).
func convertSpecials(content string) string {
	hscaleRe := regexp.MustCompile(`hscale=([0-9.]+)`)
	vscaleRe := regexp.MustCompile(`vscale=([0-9.]+)`)
	hoffsetRe := regexp.MustCompile(`hoffset=([0-9.]+)`)

	return specialRe.ReplaceAllStringFunc(content, func(match string) string {
		submatches := specialRe.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		epsFileName := submatches[1]
		// Strip .eps extension so graphicx auto-discovers .pdf
		epsFileName = strings.TrimSuffix(epsFileName, ".eps")
		params := ""
		if len(submatches) > 2 {
			params = submatches[2]
		}

		// Parse scale parameters from dvips syntax
		hscaleMatch := hscaleRe.FindStringSubmatch(params)
		vscaleMatch := vscaleRe.FindStringSubmatch(params)
		hoffsetMatch := hoffsetRe.FindStringSubmatch(params)

		var opts []string
		if len(hscaleMatch) > 1 && len(vscaleMatch) > 1 {
			opts = append(opts, fmt.Sprintf("scale=%s", vscaleMatch[1]))
		} else if len(hscaleMatch) > 1 {
			opts = append(opts, fmt.Sprintf("scale=%s", hscaleMatch[1]))
		} else if len(vscaleMatch) > 1 {
			opts = append(opts, fmt.Sprintf("scale=%s", vscaleMatch[1]))
		}

		optStr := ""
		if len(opts) > 0 {
			optStr = "[" + strings.Join(opts, ",") + "]"
		}

		result := fmt.Sprintf(`\includegraphics%s{%s}`, optStr, epsFileName)

		// dvips hoffset is absolute positioning; approximate with \hfill
		if len(hoffsetMatch) > 1 {
			result = fmt.Sprintf(`\hfill %s`, result)
		}

		return result
	})
}

// cleanupVskip removes excess \vskip spacing that was needed for dvips
// \special overlay positioning but is unnecessary with \includegraphics
// (which consumes its natural height).
func cleanupVskip(content string) string {
	// Remove \vskip after \includegraphics or \end{center}
	re1 := regexp.MustCompile(`(\\includegraphics[^\n]*\n)\\vskip\s+[0-9.]+\s*cm`)
	content = re1.ReplaceAllString(content, "${1}")

	re2 := regexp.MustCompile(`(\\end\{center\}\n)\\vskip\s+[0-9.]+\s*cm`)
	content = re2.ReplaceAllString(content, "${1}")

	// Remove \vskip before \includegraphics or \begin{center}
	re3 := regexp.MustCompile(`\\vskip\s+[0-9.]+\s*cm\n(\\includegraphics|\\begin\{center\})`)
	content = re3.ReplaceAllString(content, "${1}")

	return content
}

// wrapSideBySide detects side-by-side image patterns (from dvips hoffset
// conversion) and wraps them in \begin{center}...\end{center}.
func wrapSideBySide(content string) string {
	// Pattern: \includegraphics...\n\hfill \includegraphics...
	re := regexp.MustCompile(`(\\includegraphics[^\n]+)\n\\hfill\s+(\\includegraphics[^\n]+)`)
	content = re.ReplaceAllString(content, "\\begin{center}\n${1}\\quad ${2}\n\\end{center}")

	return content
}

// HasEpsReferences checks whether a .tex file content contains references
// that would need EPS→PDF conversion (either \special or \includegraphics
// pointing to EPS files).
func HasEpsReferences(texContent string) bool {
	if specialRe.MatchString(texContent) {
		return true
	}
	refs := epsconv.FindImageReferences(texContent)
	for _, ref := range refs {
		ext := filepath.Ext(ref)
		if ext == ".eps" || ext == ".pdf" || ext == "" {
			return true
		}
	}
	return false
}
