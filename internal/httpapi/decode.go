package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func decode(r *http.Request, target any) error {
	maxBodyBytes := int64(2 << 20)
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodyBytes))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("请求 JSON 无效: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("请求包含多段 JSON")
	}
	return nil
}

func pathParts(path string) []string {
	trimmed := strings.Trim(path, "/")
	return strings.Split(trimmed, "/")
}
