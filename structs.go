package main

type Document map[string]interface{}
type Action string

const (
	Create Action = "create"
	Index  Action = "index"
	Delete Action = "delete"
	Update Action = "update"
)

type BulkRequestItem struct {
	Action   Action
	Document map[string]interface{}
	ID       string
}

type BulkResult struct {
	Errors bool                        `json:"errors"`
	Took   int                         `json:"took"`
	Items  []map[Action]BulkResultItem `json:"items"`
}
type BulkResultItem struct {
	Index  string    `json:"_index"`
	ID     string    `json:"_id"`
	Status int       `json:"status"`
	Error  BulkError `json:"error,omitempty"`
}
type BulkError struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	IndexUUID string `json:"index_uuid"`
	Shard     string `json:"shard"`
	Index     string `json:"index"`
}

type FailedItem struct {
	ID     string
	Type   string
	Action Action
}
