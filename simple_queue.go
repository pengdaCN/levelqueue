package levelqueue

import (
	"context"
	"sync"
	"time"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

var _ SimpleQueue = (*simpleQueue)(nil)

type MainPullLifetimeStrategy int32
type PullStrategy func(*simpleQueue) error

const (
	MainPullLifetimeStrategyGeneric MainPullLifetimeStrategy = iota + 1
	MainPullLifetimeStrategyOnce
)

type simpleQueue struct {
	ldb                         *ledis.DB
	name                        string
	mu                          sync.RWMutex
	globalPopCh                 chan []byte
	retryIntervalWhenPullFailed time.Duration
	ctx                         context.Context
	globalChCtx                 context.Context
	globalChCancel              context.CancelFunc
	pullStrategy                func(*simpleQueue) error
	mainPullLifetimeStrategy    MainPullLifetimeStrategy
}

type SimpleQueueCreateOption struct {
	LdisConfig                  *lediscfg.Config
	Ldis                        *ledis.Ledis
	DbIdx                       int
	LDb                         *ledis.DB
	ctx                         context.Context
	RetryIntervalWhenPullFailed time.Duration
	MainPullLifetimeStrategy    MainPullLifetimeStrategy
	PullStrategy                PullStrategy
	closeDelay                  time.Duration
}

func (s *SimpleQueueCreateOption) setup() (err error) {
	if s.LDb == nil && s.Ldis == nil {
		if s.LdisConfig == nil {
			return NoBaseDbConfig
		}
		s.Ldis, err = ledis.Open(s.LdisConfig)
		if err != nil {
			return err
		}
	}

	if s.LDb == nil {
		if s.Ldis == nil {
			return NoConfigBaseDb
		}

		s.LDb, err = s.Ldis.Select(s.DbIdx)
		if err != nil {
			return err
		}
	}

	if s.ctx == nil {
		s.ctx = context.Background()
	}

	if s.PullStrategy == nil {
		s.PullStrategy = dftPullStrategy
	}

	return nil
}

type CreateOption = func(opts *SimpleQueueCreateOption) error

func NewSimpleQueue(name string, createOpts ...CreateOption) (SimpleQueue, error) {
	if name == "" {
		return nil, QueueNameEmpty
	}

	opts := dftCreateOption
	for _, _fn := range createOpts {
		if err := _fn(&opts); err != nil {
			return nil, err
		}
	}

	if err := opts.setup(); err != nil {
		return nil, err
	}

	var (
		ctx    = opts.ctx
		cancel context.CancelFunc
	)

	if opts.closeDelay != 0 {
		ctx, cancel = context.WithCancel(context.Background())
		go delegateCtx(opts.ctx, cancel, opts.closeDelay)
	}

	queue := simpleQueue{
		name:                        name,
		ldb:                         opts.LDb,
		ctx:                         ctx,
		retryIntervalWhenPullFailed: opts.RetryIntervalWhenPullFailed,
		pullStrategy:                opts.PullStrategy,
		mainPullLifetimeStrategy:    opts.MainPullLifetimeStrategy,
	}

	return &queue, nil
}

func delegateCtx(listenCtx context.Context, cancel context.CancelFunc, delay time.Duration) {
	<-listenCtx.Done()

	time.Sleep(delay)

	cancel()
}

func (s *simpleQueue) BreakGlobalPopCh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.globalChCancel != nil {
		s.globalChCancel()
	}

	s.globalPopCh = nil
	s.globalChCtx = nil
	s.globalChCancel = nil
}

func (s *simpleQueue) Name() string {
	return s.name
}

func (s *simpleQueue) Push(data []byte) error {
	if err := s.pullStrategy(s); err != nil {
		return err
	}

	_, err := s.ldb.RPush([]byte(s.name), data)
	return err
}

func (s *simpleQueue) Pop() ([]byte, error) {
	return s.ldb.LPop([]byte(s.name))
}

func (s *simpleQueue) PopWithTimeout(timeout time.Duration) ([]byte, error) {
	const (
		popInterval = time.Millisecond * 10
	)

	var (
		times = int(timeout / popInterval)
	)
	if times < 0 {
		times = 1
	}

	for i := 0; i < times; i++ {
		bs, err := s.Pop()
		if err != nil {
			return nil, err
		}

		if len(bs) == 0 {
			time.Sleep(time.Millisecond * 10)
			continue
		}

		return bs, nil
	}

	return nil, context.DeadlineExceeded
}

func (s *simpleQueue) BPop() []byte {
	ch := s.GlobalPopCh()

	return <-ch
}

func (s *simpleQueue) GlobalPopCh() <-chan []byte {
	s.mu.RLock()
	if s.globalPopCh != nil {
		s.mu.RUnlock()
		return s.globalPopCh
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.globalPopCh != nil {
		return s.globalPopCh
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.globalChCtx = ctx
	s.globalChCancel = cancel

	// start main customer
	ch := make(chan []byte)
	go s.mainPull(ch)

	return ch
}

func (s *simpleQueue) NewPopCh(ctx context.Context) <-chan []byte {
	return s._newPopCh(ctx, 0)
}

func (s *simpleQueue) NewPopChWithCapacity(ctx context.Context, _cap int) <-chan []byte {
	return s._newPopCh(ctx, _cap)
}

func (s *simpleQueue) Len() (int64, error) {
	return s.ldb.LLen([]byte(s.name))
}

func (s *simpleQueue) Clear() error {
	return s.ldb.LTrim([]byte(s.name), 1, 0)
}

func (s *simpleQueue) _newPopCh(ctx context.Context, _cap int) <-chan []byte {
	if _cap < 0 {
		panic("cap must be greater than 0")
	}

	ch := make(chan []byte, _cap)

	go s.minorPull(ctx, ch)

	return ch
}

func (s *simpleQueue) mainPull(ch chan []byte) {
	defer close(ch)

	gCtx := s.globalChCtx
	for {
		select {
		case <-s.ctx.Done():
			s.globalChCancel()
			switch s.mainPullLifetimeStrategy {
			case MainPullLifetimeStrategyGeneric:
			case MainPullLifetimeStrategyOnce:
				s.oncePull(ch)
			default:
			}

			return
		case <-gCtx.Done():
			return
		default:
		}

		bs, err := s.PopWithTimeout(time.Second)
		if err != nil {
			select {
			case <-s.ctx.Done():
				continue
			default:
			}

			time.Sleep(s.retryIntervalWhenPullFailed)
			continue
		}

		if len(bs) == 0 {
			continue
		}

		ch <- bs
	}
}

func (s *simpleQueue) oncePull(ch chan<- []byte) {
	for {
		bs, err := s.Pop()
		if err != nil {
			return
		}

		if len(bs) == 0 {
			return
		}

		ch <- bs
	}
}

func (s *simpleQueue) minorPull(ctx context.Context, ch chan []byte) {
	defer close(ch)

	mainCh := s.GlobalPopCh()

	for {
		select {
		case <-ctx.Done():
			return
		case bs, ok := <-mainCh:
			if !ok {
				return
			}

			ch <- bs
		}
	}
}
