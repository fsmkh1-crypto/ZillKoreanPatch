// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/HK47196/zill/internal/gamefmt/paa"
	"github.com/HK47196/zill/internal/gamefmt/pspiso"
	"github.com/HK47196/zill/internal/koreanfont"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/zillfont"
)

// forensicTargetRunes is deliberately limited to the observed bad glyph in
// message 10007 plus every custom Hangul syllable used by message 10010. This
// branch is diagnostic-only: the helpers below never alter message, font, or ISO
// payloads; they only prove what the already-existing build produced.
func forensicTargetRunes(mapping koreanslots.Mapping) []rune {
	wanted := []rune("게여나는그대가아갈길을묻노라보이고운명의문열어")
	seen := make(map[rune]struct{})
	out := make([]rune, 0, len(wanted))
	for _, r := range wanted {
		if _, ok := mapping[r]; !ok {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// logCompiledKoreanForensics intentionally does not use corpus.ParseBank.
// Compiled Korean records contain renderer-key byte pairs that are not ordinary
// retail CP932 text, so running the retail text decoder over them would make the
// diagnostic itself reject a perfectly valid Korean build. We only need the
// authenticated bank's native uint16 offset table to capture exact record bytes.
func logCompiledKoreanForensics(compiled map[string][]byte) error {
	data, ok := compiled["msgsec001.dat"]
	if !ok {
		return fmt.Errorf("compiled msgsec001.dat is missing")
	}
	if len(data) < 2 {
		return fmt.Errorf("compiled msgsec001.dat is too small")
	}
	count := int(binary.LittleEndian.Uint16(data[:2]))
	tableEnd := 2 + count*2
	if count <= 0 || tableEnd > len(data) {
		return fmt.Errorf("compiled msgsec001.dat has invalid offset table: count=%d size=%d", count, len(data))
	}
	offset := func(index int) (int, error) {
		if index < 0 || index >= count {
			return 0, fmt.Errorf("record index %d out of range 0..%d", index, count-1)
		}
		pos := 2 + index*2
		value := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
		if value < tableEnd || value > len(data) {
			return 0, fmt.Errorf("record index %d has invalid offset %#x", index, value)
		}
		return value, nil
	}
	for _, id := range []int{10007, 10010} {
		index := id % 10_000
		start, err := offset(index)
		if err != nil {
			return fmt.Errorf("ID %d: %w", id, err)
		}
		end := len(data)
		if index+1 < count {
			end, err = offset(index + 1)
			if err != nil {
				return fmt.Errorf("ID %d next offset: %w", id, err)
			}
		}
		if end < start {
			return fmt.Errorf("ID %d has descending record span %#x:%#x", id, start, end)
		}
		raw := data[start:end]
		display := raw
		if terminator := bytes.Index(raw, []byte{5, 5, 5}); terminator >= 0 {
			display = raw[:terminator+3]
		}
		fmt.Printf("FORENSIC COMPILED id=%d index=%d start=%#x end=%#x record_span=%d display_size=%d display_hex=%X raw_hex=%X\n",
			id, index, start, end, len(raw), len(display), display, raw)
	}
	return nil
}

func rasterFingerprint(r zillfont.Raster) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%dx%d:", r.Width, r.Height)
	_, _ = h.Write(r.Pixels)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sameRaster(a, b zillfont.Raster) bool {
	return a.Width == b.Width && a.Height == b.Height && bytes.Equal(a.Pixels, b.Pixels)
}

func parseBuiltPAF(member []byte) (*zillfont.PAF, error) {
	start := zillfont.RetailPAFOffset
	end := start + zillfont.PAFSize
	if start < 0 || end > len(member) {
		return nil, fmt.Errorf("PAF range %#x:%#x outside member %#x", start, end, len(member))
	}
	return zillfont.ParsePAF(member[start:end])
}

func loadForensicExpectedRasters(root string, mapping koreanslots.Mapping) (map[rune]zillfont.Raster, error) {
	data, err := os.ReadFile(filepath.Join(root, "release", "korean", "font", "glyphs.toml"))
	if err != nil {
		return nil, fmt.Errorf("read Korean raster catalog: %w", err)
	}
	catalog, err := koreanfont.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse Korean raster catalog: %w", err)
	}
	out := make(map[rune]zillfont.Raster)
	for _, r := range forensicTargetRunes(mapping) {
		raster, ok := catalog.SourceRaster(r)
		if !ok {
			return nil, fmt.Errorf("Korean raster catalog is missing %U", r)
		}
		out[r] = raster
	}
	return out, nil
}

func logFontPayloadForensics(label, root string, atlasMember, pafMember []byte, mapping koreanslots.Mapping) error {
	paf, err := parseBuiltPAF(pafMember)
	if err != nil {
		return err
	}
	expected, err := loadForensicExpectedRasters(root, mapping)
	if err != nil {
		return err
	}
	byKey := make(map[uint16]zillfont.Glyph, len(paf.Glyphs))
	for _, glyph := range paf.Glyphs {
		byKey[uint16(glyph.Key)] = glyph
	}
	for _, r := range forensicTargetRunes(mapping) {
		key := mapping[r]
		glyph, ok := byKey[uint16(key)]
		if !ok {
			return fmt.Errorf("%s PAF has no glyph for %U key=%04X", label, r, uint16(key))
		}
		actual, err := zillfont.ExtractAtlasCell(atlasMember, glyph)
		if err != nil {
			return fmt.Errorf("extract %s %U: %w", label, r, err)
		}
		want := expected[r]
		fmt.Printf("FORENSIC GLYPH stage=%s rune=%q unicode=U+%04X key=%04X paf_index=%d page=%d x=%d y=%d w=%d h=%d expected_sha=%s actual_sha=%s match=%t\n",
			label, r, r, uint16(key), glyph.Index, glyph.Page, glyph.X, glyph.Y, glyph.Width, glyph.Height,
			rasterFingerprint(want), rasterFingerprint(actual), sameRaster(want, actual))
	}
	return nil
}

func logStagedKoreanFontForensics(root, staging string, mapping koreanslots.Mapping) error {
	pair, err := paa.Open(
		filepath.Join(staging, "USRDIR", "pa.bin"),
		filepath.Join(staging, "USRDIR", "pa.arc"),
	)
	if err != nil {
		return fmt.Errorf("open rebuilt pa archive: %w", err)
	}
	defer pair.Close()
	members := pair.Members()
	if retailPAFMemberIndex >= len(members) {
		return fmt.Errorf("rebuilt pa archive has only %d members", len(members))
	}
	if members[retailAtlasMemberIndex].Name != retailAtlasMemberName || members[retailPAFMemberIndex].Name != retailPAFMemberName {
		return fmt.Errorf("rebuilt pa font member identity drift")
	}
	atlas, err := pair.Payload(retailAtlasMemberIndex)
	if err != nil {
		return err
	}
	paf, err := pair.Payload(retailPAFMemberIndex)
	if err != nil {
		return err
	}
	return logFontPayloadForensics("staged_pa", root, atlas, paf, mapping)
}

func sha256Reader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return sha256Reader(f)
}

func logFinalISOArchiveForensics(outputISO, staging string) error {
	image, err := pspiso.Open(outputISO)
	if err != nil {
		return fmt.Errorf("reopen completed ISO: %w", err)
	}
	defer image.Close()
	payloads := image.PayloadFS()
	for _, name := range []string{"pa.bin", "pa.arc", "pami.bin", "pami.arc"} {
		stagePath := filepath.Join(staging, "USRDIR", name)
		stageSHA, err := sha256File(stagePath)
		if err != nil {
			return fmt.Errorf("hash staged %s: %w", name, err)
		}
		isoFile, err := payloads.Open("PSP_GAME/USRDIR/" + name)
		if err != nil {
			return fmt.Errorf("open completed ISO %s: %w", name, err)
		}
		isoSHA, hashErr := sha256Reader(isoFile)
		closeErr := isoFile.Close()
		if hashErr != nil {
			return fmt.Errorf("hash completed ISO %s: %w", name, hashErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close completed ISO %s: %w", name, closeErr)
		}
		fmt.Printf("FORENSIC ISO_FILE path=PSP_GAME/USRDIR/%s staged_sha=%s iso_sha=%s match=%t\n",
			name, stageSHA, isoSHA, stageSHA == isoSHA)
	}
	return nil
}
