package riidoapi

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (s Server) handleRequest(ctx context.Context, req requestEnvelope) responseEnvelope {
	switch req.Method {
	case MethodStatus:
		return s.statusResponse(req.Method)
	case MethodTasks:
		return s.tasksResponse(req.Method)
	case MethodTransition:
		response, err := s.applyTransition(req.Params)
		return commandResponse(req.Method, response, err)
	case MethodEvidence:
		response, err := s.addEvidence(req.Params)
		return commandResponse(req.Method, response, err)
	case MethodValidate:
		response, err := s.validateTask(ctx, req.Params)
		return commandResponse(req.Method, response, err)
	case MethodReviewDemo:
		response, err := s.evaluateReviewDemo(req.Params)
		return commandResponse(req.Method, response, err)
	default:
		return errorResponse(req.Method, fmt.Errorf("unknown method: %s", req.Method))
	}
}

func (s Server) statusResponse(method Method) responseEnvelope {
	db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
	if err != nil {
		return errorResponse(method, err)
	}
	return okResponse(method, statusFromDB(s.config, db))
}

func (s Server) tasksResponse(method Method) responseEnvelope {
	db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
	if err != nil {
		return errorResponse(method, err)
	}
	return okResponse(method, db)
}

func commandResponse[T any](method Method, response T, err error) responseEnvelope {
	if err != nil {
		return errorResponse(method, err)
	}
	return okResponse(method, response)
}
