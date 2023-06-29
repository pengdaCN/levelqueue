package levelqueue

func dftPullStrategy(*simpleQueue) error {
	return nil
}

func onceLifetimePullStrategy(queue *simpleQueue) error {
	select {
	case <-queue.ctx.Done():
		return QueuePushDeny
	default:
		return nil
	}
}
