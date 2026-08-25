// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/pelletier/go-toml/v2"
)

const koreanLicenseLine = "# SPDX-License-Identifier: CC-BY-SA-4.0"

// KoreanEntry is one canonical Korean translation paired to the authoritative
// Japanese contributor reference. English is deliberately not stored here: it
// remains reference material in translations/messages and is never the source
// of truth for Korean compilation.
type KoreanEntry struct {
	ID       int
	Japanese string
	Korean   string
	File     string
}

// KoreanProject is the sparse canonical Korean overlay. Missing IDs mean that
// no Korean replacement has been accepted yet; the production compiler can
// therefore fail closed or preserve retail source according to its caller's
// explicit policy without inventing translations.
type KoreanProject struct {
	Entries []KoreanEntry
	byID    map[int]int
}

// KoreanSummary reports how much of the sparse Korean overlay currently exists.
type KoreanSummary struct {
	Sections int
	Records  int
}

type koreanValue struct {
	Japanese *string `toml:"japanese"`
	Korean   *string `toml:"korean"`
}

// LoadKoreanProject loads translations/korean/messages as a sparse set of
// msgsecNNN.toml files and binds every row to the authoritative Japanese field
// already authenticated by LoadProject. The directory itself and individual
// section files may be absent before translation starts, but every file that
// does exist is strict and every row must match an existing source ID and
// Japanese reference exactly.
func LoadKoreanProject(root string, source *Project) (*KoreanProject, KoreanSummary, error) {
	if source == nil {
		return nil, KoreanSummary{}, fmt.Errorf("load Korean corpus: nil source project")
	}
	dir := filepath.Join(root, "translations", "korean", "messages")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return &KoreanProject{byID: make(map[int]int)}, KoreanSummary{}, nil
	}
	if err != nil {
		return nil, KoreanSummary{}, fmt.Errorf("%s: list Korean section files: %w", dir, err)
	}

	project := &KoreanProject{byID: make(map[int]int)}
	summary := KoreanSummary{}
	seenSections := make(map[int]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, KoreanSummary{}, fmt.Errorf("%s: unexpected directory %s", dir, entry.Name())
		}
		section, ok := parseKoreanSectionFilename(entry.Name())
		if !ok {
			return nil, KoreanSummary{}, fmt.Errorf("%s: unexpected file %s; want msgsecNNN.toml", dir, entry.Name())
		}
		if _, exists := seenSections[section]; exists {
			return nil, KoreanSummary{}, fmt.Errorf("%s: duplicate Korean section %03d", dir, section)
		}
		seenSections[section] = struct{}{}
		rows, err := readKoreanFile(filepath.Join(dir, entry.Name()), section, source)
		if err != nil {
			return nil, KoreanSummary{}, err
		}
		for _, row := range rows {
			if _, exists := project.byID[row.ID]; exists {
				return nil, KoreanSummary{}, fmt.Errorf("%s: duplicate Korean ID %d", row.File, row.ID)
			}
			project.byID[row.ID] = len(project.Entries)
			project.Entries = append(project.Entries, row)
		}
		summary.Sections++
		summary.Records += len(rows)
	}

	sort.Slice(project.Entries, func(i, j int) bool { return project.Entries[i].ID < project.Entries[j].ID })
	project.byID = make(map[int]int, len(project.Entries))
	for index, row := range project.Entries {
		project.byID[row.ID] = index
	}
	return project, summary, nil
}

// Find returns one accepted Korean row by stable source ID.
func (project *KoreanProject) Find(id int) (KoreanEntry, bool) {
	if project == nil {
		return KoreanEntry{}, false
	}
	index, ok := project.byID[id]
	if !ok {
		return KoreanEntry{}, false
	}
	return project.Entries[index], true
}

// Texts returns canonical Korean strings in stable ID order for deterministic
// custom-glyph collection and slot allocation.
func (project *KoreanProject) Texts() []string {
	if project == nil {
		return nil
	}
	out := make([]string, len(project.Entries))
	for index, row := range project.Entries {
		out[index] = row.Korean
	}
	return out
}

// WithKorean returns an independent snapshot with one Korean row inserted or
// replaced. The Japanese reference is always copied from source so writers
// cannot accidentally persist stale or English-derived source text.
func (project *KoreanProject) WithKorean(source *Project, id int, korean string) (*KoreanProject, error) {
	if project == nil {
		return nil, fmt.Errorf("Korean translation update: nil Korean project")
	}
	if source == nil {
		return nil, fmt.Errorf("Korean translation update: nil source project")
	}
	item, ok := source.Find(id)
	if !ok {
		return nil, fmt.Errorf("Korean translation update: ID %d does not exist", id)
	}
	if err := validateKoreanText("Korean translation update", id, korean); err != nil {
		return nil, err
	}

	updated := &KoreanProject{Entries: append([]KoreanEntry(nil), project.Entries...), byID: make(map[int]int, len(project.byID)+1)}
	for key, value := range project.byID {
		updated.byID[key] = value
	}
	row := KoreanEntry{ID: id, Japanese: item.Translation.Japanese, Korean: korean}
	if index, exists := updated.byID[id]; exists {
		row.File = updated.Entries[index].File
		updated.Entries[index] = row
	} else {
		updated.Entries = append(updated.Entries, row)
		sort.Slice(updated.Entries, func(i, j int) bool { return updated.Entries[i].ID < updated.Entries[j].ID })
	}
	updated.byID = make(map[int]int, len(updated.Entries))
	for index, entry := range updated.Entries {
		updated.byID[entry.ID] = index
	}
	return updated, nil
}

// RenderKoreanSection emits the canonical sparse TOML for one section. An empty
// section is rejected so unfinished work does not create meaningless files.
func (project *KoreanProject) RenderKoreanSection(section int) ([]byte, error) {
	if project == nil {
		return nil, fmt.Errorf("render Korean section: nil project")
	}
	if section < 0 || section >= sectionCount {
		return nil, fmt.Errorf("render Korean section: invalid section %03d", section)
	}
	rows := make([]KoreanEntry, 0)
	for _, row := range project.Entries {
		if row.ID/10_000 == section {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("render Korean section: section %03d has no accepted rows", section)
	}
	return renderKoreanFile(rows), nil
}

func parseKoreanSectionFilename(name string) (int, bool) {
	if len(name) != len("msgsec000.toml") || !strings.HasPrefix(name, "msgsec") || !strings.HasSuffix(name, ".toml") {
		return 0, false
	}
	value := name[len("msgsec") : len("msgsec")+3]
	section, err := strconv.Atoi(value)
	if err != nil || section < 0 || section >= sectionCount || fmt.Sprintf("%03d", section) != value {
		return 0, false
	}
	return section, true
}

func readKoreanFile(path string, section int, source *Project) ([]KoreanEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: required Korean section: %w", path, err)
	}
	if !bytes.HasPrefix(data, []byte(koreanLicenseLine+"\n")) {
		return nil, fmt.Errorf("%s: required license header is missing", path)
	}
	var decoded map[string]koreanValue
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%s: invalid TOML: %w", path, err)
	}

	ids := make([]int, 0, len(decoded))
	values := make(map[int]KoreanEntry, len(decoded))
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
		item, ok := source.Find(id)
		if !ok {
			return nil, fmt.Errorf("%s: ID %d does not exist in source corpus", path, id)
		}
		if *value.Japanese != item.Translation.Japanese {
			return nil, fmt.Errorf("%s: ID %d: Japanese reference differs from canonical source", path, id)
		}
		if err := validateKoreanText(path, id, *value.Korean); err != nil {
			return nil, err
		}
		ids = append(ids, id)
		values[id] = KoreanEntry{ID: id, Japanese: *value.Japanese, Korean: *value.Korean, File: path}
	}
	sort.Ints(ids)
	rows := make([]KoreanEntry, len(ids))
	for index, id := range ids {
		rows[index] = values[id]
	}
	return rows, nil
}

func renderKoreanFile(rows []KoreanEntry) []byte {
	var output bytes.Buffer
	output.WriteString(koreanLicenseLine)
	output.WriteByte('\n')
	for _, row := range rows {
		fmt.Fprintf(&output, "\n[%q]\n", strconv.Itoa(row.ID))
		fmt.Fprintf(&output, "japanese = %s\n", strconv.Quote(row.Japanese))
		fmt.Fprintf(&output, "korean = %s\n", strconv.Quote(row.Korean))
	}
	return output.Bytes()
}

func validateKoreanText(path string, id int, text string) error {
	if text == "" {
		return fmt.Errorf("%s: ID %d: korean must be nonempty", path, id)
	}
	for _, character := range text {
		if unicode.IsControl(character) {
			return fmt.Errorf("%s: ID %d: korean contains a raw Unicode control character", path, id)
		}
	}
	return nil
}
