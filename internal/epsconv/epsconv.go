// Package epsconv provides EPS to PDF conversion using Ghostscript.
//
// Modern TeX engines (XeLaTeX, LuaLaTeX, pdfLaTeX) cannot natively include
// EPS files. This package converts EPS to PDF preserving the original
// BoundingBox dimensions, so that \includegraphics works correctly across
// all four supported engines.
package epsconv

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ConvertEpsToPdf converts an EPS file to PDF using Ghostscript,
// preserving the original BoundingBox to avoid full-page whitespace.
// If the PDF already exists with correct dimensions, the conversion is skipped.
// Returns the output PDF path, whether it was a cache hit, or an error.
func ConvertEpsToPdf(epsPath string) (pdfPath string, cached bool, err error) {
	pdfPath = strings.TrimSuffix(epsPath, ".eps") + ".pdf"

	// Read BoundingBox from EPS file
	bbWidth, bbHeight, err := readBoundingBox(epsPath)
	if err != nil {
		return "", false, fmt.Errorf("epsconv: %s: %w", filepath.Base(epsPath), err)
	}

	// Check if PDF already exists with correct dimensions (cache)
	if hasCachedPdf(pdfPath, bbWidth, bbHeight) {
		return pdfPath, true, nil
	}

	// Find ghostscript
	gsPath, err := findGhostscript()
	if err != nil {
		return "", false, fmt.Errorf("epsconv: %w", err)
	}

	// Convert EPS to PDF with correct dimensions
	cmd := exec.Command(gsPath,
		"-q", "-dNOPAUSE", "-dBATCH",
		"-sDEVICE=pdfwrite",
		fmt.Sprintf("-dDEVICEWIDTHPOINTS=%d", bbWidth),
		fmt.Sprintf("-dDEVICEHEIGHTPOINTS=%d", bbHeight),
		"-dFIXEDMEDIA",
		fmt.Sprintf("-sOutputFile=%s", pdfPath),
		epsPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", false, fmt.Errorf("epsconv: gs failed for %s: %w\n%s", filepath.Base(epsPath), err, string(output))
	}

	return pdfPath, false, nil
}

// ConvertReferencedEps scans a .tex file for \includegraphics references
// and converts any referenced EPS files to PDF. It handles both explicit
// (.pdf extension) and extensionless references.
//
// Returns the number of files converted and any errors encountered.
func ConvertReferencedEps(texPath string) (int, []error) {
	data, err := os.ReadFile(texPath)
	if err != nil {
		return 0, []error{err}
	}

	texDir := filepath.Dir(texPath)
	content := string(data)

	// Find all \includegraphics references
	refs := FindImageReferences(content)

	converted := 0
	var errs []error
	seen := make(map[string]bool)

	for _, ref := range refs {
		// Determine the EPS source path
		epsPath := resolveEpsPath(texDir, ref)
		if epsPath == "" {
			continue // No EPS file found for this reference
		}
		if seen[epsPath] {
			continue
		}
		seen[epsPath] = true

		// Convert EPS to PDF
		pdfPath, cached, err := ConvertEpsToPdf(epsPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if cached {
			fmt.Fprintf(os.Stderr, "src2tex: eps→pdf: %s (cached)\n",
				filepath.Base(epsPath))
		} else {
			fmt.Fprintf(os.Stderr, "src2tex: eps→pdf: %s → %s\n",
				filepath.Base(epsPath), filepath.Base(pdfPath))
			converted++
		}
	}

	return converted, errs
}

// readBoundingBox extracts the BoundingBox dimensions from an EPS file.
func readBoundingBox(epsPath string) (width, height int, err error) {
	f, err := os.Open(epsPath)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	bbRe := regexp.MustCompile(`^%%BoundingBox:\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if m := bbRe.FindStringSubmatch(line); len(m) == 5 {
			x1, _ := strconv.Atoi(m[1])
			y1, _ := strconv.Atoi(m[2])
			x2, _ := strconv.Atoi(m[3])
			y2, _ := strconv.Atoi(m[4])
			w := x2 - x1
			h := y2 - y1
			if w > 0 && h > 0 {
				return w, h, nil
			}
		}
		// Stop searching after PostScript comments end
		if len(line) > 0 && line[0] != '%' {
			break
		}
	}
	return 0, 0, fmt.Errorf("no valid BoundingBox found")
}

// hasCachedPdf checks whether a PDF file already exists with the expected
// MediaBox dimensions (indicating a previous successful conversion).
func hasCachedPdf(pdfPath string, expectedW, expectedH int) bool {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return false
	}
	mbRe := regexp.MustCompile(`/MediaBox\s*\[\s*0\s+0\s+(\d+)\s+(\d+)\s*\]`)
	m := mbRe.FindStringSubmatch(string(data))
	if len(m) != 3 {
		return false
	}
	w, _ := strconv.Atoi(m[1])
	h, _ := strconv.Atoi(m[2])
	return w == expectedW && h == expectedH
}

// findGhostscript locates the gs executable.
func findGhostscript() (string, error) {
	// Try common names
	for _, name := range []string{"gs", "gswin64c", "gswin32c"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("ghostscript (gs) not found in PATH; install TeX Live or Ghostscript")
}

// includegraphicsRe matches \includegraphics[...]{filename} or \includegraphics{filename}
var includegraphicsRe = regexp.MustCompile(`\\includegraphics(?:\[[^\]]*\])?\{([^}]+)\}`)

// FindImageReferences returns all file references from \includegraphics commands.
func FindImageReferences(texContent string) []string {
	matches := includegraphicsRe.FindAllStringSubmatch(texContent, -1)
	var refs []string
	for _, m := range matches {
		if len(m) >= 2 {
			refs = append(refs, m[1])
		}
	}
	return refs
}

// resolveEpsPath determines the EPS source file path for a given
// \includegraphics reference. It handles three cases:
//
//   - "file.pdf"  → look for "file.eps"
//   - "file"      → look for "file.eps" (graphicx auto-resolution)
//   - "file.eps"  → use directly
func resolveEpsPath(texDir, ref string) string {
	ext := filepath.Ext(ref)

	switch ext {
	case ".pdf":
		// Explicit .pdf reference → look for corresponding .eps
		epsPath := filepath.Join(texDir, strings.TrimSuffix(ref, ".pdf")+".eps")
		if fileExists(epsPath) {
			return epsPath
		}
	case ".eps":
		// Direct EPS reference
		epsPath := filepath.Join(texDir, ref)
		if fileExists(epsPath) {
			return epsPath
		}
	case "":
		// No extension → graphicx will search for .pdf, .png, .jpg
		// Check if there's an EPS that needs conversion
		epsPath := filepath.Join(texDir, ref+".eps")
		if fileExists(epsPath) {
			// Also check if PDF already exists (from a previous run)
			pdfPath := filepath.Join(texDir, ref+".pdf")
			if !fileExists(pdfPath) {
				return epsPath
			}
			// PDF exists — still return EPS path so cache check runs
			return epsPath
		}
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
