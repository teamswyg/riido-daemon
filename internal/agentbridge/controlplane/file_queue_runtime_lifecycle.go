package controlplane

import (
	"context"
	"errors"
	"io/fs"
	"os"
)

func (s *FileQueueSource) RegisterRuntime(ctx context.Context, rt RuntimeRegistration) error {
	if err := fileQueueContextErr(ctx); err != nil {
		return err
	}
	if rt.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.register-runtime", "empty RuntimeID")
	}
	rec := RegisteredRuntime{
		RuntimeRegistration: rt,
		LastHeartbeat:       s.now().UTC(),
	}
	return s.writeRuntimeRecord(rec)
}

func (s *FileQueueSource) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	if err := fileQueueContextErr(ctx); err != nil {
		return err
	}
	if runtimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.deregister-runtime", "empty RuntimeID")
	}
	if err := os.Remove(s.runtimePath(runtimeID)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.deregister-runtime", err, "deregister runtime")
	}
	return nil
}

func (s *FileQueueSource) Heartbeat(ctx context.Context, hb RuntimeHeartbeat) error {
	if err := fileQueueContextErr(ctx); err != nil {
		return err
	}
	if hb.RuntimeID == "" {
		return controlPlaneErrorf(ErrControlPlaneRuntime, "file-queue.heartbeat", "empty RuntimeID")
	}
	body, err := os.ReadFile(s.runtimePath(hb.RuntimeID))
	if err != nil {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.heartbeat", err, "read runtime registry")
	}
	rec, err := parseRuntimeRecord(body)
	if err != nil {
		return err
	}
	rec.LastHeartbeat = s.now().UTC()
	applyHeartbeat(&rec.RuntimeRegistration, hb)
	return s.writeRuntimeRecord(rec)
}
