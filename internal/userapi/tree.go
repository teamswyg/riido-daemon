package userapi

import "context"

// This projection deliberately omits bodies and meetings. It uses the same
// authorized one-hop selection as Context, without Context's hydration reads.
const treeQuery = `query RiidoCLIWorkTree($componentId:String!,$teamId:String!){getComponentReadScope(componentId:$componentId,teamId:$teamId){schemaVersion rootId teamId status selection{childrenLimit} works{id kind teamId reasons{kind relation source} summary{id kind title keyNumber componentType}}}}`

type TreeNode struct {
	ID            string  `json:"id"`
	TeamID        string  `json:"teamId"`
	Title         string  `json:"title"`
	KeyNumber     *string `json:"keyNumber"`
	ComponentType *string `json:"componentType"`
}

type TreeAncestor struct {
	Relation string   `json:"relation"`
	Work     TreeNode `json:"work"`
}

// WorkTree describes roles relative to Root, not edges between ancestors.
// Children are the authorized selection; an empty list is not an absence proof.
type WorkTree struct {
	Root          TreeNode       `json:"root"`
	Ancestors     []TreeAncestor `json:"ancestors"`
	Children      []TreeNode     `json:"children"`
	ChildrenLimit int            `json:"childrenLimit"`
	MaxDepth      int            `json:"maxDepth"`
	ContextStatus string         `json:"contextStatus"`
}

type treeScope struct {
	SchemaVersion string              `json:"schemaVersion"`
	RootID        string              `json:"rootId"`
	TeamID        string              `json:"teamId"`
	Status        string              `json:"status"`
	Selection     *ReadScopeSelection `json:"selection"`
	Works         []*treeWork         `json:"works"`
}

type treeWork struct {
	ID      string             `json:"id"`
	Kind    string             `json:"kind"`
	TeamID  string             `json:"teamId"`
	Reasons []*ReadScopeReason `json:"reasons"`
	Summary *struct {
		ID            string  `json:"id"`
		Kind          string  `json:"kind"`
		Title         *string `json:"title"`
		KeyNumber     *string `json:"keyNumber"`
		ComponentType *string `json:"componentType"`
	} `json:"summary"`
}

func (r Reader) Tree(ctx context.Context, id, teamID string) (*WorkTree, error) {
	if id == "" || teamID == "" {
		return nil, ErrInvalidResponse
	}
	var data struct {
		Scope *treeScope `json:"getComponentReadScope"`
	}
	if err := r.read(ctx, treeQuery, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return nil, err
	}
	return projectTree(data.Scope, id, teamID)
}

func projectTree(scope *treeScope, id, teamID string) (*WorkTree, error) {
	if scope == nil || scope.SchemaVersion != ReadScopeSchemaVersion || scope.RootID != id || scope.TeamID != teamID || scope.Selection == nil || scope.Selection.ChildrenLimit != 10 || scope.Works == nil || (scope.Status != "available" && scope.Status != "partial") {
		return nil, ErrInvalidResponse
	}
	tree := &WorkTree{Ancestors: []TreeAncestor{}, Children: []TreeNode{}, ChildrenLimit: scope.Selection.ChildrenLimit, MaxDepth: 1, ContextStatus: scope.Status}
	seen, ancestors := map[string]bool{}, map[string]TreeNode{}
	for _, row := range scope.Works {
		if !validTreeWork(row) || seen[row.Kind+":"+row.ID] {
			return nil, ErrInvalidResponse
		}
		seen[row.Kind+":"+row.ID] = true
		node := TreeNode{ID: row.ID, TeamID: row.TeamID, Title: *row.Summary.Title, KeyNumber: row.Summary.KeyNumber, ComponentType: row.Summary.ComponentType}
		roles := map[string]bool{}
		for _, reason := range row.Reasons {
			if roles[reason.Kind] {
				continue
			}
			roles[reason.Kind] = true
			switch reason.Kind {
			case "root":
				if row.Kind != "component" || row.ID != id || row.TeamID != teamID || tree.Root.ID != "" {
					return nil, ErrInvalidResponse
				}
				tree.Root = node
			case "project", "milestone", "parentTask":
				if row.Kind != "component" || row.ID == id || ancestors[reason.Kind].ID != "" {
					return nil, ErrInvalidResponse
				}
				ancestors[reason.Kind] = node
			case "child":
				if row.Kind != "component" || row.ID == id || len(tree.Children) >= tree.ChildrenLimit {
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
		if node, ok := ancestors[role]; ok {
			tree.Ancestors = append(tree.Ancestors, TreeAncestor{Relation: role, Work: node})
		}
	}
	return tree, nil
}

func validTreeWork(row *treeWork) bool {
	if row == nil || row.ID == "" || row.TeamID == "" || !validScopeReasons(row.Reasons) || row.Summary == nil || row.Summary.ID != row.ID || row.Summary.Kind != row.Kind || row.Summary.Title == nil {
		return false
	}
	return row.Kind == "component" || row.Kind == "backlog" || row.Kind == "intake"
}
