package levelqueue

import "time"

var dftCreateOption = SimpleQueueCreateOption{
	RetryIntervalWhenPullFailed: time.Second * 3,
}
