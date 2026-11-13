package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"pos/internal/application"
)

type CSVReportExporter struct{}

var _ application.ReportExporter = (*CSVReportExporter)(nil)

func NewCSVReportExporter() *CSVReportExporter {
	return &CSVReportExporter{}
}

func (e *CSVReportExporter) Format() string {
	return "csv"
}

func (e *CSVReportExporter) Export(rows []application.StudentReportRow, destination string) error {
	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("создать файл %q: %w", destination, err)
	}
	writer := csv.NewWriter(file)
	header := []string{"student_id", "student_name", "final_grade", "strategy"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return fmt.Errorf("записать заголовок CSV: %w", err)
	}
	for _, row := range rows {
		record := []string{
			row.StudentID,
			row.StudentName,
			strconv.FormatFloat(row.FinalGrade, 'f', 2, 64),
			row.Strategy,
		}
		if err := writer.Write(record); err != nil {
			file.Close()
			return fmt.Errorf("записать строку CSV: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		file.Close()
		return fmt.Errorf("записать CSV в файл %q: %w", destination, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("закрыть файл %q: %w", destination, err)
	}
	return nil
}
