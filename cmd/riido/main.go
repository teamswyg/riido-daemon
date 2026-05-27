package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "riido:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing command")
	}
	if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}
	switch args[0] {
	case "mwsd":
		return runMwsd(args[1:])
	case "task":
		return runTask(args[1:])
	case "serve":
		return runServe(args[1:])
	case "api":
		return runAPI(args[1:])
	case "bridge":
		return runBridge(args[1:])
	case "daemon":
		return runDaemon(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runTask(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing task command")
	}
	command := args[0]
	taskDBPath, err := taskdb.DefaultTaskDBPath()
	if err != nil {
		return err
	}

	switch command {
	case "list":
		for index := 1; index < len(args); index++ {
			switch args[index] {
			case "--task-db":
				index++
				if index >= len(args) {
					return fmt.Errorf("--task-db requires a path")
				}
				taskDBPath = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		db, err := taskdb.LoadTaskDB(taskDBPath)
		if err != nil {
			return err
		}
		return printJSON(db)
	case "transition":
		if len(args) < 2 {
			return fmt.Errorf("task transition requires a task id")
		}
		taskID := args[1]
		var toState string
		var eventType string
		actor := "human"
		source := "riido-cli"
		var reason string
		var provider string
		var decisionLLM string
		var approvalID string
		var commandID string
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--task-db":
				index++
				if index >= len(args) {
					return fmt.Errorf("--task-db requires a path")
				}
				taskDBPath = args[index]
			case "--to":
				index++
				if index >= len(args) {
					return fmt.Errorf("--to requires a state")
				}
				toState = args[index]
			case "--event":
				index++
				if index >= len(args) {
					return fmt.Errorf("--event requires an event type")
				}
				eventType = args[index]
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				source = args[index]
			case "--reason":
				index++
				if index >= len(args) {
					return fmt.Errorf("--reason requires a value")
				}
				reason = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				decisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				approvalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				commandID = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if toState == "" {
			return fmt.Errorf("--to is required")
		}
		if eventType == "" {
			return fmt.Errorf("--event is required")
		}
		to, err := taskdb.ParseTaskState(toState)
		if err != nil {
			return err
		}
		db, err := taskdb.LoadTaskDB(taskDBPath)
		if err != nil {
			return err
		}
		updated, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
			TaskID:  taskID,
			ToState: to,
			Event:   ir.EventType(eventType),
			Actor:   actor,
			Source:  source,
			Reason:  reason,
			Guard: taskdb.TaskMutationGuardInput{
				CommandID:   commandID,
				Provider:    provider,
				DecisionLLM: decisionLLM,
				ApprovalID:  approvalID,
			},
		}, time.Now())
		if err != nil {
			return err
		}
		if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
			return err
		}
		return printJSON(struct {
			OK         bool                            `json:"ok"`
			TaskDBPath string                          `json:"task_db_path"`
			Transition taskdb.TaskTransitionRecord     `json:"transition"`
			Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
		}{
			OK:         true,
			TaskDBPath: taskDBPath,
			Transition: transition,
			Receipt:    receipt,
		})
	case "evidence":
		if len(args) < 2 {
			return fmt.Errorf("task evidence requires a task id")
		}
		taskID := args[1]
		var command string
		var exitCode int
		var result string
		actor := "daemon"
		source := "riido-cli"
		var summary string
		var provider string
		var decisionLLM string
		var approvalID string
		var commandID string
		var validationGate string
		var providerRunID string
		var providerRunResult string
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--task-db":
				index++
				if index >= len(args) {
					return fmt.Errorf("--task-db requires a path")
				}
				taskDBPath = args[index]
			case "--command":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command requires a value")
				}
				command = args[index]
			case "--exit-code":
				index++
				if index >= len(args) {
					return fmt.Errorf("--exit-code requires a value")
				}
				value, err := strconv.Atoi(args[index])
				if err != nil {
					return fmt.Errorf("--exit-code must be an integer: %w", err)
				}
				exitCode = value
			case "--result":
				index++
				if index >= len(args) {
					return fmt.Errorf("--result requires a value")
				}
				result = args[index]
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				source = args[index]
			case "--summary":
				index++
				if index >= len(args) {
					return fmt.Errorf("--summary requires a value")
				}
				summary = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				decisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				approvalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				commandID = args[index]
			case "--validation-gate":
				index++
				if index >= len(args) {
					return fmt.Errorf("--validation-gate requires a value")
				}
				validationGate = args[index]
			case "--provider-run-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider-run-id requires a value")
				}
				providerRunID = args[index]
			case "--provider-run-result":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider-run-result requires a value")
				}
				providerRunResult = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if command == "" {
			return fmt.Errorf("--command is required")
		}
		db, err := taskdb.LoadTaskDB(taskDBPath)
		if err != nil {
			return err
		}
		updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
			TaskID:            taskID,
			Command:           command,
			ExitCode:          exitCode,
			Result:            result,
			Actor:             actor,
			Source:            source,
			Summary:           summary,
			ValidationGate:    validationGate,
			ProviderRunID:     providerRunID,
			ProviderRunResult: providerRunResult,
			Guard: taskdb.TaskMutationGuardInput{
				CommandID:   commandID,
				Provider:    provider,
				DecisionLLM: decisionLLM,
				ApprovalID:  approvalID,
			},
		}, time.Now())
		if err != nil {
			return err
		}
		if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
			return err
		}
		return printJSON(struct {
			OK         bool                            `json:"ok"`
			TaskDBPath string                          `json:"task_db_path"`
			Evidence   taskdb.TaskEvidenceRecord       `json:"evidence"`
			Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
		}{
			OK:         true,
			TaskDBPath: taskDBPath,
			Evidence:   evidence,
			Receipt:    receipt,
		})
	case "validate":
		if len(args) < 2 {
			return fmt.Errorf("task validate requires a task id")
		}
		taskID := args[1]
		var command string
		var workdir string
		var timeout time.Duration
		actor := "daemon"
		source := "riido-validation-runner"
		var summary string
		var provider string
		var decisionLLM string
		var approvalID string
		var commandID string
		var validationGate string
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--task-db":
				index++
				if index >= len(args) {
					return fmt.Errorf("--task-db requires a path")
				}
				taskDBPath = args[index]
			case "--command":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command requires a value")
				}
				command = args[index]
			case "--workdir":
				index++
				if index >= len(args) {
					return fmt.Errorf("--workdir requires a path")
				}
				workdir = args[index]
			case "--timeout-seconds":
				index++
				if index >= len(args) {
					return fmt.Errorf("--timeout-seconds requires a value")
				}
				value, err := strconv.Atoi(args[index])
				if err != nil {
					return fmt.Errorf("--timeout-seconds must be an integer: %w", err)
				}
				if value <= 0 {
					return fmt.Errorf("--timeout-seconds must be positive")
				}
				timeout = time.Duration(value) * time.Second
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				source = args[index]
			case "--summary":
				index++
				if index >= len(args) {
					return fmt.Errorf("--summary requires a value")
				}
				summary = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				decisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				approvalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				commandID = args[index]
			case "--validation-gate":
				index++
				if index >= len(args) {
					return fmt.Errorf("--validation-gate requires a value")
				}
				validationGate = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if command == "" {
			return fmt.Errorf("--command is required")
		}
		if approvalID == "" {
			return fmt.Errorf("--approval-id is required before validation command execution")
		}
		db, err := taskdb.LoadTaskDB(taskDBPath)
		if err != nil {
			return err
		}
		providerForRun, err := validationProviderForTask(db, taskID, provider)
		if err != nil {
			return err
		}
		if err := validateDecisionLLMForTask(db, taskID, decisionLLM); err != nil {
			return err
		}
		taskBeforeValidation, ok := findTaskRecord(db, taskID)
		if !ok {
			return fmt.Errorf("task %s not found", taskID)
		}
		now := time.Now()
		if commandID == "" {
			commandID = validationCommandID(taskID, now)
		}
		result, err := validation.RunCommand(context.Background(), validation.CommandRequest{
			Command:        command,
			Workdir:        workdir,
			Timeout:        timeout,
			CommandID:      commandID,
			Provider:       providerForRun,
			ValidationGate: validationGate,
			Summary:        summary,
		}, now)
		if err != nil {
			return err
		}
		updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
			TaskID:            taskID,
			Command:           result.Command,
			ExitCode:          result.ExitCode,
			Result:            result.Result,
			Actor:             actor,
			Source:            source,
			Summary:           result.Summary,
			ValidationGate:    result.ValidationGate,
			ProviderRunID:     result.ProviderRunID,
			ProviderRunResult: result.ProviderRunResult,
			Guard: taskdb.TaskMutationGuardInput{
				CommandID:   commandID,
				Provider:    providerForRun,
				DecisionLLM: decisionLLM,
				ApprovalID:  approvalID,
			},
		}, now)
		if err != nil {
			return err
		}
		var transition *taskdb.TaskTransitionRecord
		var transitionReceipt *taskdb.TaskCommandReceiptRecord
		if taskBeforeValidation.State == task.StateValidating {
			toState, eventType := validationTransitionForResult(result.Result)
			nextDB, appliedTransition, appliedReceipt, err := taskdb.ApplyGuardedTaskTransition(updated, taskdb.TaskTransitionInput{
				TaskID:  taskID,
				ToState: toState,
				Event:   eventType,
				Actor:   actor,
				Source:  source,
				Reason:  fmt.Sprintf("validation %s via %s", result.Result, result.ValidationGate),
				Guard: taskdb.TaskMutationGuardInput{
					CommandID:   commandID + ":transition",
					Provider:    providerForRun,
					DecisionLLM: decisionLLM,
					ApprovalID:  approvalID,
				},
			}, now)
			if err != nil {
				return err
			}
			updated = nextDB
			transition = &appliedTransition
			transitionReceipt = &appliedReceipt
		}
		if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
			return err
		}
		return printJSON(struct {
			OK                bool                             `json:"ok"`
			TaskDBPath        string                           `json:"task_db_path"`
			Validation        validation.CommandResult         `json:"validation"`
			Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
			Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
			Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
			TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
		}{
			OK:                evidence.Result == "passed",
			TaskDBPath:        taskDBPath,
			Validation:        result,
			Evidence:          evidence,
			Receipt:           receipt,
			Transition:        transition,
			TransitionReceipt: transitionReceipt,
		})
	default:
		printUsage()
		return fmt.Errorf("unknown task command: %s", command)
	}
}

func runServe(args []string) error {
	socketPath, err := riidoapi.DefaultSocketPath()
	if err != nil {
		return err
	}
	transport := riidoapi.LocalTransportUnixSocket
	taskDBPath, err := taskdb.DefaultTaskDBPath()
	if err != nil {
		return err
	}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--socket":
			index++
			if index >= len(args) {
				return fmt.Errorf("--socket requires a path")
			}
			socketPath = args[index]
		case "--transport":
			index++
			if index >= len(args) {
				return fmt.Errorf("--transport requires a value")
			}
			transport = riidoapi.LocalTransport(args[index])
		case "--task-db":
			index++
			if index >= len(args) {
				return fmt.Errorf("--task-db requires a path")
			}
			taskDBPath = args[index]
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fmt.Fprintf(os.Stderr, "riido serve transport=%s socket=%s task_db=%s\n", transport, socketPath, taskDBPath)
	return riidoapi.NewServer(riidoapi.Config{SocketPath: socketPath, TaskDBPath: taskDBPath, Transport: transport}).Serve(ctx)
}

func runAPI(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing api command")
	}
	command := args[0]
	socketPath, err := riidoapi.DefaultSocketPath()
	if err != nil {
		return err
	}
	transport := riidoapi.LocalTransportUnixSocket
	parseSocket := func(start int) (int, error) {
		for index := start; index < len(args); index++ {
			switch args[index] {
			case "--socket":
				index++
				if index >= len(args) {
					return index, fmt.Errorf("--socket requires a path")
				}
				socketPath = args[index]
			case "--transport":
				index++
				if index >= len(args) {
					return index, fmt.Errorf("--transport requires a value")
				}
				transport = riidoapi.LocalTransport(args[index])
			case "--help", "-h":
				printUsage()
				return index, nil
			default:
				return index, fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		return len(args), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch command {
	case "status":
		if _, err := parseSocket(1); err != nil {
			return err
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		var status riidoapi.Status
		if err := client.Request(ctx, "status", nil, &status); err != nil {
			return err
		}
		return printJSON(status)
	case "tasks":
		if _, err := parseSocket(1); err != nil {
			return err
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		var db taskdb.TaskDB
		if err := client.Request(ctx, "tasks", nil, &db); err != nil {
			return err
		}
		return printJSON(db)
	case "review-demo":
		request := riidoapi.ReviewDemoRequest{}
		for index := 1; index < len(args); index++ {
			switch args[index] {
			case "--socket":
				index++
				if index >= len(args) {
					return fmt.Errorf("--socket requires a path")
				}
				socketPath = args[index]
			case "--transport":
				index++
				if index >= len(args) {
					return fmt.Errorf("--transport requires a value")
				}
				transport = riidoapi.LocalTransport(args[index])
			case "--channel":
				index++
				if index >= len(args) {
					return fmt.Errorf("--channel requires a value")
				}
				request.DistributionChannel = args[index]
			case "--review-demo-consent-granted":
				index++
				if index >= len(args) {
					return fmt.Errorf("--review-demo-consent-granted requires a boolean")
				}
				value, err := strconv.ParseBool(args[index])
				if err != nil {
					return fmt.Errorf("--review-demo-consent-granted must be a boolean: %w", err)
				}
				request.ReviewDemoConsentGranted = value
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if request.DistributionChannel == "" {
			return fmt.Errorf("--channel is required")
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		var response riidoapi.ReviewDemoResponse
		if err := client.Request(ctx, "review-demo", request, &response); err != nil {
			return err
		}
		return printJSON(response)
	case "transition":
		if len(args) < 2 {
			return fmt.Errorf("api transition requires a task id")
		}
		request := riidoapi.TransitionRequest{
			TaskID: args[1],
			Actor:  "human",
			Source: "riido-api-cli",
		}
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--socket":
				index++
				if index >= len(args) {
					return fmt.Errorf("--socket requires a path")
				}
				socketPath = args[index]
			case "--transport":
				index++
				if index >= len(args) {
					return fmt.Errorf("--transport requires a value")
				}
				transport = riidoapi.LocalTransport(args[index])
			case "--to":
				index++
				if index >= len(args) {
					return fmt.Errorf("--to requires a state")
				}
				request.ToState = args[index]
			case "--event":
				index++
				if index >= len(args) {
					return fmt.Errorf("--event requires an event type")
				}
				request.EventType = args[index]
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				request.Actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				request.Source = args[index]
			case "--reason":
				index++
				if index >= len(args) {
					return fmt.Errorf("--reason requires a value")
				}
				request.Reason = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				request.Provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				request.DecisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				request.ApprovalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				request.CommandID = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if request.ToState == "" {
			return fmt.Errorf("--to is required")
		}
		if request.EventType == "" {
			return fmt.Errorf("--event is required")
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		var response riidoapi.TransitionResponse
		if err := client.Request(ctx, "transition", request, &response); err != nil {
			return err
		}
		return printJSON(response)
	case "evidence":
		if len(args) < 2 {
			return fmt.Errorf("api evidence requires a task id")
		}
		request := riidoapi.EvidenceRequest{
			TaskID: args[1],
			Actor:  "daemon",
			Source: "riido-api-cli",
		}
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--socket":
				index++
				if index >= len(args) {
					return fmt.Errorf("--socket requires a path")
				}
				socketPath = args[index]
			case "--transport":
				index++
				if index >= len(args) {
					return fmt.Errorf("--transport requires a value")
				}
				transport = riidoapi.LocalTransport(args[index])
			case "--command":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command requires a value")
				}
				request.Command = args[index]
			case "--exit-code":
				index++
				if index >= len(args) {
					return fmt.Errorf("--exit-code requires a value")
				}
				value, err := strconv.Atoi(args[index])
				if err != nil {
					return fmt.Errorf("--exit-code must be an integer: %w", err)
				}
				request.ExitCode = value
			case "--result":
				index++
				if index >= len(args) {
					return fmt.Errorf("--result requires a value")
				}
				request.Result = args[index]
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				request.Actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				request.Source = args[index]
			case "--summary":
				index++
				if index >= len(args) {
					return fmt.Errorf("--summary requires a value")
				}
				request.Summary = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				request.Provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				request.DecisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				request.ApprovalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				request.CommandID = args[index]
			case "--validation-gate":
				index++
				if index >= len(args) {
					return fmt.Errorf("--validation-gate requires a value")
				}
				request.ValidationGate = args[index]
			case "--provider-run-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider-run-id requires a value")
				}
				request.ProviderRunID = args[index]
			case "--provider-run-result":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider-run-result requires a value")
				}
				request.ProviderRunResult = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if request.Command == "" {
			return fmt.Errorf("--command is required")
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		var response riidoapi.EvidenceResponse
		if err := client.Request(ctx, "evidence", request, &response); err != nil {
			return err
		}
		return printJSON(response)
	case "validate":
		if len(args) < 2 {
			return fmt.Errorf("api validate requires a task id")
		}
		request := riidoapi.ValidateRequest{
			TaskID: args[1],
			Actor:  "daemon",
			Source: "riido-api-cli",
		}
		for index := 2; index < len(args); index++ {
			switch args[index] {
			case "--socket":
				index++
				if index >= len(args) {
					return fmt.Errorf("--socket requires a path")
				}
				socketPath = args[index]
			case "--transport":
				index++
				if index >= len(args) {
					return fmt.Errorf("--transport requires a value")
				}
				transport = riidoapi.LocalTransport(args[index])
			case "--command":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command requires a value")
				}
				request.Command = args[index]
			case "--workdir":
				index++
				if index >= len(args) {
					return fmt.Errorf("--workdir requires a path")
				}
				request.Workdir = args[index]
			case "--timeout-seconds":
				index++
				if index >= len(args) {
					return fmt.Errorf("--timeout-seconds requires a value")
				}
				value, err := strconv.Atoi(args[index])
				if err != nil {
					return fmt.Errorf("--timeout-seconds must be an integer: %w", err)
				}
				if value <= 0 {
					return fmt.Errorf("--timeout-seconds must be positive")
				}
				request.TimeoutSeconds = value
			case "--actor":
				index++
				if index >= len(args) {
					return fmt.Errorf("--actor requires a value")
				}
				request.Actor = args[index]
			case "--source":
				index++
				if index >= len(args) {
					return fmt.Errorf("--source requires a value")
				}
				request.Source = args[index]
			case "--summary":
				index++
				if index >= len(args) {
					return fmt.Errorf("--summary requires a value")
				}
				request.Summary = args[index]
			case "--provider":
				index++
				if index >= len(args) {
					return fmt.Errorf("--provider requires a value")
				}
				request.Provider = args[index]
			case "--decision-llm":
				index++
				if index >= len(args) {
					return fmt.Errorf("--decision-llm requires a value")
				}
				request.DecisionLLM = args[index]
			case "--approval-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--approval-id requires a value")
				}
				request.ApprovalID = args[index]
			case "--command-id":
				index++
				if index >= len(args) {
					return fmt.Errorf("--command-id requires a value")
				}
				request.CommandID = args[index]
			case "--validation-gate":
				index++
				if index >= len(args) {
					return fmt.Errorf("--validation-gate requires a value")
				}
				request.ValidationGate = args[index]
			case "--help", "-h":
				printUsage()
				return nil
			default:
				return fmt.Errorf("unknown argument: %s", args[index])
			}
		}
		if request.Command == "" {
			return fmt.Errorf("--command is required")
		}
		if request.ApprovalID == "" {
			return fmt.Errorf("--approval-id is required before validation command execution")
		}
		client := riidoapi.NewClientWithTransport(transport, socketPath)
		client.Timeout = validation.DefaultTimeout + 5*time.Second
		if request.TimeoutSeconds > 0 {
			client.Timeout = time.Duration(request.TimeoutSeconds)*time.Second + 5*time.Second
		}
		validateCtx, validateCancel := context.WithTimeout(context.Background(), client.Timeout)
		defer validateCancel()
		var response riidoapi.ValidateResponse
		if err := client.Request(validateCtx, "validate", request, &response); err != nil {
			return err
		}
		return printJSON(response)
	default:
		printUsage()
		return fmt.Errorf("unknown api command: %s", command)
	}
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func validationProviderForTask(db taskdb.TaskDB, taskID string, requested string) (string, error) {
	taskRecord, ok := findTaskRecord(db, taskID)
	if !ok {
		return "", fmt.Errorf("task %s not found", taskID)
	}
	provider := requested
	if provider == "" {
		provider = taskRecord.RecommendedProvider
	}
	if provider == "" {
		provider = db.RecommendedProvider
	}
	if provider == "" {
		return "", fmt.Errorf("task %s has no validation provider", taskID)
	}
	if !providerAvailable(db.ProviderCandidates, provider) {
		return "", fmt.Errorf("provider %s is not an available orchestration candidate for task %s", provider, taskID)
	}
	return provider, nil
}

func validateDecisionLLMForTask(db taskdb.TaskDB, taskID string, requested string) error {
	if requested == "" {
		return nil
	}
	taskRecord, ok := findTaskRecord(db, taskID)
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	recommended := taskRecord.RecommendedDecisionLLM
	if recommended == "" {
		recommended = db.RecommendedDecisionLLM
	}
	if recommended != "" && requested != recommended {
		return fmt.Errorf("decision LLM %s does not match recommended decision LLM %s for task %s", requested, recommended, taskID)
	}
	return nil
}

func findTaskRecord(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func providerAvailable(candidates []taskdb.ProviderCandidate, provider string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}

func validationCommandID(taskID string, now time.Time) string {
	return fmt.Sprintf("command:validation:%s:%s", taskID, now.UTC().Format("20060102T150405.000000000Z"))
}

func validationTransitionForResult(result string) (task.TaskState, ir.EventType) {
	if result == "passed" {
		return task.StatePatchReady, ir.EventValidationPassed
	}
	return task.StateFailed, ir.EventValidationFailed
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  riido serve [--socket PATH] [--transport unix-socket|windows-named-pipe] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido api <status|tasks> [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api review-demo --channel CHANNEL --review-demo-consent-granted true|false [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido api validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--socket PATH] [--transport unix-socket|windows-named-pipe]")
	fmt.Fprintln(os.Stderr, "  riido mwsd <snapshot|projection|sync|orchestration|projects|status> [--socket PATH] [--state PATH] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task list [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task transition <task-id> --to STATE --event EVENT --approval-id ID [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--actor ACTOR] [--source SOURCE] [--reason TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task evidence <task-id> --command COMMAND --approval-id ID [--exit-code N] [--result RESULT] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--provider-run-id ID] [--provider-run-result RESULT] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido task validate <task-id> --command COMMAND --approval-id ID [--workdir PATH] [--timeout-seconds N] [--provider PROVIDER] [--decision-llm LLM] [--command-id ID] [--validation-gate GATE] [--actor ACTOR] [--source SOURCE] [--summary TEXT] [--task-db PATH]")
	fmt.Fprintln(os.Stderr, "  riido bridge <providers|detect>")
	fmt.Fprintln(os.Stderr, "  riido daemon start [--foreground] [--socket PATH] [--pid-file PATH] [--log-file PATH] [--lock-file PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon status [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon health [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon ready [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon metrics [--socket PATH]")
	fmt.Fprintln(os.Stderr, "  riido daemon stop [--socket PATH] [--pid-file PATH] [--timeout-seconds N]")
	fmt.Fprintln(os.Stderr, "  riido daemon logs --log-file PATH [--lines N]")
}
