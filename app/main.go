package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
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
	var healthy atomic.Bool
	healthy.Store(true)

	m := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: m}

	log.Printf("Server started\n")

	// Healthcheck endpoint
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !healthy.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "SHUTTING DOWN")
			return
		}
		fmt.Fprintf(w, "OK")
	})

	// Simulate failure
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		healthy.Store(false)
		boom, _ := os.ReadFile("public/shutdown.html")
		w.Write(boom)
		log.Printf("Received shutdown request, failing health checks and waiting for drain\n")
		go func() {
			time.Sleep(12 * time.Second)
			log.Printf("Drain period complete, shutting down server\n")
			if err := s.Shutdown(context.Background()); err != nil {
				log.Fatal(err)
			}
		}()
	})

	// Dashboard
	m.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		dashboard, _ := os.ReadFile("public/dashboard.html")
		w.Write(dashboard)
		log.Printf("GET %s\n", r.URL.Path)
	})

	// Default
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, _ := os.ReadFile("public/index.html")
		count += 1
		fmt.Fprintf(w, string(index), c, count)
		//log.Printf("GET %s\n", r.URL.Path)
	})

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Printf("Exiting")
}
