package saasplane

import "net/http"

func (p *Plane) reportDeviceCredentialRejection(statusCode int, err error) {
	if p == nil || p.cfg.DeviceSecret == "" || statusCode != http.StatusUnauthorized || err == nil {
		return
	}
	p.deviceCredentialOnce.Do(func() {
		p.deviceCredentialRejected <- err
	})
}
