// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/translationexchange"
)

const geminiExportUsage = "zill gemini-export (--batch N | --section N [--start-index N]) [--count N] --out FILE"
const geminiCheckUsage = "zill gemini-check --input SOURCE.jsonl --result GEMINI.jsonl --out ACCEPTED.jsonl"

type geminiExportOptions struct {
	batch      int
	section    int
	startIndex int
	count      int
	out        string
}

type geminiCheckOptions struct {
	input  string
	result string
	out    string
}

func runGeminiExport(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseGeminiExportOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-export: %v\n", err)
		fmt.Fprintf(stderr, "zill: usage: %s\n", geminiExportUsage)
		return 2
	}
	project, _, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-export: %v\n", err)
		return 1
	}

	items, description, err := selectGeminiItems(project, options)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-export: %v\n", err)
		return 1
	}
	rows := make([]translationexchange.ExportRowV2, 0, len(items))
	for _, item := range items {
		id := item.Translation.ID
		source := translationexchange.BuildSourceV2(
			id,
			item.Record.Index,
			editablePath(id),
			item.Translation.Japanese,
			item.Translation.Text,
			map[string]string{},
		)
		rows = append(rows, source.Export)
	}
	if err := writeJSONLFile(options.out, func(w io.Writer) error {
		return translationexchange.WriteExportV2(w, rows)
	}); err != nil {
		fmt.Fprintf(stderr, "zill: gemini-export: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "OK: wrote Gemini v2 %s with %d records to %s\n", description, len(rows), options.out)
	return 0
}

func selectGeminiItems(project *corpus.Project, options geminiExportOptions) ([]corpus.Item, string, error) {
	if options.batch >= 0 {
		offset := options.batch * options.count
		if offset >= len(project.Items) {
			return nil, "", fmt.Errorf("batch %d starts at record offset %d, beyond %d records", options.batch, offset, len(project.Items))
		}
		end := min(offset+options.count, len(project.Items))
		return project.Items[offset:end], fmt.Sprintf("batch %d (project offsets %d..%d)", options.batch, offset, end-1), nil
	}
	items := make([]corpus.Item, 0, options.count)
	for _, item := range project.Items {
		if item.Translation.ID/10_000 != options.section || item.Record.Index < options.startIndex {
			continue
		}
		items = append(items, item)
		if len(items) == options.count {
			break
		}
	}
	if len(items) == 0 {
		return nil, "", fmt.Errorf("section %03d has no records at/after index %d", options.section, options.startIndex)
	}
	return items, fmt.Sprintf("section %03d from index %d", options.section, options.startIndex), nil
}

func runGeminiCheck(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseGeminiCheckOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		fmt.Fprintf(stderr, "zill: usage: %s\n", geminiCheckUsage)
		return 2
	}

	sourceFile, err := os.Open(options.input)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: open source batch: %v\n", err)
		return 1
	}
	externalSource, err := translationexchange.ReadExportV2(sourceFile)
	closeSourceErr := sourceFile.Close()
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		return 1
	}
	if closeSourceErr != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: close source batch: %v\n", closeSourceErr)
		return 1
	}

	project, _, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		return 1
	}
	canonicalSource := make([]translationexchange.SourceRowV2, 0, len(externalSource))
	for position, exported := range externalSource {
		item, exists := project.Find(exported.ID)
		if !exists {
			fmt.Fprintf(stderr, "zill: gemini-check: source position %d references unknown ID %d\n", position+1, exported.ID)
			return 1
		}
		expected := translationexchange.BuildSourceV2(
			exported.ID,
			item.Record.Index,
			editablePath(exported.ID),
			item.Translation.Japanese,
			item.Translation.Text,
			exported.Glossary,
		)
		if !translationexchange.SameExportV2(expected.Export, exported) {
			fmt.Fprintf(stderr, "zill: gemini-check: source position %d ID %d is stale or modified; regenerate the batch\n", position+1, exported.ID)
			return 1
		}
		canonicalSource = append(canonicalSource, expected)
	}

	resultFile, err := os.Open(options.result)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: open Gemini result: %v\n", err)
		return 1
	}
	results, err := translationexchange.ReadResultsV2(resultFile)
	closeResultErr := resultFile.Close()
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		return 1
	}
	if closeResultErr != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: close Gemini result: %v\n", closeResultErr)
		return 1
	}

	validation, err := translationexchange.ValidateV2(canonicalSource, results)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: REJECTED: %v\n", err)
		return 1
	}
	if err := writeJSONLFile(options.out, func(w io.Writer) error {
		return translationexchange.WriteAcceptedV2(w, validation.Rows)
	}); err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: write accepted batch: %v\n", err)
		return 1
	}
	for _, warning := range validation.Warnings {
		fmt.Fprintf(stderr, "zill: gemini-check: warning: %s\n", warning)
	}
	fmt.Fprintf(stdout, "OK: accepted %d Gemini v2 records; %d uncertain; %d warnings; reconstructed staging result: %s\n", len(validation.Rows), validation.Uncertain, len(validation.Warnings), options.out)
	return 0
}

func parseGeminiExportOptions(args []string) (geminiExportOptions, error) {
	options := geminiExportOptions{batch: -1, section: -1, startIndex: 0, count: 100}
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		name, value, hasEquals := strings.Cut(args[index], "=")
		if name != "--batch" && name != "--section" && name != "--start-index" && name != "--count" && name != "--out" {
			return geminiExportOptions{}, fmt.Errorf("unknown argument %q", args[index])
		}
		if seen[name] {
			return geminiExportOptions{}, fmt.Errorf("%s may be specified only once", name)
		}
		seen[name] = true
		if !hasEquals {
			if index+1 >= len(args) {
				return geminiExportOptions{}, fmt.Errorf("%s requires a value", name)
			}
			index++
			value = args[index]
		}
		if value == "" {
			return geminiExportOptions{}, fmt.Errorf("%s requires a value", name)
		}
		switch name {
		case "--out":
			options.out = value
		case "--batch", "--section", "--start-index", "--count":
			number, err := strconv.Atoi(value)
			if err != nil || number < 0 {
				return geminiExportOptions{}, fmt.Errorf("invalid %s %q", name, value)
			}
			switch name {
			case "--batch":
				options.batch = number
			case "--section":
				if number > 278 {
					return geminiExportOptions{}, fmt.Errorf("--section must be 0..278")
				}
				options.section = number
			case "--start-index":
				options.startIndex = number
			case "--count":
				if number < 1 || number > 150 {
					return geminiExportOptions{}, fmt.Errorf("--count must be between 1 and 150")
				}
				options.count = number
			}
		}
	}
	if options.out == "" {
		return geminiExportOptions{}, fmt.Errorf("--out is required")
	}
	if (options.batch >= 0) == (options.section >= 0) {
		return geminiExportOptions{}, fmt.Errorf("set exactly one of --batch or --section")
	}
	if options.batch >= 0 && seen["--start-index"] {
		return geminiExportOptions{}, fmt.Errorf("--start-index is valid only with --section")
	}
	return options, nil
}

func parseGeminiCheckOptions(args []string) (geminiCheckOptions, error) {
	var options geminiCheckOptions
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		name, value, hasEquals := strings.Cut(args[index], "=")
		if name != "--input" && name != "--result" && name != "--out" {
			return geminiCheckOptions{}, fmt.Errorf("unknown argument %q", args[index])
		}
		if seen[name] {
			return geminiCheckOptions{}, fmt.Errorf("%s may be specified only once", name)
		}
		seen[name] = true
		if !hasEquals {
			if index+1 >= len(args) {
				return geminiCheckOptions{}, fmt.Errorf("%s requires a value", name)
			}
			index++
			value = args[index]
		}
		if value == "" {
			return geminiCheckOptions{}, fmt.Errorf("%s requires a value", name)
		}
		switch name {
		case "--input":
			options.input = value
		case "--result":
			options.result = value
		case "--out":
			options.out = value
		}
	}
	if options.input == "" || options.result == "" || options.out == "" {
		return geminiCheckOptions{}, fmt.Errorf("--input, --result, and --out are required")
	}
	return options, nil
}

func writeJSONLFile(path string, write func(io.Writer) error) error {
	if path == "" {
		return fmt.Errorf("output path is empty")
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".zill-jsonl-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	temporaryPath := temporary.Name()
	keep := false
	defer func() {
		_ = temporary.Close()
		if !keep {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := write(temporary); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary output: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace old output: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install output: %w", err)
	}
	keep = true
	return nil
}
