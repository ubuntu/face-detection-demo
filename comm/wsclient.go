package comm

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ubuntu/face-detection-demo/messages"

	"golang.org/x/net/websocket"
)

const channelBufSize = 100

var maxID int

// Client represents a ws connection
type Client struct {
	id     int
	ws     *websocket.Conn
	server *WSServer
	ch     chan *messages.WSMessage
	doneCh chan interface{}
}

// NewClient creates a new ws client
func NewClient(ws *websocket.Conn, server *WSServer) (*Client, error) {

	if ws == nil {
		return nil, errors.New("ws cannot be nil")
	}

	if server == nil {
		return nil, errors.New("server cannot be nil")
	}

	maxID++
	ch := make(chan *messages.WSMessage, channelBufSize)
	doneCh := make(chan interface{})

	return &Client{maxID, ws, server, ch, doneCh}, nil
}

// Send a message to a client
func (c *Client) Send(msg *messages.WSMessage) {
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected", c.id)
		c.server.Err(err)
	}
}

// Done close down client connection
func (c *Client) Done() {
	close(c.doneCh)
}

// Listen Write and Read request via channel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	for {
		select {

		// send message to the client
		case msg := <-c.ch:
			log.Println("Send:", msg)
			websocket.JSON.Send(c.ws, msg)

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			return
		}
	}
}

// Listen read request via channel
func (c *Client) listenRead() {
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			return

		// read data from websocket connection
		default:
			var action messages.Action
			err := websocket.JSON.Receive(c.ws, &action)
			if err == io.EOF {
				close(c.doneCh)
			} else if err != nil {
				c.server.Err(err)
			}
			c.server.NewAction(&action)
		}
	}
}
