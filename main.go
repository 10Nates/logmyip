package main

import (
	"fmt"
	"net/http"
)

const address = ":52899"

func main() {
	handlers()
	fmt.Println("Log My IP server serving on " + address)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}

func handlers() {
	//fileserver
	fs := http.FileServer(http.Dir("src/"))
	http.Handle("/src/", http.StripPrefix("/src/", fs))

	//handle pages
	http.HandleFunc("/", home)
	http.HandleFunc("/ipinfo", ipinfow)
	http.HandleFunc("/logip", logip)
	http.HandleFunc("/rendermap.svg", rendermapw)

}
