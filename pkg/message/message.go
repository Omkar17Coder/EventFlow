package message

import "time"

type Message struct {
	ID        string
	Source    string
	Timestamp int64
	Payload   string
}

// time.Now().UnixMills();

func NewMessage(ID, Source, Payload string) *Message {
	return &Message{ID, Source, time.Now().UnixMilli(), Payload}

}
