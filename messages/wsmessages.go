package messages

import "github.com/ubuntu/face-detection-demo/datastore"

// WSMessage to be sent to clients
type WSMessage struct {
	Type                    string               `json:"type"`
	AllStats                []datastore.Stat     `json:"allstats"`
	NewStat                 *datastore.Stat      `json:"newstat"`
	RefreshScreenshot       bool                 `json:"refreshscreenshot"`
	RefreshDetectScreenshot bool                 `json:"refreshdetectscreenshot"`
	FaceDetection           bool                 `json:"facedetection"`
	RenderingMode           datastore.RenderMode `json:"renderingmode"`
	Camera                  int                  `json:"camera"`
	AvailableCameras        []int                `json:"availablecameras"`
	Broken                  bool                 `json:"broken"`
}
