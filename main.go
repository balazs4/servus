package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	logger := log.New(os.Stderr, "[servus] ", log.LstdFlags)
	watcher, watcherr := fsnotify.NewWatcher()
	if watcherr != nil {
		panic(watcherr)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				logger.Printf("ok=%t event=%s\n", ok, event)
			case err, ok := <-watcher.Errors:
				logger.Printf("ok=%t, err=%s\n", ok, err)
			}
		}
	}()

	watcher.Add("./")

	script := `<script> new EventSource(".servus").onmessage = function(ev){ console.log(ev); window.location.reload();}</script>`
	http.HandleFunc("GET /.servus", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)

		event, _ := <-watcher.Events
		fmt.Fprintf(w, "data: servus pid=%d %s\n\n", os.Getpid(), event)
	})

	http.HandleFunc("GET /{file}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		if strings.HasSuffix(name, ".html") == false {
			http.ServeFile(w, r, name)
			return
		}
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			return
		}
		w.Write([]byte(script))
		io.Copy(w, file)
	})

	logger.Printf("pid=%d url=http://localhost:3000\n", os.Getpid())
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
