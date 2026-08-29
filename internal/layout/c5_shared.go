// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "fmt"

const c5Select33Arms = 8

// c5SelectShape centralizes select decoding for new C5 forensic code. The $33
// eight-arm shape is inherited from the upstream production C5 walker rather
// than independently re-derived here. Naming it once prevents the audit from
// introducing a second unexplained magic literal, while its retail-runtime
// derivation remains an explicit open audit item.
func c5SelectShape(expr []byte) (arms int, sink bool, err error) {
	if len(expr) >= 4 && expr[0] == 2 && expr[1] == 0x20 && expr[2] == '%' {
		if _, err := fmt.Sscanf(string(expr[3:]), "%d", &arms); err != nil || arms <= 0 {
			return 0, false, fmt.Errorf("invalid $20 select arm count %q", string(expr[3:]))
		}
		return arms, false, nil
	}
	if len(expr) == 2 && expr[0] == 2 && expr[1] == 0x33 {
		return c5Select33Arms, true, nil
	}
	return 0, false, fmt.Errorf("unsupported select expression")
}

// c5PageCursor owns the three-lines-per-page transition rule. addByte returns
// true only after the byte has been counted into the current page and that byte
// is the third line break, meaning subsequent bytes belong to a new page.
// Keeping the count-before-transition rule in one helper prevents the Korean
// validator and forensic accounting from independently dropping the boundary
// newline.
type c5PageCursor struct {
	breaks int
}

func (c *c5PageCursor) addByte(b byte) (startNextPage bool) {
	if b != 10 {
		return false
	}
	c.breaks++
	if c.breaks != c5LinesPerPage {
		return false
	}
	c.breaks = 0
	return true
}
