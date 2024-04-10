package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	logger := log.New(os.Stderr, fmt.Sprintf("[servus] "), log.LstdFlags)
	logger.Printf("version=%s", getVersion())

	port, errport := strconv.Atoi(os.Getenv("PORT"))
	if errport != nil {
		port = 3000
	}

	watcher := createWatcher(logger, os.Args[1:])
	defer watcher.Close()

	http.HandleFunc("GET /.servus", serverSideEvent(logger, watcher))
	http.HandleFunc("GET /{file}", serveFile(logger))
	http.Handle("GET /", http.RedirectHandler("/index.html", http.StatusSeeOther))

	logger.Printf("pid=%d url=http://localhost:%d\n", os.Getpid(), port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		panic(err)
	}
}

var Version string

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok == false {
		return "???"
	}

	if Version != "" {
		return Version
	}

	return info.Main.Version
}

func serveFile(logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	const script = `<script> new EventSource(".servus").onmessage = function(ev){ console.log(ev); window.location.reload();}</script>`

	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			logger.Printf("[%d] %s %s, err=%s\n", http.StatusNotFound, r.Method, name, err)
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

type FsEvent struct {
	sync.Mutex
	Consumers map[int64]*chan fsnotify.Event
}

func (f *FsEvent) Subscribe(id int64, consumer *chan fsnotify.Event) {
	f.Lock()
	defer f.Unlock()
	f.Consumers[id] = consumer
}

func (f *FsEvent) Unsubscribe(id int64) {
	f.Lock()
	defer f.Unlock()
	delete(f.Consumers, id)
}

func serverSideEvent(logger *log.Logger, watcher *fsnotify.Watcher) func(w http.ResponseWriter, r *http.Request) {
	eventbroadcast := FsEvent{
		Consumers: make(map[int64]*chan fsnotify.Event),
	}

	go func() {
		for {
			event, _ := <-watcher.Events
			eventbroadcast.Lock()
			count := len(eventbroadcast.Consumers)
			if count == 0 {
				eventbroadcast.Unlock()
				continue
			}
			logger.Printf("event=%s, consumers=%d\n", event, count)
			for id, consumer := range eventbroadcast.Consumers {
				delete(eventbroadcast.Consumers, id)
				*consumer <- event
			}
			eventbroadcast.Unlock()
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		id := time.Now().UnixNano()
		channel := make(chan fsnotify.Event)
		eventbroadcast.Subscribe(id, &channel)

		select {
		case event := <-channel:
			size, err := fmt.Fprintf(w, "data: servus pid=%d %s\n\n", os.Getpid(), event)
			if err != nil {
				logger.Printf("size=%d, err=%s", size, err)
			}
		case <-r.Context().Done():
			eventbroadcast.Unsubscribe(id)
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

	for _, p := range watcher.WatchList() {
		logger.Printf("watch=%s", p)
	}

	return watcher
}
