package levelqueue

import "sync"

var _ SimpleTopic = (*simpleTopic)(nil)

type simpleTopic struct {
	name       string
	offset     int64
	offsetLock sync.RWMutex
}

func (s *simpleTopic) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopic) Push(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopic) OpenCustomer(name string) (SimpleTopicCustomer, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopic) UniqCustomer() (SimpleTopicCustomer, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTopic) Close() error {
	//TODO implement me
	panic("implement me")
}
