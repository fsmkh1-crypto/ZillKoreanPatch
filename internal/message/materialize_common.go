// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"regexp"
	"strings"
)

var materializeControl = regexp.MustCompile(`(<value:\$[0-9A-F]{2}>|<line-break>)`)

type naturalTextValidator func(id int, key, text string) error
type naturalTextEncoder func(text string) ([]byte, error)

func (p *Projection) splitSemanticWith(text string, validate naturalTextValidator) ([]string, error) {
	if text == "" {
		return nil, fmt.Errorf("message %d: canonical text must be nonempty", p.RecordID)
	}
	values := make([]string, len(p.Fragments))
	cursor := 0
	for index, node := range p.nodes {
		if node.fixed {
			if node.display != "" && !strings.HasPrefix(text[cursor:], node.display) {
				return nil, fmt.Errorf("message %d: canonical text changes fixed %s control", p.RecordID, node.kind)
			}
			cursor += len(node.display)
			continue
		}
		next := ""
		for _, following := range p.nodes[index+1:] {
			if !following.fixed {
				break
			}
			next += following.display
		}
		end := len(text)
		if next != "" {
			relative := strings.Index(text[cursor:], next)
			if relative < 0 {
				return nil, fmt.Errorf("message %d: canonical text is missing a fixed control", p.RecordID)
			}
			end = cursor + relative
		}
		values[node.fragment] = text[cursor:end]
		cursor = end
	}
	if cursor != len(text) {
		return nil, fmt.Errorf("message %d: canonical text has trailing material outside its projection", p.RecordID)
	}
	for index, value := range values {
		if err := p.validateFragmentWith(index, value, validate); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (p *Projection) validateFragmentWith(index int, value string, validate naturalTextValidator) error {
	f := p.Fragments[index]
	found := valueTag.FindAllStringSubmatch(value, -1)
	available := make(map[byte]int, len(f.Substitutions))
	for _, opcode := range f.Substitutions {
		available[opcode]++
	}
	for _, match := range found {
		var opcode byte
		_, _ = fmt.Sscanf(match[1], "%02X", &opcode)
		if available[opcode] == 0 {
			return fmt.Errorf("message %d fragment %s changes runtime substitutions", p.RecordID, f.Key)
		}
		available[opcode]--
	}
	if formatSignatureIDs[p.RecordID] {
		got := printfConversion.FindAllString(value, -1)
		if strings.Count(value, "%") != len(got) || strings.Join(got, "\x00") != strings.Join(f.FormatSignature, "\x00") {
			return fmt.Errorf("message %d fragment %s changes runtime format signature", p.RecordID, f.Key)
		}
	}
	plain := valueTag.ReplaceAllString(value, "")
	plain = strings.ReplaceAll(plain, lineBreak, "")
	if reservedMarkup.MatchString(plain) || reservedAnchor.MatchString(plain) {
		return fmt.Errorf("message %d fragment %s contains reserved markup", p.RecordID, f.Key)
	}
	return validate(p.RecordID, f.Key, plain)
}

func (p *Projection) materializeValues(values []string, layout bool, encode naturalTextEncoder) ([]byte, error) {
	var output []byte
	for _, node := range p.nodes {
		if node.fixed {
			output = append(output, node.raw...)
			continue
		}
		value := values[node.fragment]
		matches := materializeControl.FindAllStringIndex(value, -1)
		cursor := 0
		for _, match := range matches {
			encoded, err := encode(value[cursor:match[0]])
			if err != nil {
				return nil, err
			}
			output = append(output, encoded...)
			piece := value[match[0]:match[1]]
			if piece == lineBreak {
				if !layout {
					return nil, fmt.Errorf("message %d: semantic text contains a layout break", p.RecordID)
				}
				output = append(output, 10)
			} else {
				var opcode byte
				_, _ = fmt.Sscanf(piece, "<value:$%02X>", &opcode)
				output = append(output, 2, opcode)
			}
			cursor = match[1]
		}
		encoded, err := encode(value[cursor:])
		if err != nil {
			return nil, err
		}
		output = append(output, encoded...)
		for _, control := range node.kanaControl {
			output = append(output, control...)
		}
	}
	return output, nil
}
