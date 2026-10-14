package saasplane

import "net/http"

func (p *Plane) attachAuthHeaders(req *http.Request) {
	if p.cfg.DeviceSecret != "" {
		req.Header.Set("X-Riido-Device-Id", p.cfg.DeviceID)
		req.Header.Set("X-Riido-Device-Secret", p.cfg.DeviceSecret)
	}
	if p.cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.BearerToken)
	}
}
