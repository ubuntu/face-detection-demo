package appstate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

var (
	// BrokenMode signal if the application is broken
	BrokenMode bool
	// AvailableCameras list index of detected cameras
	AvailableCameras []int

	// Rootdir executable code to reach assets
	Rootdir string
	// Datadir access to write storage path
	Datadir string
)

const brokenversion = "2.0alpha1"

type versionYaml struct {
	Version string `yaml:"version"`
}

func init() {
	// Set main set of directories
	var err error
	Rootdir = os.Getenv("SNAP")
	if Rootdir == "" {
		if Rootdir, err = filepath.Abs(path.Join(filepath.Dir(os.Args[0]), "..")); err != nil {
			log.Fatal(err)
		}
	}
	Datadir = os.Getenv("SNAP_DATA")
	if Datadir == "" {
		Datadir = Rootdir
	}
}

// CheckIfBroken checks and set if app is in broken state (when matching brokenversion)
func CheckIfBroken(rootdir string) {
	yamlc := versionYaml{}
	yamlfile := path.Join(rootdir, "meta", "snap.yaml")

	// load settings
	dat, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		// no file available: can be run from trunk
		fmt.Println("Couldn't open", yamlfile, ". Probably running from master, set the app as functionning.")
		return
	}
	if err = yaml.Unmarshal(dat, &yamlc); err != nil {
		fmt.Println("Couldn't unserialized snap yaml from", yamlc, ". Setting the app as functionning.")
		return
	}
	if yamlc.Version == brokenversion {
		fmt.Println("Broken version running (", brokenversion, "). Set the app property as being broken.")
		BrokenMode = true
	}
}
