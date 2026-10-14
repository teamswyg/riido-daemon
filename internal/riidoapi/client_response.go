package riidoapi

import (
	"encoding/json"
	"fmt"
)

func decodeResponse(responseBody []byte, method string, out any) error {
	var env responseEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return fmt.Errorf("decode riido API response: %w", err)
	}
	if !env.OK {
		return responseFailure(method, env.Error)
	}
	if env.Method != Method(method) {
		return fmt.Errorf("riido API method mismatch: requested %s got %s", method, env.Method)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode riido API %s data: %w", method, err)
	}
	return nil
}

func responseFailure(method, message string) error {
	if message != "" {
		return fmt.Errorf("riido API %s failed: %s", method, message)
	}
	return fmt.Errorf("riido API %s failed", method)
}
