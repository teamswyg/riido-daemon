package main

import (
	"net/http"
	"net/url"
	"strconv"
)

func enrollSaaSDevice(cfg config, client apiClient, slot int) (saasDeviceCredential, scenario) {
	displayName := "Riido Local QA Slot " + strconv.Itoa(slot)
	body := map[string]any{
		"machine_id":   localQAMachineID(slot),
		"display_name": displayName,
		"platform":     "darwin",
		"app_version":  "local-product-acceptance",
	}
	path := "/v2/desktop/workspaces/" + url.PathEscape(*cfg.workspaceID) + "/devices/enroll"
	sc, payload := apiQueryPayload(client, saasEnrollScenarioID(slot), http.MethodPost, path, body, summarizeDeviceEnroll)
	secret, _ := payload["device_secret"].(string)
	deviceID, _ := payload["device_id"].(string)
	return saasDeviceCredential{DeviceID: deviceID, DeviceSecret: secret, DisplayName: displayName}, sc
}

func summarizeDeviceEnroll(payload map[string]any) map[string]any {
	return map[string]any{
		"device_id":              payload["device_id"],
		"device_id_present":      stringPresent(payload["device_id"]),
		"device_secret_returned": stringPresent(payload["device_secret"]),
		"display_name":           payload["display_name"],
	}
}

func saasEnrollScenarioID(slot int) string {
	return "local.saas.device_enroll." + strconv.Itoa(slot)
}
