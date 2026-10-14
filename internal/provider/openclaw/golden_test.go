package openclaw

import "testing"

func TestGoldenFullResultJSON(t *testing.T) {
	raws := goldenRawEvents(t, "full_result.json", trimGoldenTrailingNewline)

	assertGoldenFullResultCoverage(t, raws)
}

func TestGoldenNDJSONResultJSONL(t *testing.T) {
	raws := goldenRawEvents(t, "ndjson_result.jsonl", keepGoldenFixtureBytes)

	assertGoldenNDJSONCoverage(t, raws)
}
