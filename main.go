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

	script := `<script> new EventSource(".servus").onmessage = function(ev){ console.log(ev); window.location.reload();}</script>`
	http.HandleFunc("GET /.servus", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)

		<-signalchan
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
