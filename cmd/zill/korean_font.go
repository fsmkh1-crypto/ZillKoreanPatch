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

	used := make(map[cp932.GlyphKey]struct{})
	for _, item := range project.Items {
		for _, r := range item.Translation.Text {
			encoded, err := cp932.Encode(string(r))
			if err != nil {
				fmt.Fprintf(stderr, "zill: korean-slots: translation ID %d contains non-CP932 rune %U: %v\n", item.Record.ID, r, err)
				return 1
			}
			key, err := cp932.GlyphKeyFromBytes(encoded)
			if err != nil {
				fmt.Fprintf(stderr, "zill: korean-slots: translation ID %d: %v\n", item.Record.ID, err)
				return 1
			}
			used[key] = struct{}{}
		}
	}

	installedTwoByte := font.DoubleByteKeys()
	available := make([]cp932.GlyphKey, 0, len(installedTwoByte))
	usedInstalled := 0
	for _, key := range installedTwoByte {
		if _, ok := used[key]; ok {
			usedInstalled++
			continue
		}
		available = append(available, key)
	}
	sort.Slice(available, func(i, j int) bool { return available[i] < available[j] })

	fmt.Fprintf(stdout, "Installed glyph slots: %d\n", len(font.Glyphs))
	fmt.Fprintf(stdout, "Installed two-byte slots: %d\n", len(installedTwoByte))
	fmt.Fprintf(stdout, "Two-byte slots referenced by current English text: %d\n", usedInstalled)
	fmt.Fprintf(stdout, "Reusable two-byte slots for Korean PoC: %d\n", len(available))
	if len(available) > 0 {
		fmt.Fprintf(stdout, "First reusable keys:")
		limit := min(16, len(available))
		for _, key := range available[:limit] {
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
