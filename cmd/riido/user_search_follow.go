package main

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// searchFollowReserve keeps one full server search request (bounded at 5s on
// the server) plus margin inside the command timeout before another
// continuation is requested.
const searchFollowReserve = 6 * time.Second

type workSearchFunc func(context.Context, userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error)

// followWorkSearch (RIID-6014) fills one CLI page from server pages that ended
// early at the server scan budget. A first-page error fails the command as
// before; a later error or the command deadline keeps the pages already read and
// returns their continuation cursor, so the result stays incomplete but honest.
func followWorkSearch(ctx context.Context, search workSearchFunc, q userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
	page, err := search(ctx, q)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, item := range page.Items {
		seen[item.Kind+":"+item.ID] = true
	}
	for !page.Complete && len(page.Items) < page.Limit {
		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < searchFollowReserve {
			break
		}
		next := q
		next.Cursor, next.Limit = *page.NextCursor, page.Limit-len(page.Items)
		more, err := search(ctx, next)
		if err != nil {
			break
		}
		for _, item := range more.Items {
			if seen[item.Kind+":"+item.ID] {
				return nil, userapi.ErrInvalidResponse
			}
			seen[item.Kind+":"+item.ID] = true
		}
		page.Items = append(page.Items, more.Items...)
		page.Complete, page.NextCursor = more.Complete, more.NextCursor
	}
	return page, nil
}
