package chat

import (
	"fmt"
	"web-chat-server/backend/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	users    map[string]*User
	commands *Commands
}


func NewChatServer(commands *Commands) *ChatServer {
	return &ChatServer{
		users: make(map[string]*User),
		commands: commands,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("%s %s%s %v\n", r.Method, r.Host, r.RequestURI, r.Proto)
		return r.Method == http.MethodGet
	},
}

func (c *ChatServer) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln("Error on websocket connection:", err.Error())
	}

	keys := r.URL.Query()
	username := keys.Get("username")
	if strings.TrimSpace(username) == "" {
		username = fmt.Sprintf("anon-%d", utils.GetRandomI64())
	}

	user := NewUser(username, conn, c.commands)

	c.commands.join <- user

	user.Read()
}

func (c *ChatServer) Run() {
	for {
		select {
		case user := <-c.commands.join:
			c.add(user)
		case message := <-c.commands.messages:
			c.broadcast(message)
		case user := <-c.commands.leave:
			c.disconnect(user)
		}
	}
}

func (c *ChatServer) add(user *User) {
	if _, ok := c.users[user.Username]; !ok {
		c.users[user.Username] = user

		body := fmt.Sprintf("%s join the chat", user.Username)
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
		c.broadcast(NewMessage(body, "Server"))
	}
}

func Start(port string) {

	log.Printf("Chat listening on http://localhost%s\n", port)

	c := NewChatServer(NewCommands())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Go WebChat!"))
	})

	http.HandleFunc("/chat", c.Handler)

	go c.Run()

	log.Fatal(http.ListenAndServe(port, nil))
}