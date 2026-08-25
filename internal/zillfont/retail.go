// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"crypto/sha256"
	"fmt"

	"github.com/HK47196/zill/internal/koreanslots"
)

const (
	RetailAtlasMemberSize = 0x80470
	RetailPAFMemberSize   = 0x18e60
	RetailPAFOffset       = 0x4490

	retailAtlasSHA256 = "0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a"
	retailPAFSHA256   = "95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa"
)

func verifyRetailMember(label string, data []byte, size int, wantSHA string) error {
	if len(data) != size {
		return fmt.Errorf("%s size %#x, want %#x", label, len(data), size)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(data))
	if got != wantSHA {
		return fmt.Errorf("unsupported %s fingerprint %s", label, got)
	}
	return nil
}

// ParseAuthenticatedRetailPAF authenticates the complete retail jillbtn.par
// member before exposing its PAF metadata. This prevents a valid-looking PAF
// block copied from a different game/version from entering production planning.
func ParseAuthenticatedRetailPAF(jillbtn []byte) (*PAF, error) {
	if err := verifyRetailMember("2d/font/jillbtn.par", jillbtn, RetailPAFMemberSize, retailPAFSHA256); err != nil {
		return nil, err
	}
	end := RetailPAFOffset + PAFSize
	if end != len(jillbtn) {
		return nil, fmt.Errorf("retail PAF range %#x:%#x does not end at jillbtn.par EOF %#x", RetailPAFOffset, end, len(jillbtn))
	}
	return ParsePAF(jillbtn[RetailPAFOffset:end])
}

// PatchAuthenticatedRetailAtlas is the production font-patching boundary. It
// authenticates both retail font members, resolves the same renderer mapping
// used by Korean message compilation to concrete PAF cells, then returns a
// patched copy of font/zillfont.par. jillbtn.par and PAF metadata remain unchanged.
func PatchAuthenticatedRetailAtlas(zillfontMember, jillbtnMember []byte, mapping koreanslots.Mapping, rasters map[rune]Raster) ([]byte, error) {
	if err := verifyRetailMember("font/zillfont.par", zillfontMember, RetailAtlasMemberSize, retailAtlasSHA256); err != nil {
		return nil, err
	}
	paf, err := ParseAuthenticatedRetailPAF(jillbtnMember)
	if err != nil {
		return nil, err
	}
	replacements, err := paf.ReplacementPlan(mapping)
	if err != nil {
		return nil, err
	}
	patched, err := ApplyRasters(zillfontMember, replacements, rasters)
	if err != nil {
		return nil, err
	}
	return patched, nil
}
