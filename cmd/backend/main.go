package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var port string

var okResponse = map[string]string{"status": "ok"}
var errorResponse = map[string]string{"status": "error"}

var healthzResponseCode = http.StatusOK
var healthzResponse = okResponse

// defaultHandler handles the default request, returning the port and PID
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()

		delay := query.Get("delay")
		if delay != "" {
			d, err := time.ParseDuration(delay)
			if err != nil {
				fmt.Fprintf(w, "Error: %v\n", err)
				return
			}
			time.Sleep(d)
		}

		response := make(map[string]any)

		response["port"] = port
		response["pid"] = os.Getpid()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		log.Printf("Serving request from %s\n", r.RemoteAddr)

		json.NewEncoder(w).Encode(response)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// healthzHandler handles the healthz request, returning a simple status message
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(healthzResponseCode)

		json.NewEncoder(w).Encode(healthzResponse)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// healthzToggleHandler toggles the healthz response code and message
func healthzToggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if healthzResponseCode == http.StatusOK {
			healthzResponseCode = http.StatusInternalServerError
			healthzResponse = errorResponse
		} else {
			healthzResponseCode = http.StatusOK
			healthzResponse = okResponse
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(okResponse)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

}

func main() {
	flag.StringVar(&port, "port", "9000", "Defines port to start mock server")
	flag.Parse()

	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/toggle-health", healthzToggleHandler)

	fmt.Println("Server started on port", port)
	http.ListenAndServe(":"+port, nil)
}
