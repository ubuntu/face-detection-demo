package messages

import "github.com/ubuntu/face-detection-demo/datastore"

// WSMessage to be sent to clients
type WSMessage struct {
	AllStats                []datastore.Stat `json:"allstats"`
	NewStat                 datastore.Stat   `json:"newstats"`
	RefreshScreenshot       bool             `json:"refreshscreenshot"`
	RefreshDetectScreenshot bool             `json:"refreshdetectscreenshot"`
}

func (m *WSMessage) String() string {
	return m.Author + " says " + m.Body
}
