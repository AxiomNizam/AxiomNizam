package agents

import (
	"context"
	"fmt"

	"example.com/axiomnizam/internal/workflows"
)

// CloudAgentWorkflowHandler returns a workflows.StepHandler that delegates
// the step execution to a registered cloud agent.
//
// The step's Config must contain:
//
//	"agentId"  (string) – ID of the registered cloud agent to delegate to
//
// All other Config entries and the current step input are forwarded to the
// agent as task payload.
func CloudAgentWorkflowHandler(executor *Executor) workflows.StepHandler {
	return func(ctx context.Context, step *workflows.WorkflowStep, input map[string]interface{}) (map[string]interface{}, error) {
		agentID, ok := step.Config["agentId"].(string)
		if !ok || agentID == "" {
			return nil, fmt.Errorf("cloud-agent step requires 'agentId' in config")
		}

		// Build payload from config + step input
		payload := make(map[string]interface{})
		for k, v := range step.Config {
			if k != "agentId" {
				payload[k] = v
			}
		}
		for k, v := range input {
			payload[k] = v
		}

		task, err := executor.Delegate(ctx, agentID, payload)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"taskId":       task.ID,
			"remoteTaskId": task.RemoteTaskID,
			"status":       task.Status,
		}
		for k, v := range task.Result {
			result[k] = v
		}
		return result, nil
	}
}

// RegisterCloudAgentHandler registers the "cloud-agent" step type on the
// global workflow engine using the global executor.
func RegisterCloudAgentHandler() {
	workflows.GlobalWorkflowEngine.RegisterHandler("cloud-agent", CloudAgentWorkflowHandler(GlobalExecutor))
}
