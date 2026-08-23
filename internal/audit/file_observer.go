package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileObserver appends audit events to a file as newline-delimited JSON.
type FileObserver struct {
	file *os.File
	mu   *sync.Mutex
}

// NewFileObserver creates an observer that writes audit events to path.
func NewFileObserver(file *os.File) *FileObserver {
	return &FileObserver{
		file: file,
		mu:   &sync.Mutex{},
	}
}

// Observe appends an audit event to the configured file.
func (o *FileObserver) Observe(e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	data = append(data, '\n')

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, err := o.file.Write(data); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}
