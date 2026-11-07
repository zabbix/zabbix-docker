package bootstrap

import (
	"encoding/json"
	"strings"
)

// DecodeJSON decodes JSON and accepts single-quoted input for compatibility
// with existing image configuration.
func DecodeJSON(raw string) (any, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value, nil
	}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(raw, "'", `"`)), &value); err != nil {
		return nil, err
	}
	return value, nil
}
