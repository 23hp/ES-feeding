package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPollingOnce_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[{"id":1,"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z","filename":"file1.txt","file_extension":"txt","byte_size":123,"key":"key1","url":"http://example.com/file1.txt"}]`)
	}))
	defer server.Close()

	docs, err := Polling(server.URL, "", 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if docs.Get("#").Int() != 1 {
		t.Fatalf("Expected 1 document, got %d", docs.Get("#").Int())
	}
}

func TestPollingOnce_NoDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[]`)
	}))
	defer server.Close()

	docs, err := Polling(server.URL, "", 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if docs.Get("#").Int() != 0 {
		t.Fatalf("Expected 0 documents, got %d", docs.Get("#").Int())
	}
}

func TestPollingOnce_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := Polling(server.URL, "", 100)
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}

func TestPollingOnce_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `invalid json`)
	}))
	defer server.Close()

	_, err := Polling(server.URL, "", 100)
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}
