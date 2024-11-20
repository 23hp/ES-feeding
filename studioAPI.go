package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func pollingOnce(entrypoint, updatedAt, limit string) ([]Document, error) {
	req, err := http.NewRequest("GET", entrypoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	params := req.URL.Query()
	if updatedAt != "" {
		params.Add("updated_at", updatedAt)
		log.Printf("Fetching documents since: %v", updatedAt)
	} else {
		log.Println("Update history not found, fetching all documents")
	}
	if limit != "" {
		params.Add("limit", limit)
	}
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received HTTP %d", resp.StatusCode)
	}

	var docs []Document
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, err
	}
	return docs, nil
}

type Document struct {
	ID            int    `json:"id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Filename      string `json:"filename"`
	FileExtension string `json:"file_extension"`
	ByteSize      int    `json:"byte_size"`
	Key           string `json:"key"`
	URL           string `json:"url"`
}
