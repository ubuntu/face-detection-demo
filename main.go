package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ubuntu/face-detection-demo/detection"
)

func main() {
	workdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	stopDetection := make(chan bool)
	defer close(stopDetection)

	detection.StartCameraDetect(workdir, stopDetection)

}
