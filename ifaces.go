package levelqueue

import (
	"context"
	"time"
)

type SimpleQueue interface {
	Name() string
	Push(data []byte) error
	Pop() ([]byte, error)
	PopWithTimeout(timeout time.Duration) ([]byte, error)
	GlobalPopCh() <-chan []byte
	NewPopCh(ctx context.Context) <-chan []byte
	NewPopChWithCapacity(ctx context.Context, _cap int) <-chan []byte
	Len() (int64, error)
}
