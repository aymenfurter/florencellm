package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	initDatabase()
	initServiceBus()

	http.HandleFunc("/api/repository", repositoryHandler)

	port := "8081"
	fmt.Printf("Starting repository microservice on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

