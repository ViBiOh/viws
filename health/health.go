package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	httpClient := http.Client{Timeout: 5 * time.Second}

	request, err := http.NewRequest(`GET`, `http://localhost:1080/health`, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf(`HTTP/%d`, response.StatusCode)
		os.Exit(1)
	}

	os.Exit(0)
}
