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

	cameraOn   bool
	currentCam = -1
)

func init() {
	DetectCameras()
}

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

		cap := openCamera(datastore.Camera())
		if cap == nil {
			panic(fmt.Sprintf("Cannot open camera %d", currentCam))
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

// fallback to camera 0 if can't open requested camera number
func openCamera(cameraNum int) *opencv.Capture {
	currentCam = cameraNum
	cap := opencv.NewCameraCapture(currentCam)
	if cap == nil && currentCam != 0 {
		fmt.Printf("Can't open camera %d. Trying fallback to camera 0\n", currentCam)
		currentCam = 0
		cap = opencv.NewCameraCapture(currentCam)
		if cap != nil {
			datastore.SetCamera(currentCam)
			comm.WSserv.SendAllClients(&messages.WSMessage{
				Type: "newcameraactivated",
				// camera is offsetted by 1 for the client
				Camera: currentCam + 1})
		}
	}
	return cap
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

// DetectCameras detects and files the index of available cameras. Take into account current camera if already on
func DetectCameras() {
	appstate.AvailableCameras = make([]int, 0)

	for i := 0; i < 10; i++ {
		cap := opencv.NewCameraCapture(i)
		if cap != nil || (cameraOn && i == currentCam) {
			if cap != nil {
				cap.Release()
			}
			// camera is offsetted by 1 for the client
			appstate.AvailableCameras = append(appstate.AvailableCameras, i+1)
		}
	}

	if len(appstate.AvailableCameras) == 0 {
		panic("No camera detected")
	}
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

	if err := saveatomic(datadir, screenshotname, (*opencvImg)(img)); err != nil {
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
