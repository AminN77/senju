package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("senju api bootstrap ready\n"))
	})

	addr := ":" + port
	log.Printf("starting api on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(fmt.Errorf("api server stopped: %w", err))
	}
}
