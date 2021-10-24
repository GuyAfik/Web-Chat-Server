package chat

import "web-chat-server/utils"

type Message struct {
	ID     int64  `json:"id"`
	Body   string `json:"body"`
	Sender string `json:"sender"`
}

type Commands struct {
	messages chan *Message
	join     chan *User
	leave    chan *User
}

func NewCommands() *Commands {
	return &Commands{
		messages: make(chan *Message),
		join: make(chan *User),
		leave: make(chan *User),
	}
}

func NewMessage(body string, sender string) *Message {
	return &Message{
		ID:     utils.GetRandomI64(),
		Body:   body,
		Sender: sender,
	}
}