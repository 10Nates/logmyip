package main

import (
	"fmt"
	"net/http"
)

const address = ":3068"

func main() {
	mux := http.NewServeMux()
	initdb()
	handlers(mux)
	fmt.Println("Log My IP server serving on " + address)

	err := http.ListenAndServe(address, cspHandler(mux))
	if err != nil {
		panic(err)
	}
}

func handlers(mux *http.ServeMux) {
	//fileserver
	fs := http.FileServer(http.Dir("src/"))
	mux.Handle("/src/", http.StripPrefix("/src/", fs))

	//handle pages
	mux.HandleFunc("/", home)
	mux.HandleFunc("/ipinfo", ipinfow)
	mux.HandleFunc("/logip", logip)
	mux.HandleFunc("/unlog", unlogpage)
	mux.HandleFunc("/unlogip", unlogip)
	mux.HandleFunc("/rendermap.svg", rendermapw)
}

func cspHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}
