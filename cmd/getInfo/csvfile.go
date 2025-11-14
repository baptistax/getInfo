package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"
)

const csvSepLine = "sep=,"

// ensureCSVReady makes sure a CSV file exists and is in the expected format.
// - If the file does not exist or is empty, it creates a new CSV with:
//   - UTF-8 BOM (for Excel compatibility in some locales)
//   - "sep=," header line
//   - provided header row
//
// - If the file exists, it upgrades legacy files by:
//   - adding "sep=," if missing
//   - converting ';' to ',' when it looks like the old separator.
//
// It always returns the file opened in append mode.
func ensureCSVReady(path string, header []string) (*os.File, error) {
	info, statErr := os.Stat(path)

	// If the file does not exist or is empty, create and initialize it.
	if os.IsNotExist(statErr) || (statErr == nil && info.Size() == 0) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		// Write BOM for Excel compatibility (especially in some non-English locales).
		_, _ = f.Write([]byte{0xEF, 0xBB, 0xBF})
		_, _ = f.WriteString(csvSepLine + "\r\n")

		w := csv.NewWriter(f)
		w.UseCRLF = true
		w.Comma = ','

		if err := w.Write(header); err != nil {
			f.Close()
			return nil, err
		}
		w.Flush()
		if err := w.Error(); err != nil {
			f.Close()
			return nil, err
		}

		f.Close()
		return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	}

	// If the file exists, try to upgrade its format if needed.
	data, err := os.ReadFile(path)
	if err == nil {
		content := string(data)

		// Check if already has a "sep=" line (ignoring an existing BOM).
		hasSep := strings.HasPrefix(strings.TrimLeft(content, "\xEF\xBB\xBF"), "sep=")
		if !hasSep {
			firstNL := strings.IndexByte(content, '\n')
			firstLine := content
			if firstNL >= 0 {
				firstLine = strings.TrimRight(content[:firstNL], "\r\n")
			}

			// Detect old separator: if header contains ';' and no ','.
			oldComma := ','
			if strings.Contains(firstLine, ";") && !strings.Contains(firstLine, ",") {
				oldComma = ';'
			}

			conv := content
			if oldComma == ';' {
				conv = strings.ReplaceAll(content, ";", ",")
			}

			var buf bytes.Buffer
			// Always write BOM + "sep=," + CRLF + content without old BOM.
			buf.Write([]byte{0xEF, 0xBB, 0xBF})
			buf.WriteString(csvSepLine + "\r\n")
			buf.WriteString(strings.TrimLeft(conv, "\xEF\xBB\xBF"))

			_ = os.WriteFile(path, buf.Bytes(), 0644)
		}
	}

	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
}

// appendCSVRow writes a single row to an already-open CSV file.
func appendCSVRow(f *os.File, row []string) error {
	w := csv.NewWriter(f)
	w.UseCRLF = true
	w.Comma = ','

	if err := w.Write(row); err != nil {
		return err
	}

	w.Flush()
	return w.Error()
}
