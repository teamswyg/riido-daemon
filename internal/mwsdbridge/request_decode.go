package mwsdbridge

import (
	"encoding/json"
	"fmt"
)

func decodeResponse(method string, responseBody []byte, out any) error {
	var env responseEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return fmt.Errorf("decode mwsd response: %w", err)
	}
	if !env.OK {
		if env.Error != "" {
			return fmt.Errorf("mwsd %s failed: %s", method, env.Error)
		}
		return fmt.Errorf("mwsd %s failed", method)
	}
	if env.Method != Method(method) {
		return fmt.Errorf("mwsd method mismatch: requested %s got %s", method, env.Method)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode mwsd %s data: %w", method, err)
	}
	return nil
}
