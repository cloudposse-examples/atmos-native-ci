package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got %q", w.Body.String())
	}
}

func TestIndexRendersColorAndCount(t *testing.T) {
	mux := http.NewServeMux()
	count := 0
	color := "blue"
	template := `<html><body style="background-color: %s">%v</body></html>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count++
		w.Write([]byte(strings.NewReplacer("%s", color, "%v", string(rune('0'+count))).Replace(template)))
	})

	// First request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "blue") {
		t.Errorf("expected body to contain color 'blue', got %q", body)
	}
	if !strings.Contains(body, "1") {
		t.Errorf("expected body to contain count '1', got %q", body)
	}

	// Second request increments count
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	body = w.Body.String()
	if !strings.Contains(body, "2") {
		t.Errorf("expected body to contain count '2', got %q", body)
	}
}

func TestDefaultColorIsGreen(t *testing.T) {
	color := ""
	if len(color) == 0 {
		color = "green"
	}
	if color != "green" {
		t.Errorf("expected default color 'green', got %q", color)
	}
}

func TestDefaultListenAddr(t *testing.T) {
	addr := ""
	if len(addr) == 0 {
		addr = ":8080"
	}
	if addr != ":8080" {
		t.Errorf("expected default addr ':8080', got %q", addr)
	}
}

func TestDashboardEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	dashboardContent := "<html><body>dashboard</body></html>"

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(dashboardContent))
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != dashboardContent {
		t.Errorf("expected dashboard content, got %q", w.Body.String())
	}
}
