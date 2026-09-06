// SPDX-License-Identifier: GPL-3.0-or-later

package koreanfont

import "testing"

func TestBeta1BProfileAlphaQuantization(t *testing.T) {
	if KoreanAlphaGamma != 0.60 {
		t.Fatalf("KoreanAlphaGamma=%v, want 0.60", KoreanAlphaGamma)
	}
	if ProvenRenderRule != "opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2" {
		t.Fatalf("unexpected Beta 1 render rule: %q", ProvenRenderRule)
	}

	cases := []struct {
		alpha uint8
		want  uint8
	}{
		{0, 0},
		{1, 1},
		{32, 4},
		{64, 7},
		{128, 10},
		{192, 13},
		{255, 15},
	}
	for _, tc := range cases {
		if got := alphaTo4BPP(tc.alpha); got != tc.want {
			t.Errorf("alphaTo4BPP(%d)=%d, want %d", tc.alpha, got, tc.want)
		}
	}
}
