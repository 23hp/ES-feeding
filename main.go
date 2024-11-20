package main

import (
	"log"
	"os"
)

func main() {
	esEntrypoint := os.Getenv("ES_ENTRYPOINT")
	esIndex := os.Getenv("ES_INDEX")
	pollingEntrypoint := os.Getenv("POLLING_ENTRYPOINT")
	batchSize := os.Getenv("BATCH_SIZE")

	lastUpdatedAt, err := getIndexLastUpdatedAt(esEntrypoint, esIndex)
	if err != nil {
		log.Fatalf("Index unavailable: %v", err)
	}

	docs, err := pollingOnce(pollingEntrypoint, lastUpdatedAt, batchSize)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if len(docs) == 0 || (len(docs) == 1 && docs[0].UpdatedAt == lastUpdatedAt) {
		log.Println("No new documents, exiting...")
		os.Exit(0)
	}

	log.Printf("Indexing %v documents...", len(docs))
	currentUpdatedAt, err := feedElasticsearch(docs, esEntrypoint, esIndex)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Current updated at: %v", currentUpdatedAt)

	if err := setIndexMetadata(esEntrypoint, esIndex, currentUpdatedAt); err != nil {
		log.Fatalf("Unable to save metadata: %v", err)
	} else {
		log.Println("Index metadata updated")
	}
}
