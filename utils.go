package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"strings"
	"time"
)

var HttpClient = resty.New().
	SetRetryCount(3).
	SetRetryWaitTime(5 * time.Second).
	SetRetryMaxWaitTime(1 * time.Minute).
	AddRetryCondition(
		func(r *resty.Response, err error) bool {
			return r.StatusCode() > 500 || r.StatusCode() == http.StatusTooManyRequests
		},
	)

func PrintRequestError(response *resty.Response, err error) {
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", response.StatusCode())
	headers, _ := json.MarshalIndent(response.Header(), "", "  ")
	fmt.Println("  Status Code:", string(headers))
	fmt.Println("  Body       :\n", response)
}

func buildBulkListViaDocs(documents gjson.Result) []BulkRequestItem {
	var list []BulkRequestItem
	documents.ForEach(func(key, value gjson.Result) bool {
		id := value.Get("id").String()
		content, ok := value.Value().(map[string]interface{})
		if !ok {
			return false
		} else {
			delete(content, "id")
			list = append(list, BulkRequestItem{Action: Index, ID: id, Document: content})
			return true
		}
	})
	return list
}

func buildBulkListViaLogs(changelogs gjson.Result) []BulkRequestItem {
	var list []BulkRequestItem
	changelogs.ForEach(func(key, value gjson.Result) bool {
		status := value.Get("status").String()
		var action Action
		switch status {
		case "deleted":
			action = Delete
		case "updated":
			action = Update
		case "created":
			action = Index
		default:
			return false
		}
		requestItem := BulkRequestItem{Action: action, ID: value.Get("id").String()}
		content, ok := value.Get("content").Value().(map[string]interface{})
		if ok {
			requestItem.Document = content
		}
		list = append(list, requestItem)
		return true
	})
	return list
}

func buildBulkBody(list []BulkRequestItem) string {
	var body strings.Builder
	for _, item := range list {
		document := item.Document
		action := fmt.Sprintf(`{"%v":{"_id":"%s"}}`, item.Action, item.ID)
		body.WriteString(action)
		body.WriteString("\n")
		if item.Action == Delete {
			continue
		} else {
			docBytes, _ := json.Marshal(document)
			if item.Action == Update {
				body.WriteString(fmt.Sprintf(`{"doc": %s, "doc_as_upsert": true}`, docBytes))
			} else {
				body.Write(docBytes)
			}
			body.WriteString("\n")
		}
	}
	return body.String()
}

func getFailedItems(result *BulkResult) []FailedItem {
	var failures []FailedItem
	for _, item := range result.Items {
		for action, data := range item {
			if data.Error != (BulkError{}) {
				failures = append(failures, FailedItem{ID: data.ID, Type: data.Error.Type, Action: action})
			}
		}
	}
	return failures
}

func Env(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func BuildMetadata(metadata gjson.Result, key string, value any) map[string]interface{} {
	data, ok := metadata.Value().(map[string]interface{})
	if ok {
		data[key] = value
		return data
	}
	return map[string]interface{}{key: value}
}
