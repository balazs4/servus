package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	logger := log.New(os.Stderr, "[servus] ", log.LstdFlags)
	port, errport := strconv.Atoi(os.Getenv("PORT"))
	if errport != nil {
		port = 3000
	}

	watcher := createWatcher(logger, os.Args[1:])
	defer watcher.Close()

	for _, p := range watcher.WatchList() {
		logger.Printf("watching %s", p)
	}

	http.HandleFunc("GET /.servus", serverSideEvent(logger, watcher))
	http.HandleFunc("GET /{file}", serveFile(logger))

	logger.Printf("pid=%d url=http://localhost:%d\n", os.Getpid(), port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		panic(err)
	}
}

func serveFile(logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	const script = `<script> new EventSource(".servus").onmessage = function(ev){ console.log(ev); window.location.reload();}</script>`

	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			logger.Printf("name=%s, err=%s\n", name, err)
			http.NotFound(w, r)
			return
		}
		logger.Printf("[%d] %s %s\n", http.StatusOK, r.Method, name)
		if strings.HasSuffix(name, ".html") == true {
			w.Header().Add("X-Servus-Patch", fmt.Sprint(true))
			io.Copy(w, file)
			w.Write([]byte(script))
			return
		}
		http.ServeContent(w, r, name, time.Now(), file)
	}
}

func serverSideEvent(logger *log.Logger, watcher *fsnotify.Watcher) func(w http.ResponseWriter, r *http.Request) {
	eventbroker := make(map[int64]chan fsnotify.Event)

	go func() {
		for {
			//event producer
			event, _ := <-watcher.Events
			if event.Has(fsnotify.Chmod) == true {
				//ignore
				continue
			}
			logger.Printf("event=%s, eventbroker=%d\n", event, len(eventbroker))
			for _, consumer := range eventbroker {
				consumer <- event
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		id := time.Now().UnixNano()
		eventbroker[id] = make(chan fsnotify.Event)
		select {
		case event := <-eventbroker[id]:
			delete(eventbroker, id)
			size, err := fmt.Fprintf(w, "data: servus pid=%d %s\n\n", os.Getpid(), event)
			if err != nil {
				logger.Printf("size=%d, err=%s", size, err)
			}
		case <-r.Context().Done():
			delete(eventbroker, id)
		}
	}
}

func createWatcher(logger *log.Logger, path []string) *fsnotify.Watcher {
	watcher, watcherr := fsnotify.NewWatcher()
	if watcherr != nil {
		panic(watcherr)
	}

	go func() {
		for {
			err, ok := <-watcher.Errors
			logger.Printf("ok=%t, err=%s\n", ok, err)
		}
	}()
	watcher.Add(".")
	for _, p := range path {
		watcher.Add(p)
	}
	return watcher
}
