package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents output format type
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatWide Format = "wide"
	FormatTable Format = "table"
)

// Formatter handles output formatting
type Formatter struct {
	format Format
	out    io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format string, out io.Writer) *Formatter {
	f := Format(strings.ToLower(format))
	switch f {
	case FormatJSON, FormatYAML, FormatWide, FormatTable:
		return &Formatter{format: f, out: out}
	default:
		return &Formatter{format: FormatTable, out: out}
	}
}

// Print outputs data in the specified format
func (f *Formatter) Print(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.printJSON(data)
	case FormatYAML:
		return f.printYAML(data)
	case FormatWide:
		return f.printWide(data)
	case FormatTable:
		return f.printTable(data)
	default:
		return f.printTable(data)
	}
}

// printJSON outputs data as JSON
func (f *Formatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(f.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printYAML outputs data as YAML
func (f *Formatter) printYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.out)
	defer encoder.Close()
	return encoder.Encode(data)
}

// printWide outputs data in wide table format (with more columns)
func (f *Formatter) printWide(data interface{}) error {
	return f.printTable(data)
}

// printTable outputs data in table format
func (f *Formatter) printTable(data interface{}) error {
	w := tabwriter.NewWriter(f.out, 0, 0, 3, ' ', 0)

	switch v := data.(type) {
	case []map[string]interface{}:
		return f.printMapTable(w, v)
	case map[string]interface{}:
		return f.printSingleMap(w, v)
	case []interface{}:
		return f.printInterfaceSlice(w, v)
	default:
		return json.NewEncoder(f.out).Encode(data)
	}
}

// printMapTable prints a slice of maps as a table
func (f *Formatter) printMapTable(w *tabwriter.Writer, data []map[string]interface{}) error {
	if len(data) == 0 {
		fmt.Fprintln(w, "No records found")
		return w.Flush()
	}

	// Get column headers from first row
	var headers []string
	firstRow := data[0]
	for key := range firstRow {
		headers = append(headers, strings.ToUpper(key))
	}

	// Print headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range data {
		var values []string
		for _, key := range headers {
			val := row[strings.ToLower(key)]
			values = append(values, fmt.Sprintf("%v", val))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	return w.Flush()
}

// printSingleMap prints a single map as key-value pairs
func (f *Formatter) printSingleMap(w *tabwriter.Writer, data map[string]interface{}) error {
	fmt.Fprintln(w, "KEY\tVALUE")
	for key, value := range data {
		fmt.Fprintf(w, "%s\t%v\n", key, value)
	}
	return w.Flush()
}

// printInterfaceSlice prints a slice of interfaces
func (f *Formatter) printInterfaceSlice(w *tabwriter.Writer, data []interface{}) error {
	// Try to convert to maps for table format
	var maps []map[string]interface{}
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			maps = append(maps, m)
		}
	}

	if len(maps) > 0 {
		return f.printMapTable(w, maps)
	}

	// Otherwise, print as JSON
	return json.NewEncoder(f.out).Encode(data)
}

// PrintSuccess prints a success message
func PrintSuccess(out io.Writer, message string, args ...interface{}) {
	fmt.Fprintf(out, "✅ %s\n", fmt.Sprintf(message, args...))
}

// PrintError prints an error message
func PrintError(out io.Writer, message string, args ...interface{}) {
	fmt.Fprintf(out, "❌ %s\n", fmt.Sprintf(message, args...))
}

// PrintWarning prints a warning message
func PrintWarning(out io.Writer, message string, args ...interface{}) {
	fmt.Fprintf(out, "⚠️  %s\n", fmt.Sprintf(message, args...))
}

// PrintInfo prints an info message
func PrintInfo(out io.Writer, message string, args ...interface{}) {
	fmt.Fprintf(out, "ℹ️  %s\n", fmt.Sprintf(message, args...))
}

// TableWriter helps write tabular data
type TableWriter struct {
	w       *tabwriter.Writer
	headers []string
}

// NewTableWriter creates a new table writer
func NewTableWriter(out io.Writer) *TableWriter {
	return &TableWriter{
		w:       tabwriter.NewWriter(out, 0, 0, 3, ' ', 0),
		headers: []string{},
	}
}

// AddHeader adds a column header
func (tw *TableWriter) AddHeader(headers ...string) {
	tw.headers = append(tw.headers, headers...)
}

// AddRow adds a data row
func (tw *TableWriter) AddRow(values ...interface{}) {
	var strValues []string
	for _, v := range values {
		strValues = append(strValues, fmt.Sprintf("%v", v))
	}
	fmt.Fprintln(tw.w, strings.Join(strValues, "\t"))
}

// Render outputs the table
func (tw *TableWriter) Render() error {
	if len(tw.headers) > 0 {
		fmt.Fprintln(tw.w, strings.Join(tw.headers, "\t"))
	}
	return tw.w.Flush()
}
