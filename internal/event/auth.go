package event

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event interface {
	Key() []byte
	Value() ([]byte, error)
}

type AuthEvent struct {
	Type 	  string 	`json:"type"`
	Email 	  string 	`json:"email"`
	TimeStamp time.Time `json:"timestamp"`
}

func (ev AuthEvent) Key() []byte {
	return []byte(ev.Type)
}

func (ev AuthEvent) Value() ([]byte, error) {
	value, err := json.Marshal(ev)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event")
	}

	return value, nil
}