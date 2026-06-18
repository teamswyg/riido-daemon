package mwsdbridge

import "encoding/json"

type request struct {
	Method Method `json:"method"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method Method          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
