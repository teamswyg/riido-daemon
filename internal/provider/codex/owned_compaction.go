package codex

type ownedCompaction struct {
	activeID  string
	seen      map[string]struct{}
	graceUsed bool
}

func (c *ownedCompaction) observe(method codexMethod, item map[string]any) bool {
	id := stringField(item, "id")
	if id == "" || stringField(item, "type") != "contextCompaction" {
		return false
	}
	if method == codexMethodItemStarted && c.activeID == "" {
		if _, duplicate := c.seen[id]; duplicate {
			return false
		}
		if c.seen == nil {
			c.seen = make(map[string]struct{})
		}
		c.seen[id] = struct{}{}
		c.activeID = id
		return true
	}
	if method == codexMethodItemCompleted && id == c.activeID {
		c.activeID = ""
		return true
	}
	return false
}

func (d *protocolDriver) ConsumeIdleGrace() bool {
	c := &d.owned.compaction
	if d.owned.done || c.activeID == "" || c.graceUsed {
		return false
	}
	c.graceUsed = true
	return true
}
