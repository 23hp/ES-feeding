package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	updatedAtField = "updated_at"
	cursorField    = "cursor"
	waitTime       = 3 * time.Minute
	updateInterval = 5 * time.Second
)

var (
	cursor     string
	updatedAt  string
	failedList []FailedItem
)

func main() {
	esEntrypoint := os.Getenv("ES_ENTRYPOINT")
	esIndex := os.Getenv("ES_INDEX")
	pollingEntrypoint := os.Getenv("POLLING_ENTRYPOINT")
	changelogsEntrypoint := os.Getenv("CHANGELOGS_ENTRYPOINT")
	batchSize, _ := strconv.Atoi(Env("BATCH_SIZE", "1000"))

	metadata, err := getIndexMetadata(esEntrypoint, esIndex)
	if err != nil {
		log.Fatalf("Index unavailable: %v\n", err)
	}
	fmt.Printf("%s metadata:\n%s\n", esIndex, metadata.String())
	cursorResult := metadata.Get(cursorField)
	updatedAtResult := metadata.Get(updatedAtField)
	if cursorResult.Exists() {
		cursor = cursorResult.String()
	} else {
		fmt.Printf("---Start full-sync for changes after: %s\n", updatedAtResult.String())
		for {
			documents, err := Polling(pollingEntrypoint, updatedAt, batchSize)
			if err != nil {
				log.Fatalf("Polling Service unavailable: %v\n", err)
			}
			docSize := documents.Get("#").Int()
			if docSize == 0 {
				break
			}

			failures, err := bulkIndex(buildBulkListViaDocs(documents), esEntrypoint, esIndex)
			if err != nil {
				log.Printf("Index %s not ready, Error: %v\nRetry after %s\n", esIndex, err, waitTime)
				time.Sleep(waitTime)
				continue
			}
			fmt.Printf("%d changes after %s are indexed\n", batchSize, updatedAt)
			if len(failedList) > 0 {
				log.Printf("%d changes failed to sync\n", len(failedList))
				failedList = append(failedList, failures...)
			}

			updatedAt = documents.Get("@reverse|0." + updatedAtField).String()
			meta := BuildMetadata(metadata, updatedAtField, updatedAt)
			if err := setIndexMetadata(esEntrypoint, esIndex, meta); err != nil {
				log.Printf("Failed to update_at document %v\n", err)
			}
			if docSize < int64(batchSize) {
				break
			}
		}
	}

	if len(failedList) > 0 {
		log.Printf("%d items failed to index\n", len(failedList))
	}

	fmt.Printf("---Start incremental sync changes after cursor: %s, updated_at: %s\n", cursor, updatedAt)

	for {
		changelogs, err := ChangelogList(changelogsEntrypoint, cursor, updatedAt, batchSize)
		if err != nil {
			log.Printf("ChangelogList API isn't ready, Error: %v\nRetry after %s\n", err, waitTime)
			time.Sleep(waitTime)
			continue
		}
		count := changelogs.Get("#").Int()
		if count == 0 {
			fmt.Printf("No changelogs found, retry after %s\n", updateInterval)
			time.Sleep(updateInterval)
			continue
		}
		fmt.Printf("---Synchronising of %d changes after cursor: %s, updated_at: %s\n", count, cursor, updatedAt)
		failures, err := bulkIndex(buildBulkListViaLogs(changelogs), esEntrypoint, esIndex)
		if err != nil {
			log.Printf("Index %s not ready, Error: %v\nRetry after %s\n", esIndex, err, waitTime)
			time.Sleep(waitTime)
			continue
		}
		if len(failedList) > 0 {
			log.Printf("%d changes failed to sync\n", len(failedList))
			failedList = append(failedList, failures...)
		}

		cursor = changelogs.Get("@reverse|0." + cursorField).String()
		meta := BuildMetadata(metadata, cursorField, cursor)
		if err := setIndexMetadata(esEntrypoint, esIndex, meta); err != nil {
			log.Printf("Cursor %s saving failed to %v\n", cursor, err)
		}
	}
}
