package datastore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/ubuntu/face-detection-demo/appstate"

	"gopkg.in/yaml.v2"
)

// RenderMode corresponds to various rendering mode options (fun, normal…)
type RenderMode int

const (
	// NORMALRENDERING draws circles around heads
	NORMALRENDERING = iota
	// FUNRENDERING draws logos instead on top of heads
	FUNRENDERING
)

type settingsElem struct {
	FaceDetectionSetting bool
	RenderingModeSetting RenderMode
	Camera               int
}

var (
	settingsdir   string
	settings      = settingsElem{false, NORMALRENDERING, 0}
	filesavemutex = &sync.Mutex{}
)

// initialize directory where data are
func init() {
	settingsdir = path.Join(appstate.Datadir, "settings")

	// load settings
	dat, err := ioutil.ReadFile(settingsdir)
	if err != nil {
		// no file available: can be first install with defaults
		return
	}
	if err = yaml.Unmarshal(dat, &settings); err != nil {
		fmt.Println("Couldn't unserialized settings from", settingsdir, ". Reverting to defaults.")
	}
}

// FaceDetection tells if detection is on or off
func FaceDetection() bool {
	return settings.FaceDetectionSetting
}

// RenderingMode return current rendering mode
func RenderingMode() RenderMode {
	return settings.RenderingModeSetting
}

// Camera return current camera number set
func Camera() int {
	return settings.Camera
}

// SetFaceDetection save new detection state
func SetFaceDetection(faceDetection bool) {
	if faceDetection == settings.FaceDetectionSetting {
		return
	}
	settings.FaceDetectionSetting = faceDetection

	go saveToFile()
}

// SetRenderingMode save new rendering mode
func SetRenderingMode(renderingMode RenderMode) {
	if renderingMode == settings.RenderingModeSetting {
		return
	}
	settings.RenderingModeSetting = renderingMode

	go saveToFile()
}

// SetCamera save active camera number
func SetCamera(cameranum int) {
	if cameranum == settings.Camera {
		return
	}
	settings.Camera = cameranum

	go saveToFile()
}

func saveToFile() {
	data, err := yaml.Marshal(&settings)
	if err != nil {
		fmt.Println("Can't convert", settings, "to yaml:", err)
		return
	}

	filesavemutex.Lock()
	defer filesavemutex.Unlock()

	tempfile := settingsdir + ".new"
	if err = ioutil.WriteFile(tempfile, data, 0644); err != nil {
		fmt.Println("Couldn't save settings to", tempfile)
		return
	}
	defer os.Remove(tempfile)

	if err = os.Rename(tempfile, settingsdir); err != nil {
		fmt.Println("Couldn't save temp settings to", settingsdir)
	}
}
