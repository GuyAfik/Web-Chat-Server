package chat

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type User struct {
	Username string
	Conn     *websocket.Conn
	Commands *Commands
	// Global   *ChatServer
}

func NewUser(username string, connection *websocket.Conn, commands *Commands) *User {
	return &User{
		Username: username,
		Conn: connection,
		Commands: commands,
	}
}

func (u *User) Read() {
	for {
		if _, message, err := u.Conn.ReadMessage(); err != nil {
			log.Println("Error on read message:", err.Error())

			break
		} else {
			u.Commands.messages <- NewMessage(string(message), u.Username)
		}
	}

	u.Commands.leave <- u
}

func (u *User) Write(message *Message) {
	b, _ := json.Marshal(message)

	if err := u.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Println("Error on write message:", err.Error())
	}
}
