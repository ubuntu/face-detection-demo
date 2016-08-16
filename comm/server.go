package comm

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/ubuntu/face-detection-demo/messages"
)

var (
	datadir string
	rootdir string
)

// StartServer starts in a goroutine both webserver and websocket handlers
func StartServer(rootd string, datad string, actions chan<- *messages.Action) {
	datadir = datad
	rootdir = rootd
	WSserv = NewWSServer("/api", actions)
	fmt.Println("Server created", WSserv)
	go func() {
		go WSserv.Listen()
		http.HandleFunc("/data/", serveFileData)
		http.Handle("/", http.FileServer(http.Dir(path.Join(rootdir, "www"))))
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Couldn't start webserver:", err)
		}
	}()
}

func serveFileData(w http.ResponseWriter, r *http.Request) {
	fn := r.URL.Path[6:]
	filepath := path.Join(datadir, fn)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if len(filepath) > 4 && filepath[len(filepath)-4:] == ".png" {
			// return our own image
			http.ServeFile(w, r, path.Join(rootdir, "images", "fallbackscreenshot.png"))
			return
		}
		// return 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("File not found"))
		return
	}
	http.ServeFile(w, r, filepath)
}
