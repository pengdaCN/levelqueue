package extutil

import (
	"sync"

	"github.com/ledisdb/ledisdb/ledis"
	"github.com/robfig/cron/v3"
)

var (
	_globalCron = cron.New(cron.WithSeconds())
	_run        sync.Once
)

func _startGlobalCron() {
	_run.Do(_globalCron.Start)
}

func AutoCompact(ldis *ledis.Ledis) *ledis.Ledis {
	AdvAutoCompact(ldis, "0 0 4 * * *", nil)

	return ldis
}

func AdvAutoCompact(ldis *ledis.Ledis, crontab string, errHandle func(error)) *ledis.Ledis {
	_startGlobalCron()

	err := ldis.CompactStore()
	if err != nil {
		if errHandle != nil {
			errHandle(err)
		}
	}

	if _, err := _globalCron.AddFunc(crontab, func() {
		if err := ldis.CompactStore(); err != nil {
			if errHandle != nil {
				errHandle(err)
			}
		}
	}); err != nil {
		panic(err)
	}

	return ldis
}
