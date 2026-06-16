package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (s Server) Serve(ctx context.Context) error {
	transport := normalizeLocalTransport(s.config.Transport)
	if s.config.SocketPath == "" {
		return errors.New("riido API socket path is empty")
	}
	if s.config.TaskDBPath == "" {
		return errors.New("riido task DB path is empty")
	}
	listener, cleanup, err := listenLocalEndpoint(transport, s.config.SocketPath)
	if err != nil {
		return fmt.Errorf("listen riido API %s endpoint: %w", transport, err)
	}
	defer func() {
		cleanup()
	}()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept riido API connection: %w", err)
		}
		go s.handleConn(ctx, conn)
	}
}

func (s Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	var req requestEnvelope
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = writeResponse(conn, responseEnvelope{OK: false, Error: fmt.Sprintf("decode request: %v", err)})
		return
	}
	response := s.handleRequest(ctx, req)
	_ = writeResponse(conn, response)
}

func (s Server) handleRequest(ctx context.Context, req requestEnvelope) responseEnvelope {
	switch req.Method {
	case "status":
		db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, statusFromDB(s.config, db))
	case "tasks":
		db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, db)
	case "transition":
		response, err := s.applyTransition(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "evidence":
		response, err := s.addEvidence(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "validate":
		response, err := s.validateTask(ctx, req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "review-demo":
		response, err := s.evaluateReviewDemo(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	default:
		return errorResponse(req.Method, fmt.Errorf("unknown method: %s", req.Method))
	}
}

func (s Server) evaluateReviewDemo(params json.RawMessage) (ReviewDemoResponse, error) {
	var req ReviewDemoRequest
	if len(params) == 0 {
		return ReviewDemoResponse{}, errors.New("review-demo params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return ReviewDemoResponse{}, fmt.Errorf("decode review-demo params: %w", err)
	}
	channel := hostintegration.DistributionChannel(strings.TrimSpace(req.DistributionChannel))
	if !channel.Valid() {
		return ReviewDemoResponse{}, fmt.Errorf("unknown distribution channel %q", req.DistributionChannel)
	}
	mode, err := hostintegration.EvaluateReviewDemoMode(hostintegration.ReviewDemoModeInput{
		Channel: channel,
		Consent: hostintegration.ConsentState{
			ReviewDemoMode: req.ReviewDemoConsentGranted,
		},
	})
	if err != nil {
		return ReviewDemoResponse{}, err
	}
	return reviewDemoResponseFromMode(mode), nil
}
