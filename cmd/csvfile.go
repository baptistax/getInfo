package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"
)

const csvSepLine = "sep=,"

// Creates file if needed (BOM + sep=,) and upgrades old ';' CSVs.
func ensureCSVReady(path string, header []string) (*os.File, error) {
	info, statErr := os.Stat(path)
	if os.IsNotExist(statErr) || (statErr == nil && info.Size() == 0) {
		f, e := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if e != nil {
			return nil, e
		}
		_, _ = f.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM for Excel/PT-BR
		_, _ = f.WriteString(csvSepLine + "\r\n")
		w := csv.NewWriter(f)
		w.UseCRLF = true
		w.Comma = ','
		if e := w.Write(header); e != nil {
			f.Close()
			return nil, e
		}
		w.Flush()
		if e := w.Error(); e != nil {
			f.Close()
			return nil, e
		}
		f.Close()
		return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	}

	// Upgrade existing file: add "sep=," and convert ';'->',' if necessary.
	data, e := os.ReadFile(path)
	if e == nil {
		content := string(data)
		hasSep := strings.HasPrefix(strings.TrimLeft(content, "\xEF\xBB\xBF"), "sep=")
		if !hasSep {
			firstNL := strings.IndexByte(content, '\n')
			firstLine := content
			if firstNL >= 0 {
				firstLine = strings.TrimRight(content[:firstNL], "\r\n")
			}
			oldComma := ','
			if strings.Contains(firstLine, ";") && !strings.Contains(firstLine, ",") {
				oldComma = ';'
			}
			conv := content
			if oldComma == ';' {
				conv = strings.ReplaceAll(content, ";", ",")
			}
			var buf bytes.Buffer
			buf.Write([]byte{0xEF, 0xBB, 0xBF})
			buf.WriteString(csvSepLine + "\r\n")
			buf.WriteString(strings.TrimLeft(conv, "\xEF\xBB\xBF"))
			_ = os.WriteFile(path, buf.Bytes(), 0644)
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
}

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
