package collector

import (
	"encoding/csv"
	"io"
)

// NewTsvReader read Tab Separated Values using the csv package
func NewTsvReader(r io.Reader) *csv.Reader {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	return reader
}
