// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"unicode"

	"github.com/pelletier/go-toml/v2"
)

const koreanLicenseLine = "# SPDX-License-Identifier: CC-BY-SA-4.0"

var koreanMarkupToken = regexp.MustCompile(`<[^<>]+>`)
var koreanSectionFilename = regexp.MustCompile(`^msgsec([0-9]{3})(?:(?:-part([0-9]{2}))|b)?\.toml$`)

type KoreanEntry struct {
	ID       int
	Japanese string
	Korean   string
	Layout   string
	File     string
}

type KoreanProject struct {
	Entries []KoreanEntry
	byID    map[int]int
}

type KoreanSummary struct {
	Sections int
	Records  int
}

type koreanValue struct {
	Japanese *string `toml:"japanese"`
	Korean   *string `toml:"korean"`
	Layout   *string `toml:"layout"`
}

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
		section, _, ok := parseKoreanSectionFilename(entry.Name())
		if !ok {
			return nil, KoreanSummary{}, fmt.Errorf("%s: unexpected file %s; want msgsecNNN.toml, msgsecNNN-partNN.toml, or legacy msgsecNNNb.toml", dir, entry.Name())
		}
		seenSections[section] = struct{}{}
		rows, err := readKoreanFile(filepath.Join(dir, entry.Name()), section, source)
		if err != nil {
			return nil, KoreanSummary{}, err
		}
		for _, row := range rows {
			if i, exists := project.byID[row.ID]; exists {
				previous := project.Entries[i]
				if previous.Japanese == row.Japanese && previous.Korean == row.Korean && previous.Layout == row.Layout {
					continue
				}
				return nil, KoreanSummary{}, fmt.Errorf("%s: conflicting duplicate Korean ID %d (already in %s)", row.File, row.ID, previous.File)
			}
			project.byID[row.ID] = len(project.Entries)
			project.Entries = append(project.Entries, row)
		}
		summary.Records += len(rows)
	}
	summary.Sections = len(seenSections)
	sort.Slice(project.Entries, func(i, j int) bool { return project.Entries[i].ID < project.Entries[j].ID })
	project.byID = make(map[int]int, len(project.Entries))
	for i, row := range project.Entries {
		project.byID[row.ID] = i
	}
	return project, summary, nil
}

func (project *KoreanProject) Find(id int) (KoreanEntry, bool) {
	if project == nil {
		return KoreanEntry{}, false
	}
	i, ok := project.byID[id]
	if !ok {
		return KoreanEntry{}, false
	}
	return project.Entries[i], true
}

// Texts returns accepted Korean semantic text. Generated layout is deliberately
// excluded: slot planning may use RuntimeTexts, while canonical translation QA
// should operate on translator-owned text rather than machine-owned wrapping.
func (project *KoreanProject) Texts() []string {
	if project == nil {
		return nil
	}
	out := make([]string, len(project.Entries))
	for i, row := range project.Entries {
		out[i] = row.Korean
	}
	return out
}

// Layouts returns only generated layout overrides. Layout is kept separate from
// semantic Korean so line wrapping can change without rewriting translations.
func (project *KoreanProject) Layouts() map[int]string {
	out := make(map[int]string)
	if project == nil {
		return out
	}
	for _, row := range project.Entries {
		if row.Layout != "" {
			out[row.ID] = row.Layout
		}
	}
	return out
}

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
	if err := validateKoreanText("Korean translation update", id, "korean", korean); err != nil {
		return nil, err
	}
	if err := validateKoreanControls("Korean translation update", id, item.Translation.Japanese, korean, "korean"); err != nil {
		return nil, err
	}
	updated := &KoreanProject{Entries: append([]KoreanEntry(nil), project.Entries...), byID: make(map[int]int, len(project.byID)+1)}
	for key, value := range project.byID {
		updated.byID[key] = value
	}
	// Any existing generated layout belongs to the previous semantic text and is
	// invalidated by a translation edit.
	row := KoreanEntry{ID: id, Japanese: item.Translation.Japanese, Korean: korean}
	if i, exists := updated.byID[id]; exists {
		row.File = updated.Entries[i].File
		updated.Entries[i] = row
	} else {
		updated.Entries = append(updated.Entries, row)
		sort.Slice(updated.Entries, func(i, j int) bool { return updated.Entries[i].ID < updated.Entries[j].ID })
	}
	updated.byID = make(map[int]int, len(updated.Entries))
	for i, entry := range updated.Entries {
		updated.byID[entry.ID] = i
	}
	return updated, nil
}

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

func parseKoreanSectionFilename(name string) (section int, part int, ok bool) {
	match := koreanSectionFilename.FindStringSubmatch(name)
	if match == nil {
		return 0, 0, false
	}
	section64, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return 0, 0, false
	}
	section = int(section64)
	if section < 0 || section >= sectionCount {
		return 0, 0, false
	}
	if match[2] != "" {
		parsed, err := strconv.Atoi(match[2])
		if err != nil || parsed < 1 {
			return 0, 0, false
		}
		part = parsed
	}
	return section, part, true
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
		if err := validateKoreanText(path, id, "korean", *value.Korean); err != nil {
			return nil, err
		}
		if err := validateKoreanControls(path, id, *value.Japanese, *value.Korean, "korean"); err != nil {
			return nil, err
		}
		layout := ""
		if value.Layout != nil {
			layout = *value.Layout
			if layout != "" {
				if err := validateKoreanText(path, id, "layout", layout); err != nil {
					return nil, err
				}
				if err := validateKoreanControls(path, id, *value.Japanese, layout, "layout"); err != nil {
					return nil, err
				}
			}
		}
		ids = append(ids, id)
		values[id] = KoreanEntry{ID: id, Japanese: *value.Japanese, Korean: *value.Korean, Layout: layout, File: path}
	}
	sort.Ints(ids)
	rows := make([]KoreanEntry, len(ids))
	for i, id := range ids {
		rows[i] = values[id]
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
		if row.Layout != "" {
			fmt.Fprintf(&output, "layout = %s\n", strconv.Quote(row.Layout))
		}
	}
	return output.Bytes()
}

func validateKoreanText(path string, id int, field, text string) error {
	if text == "" {
		return fmt.Errorf("%s: ID %d: %s must be nonempty", path, id, field)
	}
	for _, character := range text {
		if unicode.IsControl(character) {
			return fmt.Errorf("%s: ID %d: %s contains a raw Unicode control character", path, id, field)
		}
	}
	return nil
}

// fixedKoreanControls intentionally excludes <line-break>. Line wrapping is
// build-owned layout; every other angle-bracket control remains source-owned
// and must survive in identical order.
func fixedKoreanControls(text string) []string {
	all := koreanMarkupToken.FindAllString(text, -1)
	controls := make([]string, 0, len(all))
	for _, token := range all {
		if token != "<line-break>" {
			controls = append(controls, token)
		}
	}
	return controls
}

func validateKoreanControls(path string, id int, japanese, translated, field string) error {
	want := fixedKoreanControls(japanese)
	got := fixedKoreanControls(translated)
	if len(want) != len(got) {
		return fmt.Errorf("%s: ID %d: %s fixed control token sequence differs from Japanese source: got %v, want %v", path, id, field, got, want)
	}
	for i := range want {
		if want[i] != got[i] {
			return fmt.Errorf("%s: ID %d: %s fixed control token sequence differs from Japanese source at token %d: got %q, want %q", path, id, field, i, got[i], want[i])
		}
	}
	return nil
}
