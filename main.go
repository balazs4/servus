package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, watcherr := fsnotify.NewWatcher()
	if watcherr != nil {
		panic(watcherr)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				fmt.Printf("event=%s, ok=%t\n", event, ok)
			case err, ok := <-watcher.Errors:
				fmt.Printf("err=%s, ok=%t\n", err, ok)
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

		<-watcher.Events

		data := fmt.Sprintf("data: servus pid=%d\n\n", os.Getpid())
		bytes, err := w.Write([]byte(data))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("[%d] %s", bytes, data)
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
	fmt.Printf("pid=%d url=http://localhost:3000\n", os.Getpid())
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
