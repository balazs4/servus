package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGUSR2)

	script := `<script> new EventSource('/.servus').onmessage = function(){ console.log("data"); location.reload(); }</script>`
	http.HandleFunc("GET /.servus", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Add("content-type", "text/event-stream")
		w.Header().Add("connection", "keep-alive")
		w.Header().Add("cache-control", "no-cache")

		<-signalchan
    fmt.Println("SIGUSR2");
		data := fmt.Sprintf("data: servus pid %d\n\n", os.Getpid())
		w.Write([]byte(data))
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
	fmt.Printf("pid=%d url=http://localhost:3000", os.Getpid())
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
