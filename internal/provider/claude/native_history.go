package claude

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/history"
)

var historySessionID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
var historyPathCharacters = regexp.MustCompile(`[^a-zA-Z0-9]`)

type historyRow struct {
	UUID           string    `json:"uuid"`
	ParentUUID     string    `json:"parentUuid"`
	SessionID      string    `json:"sessionId"`
	Cwd            string    `json:"cwd"`
	Type           string    `json:"type"`
	Sidechain      bool      `json:"isSidechain"`
	Meta           bool      `json:"isMeta"`
	CompactSummary bool      `json:"isCompactSummary"`
	Timestamp      time.Time `json:"timestamp"`
	Message        struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// Read only the exact retained session under its original project directory.
// Never search another project or fall back to the most recent session.
func claudeNativeHistory(req agentbridge.StartRequest) ([]agentbridge.Event, error) {
	if !historySessionID.MatchString(req.ResumeSessionID) || !filepath.IsAbs(req.Cwd) {
		return nil, errors.New("claude: invalid native history binding")
	}
	config, configured := req.Env["CLAUDE_CONFIG_DIR"]
	if !configured {
		config = os.Getenv("CLAUDE_CONFIG_DIR")
	}
	if config == "" {
		home := req.Env["HOME"]
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		config = filepath.Join(home, ".claude")
	}
	root, err := os.OpenRoot(filepath.Join(config, "projects"))
	if err != nil {
		return nil, errors.New("claude: native history directory unavailable")
	}
	defer root.Close()
	project := historyPathCharacters.ReplaceAllString(req.Cwd, "-")
	path := filepath.Join(project, req.ResumeSessionID+".jsonl")
	if len(project) > 200 {
		dir, err := root.Open(".")
		if err != nil {
			return nil, err
		}
		entries, err := dir.ReadDir(-1)
		dir.Close()
		if err != nil {
			return nil, err
		}
		var candidates []string
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), project[:200]+"-") {
				continue
			}
			candidate := filepath.Join(entry.Name(), req.ResumeSessionID+".jsonl")
			if _, err := root.Stat(candidate); err == nil {
				candidates = append(candidates, candidate)
			}
		}
		if len(candidates) != 1 {
			return nil, errors.New("claude: native history path is missing or ambiguous")
		}
		path = candidates[0]
	}
	file, err := root.Open(path)
	if err != nil {
		return nil, errors.New("claude: requested session transcript unavailable")
	}
	defer file.Close()
	before, err := file.Stat()
	if err != nil || !before.Mode().IsRegular() || before.Size() > 64<<20 {
		return nil, errors.New("claude: invalid or oversized native transcript")
	}
	messages, err := claudeHistoryMessages(io.LimitReader(file, before.Size()), req)
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return nil, errors.New("claude: native transcript changed during import; retry after the external turn finishes")
	}
	return history.External(req, messages)
}

func claudeHistoryMessages(reader io.Reader, req agentbridge.StartRequest) ([]agentbridge.NativeHistoryMessage, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 65536), 4<<20)
	rows := map[string]historyRow{}
	leaf := ""
	for scanner.Scan() {
		var row historyRow
		if json.Unmarshal(scanner.Bytes(), &row) != nil {
			return nil, errors.New("claude: incomplete native transcript")
		}
		if row.Sidechain || row.UUID == "" {
			continue
		}
		rows[row.UUID] = row
		if row.Type == "user" || row.Type == "assistant" {
			if row.SessionID != req.ResumeSessionID || row.Cwd != req.Cwd {
				return nil, errors.New("claude: native transcript owner mismatch")
			}
			leaf = row.UUID
		}
	}
	if scanner.Err() != nil {
		return nil, errors.New("claude: could not read complete native transcript")
	}
	var out []agentbridge.NativeHistoryMessage
	seen := map[string]bool{}
	for leaf != "" {
		row, ok := rows[leaf]
		if !ok || seen[leaf] {
			return nil, errors.New("claude: broken native transcript parent chain")
		}
		seen[leaf] = true
		leaf = row.ParentUUID
		if row.Meta || row.CompactSummary || (row.Type != "user" && row.Type != "assistant") {
			continue
		}
		text := claudeHistoryText(row.Message.Content)
		if text == "" {
			continue
		}
		role := row.Type
		if role == "assistant" {
			role = "agent"
		}
		out = append(out, agentbridge.NativeHistoryMessage{SessionID: row.SessionID, MessageID: row.UUID, Role: role, Body: text, ObservedAt: row.Timestamp})
	}
	slices.Reverse(out)
	return out, nil
}

func claudeHistoryText(content json.RawMessage) string {
	var text string
	if json.Unmarshal(content, &text) == nil {
		return text
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(content, &blocks) != nil {
		return ""
	}
	var parts []string
	for _, block := range blocks {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		} else if block.Type == "image" || block.Type == "document" {
			parts = append(parts, "[Non-text user input; attachment not imported.]")
		}
	}
	return strings.Join(parts, "\n")
}
