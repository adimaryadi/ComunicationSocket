package main

import (
	"ComunicationSocket/Utility"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/asim/go-micro/util/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
)

type ClientManager struct {
	clients    	map[*Client]bool
	broadcast   chan []byte
	register  	chan *Client
	unregister  chan *Client
}

type Client struct {
	id 	 	string
	socket  *websocket.Conn
	send 	chan []byte
}

type ConfigManagement struct {
	userName string `json: "userName,omitempty"`
	password string `json: "password,omitempty"`
	port     int64  `json: "port,omitempty"`
	part     string  `json: "part,omitempty"`
}

type Message struct {
	Sender    	string    `json: "sender,omitempty"`
	Recipient 	string 	  `json: "recipient, omitempty"`
	Content     string    `json: "content, omitempty"`
}

var manager   =  ClientManager{
	broadcast:   make(chan []byte),
	register: 	 make(chan *Client),
	unregister:  make(chan *Client),
	clients: 	 make(map[*Client]bool),
}

func main() {
	//fmt.Println("Starting Apps....")
	//go manager.start()
	//http.HandleFunc("/ws",wsPage)
	//http.ListenAndServe(":8080",nil)

	dat, err :=  os.Open("config.ind")
	if err == nil {
		log.Info("Running",dat)
	}
	if err != nil {
		configRun()
	}
}

func configRun() {
	reader := readerInput(&ConfigManagement{part: "username"})
	if len(reader) <= 4 {
		fmt.Println("username required 4 character")
		configRun()
		return
	}
	save := &ConfigManagement{userName: reader}
	configPassword(save)
	return
}

func configPassword(configM *ConfigManagement) {
	reader := readerInput(&ConfigManagement{part: "password"})
	if len(reader) <= 3 {
		fmt.Println("password required 4 character")
		log.Info(len(reader))
	    configPassword(configM)
		return
	}
	save := &ConfigManagement{userName: configM.userName, password: reader}
	configPort(save)
	return
}

func configPort(configM *ConfigManagement) {
//
}

func readerInput(configM *ConfigManagement) string {
	reader   := bufio.NewReader(os.Stdin)
	fmt.Print("Enter "+configM.part+": ")
	input,_ := reader.ReadString('\n')
	return input
}

func (manager *ClientManager) start() {
	for  {
		select {
			case conn := <-manager.register:
				manager.clients[conn] 	= 	true
				jsonMessage, _ := json.Marshal(&Message{Sender: conn.id ,Content: " Koneksi baru Telah Terhubung."})
				log.Info("Registrasi => ",string(jsonMessage))
				manager.send(jsonMessage,conn)
			case conn := <-manager.unregister:
				if _, ok := manager.clients[conn]; ok {
					close(conn.send)
					delete(manager.clients,conn)
					jsonMessage, _ := json.Marshal(&Message{Sender: conn.id,Content: " koneksi Terputus."})
					log.Info("Registrasi => ",string(jsonMessage))
					manager.send(jsonMessage,conn)
				}
			case message := <-manager.broadcast:
				for conn := range manager.clients {
					select {
						case conn.send <- message:
						default:
							close(conn.send)
							delete(manager.clients,conn)
						}
				}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for  {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for  {
		select {
			case message, ok := <-c.send:
				if !ok {
					c.socket.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				c.socket.WriteMessage(websocket.TextMessage,message)
		}
	}
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, error  := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res,req,nil)
	requstQuery  := req.URL.Query()
	everySpaceToPlus :=  Utility.SpaceFieldsJoin(requstQuery["Authorization"][0])
	decryption       :=  Utility.Dencrytion(requstQuery["deviceID"][0],everySpaceToPlus)
	log.Info(decryption)
	log.Info(requstQuery)
	if error != nil {
		http.NotFound(res,req)
		return
	}


	client := &Client{id: uuid.New().String(), socket: conn, send: make(chan []byte)}
	manager.register <- client
	go client.read()
	go client.write()
}