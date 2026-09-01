// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestRetailScannerSourceCompatibleRequiresAuthenticatedNULTerminator(t *testing.T) {
	if retailScannerSourceCompatible([]byte("bank-bounded-record")) {
		t.Fatal("non-NUL retail record must not inherit the captured string-scanner contract")
	}
	if !retailScannerSourceCompatible([]byte{'a', 'b', 0, 'x'}) {
		t.Fatal("retail record containing a NUL terminator should remain scanner-compatible")
	}
}

func TestU6HistoricalSinglePageC5FailuresStayOutsideC22ScannerScope(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, "release", "layout", "consumer-map.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var consumers consumersFile
	if err := decodeStrict("message consumers", raw, &consumers); err != nil {
		t.Fatal(err)
	}

	// These are the exact IDs from the asset-backed U6 failure population:
	// 9 messages / 15 branches where a broad scanner-derived wrap had produced
	// two pages even though upstream English classifies the consumer as C5
	// single-page. The production scanner hardening is C22-only now, so none of
	// these IDs may ever enter that derivation scope again.
	historical := []int{1280007, 1280008, 1280012, 1280017, 1280020, 1280021, 1280043, 1280050, 1280051}
	for _, id := range historical {
		if !slices.Contains(consumers.SinglePageC5IDs, id) {
			t.Errorf("historical C5 regression id %d is no longer authenticated as single-page C5", id)
		}
		if slices.Contains(consumers.C22IDs, id) {
			t.Errorf("historical single-page C5 id %d leaked into C22 scanner scope", id)
		}
	}
}
