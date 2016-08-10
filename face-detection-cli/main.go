package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ubuntu/face-detection-demo/comm"
	"github.com/ubuntu/face-detection-demo/messages"
)

func main() {

	enableCam := flag.Bool("enable-camera", false, "Enable the camera detection service")
	disableCam := flag.Bool("disable-camera", false, "Disable the camera detection service")

	funMode := flag.Bool("fun-mode", false, "Show some distro logos instead of the head")
	normalMode := flag.Bool("normal-mode", false, "Show some circle around detected heads")

	restart := flag.Bool("restart", false, "Restart the web server")

	flag.Parse()

	fmt.Println("camera:", *enableCam)
	fmt.Println("disable camera:", *disableCam)
	fmt.Println("mode fun:", *funMode)
	fmt.Println("normal mode:", *normalMode)
	fmt.Println("restart:", *restart)
	if len(flag.Args()) > 0 {
		errorOut("Invalid argument set")
	}

	if *enableCam && *disableCam {
		errorOut("enabling and disabling camera can't bet se at the same time")
	}

	if *funMode && *normalMode {
		errorOut("fun and normal drawing mode can't be set at the same time")
	}

	msg := createMessage(*enableCam, *disableCam, *funMode, *normalMode, *restart)

	if err := comm.SendToSocket(msg); err != nil {
		os.Exit(1)
	}
}

func errorOut(message string) {
	fmt.Println("Error:", message)
	flag.PrintDefaults()
	os.Exit(1)
}

func createMessage(enableCam bool, disableCam bool, funMode bool, normalMode bool, restart bool) *messages.Action {
	var cameraState messages.Action_CamState
	var mode messages.Action_Mode

	if enableCam {
		cameraState = messages.Action_CAMERA_ENABLE
	} else if disableCam {
		cameraState = messages.Action_CAMERA_DISABLE
	} else {
		cameraState = messages.Action_CAMERA_UNCHANGED
	}

	if funMode {
		mode = messages.Action_MODE_FUN
	} else if normalMode {
		mode = messages.Action_MODE_TRADITIONAL
	} else {
		mode = messages.Action_MODE_UNCHANGED
	}

	return &messages.Action{
		CameraState:   cameraState,
		DrawMode:      mode,
		RestartServer: restart,
	}
}
