package formatter

import (
	"encoding/json"

	"enva/internal/types"
)

// FormatJSON formats the validation result as JSON
func FormatJSON(result *types.ValidationResult) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal result"}`
	}
	return string(data)
}
