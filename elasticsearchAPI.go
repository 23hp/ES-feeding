package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func feedElasticsearch(docs []Document, esEntrypoint, index string) (string, error) {
	bulkAPIUrl := fmt.Sprintf("%s/%s/_bulk", esEntrypoint, index)
	req, err := http.NewRequest("POST", bulkAPIUrl, strings.NewReader(buildBulkBody(docs)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bulk update failed: %v", resp.Status)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data["errors"].(bool) {
		// TODO: add failed items to failure table, extract items.*.error.type
		return "", fmt.Errorf("partial bulk update failed")
	}

	return docs[len(docs)-1].UpdatedAt, nil
}

func buildBulkBody(documents []Document) string {
	var body strings.Builder
	for _, doc := range documents {
		action := fmt.Sprintf(`{"update":{"_id":"%d"}}`, doc.ID)
		body.WriteString(action)
		body.WriteString("\n")

		docBytes, _ := json.Marshal(doc)
		var docMap map[string]interface{}
		err := json.Unmarshal(docBytes, &docMap)
		if err != nil {
			break
		}
		delete(docMap, "id")
		docBytes, _ = json.Marshal(docMap)
		content := fmt.Sprintf(`{"doc": %s, "doc_as_upsert": true}`, string(docBytes))
		body.Write([]byte(content))
		body.WriteString("\n")
	}
	return body.String()
}

func getIndexLastUpdatedAt(esEntrypoint, index string) (string, error) {
	metadataUrl := fmt.Sprintf("%s/%s/_mapping?filter_path=%[2]s.mappings._meta.updated_at", esEntrypoint, index)
	resp, err := http.Get(metadataUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("index not available")
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", nil
	}
	updatedAt, ok := data[index].(map[string]interface{})["mappings"].(map[string]interface{})["_meta"].(map[string]interface{})["updated_at"].(string)
	if !ok {
		log.Printf("Metadata parsing error: %v", data)
		return "", nil
	}
	return updatedAt, nil
}

func setIndexMetadata(esEntrypoint, index, updatedAt string) error {
	metadataUrl := fmt.Sprintf("%s/%s/_mapping", esEntrypoint, index)
	bodyString := fmt.Sprintf(`{"_meta":{"updated_at":"%s"}}`, updatedAt)

	req, err := http.NewRequest("PUT", metadataUrl, strings.NewReader(bodyString))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error updating index metadata: %v", resp.Status)
	}
	return nil
}
