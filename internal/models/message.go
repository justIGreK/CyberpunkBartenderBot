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
	Time     time.Time `bson:"utc_time"`
	OriginalTime time.Time `bson:"time"`
	IsActive bool      `bson:"is_active"`
}

type Command struct {
    Command     string `json:"command"`
    Description string `json:"description"`
}

type ChatTimezone struct{
	ChatID int64 `bson:"chat_id"`
	Latitude float64 `bson:"lat"`
	Longitude float64 `bson:"long"`
	Diff_hour int `bson:"diff_hour"`
}

type UserPageState struct {
    ChatID int64 `bson:"chat_id"`
    Page   int   `bson:"page"`
}