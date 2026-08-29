// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"encoding/binary"
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// C5KnownExpansionPage records three deliberately separate byte domains for one
// C5 branch/page. StoredBytes is the exact compiled payload, including inline
// substitution opcodes such as 02 15. StaticBytes is the literal payload that
// survives without assigning semantics to substitutions. KnownMaxBytes adds
// only runtime maxima already proven elsewhere. Unknown substitutions are never
// assigned guessed expansion lengths.
type C5KnownExpansionPage struct {
	MessageID           int
	Branch              int
	Page                int
	StoredBytes          int
	StaticBytes          int
	KnownMaxBytes        int
	PlayerNameCount      int
	UnknownSubstitutions int
}

// ExceedsPageBuffer reports whether literal runtime bytes plus only proven
// substitution maxima already reach/exceed the retained 256-byte C5 contract.
func (p C5KnownExpansionPage) ExceedsPageBuffer() bool {
	return p.KnownMaxBytes >= c5PageBufferCapacityBytes
}

// StoredHeadroomBytes reports distance from the exact compiled page payload to
// 256 bytes. It is a bank-storage diagnostic only: it is not a runtime expansion
// threshold because the runtime may replace, append, or stage substitution data.
func (p C5KnownExpansionPage) StoredHeadroomBytes() int {
	return c5PageBufferCapacityBytes - p.StoredBytes
}

type c5AuditEvent struct {
	raw          []byte
	substitution bool
	opcode       byte
}

type c5AuditLeaf struct {
	events []c5AuditEvent
}

// KoreanC5KnownExpansionPages analyzes actual Korean renderer bytes while adding
// only runtime expansion bounds that are independently established. At present
// that means <value:$28>, whose player-name maximum is 16 encoded bytes. Other
// inline substitutions are counted as unknown rather than given invented bounds.
func (e *Engine) KoreanC5KnownExpansionPages(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) ([]C5KnownExpansionPage, error) {
	if source == nil || korean == nil {
		return nil, fmt.Errorf("Korean C5 known-expansion audit: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, fmt.Errorf("Korean C5 known-expansion audit: empty renderer mapping")
	}

	c5 := e.koreanC5Set()
	var pages []C5KnownExpansionPage
	for _, row := range korean.Entries {
		if _, ok := c5[row.ID]; !ok {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, fmt.Errorf("Korean C5 known-expansion audit: message %d lacks source", row.ID)
		}
		p, err := message.Project(item.Record)
		if err != nil {
			return nil, err
		}
		raw, err := p.MaterializeKorean(effectiveKoreanText(row, layouts), true, mapping)
		if err != nil {
			return nil, fmt.Errorf("message %d C5 known-expansion lowering: %w", row.ID, err)
		}
		bankData := make([]byte, 4+len(raw))
		binary.LittleEndian.PutUint16(bankData, 1)
		binary.LittleEndian.PutUint16(bankData[2:], 4)
		copy(bankData[4:], raw)
		bank, err := corpus.ParseBank("msgsec000.dat", bankData)
		if err != nil {
			return nil, fmt.Errorf("message %d C5 known-expansion parse: %w", row.ID, err)
		}
		leaves, err := walkC5Audit(bank.Records[0].Tokens, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("message %d C5 known-expansion analysis: %w", row.ID, err)
		}
		for branch, leaf := range leaves {
			for pageIndex, page := range auditC5LeafPages(leaf) {
				page.MessageID = row.ID
				page.Branch = branch + 1
				page.Page = pageIndex + 1
				pages = append(pages, page)
			}
		}
	}
	return pages, nil
}

func walkC5Audit(tokens []corpus.Token, index int, prefix []c5AuditEvent) ([]c5AuditLeaf, error) {
	out := append([]c5AuditEvent(nil), prefix...)
	nextTerm := func(start int) (int, error) {
		for i := start; i < len(tokens); i++ {
			if tokens[i].Kind == "block_terminator" {
				return i, nil
			}
		}
		return 0, fmt.Errorf("control flow has no block terminator")
	}
	for index < len(tokens) {
		t := tokens[index]
		switch t.Kind {
		case "if":
			term, err := nextTerm(index + 2)
			if err != nil {
				return nil, err
			}
			yes, err := walkC5Audit(tokens, index+2, out)
			if err != nil {
				return nil, err
			}
			no, err := walkC5Audit(tokens, term+1, out)
			if err != nil {
				return nil, err
			}
			return append(yes, no...), nil
		case "select":
			if index+1 >= len(tokens) || tokens[index+1].Kind != "expression" {
				return nil, fmt.Errorf("select lacks expression")
			}
			arms, sink, err := c5SelectShape(tokens[index+1].Raw)
			if err != nil {
				return nil, err
			}
			cursor := index + 2
			var leaves []c5AuditLeaf
			for range arms {
				arm, err := walkC5Audit(tokens, cursor, out)
				if err != nil {
					return nil, err
				}
				leaves = append(leaves, arm...)
				term, err := nextTerm(cursor)
				if err != nil {
					return nil, err
				}
				cursor = term + 1
			}
			if sink {
				rest, err := walkC5Audit(tokens, cursor, out)
				if err != nil {
					return nil, err
				}
				leaves = append(leaves, rest...)
			}
			return leaves, nil
		case "block_terminator", "archive_padding", "suffix":
			return []c5AuditLeaf{{events: out}}, nil
		case "text", "backspace", "tab", "line_break", "color", "undecodable_data":
			out = append(out, c5AuditEvent{raw: append([]byte(nil), t.Raw...)})
		case "substitution":
			opcode := byte(0)
			if len(t.Raw) >= 2 && t.Raw[0] == 2 {
				opcode = t.Raw[1]
			}
			out = append(out, c5AuditEvent{raw: append([]byte(nil), t.Raw...), substitution: true, opcode: opcode})
		case "unknown_control":
			if len(t.Raw) == 1 && t.Raw[0] == 0x7f {
				out = append(out, c5AuditEvent{raw: append([]byte(nil), t.Raw...)})
			}
		case "call", "jump":
			return nil, fmt.Errorf("unsupported %s", t.Kind)
		}
		index++
	}
	return []c5AuditLeaf{{events: out}}, nil
}

func auditC5LeafPages(leaf c5AuditLeaf) []C5KnownExpansionPage {
	pages := []C5KnownExpansionPage{{}}
	cursor := c5PageCursor{}
	for _, event := range leaf.events {
		if event.substitution {
			current := &pages[len(pages)-1]
			current.StoredBytes += len(event.raw)
			if event.opcode == 0x28 {
				current.PlayerNameCount++
				current.KnownMaxBytes += playerNameMaxEncodedBytes
			} else {
				current.UnknownSubstitutions++
			}
			continue
		}
		for _, b := range event.raw {
			current := &pages[len(pages)-1]
			current.StoredBytes++
			current.StaticBytes++
			current.KnownMaxBytes++
			if cursor.addByte(b) {
				pages = append(pages, C5KnownExpansionPage{})
			}
		}
	}
	return pages
}
