package main

func apiQuery(client apiClient, id, method, path string, body any, summarize summarizeFunc) scenario {
	out, _ := apiQueryPayload(client, id, method, path, body, summarize)
	return out
}

func apiQueryPayload(client apiClient, id, method, path string, body any, summarize summarizeFunc) (scenario, map[string]any) {
	payload, statusCode, err := client.call(method, path, body)
	out := scenario{ID: id, Method: method, Endpoint: path}
	if err != nil {
		out.Status = statusFailed
		out.FailureSummary = err.Error()
		out.Observed = map[string]any{"status_code": statusCode}
		out.Repair = apiRuntimeRepair()
		return out, payload
	}
	out.Status = statusPassed
	out.Observed = summarize(payload)
	out.Observed["status_code"] = statusCode
	return out, payload
}
