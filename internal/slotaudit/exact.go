// SPDX-License-Identifier: GPL-3.0-or-later

package slotaudit

import (
	"bytes"
	"fmt"

	"github.com/HK47196/zill/internal/cp932"
)

// ExcludeExactByteReferences removes renderer keys whose exact encoded byte
// sequence appears anywhere in any authenticated binary blob. This is a
// deliberately conservative second-stage guard for candidate slots: the
// structured CP932 literal scanner can miss isolated one-glyph references, but
// an exact raw-byte occurrence is enough to keep a key out of automatic reuse.
//
// The caller remains responsible for authenticating each blob and for auditing
// other resource classes not supplied here.
func ExcludeExactByteReferences(keys []cp932.GlyphKey, blobs ...[]byte) ([]cp932.GlyphKey, error) {
	out := make([]cp932.GlyphKey, 0, len(keys))
	for _, key := range keys {
		encoded, err := key.Bytes()
		if err != nil {
			return nil, fmt.Errorf("exact-byte audit key 0x%04X: %w", uint16(key), err)
		}
		referenced := false
		for _, blob := range blobs {
			if bytes.Contains(blob, encoded) {
				referenced = true
				break
			}
		}
		if !referenced {
			out = append(out, key)
		}
	}
	return out, nil
}
