package emitter

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates
var embeddedTemplates embed.FS

// EngineConfig describes a TeX engine's configuration.
type EngineConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	CompileCmd       string `json:"compile_cmd"`
	SupportsCJK      bool   `json:"supports_cjk"`
	SupportsFontspec *bool  `json:"supports_fontspec,omitempty"` // nil = true (default)
	PreambleTemplate string `json:"preamble_template"`
	Auto             *bool  `json:"auto,omitempty"` // true = system-managed, regenerated on init; false/nil = user-edited, preserved
}

// HasFontspec returns true if the engine supports fontspec (\setmonofont etc.).
// Defaults to true when not explicitly set.
func (c *EngineConfig) HasFontspec() bool {
	if c.SupportsFontspec == nil {
		return true
	}
	return *c.SupportsFontspec
}

// IsAuto returns true if the engine is system-managed (auto=true).
// When Auto is nil (not set in JSON), this defaults to false (user-edited).
func (c *EngineConfig) IsAuto() bool {
	if c.Auto == nil {
		return false
	}
	return *c.Auto
}

// PreambleData holds all variables available to preamble templates.
//
// fontspec では、欧文系 (\setmainfont, \setsansfont, \setmonofont) と
// CJK 系 (\setCJKmainfont, \setCJKsansfont, \setCJKmonofont) が対になる。
// -font / -commentfont 指定時は両方を出力する必要がある。
type PreambleData struct {
	PaperSize        string // e.g. "a4paper", "b5paper", "letterpaper"
	Margin           string // e.g. "20mm", "15mm", "0.75in"
	// Latin font commands (fontspec)
	MainFontLine     string // \setmainfont[Path=...]{...} — comment font (serif); empty = not set
	SansFontLine     string // \setsansfont[Path=...]{...} — code font (sans); empty = not set
	MonoFontLine     string // \setmonofont[Path=...]{...} — code font (mono)
	// CJK font commands (xeCJK / luatexja-fontspec)
	CJKMainFont      string // full \setCJKmainfont{...} or \setmainjfont{...} command
	CJKSansFont      string // font name for \setCJKsansfont (fallback)
	CJKSansFontLine  string // \setCJKsansfont[Path=...]{...} for unified fonts; empty = use CJKSansFont
	CJKMonoFont      string // font name for \setCJKmonofont (fallback)
	CJKMonoFontLine  string // \setCJKmonofont[Path=...]{...} for unified fonts; empty = use CJKMonoFont
	Header           string // \fancyhead[R]{...}
	Footer           string // \fancyfoot[R]{...}
}

// userEnginesDir returns ~/.src2tex/engines/.
func userEnginesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".src2tex", "engines")
}

// LoadEngine loads an engine config and its preamble template.
// It first looks in ~/.src2tex/engines/<name>/, then falls back to
// the embedded defaults.
func LoadEngine(name string) (*EngineConfig, *template.Template, error) {
	// Try user directory first.
	userDir := filepath.Join(userEnginesDir(), name)
	if info, err := os.Stat(userDir); err == nil && info.IsDir() {
		cfg, tmpl, err := loadEngineFromDir(userDir, name)
		if err == nil {
			return cfg, tmpl, nil
		}
		// Fall through to embedded on error.
		fmt.Fprintf(os.Stderr, "src2tex: warning: user engine %q has errors, using built-in: %v\n", name, err)
	}

	// Fall back to embedded templates.
	return loadEngineFromFS(embeddedTemplates, "templates/"+name, name)
}

func loadEngineFromDir(dir, name string) (*EngineConfig, *template.Template, error) {
	cfgPath := filepath.Join(dir, "engine.json")
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read engine.json: %w", err)
	}
	var cfg EngineConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parse engine.json: %w", err)
	}
	tmplFile := cfg.PreambleTemplate
	if tmplFile == "" {
		tmplFile = "preamble.tmpl"
	}
	tmplPath := filepath.Join(dir, tmplFile)
	tmplData, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read template %s: %w", tmplFile, err)
	}
	tmpl, err := template.New(name).Delims("<%", "%>").Parse(string(tmplData))
	if err != nil {
		return nil, nil, fmt.Errorf("parse template: %w", err)
	}
	return &cfg, tmpl, nil
}

func loadEngineFromFS(fsys fs.FS, dir, name string) (*EngineConfig, *template.Template, error) {
	cfgData, err := fs.ReadFile(fsys, dir+"/engine.json")
	if err != nil {
		return nil, nil, fmt.Errorf("unknown engine %q: %w", name, err)
	}
	var cfg EngineConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parse engine.json for %q: %w", name, err)
	}
	tmplFile := cfg.PreambleTemplate
	if tmplFile == "" {
		tmplFile = "preamble.tmpl"
	}
	tmplData, err := fs.ReadFile(fsys, dir+"/"+tmplFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read template for %q: %w", name, err)
	}
	tmpl, err := template.New(name).Delims("<%", "%>").Parse(string(tmplData))
	if err != nil {
		return nil, nil, fmt.Errorf("parse template for %q: %w", name, err)
	}
	return &cfg, tmpl, nil
}

// isBuiltinEngine returns true if the engine name matches an embedded engine.
func isBuiltinEngine(name string) bool {
	entries, err := fs.ReadDir(embeddedTemplates, "templates")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == name {
			return true
		}
	}
	return false
}

// InitEngines extracts all embedded engine templates to ~/.src2tex/engines/.
//
// Regeneration policy (same as fonts.json):
//   - Engines with "auto": true in engine.json are system-managed and their
//     files (engine.json + preamble.tmpl) will be overwritten with the latest
//     built-in versions.
//   - Engines without "auto" (or "auto": false) are user-edited and are never
//     overwritten, even if they share a name with a built-in engine.
//   - User-created custom engines (e.g. "my-lualatex") are always preserved.
//   - New built-in engines not yet present in the user directory are added
//     with "auto": true.
func InitEngines() error {
	destDir := userEnginesDir()
	if destDir == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	// Get list of built-in engines.
	entries, err := fs.ReadDir(embeddedTemplates, "templates")
	if err != nil {
		return fmt.Errorf("read embedded templates: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		engineName := e.Name()
		engineDir := filepath.Join(destDir, engineName)

		// Check if user engine already exists.
		cfgPath := filepath.Join(engineDir, "engine.json")
		if _, err := os.Stat(cfgPath); err == nil {
			// engine.json exists — check if it's auto-managed.
			cfgData, readErr := os.ReadFile(cfgPath)
			if readErr == nil {
				var cfg EngineConfig
				if json.Unmarshal(cfgData, &cfg) == nil && !cfg.IsAuto() {
					// User-edited engine: skip entirely.
					fmt.Fprintf(os.Stderr, "  %-12s (user-edited, skipped)\n", engineName)
					continue
				}
			}
			// auto=true or parse error: overwrite with latest.
		}

		// Extract (or re-extract) this built-in engine.
		if err := os.MkdirAll(engineDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", engineDir, err)
		}

		// Walk embedded files for this engine.
		embeddedDir := "templates/" + engineName
		embeddedEntries, err := fs.ReadDir(embeddedTemplates, embeddedDir)
		if err != nil {
			continue
		}
		for _, fe := range embeddedEntries {
			if fe.IsDir() {
				continue
			}
			data, err := fs.ReadFile(embeddedTemplates, embeddedDir+"/"+fe.Name())
			if err != nil {
				continue
			}

			// For engine.json, inject "auto": true into the JSON.
			if fe.Name() == "engine.json" {
				data = injectAutoFlag(data)
			}

			destPath := filepath.Join(engineDir, fe.Name())
			if err := os.WriteFile(destPath, data, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", destPath, err)
			}
		}
		fmt.Fprintf(os.Stderr, "  %-12s (updated)\n", engineName)
	}

	return nil
}

// injectAutoFlag adds "auto": true to an engine.json JSON object.
func injectAutoFlag(data []byte) []byte {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data // return as-is on parse error
	}
	raw["auto"] = true
	result, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return data
	}
	return append(result, '\n')
}

// ListEngines returns the names of available engines.
// Combines embedded engines with any user-defined ones.
func ListEngines() []string {
	seen := make(map[string]bool)
	var names []string

	// Embedded engines.
	entries, _ := fs.ReadDir(embeddedTemplates, "templates")
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
			seen[e.Name()] = true
		}
	}

	// User engines (including custom ones).
	userDir := userEnginesDir()
	if userDir != "" {
		entries, _ := os.ReadDir(userDir)
		for _, e := range entries {
			if e.IsDir() && !seen[e.Name()] {
				names = append(names, e.Name())
			}
		}
	}

	return names
}

// IsCustomEngine returns true if the engine is a user-created custom engine
// (not a built-in, or a built-in with auto=false).
func IsCustomEngine(name string) bool {
	if !isBuiltinEngine(name) {
		return true
	}
	userDir := filepath.Join(userEnginesDir(), name)
	cfgPath := filepath.Join(userDir, "engine.json")
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return false
	}
	var cfg EngineConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return false
	}
	return !cfg.IsAuto()
}

// MarginForPaper returns the appropriate margin for the given paper size.
func MarginForPaper(paper string) string {
	switch paper {
	case "b5paper", "b5":
		return "15mm"
	case "letterpaper", "letter":
		return "0.75in"
	default: // a4paper, a4, etc.
		return "20mm"
	}
}

// NormalizePaperSize converts short names to LaTeX paper size names.
func NormalizePaperSize(paper string) string {
	switch paper {
	case "a4":
		return "a4paper"
	case "b5":
		return "b5paper"
	case "letter":
		return "letterpaper"
	default:
		return paper
	}
}
