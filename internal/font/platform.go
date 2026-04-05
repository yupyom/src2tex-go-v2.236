package font

import "runtime"

// PlatformDefaults holds OS-specific default font names for preamble generation.
type PlatformDefaults struct {
	MonoFont    string   // \setmonofont default
	CJKSansFont string   // \setCJKsansfont default
	CJKMonoFont string   // \setCJKmonofont default
	MonoFallbacks    []string
	CJKSansFallbacks []string
	CJKMonoFallbacks []string
}

// GetPlatformDefaults returns font defaults appropriate for the current OS.
func GetPlatformDefaults() PlatformDefaults {
	switch runtime.GOOS {
	case "darwin":
		return PlatformDefaults{
			MonoFont:    "Courier New",
			CJKSansFont: "Hiragino Sans W3",
			CJKMonoFont: "Hiragino Sans W3",
		}
	case "windows":
		return PlatformDefaults{
			MonoFont:    "Consolas",
			CJKSansFont: "Yu Gothic",
			CJKMonoFont: "MS Gothic",
			CJKSansFallbacks: []string{"MS Gothic", "Meiryo"},
		}
	default: // linux and others
		return PlatformDefaults{
			MonoFont:    "DejaVu Sans Mono",
			CJKSansFont: "IPAexGothic",
			CJKMonoFont: "IPAexGothic",
			MonoFallbacks:    []string{"Liberation Mono", "Courier New"},
			CJKSansFallbacks: []string{"Noto Sans CJK JP"},
		}
	}
}
