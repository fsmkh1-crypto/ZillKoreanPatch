// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/gamefmt/paa"
	"github.com/HK47196/zill/internal/slotaudit"
	"github.com/HK47196/zill/internal/zillfont"
)

const (
	jillBtnMemberIndex  = 13612
	jillBtnMemberSize   = 0x18e60
	pafMemberOffset     = 0x4490
	retailEBOOTSHA256   = "2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68"
	retailBindataSHA256 = "3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e"
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
	return collectTextRendererKeys(fmt.Sprintf("%s ID %d", label, id), text, used)
}

func collectTextRendererKeys(label, text string, used map[cp932.GlyphKey]struct{}) error {
	for _, r := range text {
		encoded, err := cp932.Encode(string(r))
		if err != nil {
			return fmt.Errorf("%s contains non-CP932 rune %U: %w", label, r, err)
		}
		key, err := cp932.GlyphKeyFromBytes(encoded)
		if err != nil {
			return fmt.Errorf("%s rune %U: %w", label, r, err)
		}
		used[key] = struct{}{}
	}
	return nil
}

func mergeRendererKeys(destination map[cp932.GlyphKey]struct{}, source map[cp932.GlyphKey]struct{}) {
	for key := range source {
		destination[key] = struct{}{}
	}
}

func loadFixedRendererKeys(root string) (map[cp932.GlyphKey]struct{}, fixeddata.EquipmentTranslations, error) {
	used := make(map[cp932.GlyphKey]struct{})
	ebootData, err := os.ReadFile(filepath.Join(root, "release", "strings", "eboot.toml"))
	if err != nil {
		return nil, nil, fmt.Errorf("read release/strings/eboot.toml: %w", err)
	}
	eboot, err := fixeddata.ParseEBOOT(ebootData)
	if err != nil {
		return nil, nil, err
	}
	for offset, field := range eboot {
		if err := collectTextRendererKeys(fmt.Sprintf("EBOOT source %#x", offset), field.Source, used); err != nil {
			return nil, nil, err
		}
		if err := collectTextRendererKeys(fmt.Sprintf("EBOOT replacement %#x", offset), field.Replacement, used); err != nil {
			return nil, nil, err
		}
	}
	equipmentData, err := os.ReadFile(filepath.Join(root, "release", "strings", "equipment.toml"))
	if err != nil {
		return nil, nil, fmt.Errorf("read release/strings/equipment.toml: %w", err)
	}
	equipment, err := fixeddata.ParseEquipment(equipmentData)
	if err != nil {
		return nil, nil, err
	}
	for selector, field := range equipment {
		if err := collectTextRendererKeys(fmt.Sprintf("equipment source %d", selector), field.Source, used); err != nil {
			return nil, nil, err
		}
		if err := collectTextRendererKeys(fmt.Sprintf("equipment replacement %d", selector), field.Text, used); err != nil {
			return nil, nil, err
		}
	}
	return used, equipment, nil
}

func loadAuthenticatedRetailEBOOT(gameDir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(gameDir, "SYSDIR", "EBOOT.BIN"))
	if err != nil {
		return nil, fmt.Errorf("read retail EBOOT.BIN: %w", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != retailEBOOTSHA256 {
		return nil, fmt.Errorf("unsupported retail EBOOT.BIN fingerprint %s", got)
	}
	return data, nil
}

func loadBindataFromArchive(usrdir, archive string) ([]byte, bool, error) {
	pair, err := paa.Open(filepath.Join(usrdir, archive+".bin"), filepath.Join(usrdir, archive+".arc"))
	if err != nil {
		return nil, false, fmt.Errorf("open %s archive: %w", archive, err)
	}
	var payload []byte
	found := false
	for _, member := range pair.Members() {
		if member.Name != "data/bindata.dat" {
			continue
		}
		if found {
			_ = pair.Close()
			return nil, false, fmt.Errorf("data/bindata.dat appears more than once in %s archive", archive)
		}
		payload, err = pair.Payload(member.Index)
		if err != nil {
			_ = pair.Close()
			return nil, false, fmt.Errorf("read %s data/bindata.dat: %w", archive, err)
		}
		found = true
	}
	if err := pair.Close(); err != nil {
		return nil, false, fmt.Errorf("close %s archive: %w", archive, err)
	}
	return payload, found, nil
}

func loadRetailBindataWithSHA(gameDir, expectedSHA string) ([]byte, error) {
	usrdir := filepath.Join(gameDir, "USRDIR")
	payload, found, err := loadBindataFromArchive(usrdir, "pa")
	if err != nil {
		return nil, err
	}
	if !found {
		payload, found, err = loadBindataFromArchive(usrdir, "pami")
		if err != nil {
			return nil, err
		}
	}
	if !found {
		return nil, fmt.Errorf("retail archives do not contain data/bindata.dat")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(payload)); got != expectedSHA {
		return nil, fmt.Errorf("unsupported retail data/bindata.dat fingerprint %s", got)
	}
	return payload, nil
}

func loadRetailBindata(gameDir string) ([]byte, error) {
	return loadRetailBindataWithSHA(gameDir, retailBindataSHA256)
}

func installedReferenceCount(installed []cp932.GlyphKey, used map[cp932.GlyphKey]struct{}) int {
	count := 0
	for _, key := range installed {
		if _, ok := used[key]; ok {
			count++
		}
	}
	return count
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

	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	eboot, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	ebootScan, err := slotaudit.ScanCP932Literals(eboot)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: scan EBOOT: %v\n", err)
		return 1
	}
	bindata, err := loadRetailBindata(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	// The SHA-256 pin above authenticates retail bindata.dat. ApplyEquipment
	// remains a structural consistency check against the canonical fixed table;
	// the returned translated copy is intentionally discarded because this audit
	// scans the authenticated retail bytes.
	if _, err := fixeddata.ApplyEquipment(bindata, equipment); err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: validate bindata.dat layout: %v\n", err)
		return 1
	}
	bindataScan, err := slotaudit.ScanCP932Literals(bindata)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: scan bindata.dat: %v\n", err)
		return 1
	}

	usedAll := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(usedAll, usedEnglish)
	mergeRendererKeys(usedAll, usedJapanese)
	mergeRendererKeys(usedAll, usedFixed)
	mergeRendererKeys(usedAll, ebootScan.Keys)
	mergeRendererKeys(usedAll, bindataScan.Keys)

	installedTwoByte := font.DoubleByteKeys()
	candidates := make([]cp932.GlyphKey, 0, len(installedTwoByte))
	for _, key := range installedTwoByte {
		if _, used := usedAll[key]; !used {
			candidates = append(candidates, key)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })

	messageUnion := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(messageUnion, usedEnglish)
	mergeRendererKeys(messageUnion, usedJapanese)
	fmt.Fprintf(stdout, "Installed glyph slots: %d\n", len(font.Glyphs))
	fmt.Fprintf(stdout, "Installed two-byte slots: %d\n", len(installedTwoByte))
	fmt.Fprintf(stdout, "Two-byte slots referenced by current English message text: %d\n", installedReferenceCount(installedTwoByte, usedEnglish))
	fmt.Fprintf(stdout, "Two-byte slots referenced by retail Japanese message text: %d\n", installedReferenceCount(installedTwoByte, usedJapanese))
	fmt.Fprintf(stdout, "Two-byte slots referenced by either message corpus: %d\n", installedReferenceCount(installedTwoByte, messageUnion))
	fmt.Fprintf(stdout, "Two-byte slots referenced by canonical EBOOT/equipment fixed strings: %d\n", installedReferenceCount(installedTwoByte, usedFixed))
	fmt.Fprintf(stdout, "Recovered retail EBOOT CP932 literals: %d; installed two-byte keys referenced: %d\n", len(ebootScan.Literals), installedReferenceCount(installedTwoByte, ebootScan.Keys))
	fmt.Fprintf(stdout, "Recovered authenticated bindata.dat CP932 literals: %d; installed two-byte keys referenced: %d\n", len(bindataScan.Literals), installedReferenceCount(installedTwoByte, bindataScan.Keys))
	fmt.Fprintf(stdout, "Candidate two-byte slots after message/fixed/EBOOT/bindata audit: %d\n", len(candidates))
	fmt.Fprintln(stdout, "Safety status: AUDITED CANDIDATES ONLY; other archive/UI/script resources are not yet semantically parsed, so these are not production-safe slots yet.")
	if len(candidates) > 0 {
		fmt.Fprintf(stdout, "First audited candidate keys:")
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
