package config

import (
	"encoding/json"
	"fmt"
)

// HTTPHeaders supports YAML maps and JSON objects supplied through environment variables.
type HTTPHeaders map[string]string

func (h *HTTPHeaders) UnmarshalText(text []byte) error {
	var headers map[string]string
	if err := json.Unmarshal(text, &headers); err != nil {
		return fmt.Errorf("parse HTTP headers JSON: %w", err)
	}
	*h = headers
	return nil
}
