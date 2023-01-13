package levelqueue

import (
	"context"

	"github.com/ledisdb/ledisdb/ledis"
)

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
