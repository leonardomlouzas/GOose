package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := &http.Server {
		Addr:		":8080",
		Handler: 	mux,
	}
	mux.Handle("/", http.FileServer(http.Dir("./")))
	
	log.Println("Server starting on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
	}

}