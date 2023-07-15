package levelqueue

import "time"

var dftCreateOption = SimpleQueueCreateOption{
	RetryIntervalWhenPullFailed: time.Second * 3,
	MainPullLifetimeStrategy:    MainPullLifetimeStrategyGeneric,
	closeDelay:                  time.Second * 2,
}
