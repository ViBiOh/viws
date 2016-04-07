package main

import "net/http"
import "log"

const port = "1080"

func main() {
	http.Handle("/", http.FileServer(http.Dir("/www/")))

	log.Print("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
