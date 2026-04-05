package font

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Install downloads and installs a code font by name into fontDir.
// Use "all" to install all downloadable fonts.
// fontDir defaults to DefaultFontDir() if empty.
func Install(name, fontDir string) error {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}

	if name == "all" {
		var lastErr error
		for _, fd := range BuiltinFonts {
			if fd.GithubRepo == "" {
				continue // skip TeX Live bundled (ipaex)
			}
			fmt.Fprintf(os.Stderr, "Installing %s...\n", fd.DisplayName)
			if err := Install(fd.Name, fontDir); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
				lastErr = err
			}
		}
		return lastErr
	}

	fd := GetFontDef(name)
	if fd == nil {
		return fmt.Errorf("unknown font: %s (run 'src2tex font list' to see available fonts)", name)
	}
	if fd.GithubRepo == "" {
		fmt.Fprintf(os.Stderr, "%s is bundled with TeX Live. No download needed.\n", fd.DisplayName)
		return nil
	}
	if IsFontInstalled(name, fontDir) {
		fmt.Fprintf(os.Stderr, "%s is already installed.\n", fd.DisplayName)
		return nil
	}

	destDir := filepath.Join(fontDir, name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %v", err)
	}

	if err := downloadAndExtract(fd.GithubRepo, fd.ZipAsset, destDir, fd.RegularFile, fd.BoldFile); err != nil {
		return err
	}
	// Update fonts.json to reflect the newly installed font.
	_ = SaveFontsConfig(fontDir)
	return nil
}

// InstallCommentFont downloads and installs a comment (mincho) font by name into fontDir.
// Use "all" to install all downloadable comment fonts.
func InstallCommentFont(name, fontDir string) error {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}

	if name == "all" {
		var lastErr error
		for _, cfd := range BuiltinCommentFonts {
			if cfd.TexLive || cfd.GithubRepo == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "Installing %s...\n", cfd.DisplayName)
			if err := InstallCommentFont(cfd.Name, fontDir); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
				lastErr = err
			}
		}
		return lastErr
	}

	cfd := GetCommentFontDef(name)
	if cfd == nil {
		return fmt.Errorf("unknown comment font: %s (run 'src2tex font list' to see available fonts)", name)
	}
	if cfd.TexLive {
		fmt.Fprintf(os.Stderr, "%s is bundled with TeX Live. No download needed.\n", cfd.DisplayName)
		return nil
	}
	if cfd.GithubRepo == "" {
		return fmt.Errorf("font %s has no download source configured", name)
	}
	if IsCommentFontInstalled(name, fontDir) {
		fmt.Fprintf(os.Stderr, "%s is already installed.\n", cfd.DisplayName)
		return nil
	}

	destDir := filepath.Join(fontDir, "comment-"+name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %v", err)
	}

	if err := downloadAndExtract(cfd.GithubRepo, cfd.ZipAsset, destDir, cfd.RegularFile, cfd.BoldFile); err != nil {
		return err
	}
	// Update fonts.json to reflect the newly installed font.
	_ = SaveFontsConfig(fontDir)
	return nil
}

// downloadAndExtract fetches the latest GitHub release ZIP asset and extracts font files.
func downloadAndExtract(githubRepo, zipAssetSubstr, destDir, regularFile, boldFile string) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	fmt.Fprintf(os.Stderr, "Fetching release info from %s...\n", githubRepo)

	resp, err := http.Get(apiURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	var downloadURL, assetName string
	for _, a := range release.Assets {
		if strings.Contains(a.Name, zipAssetSubstr) {
			downloadURL = a.BrowserDownloadURL
			assetName = a.Name
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no asset matching %q found in release %s", zipAssetSubstr, release.TagName)
	}

	fmt.Fprintf(os.Stderr, "Downloading %s (%s)...\n", assetName, release.TagName)

	zipResp, err := http.Get(downloadURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer zipResp.Body.Close()

	tmpFile, err := os.CreateTemp("", "src2tex-font-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	written, err := io.Copy(tmpFile, zipResp.Body)
	if err != nil {
		return fmt.Errorf("download interrupted: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Downloaded %.1f MB\n", float64(written)/1024/1024)

	return extractFontFiles(tmpFile.Name(), destDir, regularFile, boldFile)
}

// extractFontFiles extracts .ttf and .otf files from a ZIP archive into destDir.
func extractFontFiles(zipPath, destDir, regularFile, boldFile string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %v", err)
	}
	defer r.Close()

	extracted := 0
	for _, f := range r.File {
		lower := strings.ToLower(f.Name)
		if !strings.HasSuffix(lower, ".ttf") && !strings.HasSuffix(lower, ".otf") {
			continue
		}
		baseName := filepath.Base(f.Name)
		destPath := filepath.Join(destDir, baseName)

		rc, err := f.Open()
		if err != nil {
			continue
		}
		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			continue
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			continue
		}
		extracted++
	}

	if extracted == 0 {
		return fmt.Errorf("no font files found in archive")
	}

	// Verify the required regular file is present.
	if _, err := os.Stat(filepath.Join(destDir, regularFile)); err != nil {
		// Print what was extracted to help diagnose
		fmt.Fprintf(os.Stderr, "Warning: expected file %s not found in archive.\n", regularFile)
		fmt.Fprintf(os.Stderr, "Extracted %d font files to %s\n", extracted, destDir)
		entries, _ := os.ReadDir(destDir)
		for _, e := range entries {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Name())
		}
		return fmt.Errorf("required font file not found: %s", regularFile)
	}

	fmt.Fprintf(os.Stderr, "Installed %d font file(s) to %s\n", extracted, destDir)
	return nil
}
