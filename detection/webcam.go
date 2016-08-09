package detection

import (
	"fmt"
	"path"
	"time"

	"github.com/lazywei/go-opencv/opencv"
)

var (
	stop chan bool

	// DetectionOn reports if face detection is in progress
	DetectionOn bool
)

// StartCameraDetect creates a go routine handling web cam recording and image generation
func StartCameraDetect(workdir string) {
	if DetectionOn {
		fmt.Println("Detection command received but already started")
		return
	}
	stop = make(chan bool)

	go func() {
		defer func() { DetectionOn = false }()

		cap := opencv.NewCameraCapture(0)
		if cap == nil {
			panic("cannot open camera")
		}
		defer cap.Release()
		DetectionOn = true

		detectFace(cap, workdir, stop)
	}()

}

// EndCameraDetect stop the associated goroutine turning on camera
func EndCameraDetect() {
	if !DetectionOn {
		fmt.Println("Turning off detection command received but not started")
		return
	}
	stop <- true
	close(stop)
}

func detectFace(cap *opencv.Capture, workdir string, stop <-chan bool) {
	for {
		cascade := opencv.LoadHaarClassifierCascade(path.Join(workdir, "..", "detection", "haarcascade_frontalface_alt.xml"))
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
	opencv.SaveImage("/tmp/orig.png", img, 0)
	if detectedFace {
		opencv.SaveImage("/tmp/detect.png", modifiedImg, 0)
	}
}
