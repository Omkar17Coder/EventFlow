package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   json.RawMessage        `json:"payload"`
	Retries   int                    `json:"retries"`
	CreatedAt time.Time              `json:"created_at"`
	MetaData  map[string]interface{} `json:"metadata"`
}

func NewMessage(id, msgType string, payload interface{}) (*Message, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:        id,
		Type:      msgType,
		Payload:   jsonPayload,
		CreatedAt: time.Now().UTC(),
		MetaData:  make(map[string]interface{}),
	}, nil
}
