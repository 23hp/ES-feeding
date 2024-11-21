package main

import "testing"

func TestGetFailedItems_NoFailures(t *testing.T) {
	result := &BulkResult{
		Errors: false,
		Items: []map[Action]BulkResultItem{
			{
				Create: {
					ID: "11", Status: 201,
				},
			},
			{
				Index: {
					ID: "12", Status: 200,
				},
			},
			{
				Delete: {
					ID: "13", Status: 404,
				},
			},
			{
				Update: {
					ID: "14", Status: 200,
				},
			},
		},
	}

	failures := getFailedItems(result)
	if len(failures) != 0 {
		t.Fatalf("Expected no failures, got %v", len(failures))
	}
}

func TestGetFailedItems_WithFailures(t *testing.T) {
	result := &BulkResult{
		Errors: true,
		Items: []map[Action]BulkResultItem{
			{
				Create: {
					ID: "1", Error: BulkError{Type: "unhappy_exception", Reason: "no candy"},
				},
			},
		},
	}

	failures := getFailedItems(result)
	if len(failures) != 1 {
		t.Fatalf("Expected 1 failure, got %v", len(failures))
	}
	if (failures)[0].ID != "1" {
		t.Fatalf("Expected failure ID to be 1, got %v", (failures)[0].ID)
	}
	if (failures)[0].Type != "unhappy_exception" {
		t.Fatalf("Expected failure type to be mapper_parsing_exception, got %v", (failures)[0].Type)
	}
	if (failures)[0].Action != Create {
		t.Fatalf("Expected failure action to be index, got %v", (failures)[0].Action)
	}
}
