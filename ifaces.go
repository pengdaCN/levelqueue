package levelqueue

import (
	"context"
	"time"
)

type SimpleQueue interface {
	// Name 返回队列的名字
	Name() string
	Push(data []byte) error
	// Pop 消费一个元素，不阻塞，若没有元素则返回nil
	Pop() ([]byte, error)
	// PopWithTimeout 消费一个元素，若没有元素则最大等待timeout的时间，若超时，则返回nil
	PopWithTimeout(timeout time.Duration) ([]byte, error)
	// BPop 消费一个元素，若没有元素则阻塞，直到有元素，若返回nil，则队列可能已经被关闭了
	BPop() []byte
	// GlobalPopCh 返回全局的消费队列，唯一
	GlobalPopCh() <-chan []byte
	// NewPopCh ctx 若cancel，则返回的channel 会被关闭，该channel的数据来自GlobalPopCh
	NewPopCh(ctx context.Context) <-chan []byte
	// NewPopChWithCapacity 与NewPopCh 相同，但是可以自定义channel的容量，若_cap 小于0则
	// panic
	NewPopChWithCapacity(ctx context.Context, _cap int) <-chan []byte
	// Len 返回队列中剩余的元素数量
	Len() (int64, error)
	// Clear 清空所有队列中所有的数据
	Clear() error
}

// OnceQueue 基本属于SimpleQueue一样，但是，当Ctx 被cancel，或则手动调用了Close，
// 则所有的写入操作都会被禁止，同时，在消息读取完毕后，chan会自动关闭
type OnceQueue interface {
	SimpleQueue
	Close()
}
