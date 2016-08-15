package comm

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/ubuntu/face-detection-demo/messages"

	"golang.org/x/net/websocket"
)

// we just kill the webserver on shutdown, pending requests will just be dropped

// StartServer starts in a goroutine both webserver and websocket handlers
func StartServer(rootdir string, actions chan<- *messages.Action) {
	go func() {
		server := NewWSServer("/api", actions)
		go server.Listen()
		http.Handle("/", http.FileServer(http.Dir(path.Join(rootdir, "www"))))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("Couldn't start webserver:", err)
		}
	}()
}

// WSServer maintaining the web socket server
type WSServer struct {
	match        string
	pastmessages []*messages.WSMessage
	clients      map[int]*Client
	addCh        chan *Client
	delCh        chan *Client
	sendAllCh    chan *messages.WSMessage
	doneCh       chan bool
	errCh        chan error
	actions      chan<- *messages.Action
}

// NewWSServer create a new ws server
func NewWSServer(patternURL string, actions chan<- *messages.Action) *WSServer {
	pastmessages := []*messages.WSMessage{}
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *messages.WSMessage)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &WSServer{
		match,
		pastmessages,
		clients,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
		actions,
	}
}

// TODO: load past messages and load initial list of messages

// Del removes a client from the connected list
func (s *WSServer) Del(c *Client) {
	s.delCh <- c
}

// SendAll signal a new message to send to all clients
func (s *WSServer) SendAll(msg *messages.WSMessage) {
	s.sendAllCh <- msg
}

// Done signal we are shutting down the ws server
func (s *WSServer) Done() {
	s.doneCh <- true
}

// Err signals about client errors
func (s *WSServer) Err(err error) {
	s.errCh <- err
}

func (s *WSServer) add(c *Client) {
	s.addCh <- c
}

func (s *WSServer) sendPastMessages(c *Client) {
	s.pastmessages = append(s.pastmessages, &messages.WSMessage{Author: "didrocks", Body: "hehe hehe"})
	for _, msg := range s.pastmessages {
		c.Send(msg)
	}

// NewAction is an action received by one client, sent to the main system process
func (s *WSServer) NewAction(actionmsg *messages.Action) {
	s.actions <- actionmsg
}

func (s *WSServer) sendAllClients(msg *messages.WSMessage) {
	for _, c := range s.clients {
		c.Send(msg)
	}
}

func (s *WSServer) onNewClient(ws *websocket.Conn) {
	defer func() {
		err := ws.Close()
		if err != nil {
			s.errCh <- err
		}
	}()

	client, err := NewClient(ws, s)
	if err != nil {
		fmt.Println("Couldn't accept connection:", err)
	}
	s.add(client)
	client.Listen()
}

// Listen to new ws client conn
func (s *WSServer) Listen() {
	log.Println("Start ws listener...")
	http.Handle(s.match, websocket.Handler(s.onNewClient))

	for {
		select {
		case c := <-s.addCh:
			log.Println("New client connected")
			s.clients[c.id] = c
			log.Println("Now", len(s.clients), "clients connected.")
			s.sendPastMessages(c)

		case c := <-s.delCh:
			log.Println("Disconnected client")
			delete(s.clients, c.id)

		// broadcast message to all clients
		case msg := <-s.sendAllCh:
			log.Println("Send to all clients:", msg)
			s.pastmessages = append(s.pastmessages, msg)
			s.sendAllClients(msg)

		case err := <-s.errCh:
			log.Println("Error:", err.Error())

		case <-s.doneCh:
			return
		}
	}
}
