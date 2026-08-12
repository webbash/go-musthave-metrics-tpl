package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileObserver struct {
	path string
}

func NewFileObserver(path string) *FileObserver {
	return &FileObserver{
		path: path,
	}
}

func (o *FileObserver) Observe(e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	data = append(data, '\n')

	file, err := os.OpenFile(
		o.path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}
