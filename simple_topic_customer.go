package levelqueue

import (
	"sync"
	"time"
)

var _ SimpleTopicCustomer = (*simpleTopicCustomer)(nil)

type simpleTopicCustomer struct {
	name     string
	topic    *simpleTopic
	cursor   int64
	cursorMu sync.Mutex
}

func (s *simpleTopicCustomer) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopicCustomer) Pop() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopicCustomer) PopWithTimeout(timeout time.Duration) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopicCustomer) PopCh() <-chan []byte {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopicCustomer) Cleanup() error {
	//TODO implement me
	panic("implement me")
}
