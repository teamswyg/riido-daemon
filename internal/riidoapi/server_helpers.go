package riidoapi

import (
	"encoding/json"
	"fmt"
	"net"
)

func rawParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode riido API params: %w", err)
	}
	return data, nil
}

func okResponse(method string, data any) responseEnvelope {
	payload, err := json.Marshal(data)
	if err != nil {
		return errorResponse(method, err)
	}
	return responseEnvelope{
		OK:     true,
		Method: method,
		Data:   payload,
	}
}

func errorResponse(method string, err error) responseEnvelope {
	return responseEnvelope{
		OK:     false,
		Method: method,
		Error:  err.Error(),
	}
}

func writeResponse(conn net.Conn, response responseEnvelope) error {
	encoder := json.NewEncoder(conn)
	return encoder.Encode(response)
}

type requestEnvelope struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
