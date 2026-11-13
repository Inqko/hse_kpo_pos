package export

import (
	"encoding/json"
	"fmt"
	"os"

	"pos/internal/application"
)

type JSONReportExporter struct{}

var _ application.ReportExporter = (*JSONReportExporter)(nil)

func NewJSONReportExporter() *JSONReportExporter {
	return &JSONReportExporter{}
}

func (e *JSONReportExporter) Format() string {
	return "json"
}

func (e *JSONReportExporter) Export(rows []application.StudentReportRow, destination string) error {
	if rows == nil {
		rows = []application.StudentReportRow{}
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return fmt.Errorf("сериализовать отчёт: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return fmt.Errorf("записать файл %q: %w", destination, err)
	}
	return nil
}
