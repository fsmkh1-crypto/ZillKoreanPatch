// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// koreanSourceAware mirrors the upstream English sourceAware -> preferred ->
// greedy reflow path. The only deliberate differences are Korean semantic
// projection and measurement against the actual Korean renderer mapping.
func (e *Engine) koreanSourceAware(p *message.Projection, semantic string, limit, id int, mapping koreanslots.Mapping) (string, error) {
	fragments, err := p.SplitSemanticKorean(semantic, mapping)
	if err != nil {
		return "", err
	}
	result, cursor := semantic, 0
	for i, fragment := range fragments {
		source := p.Fragments[i].SourceLayout
		flow, err := e.koreanPreferred(fragment, source, limit, id, e.has(e.consumers.C5IDs, id) || e.has(e.consumers.SinglePageC5IDs, id), mapping)
		if err != nil {
			return "", err
		}
		if flow == "" && fragment != "" {
			return "", nil
		}
		at := strings.Index(result[cursor:], fragment)
		if at < 0 {
			return "", fmt.Errorf("message %d: cannot locate projected Korean fragment", id)
		}
		at += cursor
		result = result[:at] + flow + result[at+len(fragment):]
		cursor = at + len(flow)
	}
	return result, nil
}

func (e *Engine) koreanGreedy(s string, limit, id int, mapping koreanslots.Mapping) (string, error) {
	parts := splitParts(s)
	var out []string
	start := 0
	last := -1
	for _, part := range parts {
		out = append(out, part)
		i := len(out) - 1
		if part == lineBreak || part == "<end>" {
			w, err := e.measureKoreanRenderer(strings.Join(out[start:i], ""), id, mapping)
			if err != nil {
				return "", err
			}
			if w > limit {
				if last < start {
					return "", nil
				}
				out[last] = lineBreak
				start = last + 1
				w, err = e.measureKoreanRenderer(strings.Join(out[start:i], ""), id, mapping)
				if err != nil || w > limit {
					return "", err
				}
			}
			start = i + 1
			last = -1
			continue
		}
		if strings.TrimSpace(part) == "" {
			last = i
			continue
		}
		w, err := e.measureKoreanRenderer(strings.Join(out[start:], ""), id, mapping)
		if err != nil {
			return "", err
		}
		if w > limit {
			if last < start {
				return "", nil
			}
			out[last] = lineBreak
			start = last + 1
			last = -1
			w, err = e.measureKoreanRenderer(strings.Join(out[start:], ""), id, mapping)
			if err != nil || w > limit {
				return "", err
			}
		}
	}
	w, err := e.measureKoreanRenderer(strings.Join(out[start:], ""), id, mapping)
	if err != nil {
		return "", err
	}
	if w > limit {
		if last < start {
			return "", nil
		}
		out[last] = lineBreak
		start = last + 1
		w, err = e.measureKoreanRenderer(strings.Join(out[start:], ""), id, mapping)
		if err != nil || w > limit {
			return "", err
		}
	}
	return strings.Join(out, ""), nil
}

func (e *Engine) koreanPreferred(s, source string, limit, id int, c5 bool, mapping koreanslots.Mapping) (string, error) {
	parts := splitParts(s)
	for _, part := range parts {
		if part == lineBreak || part == "<end>" {
			return e.koreanGreedy(s, limit, id, mapping)
		}
	}
	var chunks, separators []string
	var current, pending strings.Builder
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			pending.WriteString(part)
			continue
		}
		if pending.Len() > 0 && current.Len() > 0 {
			chunks = append(chunks, current.String())
			separators = append(separators, pending.String())
			current.Reset()
			pending.Reset()
		} else if pending.Len() > 0 {
			current.WriteString(pending.String())
			pending.Reset()
		}
		current.WriteString(part)
	}
	current.WriteString(pending.String())
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	if len(chunks) == 0 {
		return s, nil
	}
	greedy, err := e.koreanGreedy(s, limit, id, mapping)
	if err != nil || greedy == "" {
		return greedy, err
	}
	lines := strings.Count(greedy, lineBreak) + 1
	if c5 && lines > c5LinesPerPage && lines%c5LinesPerPage == 1 && lines < c5LinesPerPage*c5MaxPages {
		lines++
	}
	if lines > len(chunks) {
		lines = strings.Count(greedy, lineBreak) + 1
	}
	chunkAdvances := make([]int, len(chunks))
	for i, chunk := range chunks {
		advance, err := e.measureKoreanRenderer(chunk, id, mapping)
		if err != nil {
			return "", err
		}
		chunkAdvances[i] = advance
	}
	separatorAdvances := make([]int, len(separators))
	for i, separator := range separators {
		advance, err := e.measureKoreanRenderer(separator, id, mapping)
		if err != nil {
			return "", err
		}
		separatorAdvances[i] = advance
	}
	visibleCounts := make([]int, len(chunks))
	visiblePrefix := make([]int, len(chunks)+1)
	for i, chunk := range chunks {
		visibleCounts[i] = nonspaceRunes(visible(chunk))
		visiblePrefix[i+1] = visiblePrefix[i] + visibleCounts[i]
	}
	totalVisible := visiblePrefix[len(chunks)]
	sourceRatios := sourceBreakRatios(source)
	breakPenalty := func(end int) int {
		if end == len(chunks) || len(sourceRatios) == 0 || totalVisible == 0 {
			return 0
		}
		ratio := float64(visiblePrefix[end]) / float64(totalVisible)
		distance := math.Inf(1)
		for _, preferred := range sourceRatios {
			distance = min(distance, math.Abs(ratio-preferred))
		}
		return int(math.RoundToEven(distance * 1000))
	}
	breakQuality := func(chunk string) int {
		v := strings.TrimRightFunc(visible(chunk), unicode.IsSpace)
		if strongBreak.MatchString(v) {
			return 0
		}
		if mediumBreak.MatchString(v) {
			return 1
		}
		return 2
	}
	type state struct {
		cost   int
		breaks []int
	}
	states := map[int]state{0: {}}
	for line := 0; line < lines; line++ {
		next := map[int]state{}
		remain := lines - line - 1
		for start, st := range states {
			advance := 0
			for end := start + 1; end <= len(chunks); end++ {
				if end > start+1 {
					advance += separatorAdvances[end-2]
				}
				advance += chunkAdvances[end-1]
				if advance > limit {
					break
				}
				if visiblePrefix[end] <= visiblePrefix[start] {
					continue
				}
				if len(chunks)-end < remain || (end == len(chunks)) != (remain == 0) {
					continue
				}
				isBreak := end < len(chunks)
				cost := st.cost + (limit-advance)*(limit-advance)
				if isBreak {
					cost += breakQuality(chunks[end-1]) * limit * limit / 4
					cost += breakPenalty(end) * limit * limit / 2000
				}
				b := append([]int(nil), st.breaks...)
				if isBreak {
					b = append(b, end)
				}
				old, ok := next[end]
				if !ok || cost < old.cost || cost == old.cost && lexLess(b, old.breaks) {
					next[end] = state{cost, b}
				}
			}
		}
		states = next
	}
	best, ok := states[len(chunks)]
	if !ok {
		return greedy, nil
	}
	set := map[int]bool{}
	for _, b := range best.breaks {
		set[b] = true
	}
	var b strings.Builder
	for i, chunk := range chunks {
		if i > 0 {
			if set[i] {
				b.WriteString(lineBreak)
			} else {
				b.WriteString(separators[i-1])
			}
		}
		b.WriteString(chunk)
	}
	return b.String(), nil
}
