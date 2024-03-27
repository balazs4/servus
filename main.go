package main

import "net/http"

func main() {
  h := http.NewServeMux()
	h.HandleFunc("GET /{file}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		http.ServeFile(w, r, name)
	})
  http.ListenAndServe(":3000", h)
}
