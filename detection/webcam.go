package detection

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/ubuntu/face-detection-demo/appstate"
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

		// TODO: should check this one exists, if not, fallback to 0 (and let if crash if none)
		cap := opencv.NewCameraCapture(datastore.Camera())
		if cap == nil {
			panic("cannot open camera")
		}
		defer cap.Release()
		cameraOn = true
		datastore.SetFaceDetection(true)
		comm.WSserv.SendAllClients(&messages.WSMessage{
			Type:          "facedetection",
			FaceDetection: datastore.FaceDetection(),
		})

		detectFace(cap, rootdir, stop)
	}()

}

// EndCameraDetect stop the associated goroutine turning on camera
func EndCameraDetect() {
	datastore.SetFaceDetection(false)
	comm.WSserv.SendAllClients(&messages.WSMessage{
		Type:          "facedetection",
		FaceDetection: datastore.FaceDetection(),
	})
	if !cameraOn {
		fmt.Println("Turning off detection command received but not started")
		return
	}
	close(stop)
}

// RestartCamera stops and restarts the camera in a sync fashion (wait for the camera to stop before sending the Start signal)
func RestartCamera(rootdir string, shutdown <-chan interface{}, wg *sync.WaitGroup) {
	if !cameraOn {
		// check again after a second in case of a start + restart is issued.
		// FIXME: this should be way better handled
		time.Sleep(time.Second * 1)
		if !cameraOn {
			StartCameraDetect(rootdir, shutdown, wg)
			return
		}
	}
	close(stop)
	for {
		if !cameraOn {
			break
		}
		time.Sleep(1 * time.Second)
	}
	StartCameraDetect(rootdir, shutdown, wg)
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
	np := len(faces)
	if appstate.BrokenMode {
		np = -np
	}
	s := &datastore.Stat{TimeStamp: time.Now(), NumPersons: np}
	datastore.DB.Add(*s)

	// save raw image
	savefn := func(filepath string) error {
		opencv.SaveImage(filepath, img, 0)
		return nil
	}
	if err := saveatomic(datadir, screenshotname, savefn); err != nil {
		fmt.Println(err)
	}

	// save image with face detection if any
	if detectedFace {
		dest.Save()
	}

	// send messages to clients
	comm.WSserv.SendAllClients(&messages.WSMessage{
		Type:                    "newstat",
		NewStat:                 s,
		RefreshScreenshot:       true,
		RefreshDetectScreenshot: detectedFace})
}
