// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"path/filepath"

	"github.com/HK47196/zill/internal/gamefmt/paa"
)

const paaForensicAlignment = uint64(0x10)

func alignForensic(value uint64) uint64 {
	return (value + paaForensicAlignment - 1) &^ (paaForensicAlignment - 1)
}

func archiveReplacementSizes(a *archive) (map[int]uint64, error) {
	sizes := make(map[int]uint64, len(a.replacements))
	for _, replacement := range a.replacements {
		if _, exists := sizes[replacement.Index]; exists {
			return nil, fmt.Errorf("archive %s has duplicate replacement for member %d", a.name, replacement.Index)
		}
		sizes[replacement.Index] = uint64(len(replacement.Payload))
	}
	return sizes, nil
}

func simulateArchiveOffsets(a *archive, growIndex int, growBy uint64) (map[int]uint64, map[int]uint64, error) {
	replacements, err := archiveReplacementSizes(a)
	if err != nil {
		return nil, nil, err
	}
	offsets := make(map[int]uint64, len(a.pair.Members()))
	sizes := make(map[int]uint64, len(a.pair.Members()))
	position := uint64(0x10) // PAA payload archive prefix.
	for _, member := range a.pair.Members() {
		offset := alignForensic(position)
		size := uint64(member.Size)
		if replacementSize, ok := replacements[member.Index]; ok {
			size = replacementSize
		}
		if member.Index == growIndex {
			size += growBy
		}
		offsets[member.Index] = offset
		sizes[member.Index] = size
		position = offset + size
	}
	return offsets, sizes, nil
}

// logArchiveLayoutCounterfactual compares the current known-good L71 build
// against the exact one-byte-larger msgsec001.dat size used by the failing
// baseline. PAA rebuild is deterministic and aligns every member to 16 bytes,
// so this proves whether that one-byte delta can physically move msgsec021.dat.
func logArchiveLayoutCounterfactual(archives []*archive, owners map[string]bankOwner, compiled map[string][]byte) error {
	one, ok := owners["msgsec001.dat"]
	if !ok {
		return fmt.Errorf("msgsec001.dat owner missing")
	}
	twentyOne, ok := owners["msgsec021.dat"]
	if !ok {
		return fmt.Errorf("msgsec021.dat owner missing")
	}
	for _, target := range []struct {
		name  string
		owner bankOwner
	}{
		{"msgsec001.dat", one},
		{"msgsec021.dat", twentyOne},
	} {
		members := target.owner.archive.pair.Members()
		if target.owner.index < 0 || target.owner.index >= len(members) {
			return fmt.Errorf("%s owner index %d out of range", target.name, target.owner.index)
		}
		member := members[target.owner.index]
		fmt.Printf("FORENSIC BANK_OWNER bank=%s archive=%s member_index=%d member_name=%q retail_offset=%#x retail_size=%d\n",
			target.name, target.owner.archive.name, target.owner.index, member.Name, member.Offset, member.Size)
	}
	fmt.Printf("FORENSIC BANK_RELATION same_archive=%t msgsec001_archive=%s msgsec021_archive=%s msgsec001_index=%d msgsec021_index=%d current_msgsec001_size=%d failing_counterfactual_size=%d current_msgsec021_size=%d\n",
		one.archive == twentyOne.archive, one.archive.name, twentyOne.archive.name, one.index, twentyOne.index,
		len(compiled["msgsec001.dat"]), len(compiled["msgsec001.dat"])+1, len(compiled["msgsec021.dat"]))

	for _, a := range archives {
		growIndex := -1
		if a == one.archive {
			growIndex = one.index
		}
		currentOffsets, currentSizes, err := simulateArchiveOffsets(a, -1, 0)
		if err != nil {
			return err
		}
		failingOffsets, failingSizes, err := simulateArchiveOffsets(a, growIndex, 1)
		if err != nil {
			return err
		}
		for _, target := range []struct {
			bank  string
			owner bankOwner
		}{
			{"msgsec001.dat", one},
			{"msgsec021.dat", twentyOne},
		} {
			if target.owner.archive != a {
				continue
			}
			idx := target.owner.index
			currentEnd := currentOffsets[idx] + currentSizes[idx]
			failingEnd := failingOffsets[idx] + failingSizes[idx]
			fmt.Printf("FORENSIC ARCHIVE_SIM archive=%s bank=%s member_index=%d current_offset=%#x current_size=%d current_end=%#x current_end_mod16=%d failing_offset=%#x failing_size=%d failing_end=%#x failing_end_mod16=%d offset_delta=%d\n",
				a.name, target.bank, idx,
				currentOffsets[idx], currentSizes[idx], currentEnd, currentEnd%16,
				failingOffsets[idx], failingSizes[idx], failingEnd, failingEnd%16,
				int64(failingOffsets[idx])-int64(currentOffsets[idx]))
		}
	}
	if one.archive != twentyOne.archive {
		fmt.Printf("FORENSIC ARCHIVE_VERDICT msgsec001_cannot_shift_msgsec021=true reason=%q\n", "message banks are stored in different PAA archives")
		return nil
	}
	currentOffsets, _, err := simulateArchiveOffsets(one.archive, -1, 0)
	if err != nil {
		return err
	}
	failingOffsets, _, err := simulateArchiveOffsets(one.archive, one.index, 1)
	if err != nil {
		return err
	}
	delta := int64(failingOffsets[twentyOne.index]) - int64(currentOffsets[twentyOne.index])
	fmt.Printf("FORENSIC ARCHIVE_VERDICT msgsec001_cannot_shift_msgsec021=%t msgsec021_offset_delta_if_10010_restored=%d current_offset=%#x failing_offset=%#x\n",
		delta == 0, delta, currentOffsets[twentyOne.index], failingOffsets[twentyOne.index])
	return nil
}

func logStagedMessageArchiveForensics(staging string) error {
	for _, archiveName := range []string{"pa", "pami"} {
		pair, err := paa.Open(
			filepath.Join(staging, "USRDIR", archiveName+".bin"),
			filepath.Join(staging, "USRDIR", archiveName+".arc"),
		)
		if err != nil {
			return fmt.Errorf("open staged %s archive: %w", archiveName, err)
		}
		members := pair.Members()
		for _, member := range members {
			base := filepath.Base(member.Name)
			if base != "msgsec001.dat" && base != "msgsec021.dat" {
				continue
			}
			fmt.Printf("FORENSIC STAGED_MEMBER archive=%s bank=%s member_index=%d offset=%#x size=%d end=%#x end_mod16=%d\n",
				archiveName, base, member.Index, member.Offset, member.Size, uint64(member.Offset)+uint64(member.Size), (uint64(member.Offset)+uint64(member.Size))%16)
		}
		if err := pair.Close(); err != nil {
			return err
		}
	}
	return nil
}
