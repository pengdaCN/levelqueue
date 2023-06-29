package levelqueue

import "errors"

var (
	NoBaseDbConfig = errors.New("not configuration")
	NoConfigBaseDb = errors.New("not config base db")
	QueueNameEmpty = errors.New("queue name cannot be empty")
	QueuePushDeny  = errors.New("push is deny")
)
