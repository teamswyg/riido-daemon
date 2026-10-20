package codex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexHostDiscoveryContractReachesStartAndResumeWire(t *testing.T) {
	for _, resume := range []string{"", "existing-native-thread"} {
		t.Run("resume="+resume, func(t *testing.T) {
			prompt, system, _, _ := agentbridge.ApplyRuntimeInstructionContract(Name, "Summarize my recent repositories.", "", "")
			d := &protocolDriver{req: agentbridge.StartRequest{Prompt: prompt, SystemPrompt: system, ResumeSessionID: resume}}
			method, params := d.threadStartRequest()
			want := codexMethodThreadStart
			if resume != "" {
				want = codexMethodThreadResume
			}
			if method != want || params["developerInstructions"] != system || prompt != "Summarize my recent repositories." {
				t.Fatal("host discovery contract lost at native request boundary")
			}
			wire, err := json.Marshal(params)
			if err != nil || !strings.Contains(string(wire), "Riido host context discovery contract:") || !strings.Contains(string(wire), "includeOutputs=false") {
				t.Fatal("native wire lost discovery/read budget instructions")
			}
			if _, ok := params["dynamicTools"]; ok {
				t.Fatal("instruction tuning must not fabricate or grant native tools")
			}
		})
	}
}
