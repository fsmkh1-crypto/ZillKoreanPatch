// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"os"
	"strings"
	"testing"
)

func TestProductionBuildDoesNotCallKoreanAlphaHelpers(t *testing.T) {
	data, err := os.ReadFile("build.go")
	if err != nil {
		t.Fatalf("read production build.go: %v", err)
	}
	text := string(data)
	for _, forbidden := range []string{
		"BuildKoreanAlphaISOOnly",
		"BuildKoreanAlphaPlaceholderProject",
		"compileKoreanBanksWithPlan",
		"prepareKoreanMobileFontReplacements",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("production build.go references alpha-only helper %q", forbidden)
		}
	}
}
