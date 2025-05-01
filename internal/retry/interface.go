package retry

import "learningGolang/pkg/message"

type RetryManager interface {
	HandleFailedMessage(msg message.Message) error
	EnqueueIntoRetryQueue(msg message.Message) error
	SetMainChannel(ch chan message.Message)
	StartRetryManager()
	StopRetryManager()
}

