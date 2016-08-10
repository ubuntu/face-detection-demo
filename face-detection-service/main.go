package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/detection"
	"github.com/ubuntu/face-detection-demo/messages"
)

func main() {
	workdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	/*
	   16:21:21   jdstrand | didrocks, zyga: /run/shm/snap.$SNAP_NAME.** is in the default template and can be used for cross-app           │ bloodearnest
	                       | communications, but not cross-snap communications
	*/

	wg := new(sync.WaitGroup)

	actions := make(chan *messages.Action, 2)
	quit := make(chan interface{})
	comm.StartSocketListener(actions, quit, wg)

mainloop:
	for {
		fmt.Println("main loop")
		select {
		case action := <-actions:
			fmt.Println("new action received")
			if action.CameraState == messages.Action_CAMERA_ENABLE {
				detection.StartCameraDetect(workdir)
				fmt.Println("Received camera on")
			} else if action.CameraState == messages.Action_CAMERA_DISABLE {
				detection.EndCameraDetect()
				fmt.Println("Received camera off")
			}
			if action.QuitServer {
				fmt.Println("quit server")
				// signal all main goroutines to exits
				close(quit)
				break mainloop
			}
		case <-time.After(5 * time.Second):
			fmt.Println("timeout")
			var foo *messages.Action
			if detection.DetectionOn {
				foo = &messages.Action{
					DrawMode:    messages.Action_MODE_UNCHANGED,
					CameraState: messages.Action_CAMERA_DISABLE,
					QuitServer:  false,
				}
				fmt.Println("Switch camera off")
			} else {
				foo = &messages.Action{
					DrawMode:    messages.Action_MODE_UNCHANGED,
					CameraState: messages.Action_CAMERA_ENABLE,
					QuitServer:  false,
				}
				fmt.Println("Switch camera on")
			}
			actions <- foo
		}
	}

	wg.Wait()

}
