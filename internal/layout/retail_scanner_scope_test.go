// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "testing"

func TestRetailScannerSourceCompatibleRequiresAuthenticatedNULTerminator(t *testing.T) {
	if retailScannerSourceCompatible([]byte("bank-bounded-record")) {
		t.Fatal("non-NUL retail record must not inherit the captured string-scanner contract")
	}
	if !retailScannerSourceCompatible([]byte{'a', 'b', 0, 'x'}) {
		t.Fatal("retail record containing a NUL terminator should remain scanner-compatible")
	}
}
