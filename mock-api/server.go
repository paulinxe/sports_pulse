package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// Start starts the HTTP server
func Start() error {
	// TODO: Set up routes
	// GET /competitions/{id}/matches
	// GET /matches/{id}
	
	port := getPort()
	slog.Info("Starting HTTP server", "port", port)
	
	http.HandleFunc("/health", healthHandler)
	
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

