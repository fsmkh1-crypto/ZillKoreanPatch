// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// MaterializeKoreanRecordForScannerAudit lowers one selected Korean record with
// the same record-local diagnostic overrides used by CompileBankKorean. It is
// intentionally exported only as a forensic/test seam: callers can inspect the
// exact current compiler bytes without manufacturing a synthetic bank.
func MaterializeKoreanRecordForScannerAudit(source corpus.Record, replacement KoreanRecord, mapping koreanslots.Mapping) ([]byte, error) {
	return materializeKoreanRecord(source, replacement, mapping)
}

func materializeKoreanRecord(source corpus.Record, replacement KoreanRecord, mapping koreanslots.Mapping) ([]byte, error) {
	var err error
	replacement, err = applyKoreanRuntimeDiagnostics(source.ID, replacement)
	if err != nil {
		return nil, err
	}

	projection, err := Project(source)
	if err != nil {
		return nil, fmt.Errorf("projection: %w", err)
	}
	semantic, err := projection.MaterializeKorean(replacement.Text, false, mapping)
	if err != nil {
		return nil, fmt.Errorf("Korean semantic text: %w", err)
	}
	if replacement.Layout == "" {
		return semantic, nil
	}
	if !preservesSemantics(replacement.Text, replacement.Layout) {
		return nil, fmt.Errorf("Korean layout changes semantic/control text; only layout boundaries may replace semantic whitespace")
	}
	materialized, err := projection.MaterializeKorean(replacement.Layout, true, mapping)
	if err != nil {
		return nil, fmt.Errorf("Korean layout: %w", err)
	}
	return materialized, nil
}

func applyKoreanRuntimeDiagnostics(id int, replacement KoreanRecord) (KoreanRecord, error) {
	if text, ok := characterChoiceBufferDiagnostic[id]; ok {
		replacement.Text = text
		replacement.Layout = ""
	}
	if id == 10010 {
		replacement.Text = strings.Replace(replacement.Text, "<value:$15>", "<value:$15> ", 1)
		if replacement.Layout != "" {
			replacement.Layout = strings.Replace(replacement.Layout, "<value:$15>", "<value:$15> ", 1)
		}
	}
	if id == 210065 {
		const semantic = "광대한 대지 바이아시온 대륙. 너무나 넓어 지도에도 기록되지 않고 여행자에게조차 알려지지 않은 작은 마을이 있다…. 마을의 이름은 미이스. 그곳에는 작은 신전과 숲, 그리고 평온한 일상 정도뿐이었다. 위대한 혼의 이야기는 여기서 시작된다…….<end>"
		if replacement.Text != semantic {
			return replacement, fmt.Errorf("combined diagnostic semantic precondition failed")
		}
		replacement.Layout = opening210065SafeLayout
	}
	return replacement, nil
}

// RetailScannerMetrics mirrors the observable contract of z_un_089661DC from
// docs/audit/fixtures/runtime-20260829-freeze-disassembly.txt. MaxSpan is the
// maximum number of ordinary encoded bytes between CR/LF boundaries. ESC (0x1B)
// control sequences are skipped as a three-byte unit by the retail scanner and
// therefore do not contribute to the span count.
type RetailScannerMetrics struct {
	MaxSpan    int
	LineBreaks int
	Terminated bool
}

// AnalyzeRetailStringScanner reproduces the max-span part of z_un_089661DC.
// It is deliberately byte-oriented: Korean custom glyphs are two-byte renderer
// keys, exactly as the retail routine sees them.
func AnalyzeRetailStringScanner(raw []byte) (RetailScannerMetrics, error) {
	var out RetailScannerMetrics
	span := 0
	finishLine := func() {
		if span > out.MaxSpan {
			out.MaxSpan = span
		}
		out.LineBreaks++
		span = 0
	}

	for i := 0; i < len(raw); {
		b := raw[i]
		if b == 0 {
			if span > out.MaxSpan {
				out.MaxSpan = span
			}
			out.Terminated = true
			return out, nil
		}
		switch b {
		case 0x1B:
			// The captured MIPS path advances one byte in the branch delay slot
			// and two more at the ESC target: the encoded control is 3 bytes.
			if i+2 >= len(raw) {
				return out, fmt.Errorf("truncated ESC control at byte %d", i)
			}
			i += 3
		case '\r':
			if i+1 < len(raw) && raw[i+1] == '\n' {
				i += 2
			} else {
				i++
			}
			finishLine()
		case '\n':
			i++
			finishLine()
		default:
			span++
			i++
		}
	}
	return out, fmt.Errorf("materialized record is not NUL terminated")
}
