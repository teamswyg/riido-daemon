package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func runAPIReviewDemo(args []string, config apiCLIConfig) error {
	request := riidoapi.ReviewDemoRequest{}
	for index := 0; index < len(args); index++ {
		handled, err := parseAPIConnectionFlag(args, &index, &config)
		if err != nil {
			if isCLIHelp(err) {
				return nil
			}
			return err
		}
		if handled {
			continue
		}
		switch args[index] {
		case "--channel":
			request.DistributionChannel, err = cliRequiredArg(args, &index, "--channel", "value")
		case "--review-demo-consent-granted":
			request.ReviewDemoConsentGranted, err = cliRequiredBool(args, &index, "--review-demo-consent-granted")
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
		if err != nil {
			return err
		}
	}
	if request.DistributionChannel == "" {
		return fmt.Errorf("--channel is required")
	}
	var response riidoapi.ReviewDemoResponse
	if err := requestAPI(config, 5*time.Second, riidoapi.MethodReviewDemo, request, &response); err != nil {
		return err
	}
	return printJSON(response)
}
