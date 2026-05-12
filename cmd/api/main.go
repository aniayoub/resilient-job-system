package main

import (
	"log"
	"net/http"

	"github/aniayoub/resilient-job-system/internal/httpapi"
	"github/aniayoub/resilient-job-system/internal/store"
)

func main() {
	// Initialize an in-memory store for jobs
	store := store.NewStore()

	handler := httpapi.NewHander(store)

	handler.RegisterRoutes()

	// Listen to http requests and handle job creation, status retrieval, etc.
	log.Println("starting server on :8080")

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
