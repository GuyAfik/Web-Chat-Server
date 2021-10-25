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
}

func NewUser(username string, connection *websocket.Conn, commands *Commands) *User {
	return &User{
		Username: username,
		Conn:     connection,
		Commands: commands,
	}
}

func (u *User) HandleUser() {
	for {
		if _, message, err := u.Read(); err != nil {
			log.Println("Error on read message:", err.Error())
			break
		} else {
			u.Commands.messages <- NewMessage(string(message), u.Username)
		}
	}

	u.Commands.leave <- u
}


func (u *User) Read() (int, []byte, error) {
	return u.Conn.ReadMessage()
}

func (u *User) Write(message *Message) {
	jsonMessage, _ := json.Marshal(message)

	if err := u.Conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
		log.Println("Error on write message:", err.Error())
	}
}
