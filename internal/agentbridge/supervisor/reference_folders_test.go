package supervisor

import (
	"context"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"testing"
)

func TestReferenceFoldersRequireActualExecutionBinding(t *testing.T) {
	actor := &Actor{}
	req := &bridge.TaskRequest{Cwd: "/original/workdir", ReferenceFolders: &assignmentcontract.ReferenceFolders{DeviceID: "device", RuntimeID: "runtime", Paths: []string{"/reference"}}, Metadata: map[string]string{"execution_device_id": "device"}}
	status := runtimeactor.Status{RuntimeID: "runtime"}
	if _, err := actor.prepareWorkspace(context.Background(), status, req); err != nil {
		t.Fatal(err)
	}
	if req.Cwd != "/original/workdir" {
		t.Fatal("reference changed cwd")
	}
	req.Metadata["execution_device_id"] = "another-device"
	if _, err := actor.prepareWorkspace(context.Background(), status, req); err == nil {
		t.Fatal("accepted another device")
	}
	req.Metadata["execution_device_id"] = "device"
	status.RuntimeID = "another-runtime"
	if _, err := actor.prepareWorkspace(context.Background(), status, req); err == nil {
		t.Fatal("accepted another runtime")
	}
}
