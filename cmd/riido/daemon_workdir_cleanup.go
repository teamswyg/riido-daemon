package main

import (
	"context"
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func startWorkdirCleanupLoop(ctx lifecycle.Context, cleaner workdir.Cleaner, settings daemonSettings, log logging.Logger) func() {
	if settings.WorkdirRetention <= 0 {
		return func() {}
	}
	if settings.WorkdirCleanupEvery <= 0 {
		settings.WorkdirCleanupEvery = time.Hour
	}
	cleanupCtx, cancel := lifecycle.WithCancel(ctx)
	runCleanup := func() {
		cutoff := time.Now().UTC().Add(-settings.WorkdirRetention)
		result, err := cleaner.CleanupArchivedBefore(cleanupCtx.Context(), workdir.CleanupRequest{ArchivedBefore: cutoff})
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("workdir cleanup error: %v", err)
			}
			return
		}
		if len(result.Removed) > 0 {
			log.Printf("workdir cleanup removed=%d scanned=%d retention=%s", len(result.Removed), result.ScannedArchiveRecords, settings.WorkdirRetention)
		}
	}
	runCleanup()
	go func() {
		ticker := time.NewTicker(settings.WorkdirCleanupEvery)
		defer ticker.Stop()
		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				runCleanup()
			}
		}
	}()
	return cancel
}
