package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
)

func bulkIndex(list []BulkRequestItem, esEntrypoint, index string) ([]FailedItem, error) {
	body := buildBulkBody(list)
	result := &BulkResult{}
	bulkAPIUrl := fmt.Sprintf("%s/%s/_bulk", esEntrypoint, index)
	resp, err := HttpClient.R().
		SetHeader("Content-Type", "application/x-ndjson").
		SetBody(body).
		SetResult(result).
		Post(bulkAPIUrl)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("index failed, HTTP %d", resp.StatusCode())
	}

	return getFailedItems(result), nil
}

func getIndexMetadata(esEntrypoint, index string) (gjson.Result, error) {
	filterPath := index + ".mappings._meta"
	metadataUrl := fmt.Sprintf("%s/%s/_mapping", esEntrypoint, index)
	resp, err := HttpClient.R().
		SetQueryParams(map[string]string{
			"filter_path": filterPath,
		}).
		Get(metadataUrl)
	if err != nil {
		return gjson.Result{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		return gjson.Result{}, fmt.Errorf("failed to get %s metadata: %s\n", index, resp.String())
	}
	metadata := gjson.Get(resp.String(), filterPath)
	return metadata, nil
}

// elasticsearch will overwrite the existing metadata
func setIndexMetadata(esEntrypoint, index string, data map[string]interface{}) error {
	metadataUrl := fmt.Sprintf("%s/%s/_mapping", esEntrypoint, index)
	body := map[string]interface{}{
		"_meta": data,
	}
	resp, err := HttpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Put(metadataUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to update %s metadata: %s\n", index, resp.String())
	}
	return nil
}
