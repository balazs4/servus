package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
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
		w.Write([]byte(fmt.Sprintf("<h2>%s</h2>", time.Now())))
		io.Copy(w, file)
	})
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
