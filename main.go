package main

import (
	"fmt"
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

		w.Write([]byte(fmt.Sprintf("<h2>%s</h2>", time.Now())))
		file, err := os.Open(name)
		if err != nil {
			return
		}
		file.WriteTo(w)
	})
	http.ListenAndServe(":3000", nil)
}
