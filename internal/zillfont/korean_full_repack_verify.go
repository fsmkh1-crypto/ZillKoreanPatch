// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"bytes"
	"fmt"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// VerifyFullRepackSemantics proves the semantic invariants expected from the
// Korean mobile full-font transform against the authenticated retail inputs.
//
// Stock glyphs may move to a different atlas position/page, but their renderer
// key, BST links, dimensions, bearings, advance and raster pixels must remain
// identical. Mapped custom slots keep the retail key/BST links but must expose
// exactly the requested Korean raster and target Korean metrics. This verifier
// deliberately re-parses and re-extracts the produced artifacts instead of
// trusting the transform's in-memory placement decisions.
func VerifyFullRepackSemantics(retailAtlas, retailPAF, patchedAtlas, patchedPAF []byte, mapping koreanslots.Mapping, koreanRasters map[rune]Raster) error {
	// The upstream English font transform authenticates the complete output with
	// result_sha256. Korean cannot use one frozen result hash because its renderer
	// mapping is corpus-derived, so enforce the equivalent engine-facing boundary:
	// only atlas image payloads and explicitly modeled PAF geometry/metric fields
	// may differ. Container headers, inter-page bytes, keys, BST links, reserved
	// tails and every other byte remain retail-exact.
	if err := verifyFullRepackContainerMutationSurface(retailAtlas, retailPAF, patchedAtlas, patchedPAF); err != nil {
		return err
	}

	original, err := ParseAuthenticatedRetailPAF(retailPAF)
	if err != nil {
		return fmt.Errorf("verify full repack retail PAF: %w", err)
	}
	if len(patchedPAF) < RetailPAFOffset {
		return fmt.Errorf("verify full repack patched PAF member size %#x is smaller than payload offset %#x", len(patchedPAF), RetailPAFOffset)
	}
	result, err := ParsePAF(patchedPAF[RetailPAFOffset:])
	if err != nil {
		return fmt.Errorf("verify full repack patched PAF: %w", err)
	}
	if len(original.Glyphs) != len(result.Glyphs) {
		return fmt.Errorf("verify full repack glyph count changed: %d -> %d", len(original.Glyphs), len(result.Glyphs))
	}
	if len(mapping) != len(koreanRasters) {
		return fmt.Errorf("verify full repack has %d mappings for %d Korean rasters", len(mapping), len(koreanRasters))
	}

	byKey := make(map[cp932.GlyphKey]rune, len(mapping))
	for r, key := range mapping {
		if prior, exists := byKey[key]; exists {
			return fmt.Errorf("verify full repack maps %U and %U to renderer key 0x%04X", prior, r, uint16(key))
		}
		byKey[key] = r
	}

	matched := make(map[rune]struct{}, len(mapping))
	for index := range original.Glyphs {
		before := original.Glyphs[index]
		after := result.Glyphs[index]
		if before.Index != after.Index || before.Key != after.Key {
			return fmt.Errorf("verify full repack glyph %d identity changed: index/key (%d,0x%04X) -> (%d,0x%04X)", index, before.Index, uint16(before.Key), after.Index, uint16(after.Key))
		}
		if before.Left != after.Left || before.Right != after.Right {
			return fmt.Errorf("verify full repack glyph %d key 0x%04X BST links changed: (%d,%d) -> (%d,%d)", index, uint16(before.Key), before.Left, before.Right, after.Left, after.Right)
		}

		if r, custom := byKey[before.Key]; custom {
			expected, ok := koreanRasters[r]
			if !ok {
				return fmt.Errorf("verify full repack missing expected raster for %U", r)
			}
			if int(after.Width) != expected.Width || int(after.Height) != expected.Height {
				return fmt.Errorf("verify full repack custom %U key 0x%04X dimensions are %dx%d, want %dx%d", r, uint16(after.Key), after.Width, after.Height, expected.Width, expected.Height)
			}
			if after.BearingX != KoreanTargetBearingX || after.BearingY != KoreanTargetBearingY || after.Advance != KoreanTargetAdvance {
				return fmt.Errorf("verify full repack custom %U key 0x%04X metrics are bearing=(%d,%d) advance=%d, want (%d,%d) advance=%d", r, uint16(after.Key), after.BearingX, after.BearingY, after.Advance, KoreanTargetBearingX, KoreanTargetBearingY, KoreanTargetAdvance)
			}
			actual, err := ExtractAtlasCell(patchedAtlas, after)
			if err != nil {
				return fmt.Errorf("verify full repack custom %U raster: %w", r, err)
			}
			if !sameRaster(actual, expected) {
				return fmt.Errorf("verify full repack custom %U key 0x%04X raster differs from requested Korean raster", r, uint16(after.Key))
			}
			matched[r] = struct{}{}
			continue
		}

		if before.Width != after.Width || before.Height != after.Height || before.BearingX != after.BearingX || before.BearingY != after.BearingY || before.Advance != after.Advance {
			return fmt.Errorf("verify full repack stock glyph %d key 0x%04X metrics changed", index, uint16(before.Key))
		}
		if before.Width == 0 || before.Height == 0 {
			continue
		}
		beforeRaster, err := ExtractAtlasCell(retailAtlas, before)
		if err != nil {
			return fmt.Errorf("verify full repack stock glyph %d retail raster: %w", index, err)
		}
		afterRaster, err := ExtractAtlasCell(patchedAtlas, after)
		if err != nil {
			return fmt.Errorf("verify full repack stock glyph %d patched raster: %w", index, err)
		}
		if !sameRaster(beforeRaster, afterRaster) {
			return fmt.Errorf("verify full repack stock glyph %d key 0x%04X raster changed", index, uint16(before.Key))
		}
	}

	if len(matched) != len(mapping) {
		return fmt.Errorf("verify full repack matched %d/%d custom mappings", len(matched), len(mapping))
	}
	return nil
}

func verifyFullRepackContainerMutationSurface(retailAtlas, retailPAF, patchedAtlas, patchedPAF []byte) error {
	if len(patchedAtlas) != len(retailAtlas) {
		return fmt.Errorf("verify full repack atlas member size changed: %#x -> %#x", len(retailAtlas), len(patchedAtlas))
	}
	if len(patchedPAF) != len(retailPAF) {
		return fmt.Errorf("verify full repack PAF member size changed: %#x -> %#x", len(retailPAF), len(patchedPAF))
	}

	atlasMutable := make([]bool, len(retailAtlas))
	pageBytes := AtlasSize * AtlasSize / 2
	for page, startBase := range gimStarts {
		start := startBase + imageDataOffset
		end := start + pageBytes
		if start < 0 || end > len(atlasMutable) {
			return fmt.Errorf("verify full repack atlas page %d mutable payload [%#x,%#x) is outside member size %#x", page, start, end, len(atlasMutable))
		}
		for offset := start; offset < end; offset++ {
			atlasMutable[offset] = true
		}
	}
	for offset := range retailAtlas {
		if !atlasMutable[offset] && retailAtlas[offset] != patchedAtlas[offset] {
			return fmt.Errorf("verify full repack changed immutable atlas/container byte %#x", offset)
		}
	}

	pafMutable := make([]bool, len(retailPAF))
	for index := 0; index < GlyphCount; index++ {
		record := RetailPAFOffset + RecordOffset + index*RecordStride
		if record < 0 || record+RecordStride > len(pafMutable) {
			return fmt.Errorf("verify full repack PAF record %d is outside member size %#x", index, len(pafMutable))
		}
		// width, height, x, y, bearing x/y and advance
		for offset := record + 2; offset < record+0x10; offset++ {
			pafMutable[offset] = true
		}
		// page index. Key, BST links and reserved tail are deliberately immutable.
		for offset := record + 0x18; offset < record+0x1c; offset++ {
			pafMutable[offset] = true
		}
	}
	for offset := range retailPAF {
		if !pafMutable[offset] && retailPAF[offset] != patchedPAF[offset] {
			return fmt.Errorf("verify full repack changed immutable PAF/container byte %#x", offset)
		}
	}
	return nil
}

func sameRaster(left, right Raster) bool {
	return left.Width == right.Width && left.Height == right.Height && bytes.Equal(left.Pixels, right.Pixels)
}
