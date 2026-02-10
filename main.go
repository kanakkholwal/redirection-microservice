package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	if err := connectDB(); err != nil {
		log.Fatal("DB connection failed:", err)
	}
	http.HandleFunc("/links", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createLinkHandler(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			deleteLinkHandler(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createDomainHandler(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			deleteDomainHandler(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/", redirectHandler)

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	path := strings.TrimPrefix(r.URL.Path, "/")

	if path == "" {
		http.NotFound(w, r)
		return
	}

	dest, err := lookupDestination(host, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, dest, http.StatusMovedPermanently)
}

type CreateLinkRequest struct {
	Domain      string `json:"domain"`
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}

func createLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.Slug == "" || req.Destination == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if err := upsertLink(req.Domain, req.Slug, req.Destination); err != nil {
		http.Error(w, "failed to create link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type DeleteLinkRequest struct {
	Domain string `json:"domain"`
	Slug   string `json:"slug"`
}

func deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.Slug == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if err := deleteLink(req.Domain, req.Slug); err != nil {
		http.Error(w, "failed to delete link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type CreateDomainRequest struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"` // "active" or "inactive"
}

func createDomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Hostname == "" {
		http.Error(w, "hostname required", http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "inactive"
	}

	if status != "active" && status != "inactive" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := upsertDomain(req.Hostname, status); err != nil {
		http.Error(w, "failed to create domain", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type DeleteDomainRequest struct {
	Hostname string `json:"hostname"`
}

func deleteDomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Hostname == "" {
		http.Error(w, "hostname required", http.StatusBadRequest)
		return
	}

	if err := deactivateDomain(req.Hostname); err != nil {
		http.Error(w, "failed to deactivate domain", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
