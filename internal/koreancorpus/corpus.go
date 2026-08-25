// SPDX-License-Identifier: GPL-3.0-or-later

// Package koreancorpus loads the contributor-owned Korean message overlay.
//
// Korean is intentionally stored separately from the existing English corpus.
// Each sparse section file carries the authenticated Japanese display string so
// stale or mis-addressed translations fail closed before retail compilation.
package koreancorpus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/pelletier/go-toml/v2"
)

const licenseLine = "# SPDX-License-Identifier: CC-BY-SA-4.0"

// Record is one canonical Korean replacement. Layout is optional generated
// reflow text; when empty, Text is compiled directly.
type Record struct {
	ID       int
	Japanese string
	Text     string
	Layout   string
	File     string
}

// Overlay is the complete sparse Korean replacement set in stable ID order.
type Overlay struct {
	Records []Record
	byID    map[int]int
}

type row struct {
	Japanese *string `toml:"japanese"`
	Korean   *string `toml:"korean"`
	Layout   *string `toml:"layout"`
}

// Load reads translations/korean/messages/msgsecNNN.toml. The directory is
// sparse by design: only sections with at least one Korean replacement need a
// file, which keeps incremental bulk translation and review practical.
func Load(root string, project *corpus.Project) (*Overlay, error) {
	if project == nil {
		return nil, fmt.Errorf("load Korean corpus: nil source project")
	}
	dir := filepath.Join(root, "translations", "korean", "messages")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%s: list Korean sections: %w", dir, err)
	}
	overlay := &Overlay{byID: make(map[int]int)}
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("%s: unexpected directory %s", dir, entry.Name())
		}
		section, ok := sectionFromName(entry.Name())
		if !ok {
			return nil, fmt.Errorf("%s: unexpected file %s; want msgsecNNN.toml", dir, entry.Name())
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		records, err := parseSection(data, path, section, project)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			if _, exists := overlay.byID[record.ID]; exists {
				return nil, fmt.Errorf("%s: duplicate Korean ID %d", path, record.ID)
			}
			overlay.byID[record.ID] = len(overlay.Records)
			overlay.Records = append(overlay.Records, record)
		}
	}
	sort.Slice(overlay.Records, func(i, j int) bool { return overlay.Records[i].ID < overlay.Records[j].ID })
	for index, record := range overlay.Records {
		overlay.byID[record.ID] = index
	}
	return overlay, nil
}

// Find returns one canonical Korean replacement by stable source ID.
func (overlay *Overlay) Find(id int) (Record, bool) {
	if overlay == nil {
		return Record{}, false
	}
	index, ok := overlay.byID[id]
	if !ok {
		return Record{}, false
	}
	return overlay.Records[index], true
}

// EffectiveTexts returns the exact natural text that will require glyph slots.
// Generated layout takes precedence when present.
func (overlay *Overlay) EffectiveTexts() []string {
	if overlay == nil {
		return nil
	}
	out := make([]string, 0, len(overlay.Records))
	for _, record := range overlay.Records {
		if record.Layout != "" {
			out = append(out, record.Layout)
		} else {
			out = append(out, record.Text)
		}
	}
	return out
}

// Sections groups replacements by retail message section while preserving ID
// order inside each section.
func (overlay *Overlay) Sections() map[int][]Record {
	out := make(map[int][]Record)
	if overlay == nil {
		return out
	}
	for _, record := range overlay.Records {
		section := record.ID / 10_000
		out[section] = append(out[section], record)
	}
	return out
}

func parseSection(data []byte, path string, section int, project *corpus.Project) ([]Record, error) {
	if !bytes.HasPrefix(data, []byte(licenseLine+"\n")) {
		return nil, fmt.Errorf("%s: required license header is missing", path)
	}
	var decoded map[string]row
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%s: invalid TOML: %w", path, err)
	}
	ids := make([]int, 0, len(decoded))
	values := make(map[int]Record, len(decoded))
	for key, value := range decoded {
		id, err := strconv.Atoi(key)
		if err != nil || key != strconv.Itoa(id) {
			return nil, fmt.Errorf("%s: record key %q is not a decimal ID", path, key)
		}
		if id/10_000 != section {
			return nil, fmt.Errorf("%s: ID %d belongs to section %03d", path, id, id/10_000)
		}
		if value.Japanese == nil || value.Korean == nil {
			return nil, fmt.Errorf("%s: ID %d: japanese and korean are required", path, id)
		}
		if *value.Korean == "" {
			return nil, fmt.Errorf("%s: ID %d: korean must be nonempty", path, id)
		}
		source, ok := project.Find(id)
		if !ok {
			return nil, fmt.Errorf("%s: ID %d does not exist in canonical source corpus", path, id)
		}
		if *value.Japanese != source.Translation.Japanese {
			return nil, fmt.Errorf("%s: ID %d: japanese differs from canonical source", path, id)
		}
		layout := ""
		if value.Layout != nil {
			layout = *value.Layout
		}
		if err := validateText(path, id, "korean", *value.Korean); err != nil {
			return nil, err
		}
		if layout != "" {
			if err := validateText(path, id, "layout", layout); err != nil {
				return nil, err
			}
		}
		ids = append(ids, id)
		values[id] = Record{ID: id, Japanese: *value.Japanese, Text: *value.Korean, Layout: layout, File: path}
	}
	sort.Ints(ids)
	out := make([]Record, len(ids))
	for index, id := range ids {
		out[index] = values[id]
	}
	return out, nil
}

func validateText(path string, id int, field, text string) error {
	if strings.IndexByte(text, 0) >= 0 {
		return fmt.Errorf("%s: ID %d: %s contains NUL", path, id, field)
	}
	for _, r := range text {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s: ID %d: %s contains raw Unicode control character %U", path, id, field, r)
		}
	}
	return nil
}

func sectionFromName(name string) (int, bool) {
	if len(name) != len("msgsec000.toml") || !strings.HasPrefix(name, "msgsec") || !strings.HasSuffix(name, ".toml") {
		return 0, false
	}
	digits := name[len("msgsec") : len("msgsec")+3]
	section, err := strconv.Atoi(digits)
	if err != nil || section < 0 || section >= 279 || fmt.Sprintf("%03d", section) != digits {
		return 0, false
	}
	return section, true
}
