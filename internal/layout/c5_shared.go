// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "fmt"

const c5Select33Arms = 8

// c5SelectShape centralizes the C5 select grammar used by both the production
// validator and forensic audit. The $33 eight-arm shape is inherited from the
// upstream C5 walker. It is intentionally named here rather than duplicated as
// a magic literal; its independent retail-runtime derivation remains an audit
// item and must not be treated as newly proven by this refactor.
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
// Keeping the count-before-transition rule in one helper prevents the validator
// and forensic accounting from independently dropping the boundary newline.
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
