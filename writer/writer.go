package writer

import (
	"fmt"
	"sync"
)

// Writer write received log
type Writer interface {

	// WriteLine write the log as newline terminated string
	// The implementation should be thread safe
	WriteLine(string) error
}

type stdout struct {
	mu sync.Mutex
}

// NewStdoutWriter return writer that write to stdout
func NewStdoutWriter() (Writer, error) {
	return new(stdout), nil
}

func (w *stdout) WriteLine(log string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Println(log)
	return nil
}
