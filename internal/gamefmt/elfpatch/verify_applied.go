// SPDX-License-Identifier: GPL-3.0-or-later

package elfpatch

import (
	"bytes"
	"fmt"
)

// VerifyApplied checks that every guarded executable patch still contains its
// declared replacement bytes. This is intentionally span-local rather than a
// whole-file fingerprint check so later, disjoint localization overlays may
// modify fixed strings without being mistaken for damage to the runtime patch.
//
// Korean builds use this after their sparse EBOOT string overlay. It prevents a
// localization field from silently clobbering message-arena or wide-offset
// instructions after the manifest itself has already been authenticated.
func VerifyApplied(result []byte, manifest Manifest) error {
	if err := manifest.Validate(); err != nil {
		return err
	}
	for _, patch := range manifest.Patches {
		replacement, err := decodeBytes(patch.Replacement)
		if err != nil {
			return fmt.Errorf("patch %q replacement bytes: %w", patch.ID, err)
		}
		if patch.Offset > uint64(len(result)) || uint64(len(replacement)) > uint64(len(result))-patch.Offset {
			return fmt.Errorf("patch %q replacement span is outside the executable", patch.ID)
		}
		start := int(patch.Offset)
		if !bytes.Equal(result[start:start+len(replacement)], replacement) {
			return fmt.Errorf("patch %q was modified after executable patching", patch.ID)
		}
	}
	return nil
}
