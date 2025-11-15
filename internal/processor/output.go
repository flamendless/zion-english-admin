package processor

import "strings"

type OutputFormat int

const (
	OutputFormatUndefined OutputFormat = iota
	OutputFormatCSV
	OutputFormatXLSX
)

type RecordWriter interface {
	WriteRecords(records []ClassRecord, outputPath string, colIndices ColumnIndices) error
}

func GetWriter(format OutputFormat) RecordWriter {
	switch format {
	case OutputFormatCSV:
		return &CSVWriter{}
	case OutputFormatXLSX:
		return &XLSXWriter{}
	default:
		return &XLSXWriter{}
	}
}

func GetWriterFromPath(outputPath string) RecordWriter {
	format := GetOutputFormatFromPath(outputPath)
	return GetWriter(format)
}

func GetOutputFormatFromPath(outputPath string) OutputFormat {
	lowerPath := strings.ToLower(outputPath)
	if strings.HasSuffix(lowerPath, ".csv") {
		return OutputFormatCSV
	}
	if strings.HasSuffix(lowerPath, ".xlsx") {
		return OutputFormatXLSX
	}
	return OutputFormatUndefined
}

func SaveRecords(records []ClassRecord, outputPath string, colIndices ColumnIndices) error {
	writer := GetWriterFromPath(outputPath)
	return writer.WriteRecords(records, outputPath, colIndices)
}

func SaveRecordsWithFormat(records []ClassRecord, outputPath string, colIndices ColumnIndices, format OutputFormat) error {
	writer := GetWriter(format)
	return writer.WriteRecords(records, outputPath, colIndices)
}
