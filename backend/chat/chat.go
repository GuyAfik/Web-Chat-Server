package chat

import (
	"fmt"
	"web-chat-server/backend/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const ServerSender = "Server"

type ChatServer struct {
	users    map[string]*User
	commands *Commands
	upgrader *websocket.Upgrader
}


func NewChatServer() *ChatServer {
	return &ChatServer{
		users: make(map[string]*User),
		commands: NewCommands(),
		upgrader: &websocket.Upgrader{
			ReadBufferSize: 512,
			WriteBufferSize: 512,
			CheckOrigin: func(r *http.Request) bool {
				log.Printf("%s %s%s %v\n", r.Method, r.Host, r.RequestURI, r.Proto)
				return r.Method == http.MethodGet
			},
		},
	}
}

func (c *ChatServer) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln("Error on websocket connection:", err.Error())
	}

	keys := r.URL.Query()
	username := keys.Get("username")
	if strings.TrimSpace(username) == "" {
		username = fmt.Sprintf("anonnymous-%d", utils.GetRandomI64())
	}
	user := NewUser(username, conn, c.commands)

	c.commands.join <- user

	user.HandleUser()
}

func (c *ChatServer) Run() {
	for {
		select {
		case user := <-c.commands.join:
			c.add(user)
		case message := <-c.commands.messages:
			c.processUserRequest(message)
		case user := <-c.commands.leave:
			c.disconnect(user)
		}
	}
}

func (c *ChatServer) processUserRequest(message *Message) {
	username := message.Sender
	switch message.Body {
	case "!whoiswithme":
		c.users[username].Write(NewMessage(c.whoIsWithUser(username), ServerSender))
	case "!whoami":
		c.users[username].Write(NewMessage(fmt.Sprintf("You are the user: %s", username), ServerSender))
	default: 
		c.broadcast(message)
	}

}


func (c *ChatServer) whoIsWithUser(excludedUsername string) string {
	otherClients := make([]string, 0)
	for username := range c.users {
		if username != excludedUsername {
			otherClients = append(otherClients, username)
		} 
	}
	if len(otherClients) == 0 {
		return "No other clients"
	}
	return utils.StringArrayToString(otherClients, ", ")
}

func (c *ChatServer) add(user *User) {
	if _, ok := c.users[user.Username]; !ok {
		c.users[user.Username] = user

		body := fmt.Sprintf("%s joined the chat", user.Username)
		c.broadcast(NewMessage(body, "Server"))
	}
}

func (c *ChatServer) broadcast(message *Message) {
	log.Printf("Broadcast message: %v\n", message)
	for _, user := range c.users {
		user.Write(message)
	}
}

func (c *ChatServer) disconnect(user *User) {
	if _, ok := c.users[user.Username]; ok {
		defer user.Conn.Close()
		delete(c.users, user.Username)

		body := fmt.Sprintf("%s left the chat", user.Username)
		c.broadcast(NewMessage(body, ServerSender))
	}
}

func Start(port string) {

	log.Printf("Chat listening on http://localhost%s\n", port)

	chatServer := NewChatServer()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Go WebChat!"))
	})

	http.HandleFunc("/chat", chatServer.Handler)

	go chatServer.Run()

	log.Fatal(http.ListenAndServe(port, nil))
}