package messages

// WSMessage to be sent to clients
type WSMessage struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

func (m *WSMessage) String() string {
	return m.Author + " says " + m.Body
}
