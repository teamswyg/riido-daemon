package controlplane

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func (s *FileQueueSource) moveTaskToClaim(path, runtimeID string) (string, error) {
	claimsDir := filepath.Join(s.dir, "claims")
	if err := os.MkdirAll(claimsDir, 0o755); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "create claims dir")
	}
	claimPath, err := s.reserveClaimPath(claimsDir, runtimeID)
	if err != nil {
		return "", err
	}
	if err := os.Rename(path, claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "rename task to claim")
	}
	return claimPath, nil
}

func (s *FileQueueSource) reserveClaimPath(claimsDir, runtimeID string) (string, error) {
	runtimeHash := sha256.Sum256([]byte(runtimeID))
	tmp, err := os.CreateTemp(claimsDir, fmt.Sprintf("%020d-%x-*.json", s.now().UTC().UnixNano(), runtimeHash[:4]))
	if err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "reserve claim path")
	}
	claimPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(claimPath)
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "close claim path reservation")
	}
	if err := os.Remove(claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "release claim path reservation")
	}
	return claimPath, nil
}
