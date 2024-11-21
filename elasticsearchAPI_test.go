package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBulkIndex_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		file, err := os.Open("./testdata/bulkResponse.json")
		if err != nil {
			t.Fatalf("Failed to open file: %v", err)
		}
		defer file.Close()
		byteValue, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		w.Write(byteValue)
	}))
	defer server.Close()

	res, err := bulkIndex([]BulkRequestItem{}, server.URL, "test-index")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(res) != 1 || res[0].ID != "another_new_doc" || res[0].Type != "version_conflict_engine_exception" || res[0].Action != Create {
		t.Fatalf("Expected to have an error ")
	}
}

func TestBuildBulkBody_EmptyList(t *testing.T) {
	body := buildBulkBody([]BulkRequestItem{})
	if body != "" {
		t.Fatalf("Expected empty body, got %v", body)
	}
}

func TestBuildBulkBody_ValidList(t *testing.T) {
	bulkList := []BulkRequestItem{{Action: Create, ID: "1", Document: Document{"name": "test1"}}}
	body := buildBulkBody(bulkList)
	expected := `{"create":{"_id":"1"}}` + "\n" + `{"name":"test1"}` + "\n"
	if body != expected {
		t.Fatalf("Expected %v, got %v", expected, body)
	}
	bulkList = []BulkRequestItem{{Action: Update, ID: "2", Document: Document{"age": "22"}}}
	body = buildBulkBody(bulkList)
	expected = `{"update":{"_id":"2"}}` + "\n" + `{"doc": {"age":"22"}, "doc_as_upsert": true}` + "\n"
	if body != expected {
		t.Fatalf("Expected %v, got %v", expected, body)
	}
	bulkList = []BulkRequestItem{{Action: Index, ID: "3", Document: Document{"color": "red"}}}
	body = buildBulkBody(bulkList)
	expected = `{"index":{"_id":"3"}}` + "\n" + `{"color":"red"}` + "\n"
	if body != expected {
		t.Fatalf("Expected %v, got %v", expected, body)
	}
	bulkList = []BulkRequestItem{{Action: Delete, ID: "4", Document: Document{"size": "M"}}}
	body = buildBulkBody(bulkList)
	expected = `{"delete":{"_id":"4"}}` + "\n"
	if body != expected {
		t.Fatalf("Expected %v, got %v", expected, body)
	}
}
