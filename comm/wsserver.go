package comm

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ubuntu/face-detection-demo/appstate"
	"github.com/ubuntu/face-detection-demo/datastore"
	"github.com/ubuntu/face-detection-demo/messages"

	"golang.org/x/net/websocket"
)

// we just kill the webserver on shutdown, pending requests will just be dropped

// WSserv main ws client connections
var WSserv *WSServer

// WSServer maintaining the web socket server
type WSServer struct {
	patternURL string
	clients    map[int]*Client
	addCh      chan *Client
	delCh      chan *Client
	sendAllCh  chan *messages.WSMessage
	doneCh     chan interface{}
	errCh      chan error
	actions    chan<- *messages.Action
}

// NewWSServer create a new ws server
func NewWSServer(patternURL string, actions chan<- *messages.Action) *WSServer {
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *messages.WSMessage)
	doneCh := make(chan interface{})
	errCh := make(chan error)

	return &WSServer{
		patternURL,
		clients,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
		actions,
	}
}

// Del removes a client from the connected list
func (s *WSServer) Del(c *Client) {
	s.delCh <- c
}

// SendAllClients signal a new message to send to all clients
func (s *WSServer) SendAllClients(msg *messages.WSMessage) {
	s.sendAllCh <- msg
}

// Done signal we are shutting down the ws server
func (s *WSServer) Done() {
	close(s.doneCh)
}

// Err signals about client errors
func (s *WSServer) Err(err error) {
	s.errCh <- err
}

// NewAction is an action received by one client, sent to the main system process
func (s *WSServer) NewAction(actionmsg *messages.Action) {
	s.actions <- actionmsg
}

func (s *WSServer) add(c *Client) {
	s.addCh <- c
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

	// Main loop for client
	client.Listen()
}

// Listen to new ws client conn
func (s *WSServer) Listen() {
	log.Println("Start ws listener...")
	http.Handle(s.patternURL, websocket.Handler(s.onNewClient))

	for {
		select {
		// new client connected
		case c := <-s.addCh:
			log.Println("New client connected")
			s.clients[c.id] = c
			log.Println("Now", len(s.clients), "clients connected.")
			// send all stats messages
			c.Send(&messages.WSMessage{
				Type:          "init",
				AllStats:      datastore.DB.Stats,
				FaceDetection: datastore.FaceDetection(),
				RenderingMode: datastore.RenderingMode(),
				// camera is offsetted by 1 for the client
				Camera:           datastore.Camera() + 1,
				AvailableCameras: appstate.AvailableCameras,
				Broken:           appstate.BrokenMode})

		// client disconnected
		case c := <-s.delCh:
			log.Println("Disconnected client")
			delete(s.clients, c.id)

		// broadcast message to all clients
		case msg := <-s.sendAllCh:
			log.Println("Send to all clients:", msg)
			for _, c := range s.clients {
				c.Send(msg)
			}

		// error reported
		case err := <-s.errCh:
			log.Println("Error:", err.Error())

		// server shutdown
		case <-s.doneCh:
			return
		}
	}
}
