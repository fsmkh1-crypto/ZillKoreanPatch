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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/pelletier/go-toml/v2"
)

const licenseLine = "# SPDX-License-Identifier: CC-BY-SA-4.0"

var sectionFilename = regexp.MustCompile(`^msgsec([0-9]{3})(?:(?:-part([0-9]{2}))|b)?\.toml$`)

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

// Load reads the sparse Korean overlay. Large sections may be split into
// msgsecNNN-partNN.toml files; legacy msgsecNNNb.toml is accepted while older
// translation branches are being consolidated. Every row remains bound to the
// canonical Japanese source and duplicate IDs must be byte-for-byte identical.
func Load(root string, project *corpus.Project) (*Overlay, error) {
	if project == nil {
		return nil, fmt.Errorf("load Korean corpus: nil source project")
	}
	sources := make(map[int]string, len(project.Items))
	for _, item := range project.Items {
		if _, exists := sources[item.Record.ID]; exists {
			return nil, fmt.Errorf("load Korean corpus: duplicate canonical source ID %d", item.Record.ID)
		}
		sources[item.Record.ID] = item.Translation.Japanese
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
			return nil, fmt.Errorf("%s: unexpected file %s; want msgsecNNN.toml, msgsecNNN-partNN.toml, or legacy msgsecNNNb.toml", dir, entry.Name())
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		records, err := parseSection(data, path, section, sources)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			if index, exists := overlay.byID[record.ID]; exists {
				previous := overlay.Records[index]
				if previous.Japanese == record.Japanese && previous.Text == record.Text && previous.Layout == record.Layout {
					continue
				}
				return nil, fmt.Errorf("%s: conflicting duplicate Korean ID %d (already in %s)", path, record.ID, previous.File)
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

func parseSection(data []byte, path string, section int, sources map[int]string) ([]Record, error) {
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
		source, ok := sources[id]
		if !ok {
			return nil, fmt.Errorf("%s: ID %d does not exist in canonical source corpus", path, id)
		}
		if *value.Japanese != source {
			return nil, fmt.Errorf("%s: ID %d: japanese differs from canonical source", path, id)
		}
		layout := ""
		if value.Layout != nil {
			layout = *value.Layout
		}
		if err := validateText(path, id, "korean", *value.Korean); err != nil {
			return nil, err
		}
		if err := validateControlContract(path, id, source, *value.Korean, "korean"); err != nil {
			return nil, err
		}
		if layout != "" {
			if err := validateText(path, id, "layout", layout); err != nil {
				return nil, err
			}
			if err := validateControlContract(path, id, source, layout, "layout"); err != nil {
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
	match := sectionFilename.FindStringSubmatch(name)
	if match == nil {
		return 0, false
	}
	section, err := strconv.Atoi(match[1])
	if err != nil || section < 0 || section >= 279 || fmt.Sprintf("%03d", section) != match[1] {
		return 0, false
	}
	if match[2] != "" {
		part, err := strconv.Atoi(match[2])
		if err != nil || part < 1 {
			return 0, false
		}
	}
	return section, true
}
