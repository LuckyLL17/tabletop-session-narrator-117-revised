package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 2<<20))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("请求 JSON 无效: %w", err)
	}
	return nil
}

func pathParts(path string) []string {
	trimmed := strings.Trim(path, "/")
	return strings.Split(trimmed, "/")
}
