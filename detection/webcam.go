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
	// we can stop in two ways, hence the use of this channel
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
	comm.WSserv.SendAllClients(&messages.WSMessage{
		FaceDetection: datastore.FaceDetection(),
		RenderingMode: datastore.RenderingMode()})
	if !cameraOn {
		fmt.Println("Turning off detection command received but not started")
		return
	}
	close(stop)
}

func detectFace(cap *opencv.Capture, rootdir string, stop <-chan interface{}) {
	nextFrameSec := time.Now()
	cascade := opencv.LoadHaarClassifierCascade(path.Join(rootdir, "frontfacedetection.xml"))
	for {

		select {
		case <-stop:
			fmt.Println("Stop processing webcam events")
			cascade.Release()
			return
		default:
		}

		if cap.GrabFrame() {

			// we dropped all grab framesand only take one every X seconds (no support in opencv go binding for CV_CAP_PROP_BUFFERSIZE)
			// if we didn't grab them one after another, we'll have past frames when proceeding
			if time.Now().Before(nextFrameSec) {
				continue
			}

			// treat framee
			img := cap.RetrieveFrame(1)
			if img != nil {
				faces := cascade.DetectObjects(img)
				drawAndSaveFaces(img, faces)
			}

		}
		cascade.Release()
		cascade = opencv.LoadHaarClassifierCascade(path.Join(rootdir, "frontfacedetection.xml"))
		nextFrameSec = time.Now().Add(time.Duration(5 * time.Second))
	}

}

func drawAndSaveFaces(img *opencv.IplImage, faces []*opencv.Rect) {
	// save raw image before modifications
	detectedFace := false

	dest := RenderedImage{RenderingMode: datastore.RenderingMode()}

	for num, face := range faces {
		fmt.Println("face detected")
		detectedFace = true
		dest.DrawFace(face, num, img)
	}

	// store and save stat
	s := &datastore.Stat{TimeStamp: time.Now(), NumPersons: len(faces)}
	comm.WSserv.SendAllClients(&messages.WSMessage{NewStat: s,
		FaceDetection: datastore.FaceDetection(),
		RenderingMode: datastore.RenderingMode()})
	datastore.DB.Add(*s)

	opencv.SaveImage("/tmp/orig.png", img, 0)

	// save image with face detection if any
	if detectedFace {
		dest.Save()
	}
}
