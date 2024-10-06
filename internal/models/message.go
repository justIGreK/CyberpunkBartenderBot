package models

import (
	"time"
)

type MessageParts struct {
	Time   string
	Action string
}
type Reminder struct {
	ID       string    `bson:"_id,omitempty"`
	ChatID   int64     `bson:"chat_id"`
	Action   string    `bson:"action"`
	Time     time.Time `bson:"time"`
	IsActive bool      `bson:"is_active"` // Новое поле для отслеживания статуса
}

type Command struct {
    Command     string `json:"command"`
    Description string `json:"description"`
}