package saasplane

import (
	"net/http/httptest"
	"sync"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type fakeAssignmentServer struct {
	t            *testing.T
	server       *httptest.Server
	mu           sync.Mutex
	bearerToken  string
	deviceID     string
	deviceSecret string

	assignmentsByAgent  map[string][]assignmentcontract.Assignment
	assignmentsByID     map[string]assignmentcontract.Assignment
	activeByAgent       map[string]assignmentcontract.Assignment
	cancelByAgent       map[string]assignmentcontract.Assignment
	staleHeartbeatIDs   map[string]bool
	requestCounts       map[string]int
	transientFailures   map[string]int
	transientStatuses   map[string]int
	bindings            []assignmentcontract.AgentRuntimeBinding
	pollRequestsByAgent map[string][]assignmentcontract.PollRequest
	runtimeSnapshots    []DeviceRuntimeSnapshotSyncRequest
	events              []assignmentcontract.AgentEventRequest
	heartbeats          []assignmentcontract.AgentHeartbeatRequest
	toolApprovals       []assignmentcontract.ToolApprovalRequest
	toolApprovalWaits   []assignmentcontract.ToolApprovalWaitRequest
	toolDecision        *assignmentcontract.ToolApprovalDecision
	toolApprovalStatus  assignmentcontract.ApprovalStatus
}
