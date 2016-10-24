package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ubuntu/face-detection-demo/appstate"
	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/datastore"
	"github.com/ubuntu/face-detection-demo/detection"
	"github.com/ubuntu/face-detection-demo/messages"
)

var (
	wgwebcam         *sync.WaitGroup
	wgservices       *sync.WaitGroup
	shutdownwebcam   chan interface{}
	shutdownservices chan interface{}
)

//go:generate protoc --go_out=../messages/ --proto_path ../messages/ ../messages/communication.proto
func main() {

	// always starts even if socket exists
	deletesocket := flag.Bool("force", false, "Try force starting even if another daemon is running")
	flag.Parse()

	// check if we are in broken mode and remove database if it's the case
	appstate.CheckIfBroken(appstate.Rootdir)
	if appstate.BrokenMode {
		datastore.WipeDB(appstate.Datadir)
		detection.WipeScreenshots(appstate.Datadir)
	}

	// channels synchronization
	wgwebcam = new(sync.WaitGroup)
	wgservices = new(sync.WaitGroup)
	shutdownwebcam = make(chan interface{})
	shutdownservices = make(chan interface{})

	// handle user generated stop requests
	userstop := make(chan os.Signal)
	signal.Notify(userstop, syscall.SIGINT, syscall.SIGTERM)

	actions := make(chan *messages.Action, 2)

	// prepare settings and data
	datastore.StartDB(appstate.Datadir, shutdownservices, wgservices)

	// starts external communications channel
	comm.StartSocketListener(actions, shutdownservices, *deletesocket, wgservices)
	comm.StartServer(appstate.Rootdir, appstate.Datadir, actions)

	// starts camera if it was already started last time
	if datastore.FaceDetection() {
		detection.StartCameraDetect(appstate.Rootdir, shutdownwebcam, wgwebcam)
	}

mainloop:
	for {
		select {
		case action := <-actions:
			fmt.Println("new action received")
			if processaction(action) {
				break mainloop
			}
		case <-userstop:
			quit()
			break mainloop
		}
	}

	// Ensure webcam and services stopped
	wgwebcam.Wait()
	wgservices.Wait()
}

// process action and return true if we need to quit (exit mainloop)
// TODO: use quit channel (renamed userstop to quit) and send data there. Remove the bool True/False
func processaction(action *messages.Action) bool {
	if action.FaceDetection == messages.Action_FACEDETECTION_ENABLE {
		detection.StartCameraDetect(appstate.Rootdir, shutdownwebcam, wgwebcam)
		fmt.Println("Received camera on")
	} else if action.FaceDetection == messages.Action_FACEDETECTION_DISABLE {
		detection.EndCameraDetect()
		fmt.Println("Received camera off")
	}
	if action.RenderingMode == messages.Action_RENDERINGMODE_FUN {
		datastore.SetRenderingMode(datastore.FUNRENDERING)
		comm.WSserv.SendAllClients(&messages.WSMessage{
			Type:          "renderingmode",
			RenderingMode: datastore.RenderingMode()})
	} else if action.RenderingMode == messages.Action_RENDERINGMODE_NORMAL {
		datastore.SetRenderingMode(datastore.NORMALRENDERING)
		comm.WSserv.SendAllClients(&messages.WSMessage{
			Type:          "renderingmode",
			RenderingMode: datastore.RenderingMode()})
	}
	// camera is offsetted by 1 for the client (0, protobuf default means no change)
	cameranum := int(action.Camera) - 1
	if cameranum > -1 && cameranum != datastore.Camera() {
		datastore.SetCamera(cameranum)
		comm.WSserv.SendAllClients(&messages.WSMessage{
			Type:   "newcameraactivated",
			Camera: cameranum + 1})
		if datastore.FaceDetection() {
			fmt.Println("Change active camera")
			go detection.RestartCamera(appstate.Rootdir, shutdownwebcam, wgwebcam)
		}
	}
	if action.QuitServer {
		quit()
		return true
	}
	return false
}

func quit() {
	fmt.Println("quit server")
	// wait for webcam to shutdown, then ask services to shutdown
	close(shutdownwebcam)
	wgwebcam.Wait()
	close(shutdownservices)
}
