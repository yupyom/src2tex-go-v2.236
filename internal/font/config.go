// Package font — config.go handles loading and saving font definitions via ~/.src2tex/fonts.json.
//
// The fonts.json file serves as both a human-readable configuration and a
// reference for users to see which fonts are available and where they are stored.
// It is automatically generated/updated when fonts are installed or when
// SaveFontsConfig is called, and includes all builtin and comment font definitions
// along with their resolved paths on the current system.
//
// Entries managed by the system have "auto": true. When fonts.json is regenerated,
// only auto-managed entries are overwritten. User-edited entries (where "auto" is
// false or absent) are always preserved, even for builtin font names.
package font

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FontConfigFile represents the ~/.src2tex/fonts.json file structure.
type FontConfigFile struct {
	Comment      string            `json:"_comment"`
	CodeFonts    []FontConfigEntry `json:"code_fonts"`
	CommentFonts []FontConfigEntry `json:"comment_fonts"`
}

// FontConfigEntry is a single font entry in fonts.json, designed to be
// human-readable and editable.
type FontConfigEntry struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	License     string `json:"license,omitempty"`
	Unified     bool   `json:"unified,omitempty"`
	RegularFile string `json:"regular_file"`
	BoldFile    string `json:"bold_file,omitempty"`
	FontDir     string `json:"font_dir,omitempty"` // resolved absolute path; empty = ~/.src2tex/fonts/{name}/
	TexLive     bool   `json:"texlive,omitempty"`
	Installed   bool   `json:"installed"`
	Auto        bool   `json:"auto,omitempty"` // true = system-managed, regenerated on init; false/absent = user-edited, preserved
	Description string `json:"description,omitempty"`
}

// configPath returns the path to ~/.src2tex/fonts.json.
func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".src2tex", "fonts.json")
}

// LoadCustomFonts loads user-defined font definitions from ~/.src2tex/fonts.json.
// Returns nil if the file does not exist or is invalid.
// This reads the code_fonts section and converts entries to FontDef.
func LoadCustomFonts() []FontDef {
	path := configPath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var config FontConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	var result []FontDef
	for _, entry := range config.CodeFonts {
		// Skip builtin fonts that are already in BuiltinFonts.
		if GetFontDef(entry.Name) != nil {
			continue
		}
		fd := FontDef{
			Name:        entry.Name,
			DisplayName: entry.DisplayName,
			License:     entry.License,
			Unified:     entry.Unified,
			Description: entry.Description,
			RegularFile: entry.RegularFile,
			BoldFile:    entry.BoldFile,
		}
		result = append(result, fd)
	}
	return result
}

// LookupFont finds a font definition by name.
// Search order: BuiltinFonts → fonts.json custom fonts.
func LookupFont(name string) *FontDef {
	if fd := GetFontDef(name); fd != nil {
		return fd
	}
	for _, cf := range LoadCustomFonts() {
		if cf.Name == name {
			result := cf
			return &result
		}
	}
	return nil
}

// SaveFontsConfig generates ~/.src2tex/fonts.json with all builtin and installed fonts.
// This provides a human-readable reference for users to see and customize their font setup.
//
// Regeneration policy:
//   - Entries with "auto": true are system-managed and will be regenerated from
//     builtin definitions (font_dir, installed status, etc. are refreshed).
//   - Entries without "auto" (or "auto": false) are user-edited and are always
//     preserved exactly as-is, even if their name matches a builtin font.
//   - This allows users to override a builtin font's font_dir by setting "auto": false.
func SaveFontsConfig(fontDir string) error {
	if fontDir == "" {
		fontDir = DefaultFontDir()
	}

	path := configPath()
	if path == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	// Load existing entries to preserve user-edited ones.
	existing := loadExistingConfig(path)

	// Collect user-edited entries (auto == false) keyed by name.
	userCodeFonts := make(map[string]FontConfigEntry)
	userCommentFonts := make(map[string]FontConfigEntry)
	if existing != nil {
		for _, entry := range existing.CodeFonts {
			if !entry.Auto {
				userCodeFonts[entry.Name] = entry
			}
		}
		for _, entry := range existing.CommentFonts {
			if !entry.Auto {
				userCommentFonts[entry.Name] = entry
			}
		}
	}

	// Build code font entries from BuiltinFonts.
	var codeFonts []FontConfigEntry
	for _, fd := range BuiltinFonts {
		// If user has a non-auto override for this font, keep theirs.
		if userEntry, ok := userCodeFonts[fd.Name]; ok {
			codeFonts = append(codeFonts, userEntry)
			delete(userCodeFonts, fd.Name) // consumed
			continue
		}
		entry := FontConfigEntry{
			Name:        fd.Name,
			DisplayName: fd.DisplayName,
			License:     fd.License,
			Unified:     fd.Unified,
			RegularFile: fd.RegularFile,
			BoldFile:    fd.BoldFile,
			Auto:        true,
			Description: fd.Description,
		}
		if fd.Name == "ipaex" {
			entry.TexLive = true
			entry.Installed = isTexLiveAvailable()
			if dir, ok := findTexLiveFont(fd.RegularFile); ok {
				entry.FontDir = dir
			}
		} else {
			dir := filepath.Join(fontDir, fd.Name)
			entry.FontDir = dir
			entry.Installed = IsFontInstalled(fd.Name, fontDir)
		}
		codeFonts = append(codeFonts, entry)
	}
	// Append remaining user-edited code fonts (not matching any builtin).
	for _, entry := range userCodeFonts {
		codeFonts = append(codeFonts, entry)
	}

	// Build comment font entries from BuiltinCommentFonts.
	var commentFonts []FontConfigEntry
	for _, cfd := range BuiltinCommentFonts {
		// If user has a non-auto override, keep theirs.
		if userEntry, ok := userCommentFonts[cfd.Name]; ok {
			commentFonts = append(commentFonts, userEntry)
			delete(userCommentFonts, cfd.Name)
			continue
		}
		entry := FontConfigEntry{
			Name:        cfd.Name,
			DisplayName: cfd.DisplayName,
			License:     cfd.License,
			RegularFile: cfd.RegularFile,
			BoldFile:    cfd.BoldFile,
			Auto:        true,
			Description: cfd.Description,
		}
		if cfd.TexLive {
			entry.TexLive = true
			if dir, ok := findTexLiveFont(cfd.RegularFile); ok {
				entry.FontDir = dir
				entry.Installed = true
			}
		} else {
			dir := filepath.Join(fontDir, "comment-"+cfd.Name)
			entry.FontDir = dir
			entry.Installed = IsCommentFontInstalled(cfd.Name, fontDir)
		}
		commentFonts = append(commentFonts, entry)
	}
	// Append remaining user-edited comment fonts.
	for _, entry := range userCommentFonts {
		commentFonts = append(commentFonts, entry)
	}

	config := FontConfigFile{
		Comment:      "src2tex font configuration. Entries with \"auto\": true are system-managed and will be regenerated. Remove \"auto\" or set it to false to prevent overwriting your edits.",
		CodeFonts:    codeFonts,
		CommentFonts: commentFonts,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal fonts.json: %w", err)
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write fonts.json: %w", err)
	}

	fmt.Fprintf(os.Stderr, "src2tex: font config saved to %s\n", path)
	return nil
}

// loadExistingConfig reads an existing fonts.json file.
func loadExistingConfig(path string) *FontConfigFile {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var config FontConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	return &config
}
