package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	c := os.Getenv("COLOR")
	if len(c) == 0 {
		c = "green"
	}

	addr := os.Getenv("LISTEN")
	if len(addr) == 0 {
		addr = ":8080"
	}

	count := 0

	m := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: m}

	log.Printf("Server started\n")

	// Healthcheck endpoint
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	// Simulate failure
	boom, _ := os.ReadFile("public/shutdown.html")
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(boom))
		log.Printf("Received shutdown request\n")
		go func() {
			if err := s.Shutdown(context.Background()); err != nil {
				log.Fatal(err)
			}
		}()
	})

	// Dashboard
	dashboard, _ := os.ReadFile("public/dashboard.html")
	m.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(dashboard))
		log.Printf("GET %s\n", r.URL.Path)
	})

	// Default
	index, _ := os.ReadFile("public/index.html")
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count += 1
		fmt.Fprintf(w, string(index), c, count)
		//log.Printf("GET %s\n", r.URL.Path)
	})

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Printf("Exiting")
}
