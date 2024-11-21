package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
)

func Polling(entrypoint, updatedAt string, limit int) (gjson.Result, error) {
	params := make(map[string]string)
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if updatedAt != "" {
		params["updated_at"] = updatedAt
	}
	resp, err := HttpClient.R().
		SetQueryParams(params).
		SetHeader("Accept", "application/json").
		Get(entrypoint)
	if err != nil {
		return gjson.Result{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return gjson.Result{}, fmt.Errorf("HTTP %d\n%s", resp.StatusCode(), resp.String())
	}
	return gjson.Parse(resp.String()), nil
}

func ChangelogList(entrypoint, cursor, docUpdatedAt string, limit int) (gjson.Result, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if cursor != "" {
		params["cursor"] = cursor
	} else if docUpdatedAt != "" {
		params["doc_updated_at"] = docUpdatedAt
	}
	resp, err := HttpClient.R().
		SetQueryParams(params).
		SetHeader("Accept", "application/json").
		Get(entrypoint)
	if err != nil {
		return gjson.Result{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return gjson.Result{}, fmt.Errorf("HTTP %d\n%s", resp.StatusCode(), resp.String())
	}

	return gjson.Parse(resp.String()), nil
}
