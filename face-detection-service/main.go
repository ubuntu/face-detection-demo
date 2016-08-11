package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/datastore"
	"github.com/ubuntu/face-detection-demo/detection"
	"github.com/ubuntu/face-detection-demo/messages"
)

var (
	wg       *sync.WaitGroup
	shutdown chan interface{}
	rootdir  string
)

//go:generate protoc --go_out=../messages/ --proto_path ../messages/ ../messages/communication.proto
func main() {
	var err error

	// Set main set of directories
	if rootdir, err = filepath.Abs(path.Join(filepath.Dir(os.Args[0]), "..")); err != nil {
		log.Fatal(err)
	}
	datadir := os.Getenv("SNAP_DATA")
	if datadir == "" {
		datadir = rootdir
	}

	datastore.LoadSettings(datadir)

	// channels synchronization
	wg = new(sync.WaitGroup)
	shutdown = make(chan interface{})

	// handle user generated stop requests
	userstop := make(chan os.Signal)
	signal.Notify(userstop, syscall.SIGINT, syscall.SIGTERM)

	actions := make(chan *messages.Action, 2)

	// starts external communications channel
	comm.StartSocketListener(actions, shutdown, wg)

	// starts camera if it was already started last time
	if datastore.FaceDetection() {
		detection.StartCameraDetect(rootdir, shutdown, wg)
	}

mainloop:
	for {
		fmt.Println("main loop")
		select {
		case action := <-actions:
			fmt.Println("new action received")
			if processaction(action) {
				break mainloop
			}
		case <-userstop:
			quit()
			break mainloop
		case <-time.After(5 * time.Second):
			fmt.Println("timeout")
			var foo *messages.Action
			if datastore.FaceDetection() {
				foo = &messages.Action{
					FaceDetection: messages.Action_FACEDETECTION_DISABLE,
					RenderingMode: messages.Action_RENDERINGMODE_UNCHANGED,
					QuitServer:    false,
				}
				fmt.Println("Switch camera off")
			} else {
				foo = &messages.Action{
					FaceDetection: messages.Action_FACEDETECTION_ENABLE,
					RenderingMode: messages.Action_RENDERINGMODE_UNCHANGED,
					QuitServer:    false,
				}
				fmt.Println("Switch camera on")
			}
			actions <- foo
		}
	}

	wg.Wait()
}

// process action and return true if we need to quit (exit mainloop)
func processaction(action *messages.Action) bool {
	if action.FaceDetection == messages.Action_FACEDETECTION_ENABLE {
		detection.StartCameraDetect(rootdir, shutdown, wg)
		fmt.Println("Received camera on")
	} else if action.FaceDetection == messages.Action_FACEDETECTION_DISABLE {
		detection.EndCameraDetect()
		fmt.Println("Received camera off")
	}
	if action.RenderingMode == messages.Action_RENDERINGMODE_FUN {
		datastore.SetRenderingMode(datastore.FUNRENDERING)
	} else if action.RenderingMode == messages.Action_RENDERINGMODE_NORMAL {
		datastore.SetRenderingMode(datastore.NORMALRENDERING)
	}
	if action.QuitServer {
		quit()
		return true
	}
	return false
}

func quit() {
	fmt.Println("quit server")
	// signal all main goroutines to exits
	close(shutdown)
}
