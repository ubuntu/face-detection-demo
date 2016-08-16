package comm

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/ubuntu/face-detection-demo/messages"
)

// StartServer starts in a goroutine both webserver and websocket handlers
func StartServer(rootdir string, actions chan<- *messages.Action) {
	WSserv = NewWSServer("/api", actions)
	fmt.Println("Server created", WSserv)
	go func() {
		go WSserv.Listen()
		http.Handle("/", http.FileServer(http.Dir(path.Join(rootdir, "www"))))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("Couldn't start webserver:", err)
		}
	}()
}
