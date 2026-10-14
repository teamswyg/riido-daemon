package taskdbplane

import (
	"sort"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func claimCandidates(db taskdb.TaskDB) []taskdb.TaskRecord {
	out := make([]taskdb.TaskRecord, 0, len(db.Tasks))
	for _, record := range db.Tasks {
		if record.State.Code() == task.TaskStateCodeQueued {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i].UpdatedAt
		right := out[j].UpdatedAt
		if left != right {
			return left < right
		}
		return out[i].ID < out[j].ID
	})
	return out
}
