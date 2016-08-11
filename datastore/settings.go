package datastore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"
)

// Rendermode corresponds to various rendering mode options (fun, normal…)
type Rendermode int

const (
	// NORMALMODE draws circles around heads
	NORMALMODE = iota
	// FUNMODE draws logos instead on top of heads
	FUNMODE
)

type settings struct {
	Detection bool
	Mode      Rendermode
}

var (
	datadir       string
	settingsdir   string
	values        = settings{false, NORMALMODE}
	valuemutex    = &sync.Mutex{}
	filesavemutex = &sync.Mutex{}
)

// LoadSettings initialize directory where data are
func LoadSettings(dir string) {
	datadir = dir
	settingsdir = path.Join(datadir, "settings")

	// load settings
	dat, err := ioutil.ReadFile(settingsdir)
	if err != nil {
		// no file available: can be first install with defaults
		return
	}
	if err = yaml.Unmarshal(dat, &values); err != nil {
		fmt.Println("Couldn't unserialized settings from", settingsdir, ". Reverting to defaults.")
	}
}

// GetDetection tells if detection is on or off
func GetDetection() bool {
	valuemutex.Lock()
	val := values.Detection
	valuemutex.Unlock()
	return val
}

// GetRenderingMode return current rendering mode
func GetRenderingMode() Rendermode {
	valuemutex.Lock()
	val := values.Mode
	valuemutex.Unlock()
	return val
}

// SetDetection save new detection state
func SetDetection(newDetectionValue bool) {
	valuemutex.Lock()
	defer valuemutex.Unlock()
	if newDetectionValue == values.Detection {
		return
	}
	values.Detection = newDetectionValue

	go saveToFile()
}

// SetRenderMode save new rendering mode
func SetRenderMode(newRenderModeValue Rendermode) {
	valuemutex.Lock()
	defer valuemutex.Unlock()
	if newRenderModeValue == values.Mode {
		return
	}
	values.Mode = newRenderModeValue

	go saveToFile()
}

func saveToFile() {
	data, err := yaml.Marshal(&values)
	if err != nil {
		fmt.Println("Can't convert", values, "to yaml:", err)
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
