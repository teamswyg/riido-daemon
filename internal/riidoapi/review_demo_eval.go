package riidoapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

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
		Consent: hostintegration.ConsentState{ReviewDemoMode: req.ReviewDemoConsentGranted},
	})
	if err != nil {
		return ReviewDemoResponse{}, err
	}
	return reviewDemoResponseFromMode(mode), nil
}
