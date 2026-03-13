package csvutil

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
)

// Writer wraps csv.Writer with a bytes.Buffer for in-memory CSV generation.
type Writer struct {
	buf bytes.Buffer
	w   *csv.Writer
}

// New creates a new CSV writer.
func New() *Writer {
	cw := &Writer{}
	cw.w = csv.NewWriter(&cw.buf)
	return cw
}

// Row writes a single row of string values.
func (c *Writer) Row(fields ...string) {
	c.w.Write(fields)
}

// EmptyRow writes a blank row.
func (c *Writer) EmptyRow() {
	c.w.Write([]string{""})
}

// Flush flushes the underlying csv.Writer.
func (c *Writer) Flush() {
	c.w.Flush()
}

// Bytes returns the CSV content as bytes (call Flush first).
func (c *Writer) Bytes() []byte {
	return c.buf.Bytes()
}

// ServeHTTP writes the CSV as an HTTP response with Content-Disposition attachment.
func (c *Writer) ServeHTTP(w http.ResponseWriter, filename string) {
	c.w.Flush()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(c.buf.Bytes())
}

// F formats a float64 with 2 decimal places.
func F(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// F1 formats a float64 with 1 decimal place.
func F1(v float64) string {
	return fmt.Sprintf("%.1f", v)
}
