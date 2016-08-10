package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/detection"
	"github.com/ubuntu/face-detection-demo/messages"
)

var shutdown chan interface{}

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
	shutdown = make(chan interface{})

	// handle user generated stop requests
	userstop := make(chan os.Signal)
	signal.Notify(userstop, syscall.SIGINT, syscall.SIGTERM)

	actions := make(chan *messages.Action, 2)
	comm.StartSocketListener(actions, shutdown, wg)

mainloop:
	for {
		fmt.Println("main loop")
		select {
		case action := <-actions:
			fmt.Println("new action received")
			if action.CameraState == messages.Action_CAMERA_ENABLE {
				detection.StartCameraDetect(workdir, shutdown, wg)
				fmt.Println("Received camera on")
			} else if action.CameraState == messages.Action_CAMERA_DISABLE {
				detection.EndCameraDetect()
				fmt.Println("Received camera off")
			}
			if action.QuitServer {
				quit()
				break mainloop
			}
		case <-userstop:
			quit()
			break mainloop
		case <-time.After(225 * time.Second):
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

func quit() {
	fmt.Println("quit server")
	// signal all main goroutines to exits
	close(shutdown)
}
