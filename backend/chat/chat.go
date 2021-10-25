package chat

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"web-chat-server/backend/utils"

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
		users:    make(map[string]*User),
		commands: NewCommands(),
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  512,
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
	username := strings.TrimSpace(keys.Get("username"))
	if username == "" {
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
	senderUser := message.Sender
	parsedMessageBody := utils.ParseMessageBody(message.Body, ":")
	switch parsedMessageBody[0] {
	case "!whoiswithme":
		c.response(c.whoIsWithUser(senderUser), ServerSender, senderUser)
	case "!whoami":
		c.response(fmt.Sprintf("You are the user %s", senderUser), ServerSender, senderUser)
	case "!privatemessage":
		if len(parsedMessageBody) < 3 {
			c.response(
				fmt.Sprintf("Message Body: %s does not have enough arguments", message.Body),
				ServerSender,
			)
		} else {
			body := parsedMessageBody[len(parsedMessageBody)-1]
			c.response(
				body,
				message.Sender,
				append(parsedMessageBody[1:len(parsedMessageBody)-1], senderUser)...,
			)
		}
	default:
		c.response(message.Body, senderUser)
	}

}

func (c *ChatServer) response(body, sender string, usernames ...string) {
	responseMessage := NewMessage(body, sender)
	if len(usernames) > 0 {
		if sender != ServerSender {
			responseMessage.Sender = fmt.Sprintf("Private message [%s]", sender)
		}
		c.privateMessage(responseMessage, usernames...)
	} else {
		c.broadcast(responseMessage)
	}
}

func (c *ChatServer) privateMessage(message *Message, usernames ...string) {
	log.Printf("Private message: %v, users: %v\n", message, usernames)
	for _, username := range usernames {
		if user, ok := c.users[username]; ok {
			user.Write(message)
		}
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
	return fmt.Sprintf("Other connected clients: %s", utils.StringArrayToString(otherClients, ", "))
}

func (c *ChatServer) add(user *User) {
	if _, ok := c.users[user.Username]; !ok {
		c.users[user.Username] = user

		body := fmt.Sprintf("%s joined the chat", user.Username)
		c.broadcast(NewMessage(body, ServerSender))
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

		body := fmt.Sprintf("%s has left the chat server :-(", user.Username)
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
