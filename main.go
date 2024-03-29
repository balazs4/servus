package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	logger := log.New(os.Stderr, "[servus] ", log.LstdFlags)
	port, errport := strconv.Atoi(os.Getenv("PORT"))
	if errport != nil {
		port = 3000
	}

	working_dir := "./"
	if len(os.Args) == 2 {
		working_dir = os.Args[1]
	}

	watcher := Watcher(logger, working_dir)
	defer watcher.Close()

	logger.Printf("watching %s", watcher.WatchList()[0])

	http.HandleFunc("GET /.servus", ServerSideEvent(logger, watcher))
	http.HandleFunc("GET /{file}", ServeFile(logger, working_dir))

	logger.Printf("pid=%d url=http://localhost:%d\n", os.Getpid(), port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		panic(err)
	}
}

func ServeFile(logger *log.Logger, working_dir string) func(w http.ResponseWriter, r *http.Request) {
	const script = `<script> new EventSource(".servus").onmessage = function(ev){ console.log(ev); window.location.reload();}</script>`

	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		if strings.HasSuffix(name, ".html") == false {
			http.ServeFile(w, r, name)
			return
		}
		file, err := os.Open(path.Join(working_dir, name))
		defer file.Close()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(script))
		io.Copy(w, file)
	}
}

func ServerSideEvent(logger *log.Logger, watcher *fsnotify.Watcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)

		event, _ := <-watcher.Events
		size, err := fmt.Fprintf(w, "data: servus pid=%d %s\n\n", os.Getpid(), event)
		if err != nil {
			logger.Printf("size=%d, err=%s", size, err)
		}
	}
}

func Watcher(logger *log.Logger, working_dir string) *fsnotify.Watcher {
	watcher, watcherr := fsnotify.NewWatcher()
	if watcherr != nil {
		panic(watcherr)
	}

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
	watcher.Add(working_dir)
	return watcher
}
