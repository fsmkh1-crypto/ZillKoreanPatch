// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/gamefmt/paa"
	"github.com/HK47196/zill/internal/zillfont"
)

const (
	jillBtnMemberIndex = 13612
	jillBtnMemberSize  = 0x18e60
	pafMemberOffset    = 0x4490
)

func loadRetailPAF(gameDir string) (*zillfont.PAF, error) {
	indexPath := filepath.Join(gameDir, "USRDIR", "pa.bin")
	archivePath := filepath.Join(gameDir, "USRDIR", "pa.arc")
	pair, err := paa.Open(indexPath, archivePath)
	if err != nil {
		return nil, err
	}
	defer pair.Close()

	members := pair.Members()
	if len(members) != 14231 {
		return nil, fmt.Errorf("PAA member count %d, want 14231", len(members))
	}
	member := members[jillBtnMemberIndex]
	if member.Name != "2d/font/jillbtn.par" {
		return nil, fmt.Errorf("PAA member %d is %q, want %q", jillBtnMemberIndex, member.Name, "2d/font/jillbtn.par")
	}
	if member.Size != jillBtnMemberSize {
		return nil, fmt.Errorf("%s member size %#x, want %#x", member.Name, member.Size, jillBtnMemberSize)
	}
	payload, err := pair.Payload(jillBtnMemberIndex)
	if err != nil {
		return nil, err
	}
	end := pafMemberOffset + zillfont.PAFSize
	if end != len(payload) {
		return nil, fmt.Errorf("PAF range %#x:%#x does not end at %s EOF (%#x bytes)", pafMemberOffset, end, member.Name, len(payload))
	}
	return zillfont.ParsePAF(payload[pafMemberOffset:end])
}

func runFontStatus(args []string, stdout, stderr io.Writer) int {
	gameDir, ok := parseRequiredGameDir("font-status", args, stderr)
	if !ok {
		return 2
	}
	font, err := loadRetailPAF(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: font-status: %v\n", err)
		return 1
	}
	pages := font.PageCounts()
	fmt.Fprintf(stdout, "Target font: ULJM-05410 v1.03\n")
	fmt.Fprintf(stdout, "Glyph slots: %d\n", len(font.Glyphs))
	fmt.Fprintf(stdout, "Two-byte CP932 renderer slots: %d\n", len(font.DoubleByteKeys()))
	fmt.Fprintf(stdout, "BST root: %d\n", zillfont.BSTRoot)
	fmt.Fprintf(stdout, "PAF member-relative range: %#x:%#x\n", pafMemberOffset, pafMemberOffset+zillfont.PAFSize)
	fmt.Fprintf(stdout, "Pages: 0=%d 1=%d 2=%d 3=%d\n", pages[0], pages[1], pages[2], pages[3])
	return 0
}

func collectRendererKeys(id int, label, text string, used map[cp932.GlyphKey]struct{}) error {
	for _, r := range text {
		encoded, err := cp932.Encode(string(r))
		if err != nil {
			return fmt.Errorf("%s ID %d contains non-CP932 rune %U: %w", label, id, r, err)
		}
		key, err := cp932.GlyphKeyFromBytes(encoded)
		if err != nil {
			return fmt.Errorf("%s ID %d rune %U: %w", label, id, r, err)
		}
		used[key] = struct{}{}
	}
	return nil
}

func runKoreanSlots(root string, args []string, stdout, stderr io.Writer) int {
	gameDir, ok := parseRequiredGameDir("korean-slots", args, stderr)
	if !ok {
		return 2
	}
	font, err := loadRetailPAF(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	project, _, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}

	usedEnglish := make(map[cp932.GlyphKey]struct{})
	usedJapanese := make(map[cp932.GlyphKey]struct{})
	for _, item := range project.Items {
		if err := collectRendererKeys(item.Record.ID, "English translation", item.Translation.Text, usedEnglish); err != nil {
			fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
			return 1
		}
		if err := collectRendererKeys(item.Record.ID, "Japanese source", item.Translation.Japanese, usedJapanese); err != nil {
			fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
			return 1
		}
	}

	installedTwoByte := font.DoubleByteKeys()
	candidates := make([]cp932.GlyphKey, 0, len(installedTwoByte))
	englishInstalled := 0
	japaneseInstalled := 0
	unionInstalled := 0
	for _, key := range installedTwoByte {
		_, english := usedEnglish[key]
		_, japanese := usedJapanese[key]
		if english {
			englishInstalled++
		}
		if japanese {
			japaneseInstalled++
		}
		if english || japanese {
			unionInstalled++
			continue
		}
		candidates = append(candidates, key)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })

	fmt.Fprintf(stdout, "Installed glyph slots: %d\n", len(font.Glyphs))
	fmt.Fprintf(stdout, "Installed two-byte slots: %d\n", len(installedTwoByte))
	fmt.Fprintf(stdout, "Two-byte slots referenced by current English message text: %d\n", englishInstalled)
	fmt.Fprintf(stdout, "Two-byte slots referenced by retail Japanese message text: %d\n", japaneseInstalled)
	fmt.Fprintf(stdout, "Two-byte slots referenced by either message corpus: %d\n", unionInstalled)
	fmt.Fprintf(stdout, "Candidate two-byte slots unreferenced by both message corpora: %d\n", len(candidates))
	fmt.Fprintln(stdout, "Safety status: CANDIDATES ONLY; UI/ELF/fixed-data references outside message banks are not yet excluded.")
	if len(candidates) > 0 {
		fmt.Fprintf(stdout, "First candidate keys:")
		limit := min(16, len(candidates))
		for _, key := range candidates[:limit] {
			fmt.Fprintf(stdout, " %04X", uint16(key))
		}
		fmt.Fprintln(stdout)
	}
	return 0
}

func parseRequiredGameDir(command string, args []string, stderr io.Writer) (string, bool) {
	gameDir := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--game-dir":
			if index+1 >= len(args) {
				fmt.Fprintf(stderr, "zill: %s: --game-dir requires a path\n", command)
				return "", false
			}
			index++
			gameDir = args[index]
		default:
			fmt.Fprintf(stderr, "zill: %s: unknown argument %q\n", command, args[index])
			return "", false
		}
	}
	if gameDir == "" {
		fmt.Fprintf(stderr, "zill: usage: zill %s --game-dir PSP_GAME\n", command)
		return "", false
	}
	return gameDir, true
}
