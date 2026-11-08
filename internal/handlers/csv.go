package handlers

import (
	"encoding/csv"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode"
)

func safeCSVCell(value string) string {
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") {
		return "'" + value
	}
	return value
}
func encodeCSV(w io.Writer, header []string, rows [][]string) error {
	out := csv.NewWriter(w)
	if err := out.Write(header); err != nil {
		return err
	}
	for _, row := range rows {
		safe := make([]string, len(row))
		for i, v := range row {
			safe[i] = safeCSVCell(v)
		}
		if err := out.Write(safe); err != nil {
			return err
		}
	}
	out.Flush()
	return out.Error()
}
func writeCSV(w http.ResponseWriter, filename string, header []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	if err := encodeCSV(w, header, rows); err != nil {
		slog.Error("CSV response failed", "err", err)
	}
}
