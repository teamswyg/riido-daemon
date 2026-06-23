package main

type coverageSnapshot struct {
	Summary coverageSummary       `json:"summary"`
	Rows    []coverageSnapshotRow `json:"rows"`
}

func writeCoverageSnapshot(path string, rows []coverageRow, summary coverageSummary) error {
	return writeJSON(path, coverageSnapshot{Summary: summary, Rows: coverageSnapshotRows(rows)})
}

type coverageSnapshotRow struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Tier       string          `json:"tier"`
	Surface    string          `json:"surface"`
	Status     string          `json:"status"`
	Evidence   string          `json:"evidence,omitempty"`
	ExpiresAt  string          `json:"expires_at,omitempty"`
	Repair     *repairEvidence `json:"repair,omitempty"`
	Detail     string          `json:"detail,omitempty"`
	Screenshot string          `json:"screenshot,omitempty"`
}

func coverageSnapshotRows(rows []coverageRow) []coverageSnapshotRow {
	out := make([]coverageSnapshotRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, coverageSnapshotRow{
			ID: row.ID, Title: row.Title, Tier: row.Tier, Surface: row.Surface,
			Status: row.Status, Evidence: row.Evidence, ExpiresAt: row.ExpiresAt,
			Repair: coverageSnapshotRepair(row.Repair), Detail: row.Detail,
			Screenshot: row.Screenshot,
		})
	}
	return out
}

func coverageSnapshotRepair(repair repairEvidence) *repairEvidence {
	if repair.Class == "" {
		return nil
	}
	return &repair
}
