package detection

import (
	"fmt"
	"path"
	"time"

	"github.com/lazywei/go-opencv/opencv"
)

// StartCameraDetect creates a go routine handling web cam recording and image generation
func StartCameraDetect(workdir string, stop <-chan bool) {
	cap := opencv.NewCameraCapture(0)

	if cap == nil {
		panic("cannot open camera")
	}
	defer cap.Release()

	detectFace(cap, workdir, stop)

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
			break
		case <-time.After(2 * time.Second):
			break
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
