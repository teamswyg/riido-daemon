package userapi

import "context"

// Keep the lightweight one-hop projection; do not hydrate bodies or meetings.
const treeV2Query = `query RiidoCLIWorkTree($componentId:String!,$workspaceId:String!,$teamId:String){getComponentReadScope(componentId:$componentId,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2){schemaVersion workspaceId rootId teamId status selection{childrenLimit} works{id kind workspaceId teamId reasons{kind relation source} summary{id kind title keyNumber componentType}}}}`

type TreeNodeV2 struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspaceId"`
	TeamID        *string `json:"teamId"`
	Kind          string  `json:"kind"`
	Title         string  `json:"title"`
	KeyNumber     *string `json:"keyNumber"`
	ComponentType *string `json:"componentType"`
}
type TreeAncestorV2 struct {
	Relation string     `json:"relation"`
	Work     TreeNodeV2 `json:"work"`
}
type WorkTreeV2 struct {
	Root          TreeNodeV2       `json:"root"`
	Ancestors     []TreeAncestorV2 `json:"ancestors"`
	Children      []TreeNodeV2     `json:"children"`
	ChildrenLimit int              `json:"childrenLimit"`
	MaxDepth      int              `json:"maxDepth"`
	ContextStatus string           `json:"contextStatus"`
}
type treeWorkV2 struct {
	treeWork
	WorkspaceID string  `json:"workspaceId"`
	TeamID      *string `json:"teamId"`
}

func (v *treeWorkV2) UnmarshalJSON(raw []byte) error {
	type plain treeWorkV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

type treeScopeV2 struct {
	treeScope
	WorkspaceID string        `json:"workspaceId"`
	TeamID      *string       `json:"teamId"`
	Works       []*treeWorkV2 `json:"works"`
}

func (v *treeScopeV2) UnmarshalJSON(raw []byte) error {
	type plain treeScopeV2
	return decodeV2Locator(raw, (*plain)(v), "teamId")
}

func (r Reader) TreeV2(ctx context.Context, locator WorkLocator) (*WorkTreeV2, error) {
	if !locator.valid() {
		return nil, ErrInvalidRequest
	}
	var data struct {
		Scope *treeScopeV2 `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, treeV2Query, locator.variables(), &data); err != nil {
		return nil, err
	}
	s := data.Scope
	if s == nil || s.SchemaVersion != "riido-read-scope.v2" || s.RootID != locator.ID || s.WorkspaceID != locator.WorkspaceID || !sameLocatorTeam(s.TeamID, locator.TeamID) || s.Selection == nil || s.Selection.ChildrenLimit != 10 || s.Works == nil || (s.Status != "available" && s.Status != "partial") {
		return nil, ErrInvalidResponse
	}
	tree := &WorkTreeV2{Ancestors: []TreeAncestorV2{}, Children: []TreeNodeV2{}, ChildrenLimit: 10, MaxDepth: 1, ContextStatus: s.Status}
	seen, ancestors := map[string]bool{}, map[string]TreeNodeV2{}
	for _, row := range s.Works {
		if row == nil || row.ID == "" || row.WorkspaceID != locator.WorkspaceID || seen[row.ID] || !validScopeReasons(row.Reasons) || row.Summary == nil || row.Summary.ID != row.ID || row.Summary.Kind != row.Kind || row.Summary.Title == nil {
			return nil, ErrInvalidResponse
		}
		typ := "task"
		if row.Kind == "component" {
			if row.Summary.ComponentType == nil {
				return nil, ErrInvalidResponse
			}
			typ = *row.Summary.ComponentType
		} else if row.Kind != "backlog" && row.Kind != "intake" {
			return nil, ErrInvalidResponse
		}
		if !validLocatorTeam(2, row.Kind, typ, row.TeamID) {
			return nil, ErrInvalidResponse
		}
		seen[row.ID] = true
		node := TreeNodeV2{ID: row.ID, WorkspaceID: row.WorkspaceID, TeamID: row.TeamID, Kind: row.Kind, Title: *row.Summary.Title, KeyNumber: row.Summary.KeyNumber, ComponentType: row.Summary.ComponentType}
		roles := map[string]bool{}
		for _, reason := range row.Reasons {
			if roles[reason.Kind] {
				continue
			}
			roles[reason.Kind] = true
			switch reason.Kind {
			case "root":
				if row.ID != locator.ID || !sameLocatorTeam(row.TeamID, locator.TeamID) || (row.Kind == "component" && typ != locator.ComponentType) || tree.Root.ID != "" {
					return nil, ErrInvalidResponse
				}
				tree.Root = node
			case "project", "milestone", "parentTask":
				if row.Kind != "component" || row.ID == locator.ID || ancestors[reason.Kind].ID != "" || (reason.Kind == "project" && typ != "project") || (reason.Kind == "milestone" && typ != "milestone") || (reason.Kind == "parentTask" && typ != "task" && typ != "subtask") {
					return nil, ErrInvalidResponse
				}
				ancestors[reason.Kind] = node
			case "child":
				if row.ID == locator.ID || len(tree.Children) >= tree.ChildrenLimit {
					return nil, ErrInvalidResponse
				}
				tree.Children = append(tree.Children, node)
			}
		}
	}
	if tree.Root.ID == "" {
		return nil, ErrInvalidResponse
	}
	for _, role := range []string{"project", "milestone", "parentTask"} {
		if n, ok := ancestors[role]; ok {
			tree.Ancestors = append(tree.Ancestors, TreeAncestorV2{Relation: role, Work: n})
		}
	}
	return tree, nil
}
