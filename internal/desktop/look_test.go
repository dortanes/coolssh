package desktop

import (
	"testing"

	"github.com/dortanes/coolssh/internal/preferences"
)

func TestLookResolvesTheStyleForEachPlatform(t *testing.T) {
	tests := []struct {
		name          string
		chosen        preferences.Style
		platform      string
		micaSupported bool
		want          Look
	}{
		{"automatic on macOS is Liquid Glass", preferences.AutomaticStyle, "darwin", false, Look{preferences.LiquidGlassStyle, true}},
		{"automatic on Windows 11 is Mica", preferences.AutomaticStyle, "windows", true, Look{preferences.MicaStyle, true}},
		{"Windows without Mica has no material", preferences.AutomaticStyle, "windows", false, Look{preferences.MicaStyle, false}},
		{"automatic on Linux is shadcn", preferences.AutomaticStyle, "linux", false, Look{preferences.ShadcnStyle, false}},
		{"a style off its platform keeps the platform's material", preferences.MicaStyle, "darwin", false, Look{preferences.MicaStyle, true}},
		{"shadcn on macOS keeps the material it paints over", preferences.ShadcnStyle, "darwin", false, Look{preferences.ShadcnStyle, true}},
		{"Liquid Glass on Linux is solid", preferences.LiquidGlassStyle, "linux", false, Look{preferences.LiquidGlassStyle, false}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := lookFor(test.chosen, test.platform, test.micaSupported); got != test.want {
				t.Errorf("lookFor(%q, %q, %v) = %+v, want %+v", test.chosen, test.platform, test.micaSupported, got, test.want)
			}
		})
	}
}
