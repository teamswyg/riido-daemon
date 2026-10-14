package saasplane

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeJSONResponse(resp *http.Response, out any, method, path string) error {
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("saasplane: decode %s %s %s: %w", method, path, resp.Status, err)
	}
	return nil
}
