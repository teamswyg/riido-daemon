package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
)

func (p *Plane) postJSON(ctx context.Context, path string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return p.doJSON(ctx, http.MethodPost, path, body, out)
}

func (p *Plane) getJSON(ctx context.Context, path string, out any) error {
	return p.doJSON(ctx, http.MethodGet, path, nil, out)
}
