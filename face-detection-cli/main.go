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

	funMode := flag.Bool("fun", false, "Show some distro logos instead of the head")
	normalMode := flag.Bool("normal", false, "Show some circle around detected heads")

	camera := flag.Int("camera", 0, "Change active camera number")

	quit := flag.Bool("quit", false, "Force the web server to shutdown")

	flag.Parse()
	if len(flag.Args()) > 0 {
		errorOut("Invalid argument set")
	}

	if *enableCam && *disableCam {
		errorOut("enabling and disabling camera can't bet se at the same time")
	}

	if *funMode && *normalMode {
		errorOut("fun and normal rendering mode can't be set at the same time")
	}

	msg := createMessage(*enableCam, *disableCam, *funMode, *normalMode, *camera, *quit)

	if err := comm.SendToSocket(msg); err != nil {
		os.Exit(1)
	}
}

func errorOut(message string) {
	fmt.Println("Error:", message)
	flag.PrintDefaults()
	os.Exit(1)
}

func createMessage(enablefd bool, disablefd bool, fun bool, normal bool, camera int, quit bool) *messages.Action {
	var cameraState messages.Action_FaceDetectionState
	var renderingMode messages.Action_RenderingMode

	if enablefd {
		cameraState = messages.Action_FACEDETECTION_ENABLE
	} else if disablefd {
		cameraState = messages.Action_FACEDETECTION_DISABLE
	} else {
		cameraState = messages.Action_FACEDETECTION_UNCHANGED
	}

	if fun {
		renderingMode = messages.Action_RENDERINGMODE_FUN
	} else if normal {
		renderingMode = messages.Action_RENDERINGMODE_NORMAL
	} else {
		renderingMode = messages.Action_RENDERINGMODE_UNCHANGED
	}

	return &messages.Action{
		FaceDetection: cameraState,
		RenderingMode: renderingMode,
		Camera:        int32(camera),
		QuitServer:    quit,
	}
}
