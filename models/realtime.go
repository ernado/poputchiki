package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Message struct {
	Id          bson.ObjectId `json:"id"          bson:"_id"`
	Chat        bson.ObjectId `json:"chat"        bson:"chat"`
	User        bson.ObjectId `json:"-"           bson:"user"`
	Origin      bson.ObjectId `json:"origin"      bson:"origin"`
	Destination bson.ObjectId `json:"destination" bson:"destination"`
	Read        bool          `json:"read"        bson:"read"`
	Time        time.Time     `json:"time"        bson:"time"`
	Text        string        `json:"text"        bson:"text"`
	Invite      bool          `json:"invite"      bson:"invite"`
}

type RealtimeEvent struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
	Time time.Time   `json:"time"`
}

type Dialog struct {
	message *Message `json:"message"`
	user    *User    `json:"user"`
}

type UnreadCount struct {
	Count int `json:"count"`
}

type ProgressMessage struct {
	Id       bson.ObjectId `json:"id,omitempty"`
	Progress float32       `json:"progress"`
}

type MessageSendBlacklisted struct {
	Id bson.ObjectId `json:"id"`
}
