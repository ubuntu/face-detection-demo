package detection

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/datastore"
	"github.com/ubuntu/face-detection-demo/messages"
)

var (
	stop chan interface{}

	cameraOn bool
)

// StartCameraDetect creates a go routine handling web cam recording and image generation
func StartCameraDetect(rootdir string, shutdown <-chan interface{}, wg *sync.WaitGroup) {
	if cameraOn {
		fmt.Println("Detection command received but already started")
		return
	}
	stop = make(chan interface{})

	// send the main quit channel to stop if we got a shutdown request
	go func() {
		select {
		case <-shutdown:
			close(stop)
		case <-stop:
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { cameraOn = false }()
		defer fmt.Println("Stop camera")

		cap := opencv.NewCameraCapture(0)
		if cap == nil {
			panic("cannot open camera")
		}
		defer cap.Release()
		cameraOn = true
		datastore.SetFaceDetection(true)

		detectFace(cap, rootdir, stop)
	}()

}

// EndCameraDetect stop the associated goroutine turning on camera
func EndCameraDetect() {
	datastore.SetFaceDetection(false)
	if !cameraOn {
		fmt.Println("Turning off detection command received but not started")
		return
	}
	close(stop)
}

func detectFace(cap *opencv.Capture, rootdir string, stop <-chan interface{}) {
	for {
		cascade := opencv.LoadHaarClassifierCascade(path.Join(rootdir, "frontfacedetection.xml"))
		if cap.GrabFrame() {
			img := cap.RetrieveFrame(1)
			if img != nil {
				faces := cascade.DetectObjects(img)
				drawAndSaveFaces(img, faces)
			}
		}

		// check if we need to exit. Add some timeouts before grabbing the next camera image
		select {
		// we received the signal of cancelation in this channel
		case <-stop:
			fmt.Println("Stop processing webcam events")
			return
		case <-time.After(2 * time.Second):
			continue
		}
	}
}

func drawAndSaveFaces(img *opencv.IplImage, faces []*opencv.Rect) {
	// save raw image before modifications
	modifiedImg := img.Clone()
	detectedFace := false
	for num, face := range faces {
		fmt.Println("face detected")
		detectedFace = true
		drawFace(modifiedImg, face, num)
	}

	// store and save stat
	s := datastore.Stat{TimeStamp: time.Now(), NumPersons: len(faces)}
	comm.WSserv.SendAllClients(&messages.WSMessage{NewStat: s})
	datastore.DB.Add(s)

	opencv.SaveImage("/tmp/orig.png", img, 0)
	if detectedFace {
		opencv.SaveImage("/tmp/detect.png", modifiedImg, 0)
	}
}
