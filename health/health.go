package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	httpClient := http.Client{Timeout: 5 * time.Second}

	request, err := http.NewRequest(`GET`, `http://localhost:1080/health`, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf(`HTTP/%d`, response.StatusCode)
	}

	log.Print(`Health succeed`)
}
