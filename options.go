package levelqueue

import (
	"context"
	"time"

	"github.com/ledisdb/ledisdb/ledis"
)

func SetCloseDelay(dur time.Duration) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.closeDelay = dur

		return nil
	}
}

func DisableCloseDelay() CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.closeDelay = 0

		return nil
	}
}

func WithOnceLifetime() CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.MainPullLifetimeStrategy = MainPullLifetimeStrategyOnce
		opts.PullStrategy = onceLifetimePullStrategy

		return nil
	}
}

func WithDir(dir string) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		ldis, err := Open(dir)
		if err != nil {
			return err
		}

		opts.Ldis = ldis
		return nil
	}
}

func WithOwnLedis(ldis *ledis.Ledis) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.Ldis = ldis

		return nil
	}
}

func WithDBIdx(idx int) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.DbIdx = idx

		return nil
	}
}

func WithContext(ctx context.Context) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.ctx = ctx

		return nil
	}
}

func SetRetryInterval(interval time.Duration) CreateOption {
	return func(opts *SimpleQueueCreateOption) error {
		opts.RetryIntervalWhenPullFailed = interval

		return nil
	}
}
