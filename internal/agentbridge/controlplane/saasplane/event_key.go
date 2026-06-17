package saasplane

import (
	"context"
	"maps"
	"strconv"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (p *Plane) assignmentEventMetadata(
	ctx context.Context,
	assignment assignmentcontract.Assignment,
	metadata map[string]string,
) (map[string]string, error) {
	keyName := metadatakeys.AssignmentEventKey.String()
	if strings.TrimSpace(metadata[keyName]) != "" {
		return metadata, nil
	}
	seq, err := p.nextAssignmentEventSeq(ctx)
	if err != nil {
		return nil, err
	}
	out := cloneEventMetadata(metadata)
	out[keyName] = assignmentEventKey(p.cfg.DaemonID, assignment.ID, seq)
	return out, nil
}

func (p *Plane) nextAssignmentEventSeq(ctx context.Context) (uint64, error) {
	var seq uint64
	err := p.withState(ctx, func(s *planeState) {
		s.nextAssignmentEventSeq++
		seq = s.nextAssignmentEventSeq
	})
	return seq, err
}

func assignmentEventKey(daemonID, assignmentID string, seq uint64) string {
	if daemonID == "" {
		daemonID = "daemon"
	}
	if assignmentID == "" {
		assignmentID = "assignment"
	}
	return daemonID + ":" + assignmentID + ":" + strconv.FormatUint(seq, 10)
}

func cloneEventMetadata(metadata map[string]string) map[string]string {
	out := make(map[string]string, len(metadata)+1)
	maps.Copy(out, metadata)
	return out
}
