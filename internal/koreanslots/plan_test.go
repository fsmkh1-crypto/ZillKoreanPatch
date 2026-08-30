// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"reflect"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func mustPlanKey(t *testing.T, text string) cp932.GlyphKey {
	t.Helper()
	encoded, err := cp932.Encode(text)
	if err != nil {
		t.Fatal(err)
	}
	key, err := cp932.GlyphKeyFromBytes(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestBuildPlanPreservesRuntimeStockAndReservations(t *testing.T) {
	stock := mustPlanKey(t, "日")
	reserved := mustPlanKey(t, "本")
	free := mustPlanKey(t, "語")

	plan, err := BuildPlan([]string{"日한"}, []cp932.GlyphKey{free, stock, reserved}, []cp932.GlyphKey{reserved})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.CustomRunes, []rune{'한'}) {
		t.Fatalf("custom runes = %#v", plan.CustomRunes)
	}
	if !reflect.DeepEqual(plan.RequiredStock, []cp932.GlyphKey{stock}) {
		t.Fatalf("required stock = %#v", plan.RequiredStock)
	}
	if !reflect.DeepEqual(plan.Candidates, []cp932.GlyphKey{free}) {
		t.Fatalf("candidates = %#v", plan.Candidates)
	}
	if got := plan.Mapping['한']; got != free {
		t.Fatalf("mapping = 0x%04X, want 0x%04X", uint16(got), uint16(free))
	}
}

func TestBuildPlanHasNoWholeBlobOwnershipInput(t *testing.T) {
	first := mustPlanKey(t, "本")
	second := mustPlanKey(t, "語")
	plan, err := BuildPlan([]string{"한"}, []cp932.GlyphKey{second, first}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Both installed keys remain eligible. Whole-blob byte aliases are audited
	// separately and cannot silently remove renderer slots from production.
	if !reflect.DeepEqual(plan.Candidates, []cp932.GlyphKey{first, second}) {
		t.Fatalf("candidates = %#v", plan.Candidates)
	}
}

func TestBuildPlanFailsWhenSafeCapacityIsInsufficient(t *testing.T) {
	key := mustPlanKey(t, "本")
	if _, err := BuildPlan([]string{"한글"}, []cp932.GlyphKey{key}, nil); err == nil {
		t.Fatal("BuildPlan unexpectedly succeeded")
	}
}

func TestBuildPlanIsDeterministicAcrossInputOrderAndDuplicates(t *testing.T) {
	a := mustPlanKey(t, "本")
	b := mustPlanKey(t, "語")
	left, err := BuildPlan([]string{"글한"}, []cp932.GlyphKey{b, a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	right, err := BuildPlan([]string{"한글"}, []cp932.GlyphKey{a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("plans differ:\nleft=%#v\nright=%#v", left, right)
	}
}
