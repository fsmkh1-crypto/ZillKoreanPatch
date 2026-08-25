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

const geminiExportUsage = "zill gemini-export --batch N [--count N] --out FILE"
const geminiCheckUsage = "zill gemini-check --input SOURCE.jsonl --result GEMINI.jsonl --out ACCEPTED.jsonl"

type geminiExportOptions struct {
	batch int
	count int
	out   string
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
	offset := options.batch * options.count
	if offset >= len(project.Items) {
		fmt.Fprintf(stderr, "zill: gemini-export: batch %d starts at record offset %d, beyond %d records\n", options.batch, offset, len(project.Items))
		return 1
	}
	end := min(offset+options.count, len(project.Items))
	rows := make([]translationexchange.ExportRow, 0, end-offset)
	for _, item := range project.Items[offset:end] {
		id := item.Translation.ID
		rows = append(rows, translationexchange.ExportRow{
			Schema:           translationexchange.Schema,
			ID:               id,
			Section:          id / 10_000,
			RecordIndex:      item.Record.Index,
			SourceFile:       editablePath(id),
			Japanese:         item.Translation.Japanese,
			EnglishReference: item.Translation.Text,
			// Speaker and scene context are deliberately null until a source can
			// establish them. Storage neighbors are not safe dialogue chronology.
			Speaker: nil,
			Context: nil,
		})
	}
	if err := writeJSONLFile(options.out, func(w io.Writer) error {
		return translationexchange.WriteExport(w, rows)
	}); err != nil {
		fmt.Fprintf(stderr, "zill: gemini-export: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "OK: wrote Gemini batch %d with %d records (project offsets %d..%d) to %s\n", options.batch, len(rows), offset, end-1, options.out)
	return 0
}

func runGeminiCheck(args []string, stdout, stderr io.Writer) int {
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
	source, err := translationexchange.ReadExport(sourceFile)
	closeSourceErr := sourceFile.Close()
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		return 1
	}
	if closeSourceErr != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: close source batch: %v\n", closeSourceErr)
		return 1
	}

	resultFile, err := os.Open(options.result)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: open Gemini result: %v\n", err)
		return 1
	}
	results, err := translationexchange.ReadResults(resultFile)
	closeResultErr := resultFile.Close()
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: %v\n", err)
		return 1
	}
	if closeResultErr != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: close Gemini result: %v\n", closeResultErr)
		return 1
	}

	validation, err := translationexchange.Validate(source, results)
	if err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: REJECTED: %v\n", err)
		return 1
	}
	if err := writeJSONLFile(options.out, func(w io.Writer) error {
		return translationexchange.WriteResults(w, validation.Rows)
	}); err != nil {
		fmt.Fprintf(stderr, "zill: gemini-check: write accepted batch: %v\n", err)
		return 1
	}
	for _, warning := range validation.Warnings {
		fmt.Fprintf(stderr, "zill: gemini-check: warning: %s\n", warning)
	}
	fmt.Fprintf(stdout, "OK: accepted %d records; %d uncertain; %d warnings; normalized result: %s\n", len(validation.Rows), validation.Uncertain, len(validation.Warnings), options.out)
	return 0
}

func parseGeminiExportOptions(args []string) (geminiExportOptions, error) {
	options := geminiExportOptions{batch: -1, count: 100}
	seenBatch, seenCount, seenOut := false, false, false
	for index := 0; index < len(args); index++ {
		name, value, hasEquals := strings.Cut(args[index], "=")
		next := func() (string, error) {
			if hasEquals {
				if value == "" {
					return "", fmt.Errorf("%s requires a value", name)
				}
				return value, nil
			}
			if index+1 >= len(args) {
				return "", fmt.Errorf("%s requires a value", name)
			}
			index++
			return args[index], nil
		}
		switch name {
		case "--batch":
			if seenBatch {
				return geminiExportOptions{}, fmt.Errorf("--batch may be specified only once")
			}
			seenBatch = true
			raw, err := next()
			if err != nil {
				return geminiExportOptions{}, err
			}
			options.batch, err = strconv.Atoi(raw)
			if err != nil || options.batch < 0 {
				return geminiExportOptions{}, fmt.Errorf("invalid --batch %q", raw)
			}
		case "--count":
			if seenCount {
				return geminiExportOptions{}, fmt.Errorf("--count may be specified only once")
			}
			seenCount = true
			raw, err := next()
			if err != nil {
				return geminiExportOptions{}, err
			}
			options.count, err = strconv.Atoi(raw)
			if err != nil || options.count < 1 || options.count > 150 {
				return geminiExportOptions{}, fmt.Errorf("--count must be between 1 and 150, got %q", raw)
			}
		case "--out":
			if seenOut {
				return geminiExportOptions{}, fmt.Errorf("--out may be specified only once")
			}
			seenOut = true
			var err error
			options.out, err = next()
			if err != nil {
				return geminiExportOptions{}, err
			}
		default:
			return geminiExportOptions{}, fmt.Errorf("unknown argument %q", args[index])
		}
	}
	if !seenBatch || !seenOut || options.out == "" {
		return geminiExportOptions{}, fmt.Errorf("--batch and --out are required")
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
	// Remove only the non-canonical exchange output path. Repository TOML files
	// are never passed here by the command itself.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace old output: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install output: %w", err)
	}
	keep = true
	return nil
}
