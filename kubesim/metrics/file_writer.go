package metrics

import (
	"os"
)

// FileWriter writes metrics to a file.
type FileWriter struct {
	file      *os.File
	formatter Formatter
}

// NewFileWriter creates a new FileWriter instance with a file at the given path, and a formatter
// that formats metrics to a string. Returns err if failed to create a file.
func NewFileWriter(path string, formatter Formatter) (*FileWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:      file,
		formatter: formatter,
	}, nil
}

// FileName returns the name of file underlying this FileWriter.
func (w *FileWriter) FileName() string { return w.file.Name() }

func (w *FileWriter) Write(metrics Metrics) error {
	str, err := w.formatter.Format(metrics)
	if err != nil {
		return err
	}
	w.file.WriteString(str)
	w.file.Write([]byte{'\n'})

	return nil
}

var _ = Writer(&FileWriter{})
