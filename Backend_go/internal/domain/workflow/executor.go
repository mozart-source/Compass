package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

// DefaultWorkflowExecutor is the standard implementation of WorkflowExecutor
type DefaultWorkflowExecutor struct {
	repo         Repository
	logger       *logrus.Logger
	notifier     notification.DomainNotifier
	rolesService roles.Service
}

// NewDefaultExecutor creates a new workflow executor
func NewDefaultExecutor(repo Repository, logger *logrus.Logger, notifier notification.DomainNotifier, rolesService roles.Service) *DefaultWorkflowExecutor {
	return &DefaultWorkflowExecutor{
		repo:         repo,
		logger:       logger,
		notifier:     notifier,
		rolesService: rolesService,
	}
}

// ExecuteStep handles the execution of a workflow step
func (e *DefaultWorkflowExecutor) ExecuteStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithFields(logrus.Fields{
		"step_id":      step.ID,
		"execution_id": execution.ExecutionID,
		"step_type":    step.StepType,
	}).Info("Executing workflow step")

	// For manual/approval steps, we just ensure they are pending. For others, we set them to active.
	if step.StepType != StepTypeApproval && step.StepType != StepTypeManual {
		execution.Status = StepStatusActive
		if err := e.repo.UpdateStepExecution(ctx, execution); err != nil {
			return fmt.Errorf("failed to update step execution status: %w", err)
		}
	}

	// Execute the appropriate logic based on step type
	var err error
	switch step.StepType {
	case StepTypeManual:
		err = e.executeManualStep(ctx, step, execution)
	case StepTypeAutomated:
		err = e.executeAutomatedStep(ctx, step, execution)
	case StepTypeApproval:
		err = e.executeApprovalStep(ctx, step, execution)
	case StepTypeNotification:
		err = e.executeNotificationStep(ctx, step, execution)
	case StepTypeIntegration:
		err = e.executeIntegrationStep(ctx, step, execution)
	case StepTypeDecision:
		err = e.executeDecisionStep(ctx, step, execution)
	case StepTypeAITask:
		err = e.executeAIStep(ctx, step, execution)
	default:
		err = fmt.Errorf("unsupported step type: %s", step.StepType)
	}

	// Update execution based on result
	completedTime := time.Now()
	execution.UpdatedAt = completedTime

	if err != nil {
		execution.Status = StepStatusFailed
		errStr := err.Error()
		execution.Error = &errStr
		e.logger.WithError(err).WithFields(logrus.Fields{
			"step_id":      step.ID,
			"execution_id": execution.ExecutionID,
		}).Error("Step execution failed")
	} else if step.StepType != StepTypeApproval && step.StepType != StepTypeManual {
		// Only auto-complete non-manual steps
		execution.Status = StepStatusCompleted
		result := map[string]interface{}{
			"completed_at": completedTime,
			"duration":     completedTime.Sub(execution.StartedAt).Seconds(),
		}
		resultJSON, _ := json.Marshal(result)
		execution.Result = datatypes.JSON(resultJSON)
		execution.CompletedAt = &completedTime
	}

	// Save step execution status
	if err := e.repo.UpdateStepExecution(ctx, execution); err != nil {
		e.logger.WithError(err).Error("Failed to update step execution")
		return fmt.Errorf("failed to update step execution: %w", err)
	}

	// If step was successfully and automatically completed, process next steps
	if err == nil && execution.Status == StepStatusCompleted {
		if err := e.processTransitions(ctx, step, execution, "on_approve"); err != nil {
			e.logger.WithError(err).Error("Failed to process next steps")
			// Continue execution even if next steps processing fails
		}
	}

	// Check if workflow is complete
	if err := e.checkWorkflowCompletion(ctx, execution.ExecutionID); err != nil {
		e.logger.WithError(err).Error("Failed to check workflow completion")
		// Continue execution even if completion check fails
	}

	return err
}

// ValidateTransition checks if a transition from one step to another is valid
func (e *DefaultWorkflowExecutor) ValidateTransition(ctx context.Context, fromStep, toStep *WorkflowStep) error {
	// List transitions from the source step
	filter := &WorkflowTransitionFilter{
		FromStepID: &fromStep.ID,
	}
	transitions, _, err := e.repo.ListTransitions(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list transitions: %w", err)
	}

	// Check if there's a transition to the target step
	for _, transition := range transitions {
		if transition.ToStepID == toStep.ID {
			return nil // Valid transition found
		}
	}

	return fmt.Errorf("no valid transition from step %s to step %s", fromStep.ID, toStep.ID)
}

// executeManualStep handles manual steps which require user interaction
func (e *DefaultWorkflowExecutor) executeManualStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Executing manual step - setting to pending")
	// Manual steps are also set to pending and wait for a user to mark them as complete.
	execution.Status = StepStatusPending
	if err := e.repo.UpdateStepExecution(ctx, execution); err != nil {
		return err
	}

	// Notify assigned user or role
	go e.notifyAssignees(context.Background(), step)

	return nil
}

// executeAutomatedStep handles automated steps
func (e *DefaultWorkflowExecutor) executeAutomatedStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Executing automated step")

	// Simulate processing time
	time.Sleep(time.Millisecond * 200)

	// For demonstration, we'll just mark it as successful
	return nil
}

// executeApprovalStep handles approval steps
func (e *DefaultWorkflowExecutor) executeApprovalStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Executing approval step - setting to pending")
	// For approval steps, we just set them to pending and wait for external approval.
	execution.Status = StepStatusPending
	if err := e.repo.UpdateStepExecution(ctx, execution); err != nil {
		return err
	}

	// Notify assigned user or role
	go e.notifyAssignees(context.Background(), step)

	return nil
}

// executeNotificationStep handles notification steps
func (e *DefaultWorkflowExecutor) executeNotificationStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Sending notification")

	// Simulate sending a notification
	time.Sleep(time.Millisecond * 50)

	return nil
}

// executeIntegrationStep handles integration with external systems
func (e *DefaultWorkflowExecutor) executeIntegrationStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Executing integration step")

	// Simulate integration with external system
	time.Sleep(time.Millisecond * 300)

	return nil
}

// executeDecisionStep handles decision branches
func (e *DefaultWorkflowExecutor) executeDecisionStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Evaluating decision step")

	// Simulate decision logic
	time.Sleep(time.Millisecond * 100)

	// Decision outcome would determine next step via transitions
	return nil
}

// executeAIStep handles AI-powered tasks
func (e *DefaultWorkflowExecutor) executeAIStep(ctx context.Context, step *WorkflowStep, execution *WorkflowStepExecution) error {
	e.logger.WithField("step_id", step.ID).Info("Executing AI step")

	// Simulate AI processing
	time.Sleep(time.Millisecond * 400)

	return nil
}

// ProcessTransitions exposes the processTransitions functionality from the executor
func (e *DefaultWorkflowExecutor) ProcessTransitions(ctx context.Context, currentStep *WorkflowStep, execution *WorkflowStepExecution, onEvent string) error {
	return e.processTransitions(ctx, currentStep, execution, onEvent)
}

// processTransitions finds and executes the next steps in the workflow based on an event
func (e *DefaultWorkflowExecutor) processTransitions(ctx context.Context, currentStep *WorkflowStep, execution *WorkflowStepExecution, onEvent string) error {
	// List transitions from the current step for the given event
	filter := &WorkflowTransitionFilter{
		FromStepID: &currentStep.ID,
		OnEvent:    &onEvent,
	}
	transitions, _, err := e.repo.ListTransitions(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list transitions: %w", err)
	}

	// If no transitions for this event, the path might be complete
	if len(transitions) == 0 {
		e.logger.WithFields(logrus.Fields{
			"workflow_execution_id": execution.ExecutionID,
			"on_event":              onEvent,
		}).Info("No more steps to process for this event.")

		// Only mark as complete on an approval event, not on rejection.
		if onEvent == "on_approve" {
			go e.completeWorkflow(context.Background(), execution.ExecutionID, currentStep.WorkflowID)
		}
		return nil
	}

	// Check if workflow execution exists
	if _, err := e.repo.GetExecutionByID(ctx, execution.ExecutionID); err != nil {
		return fmt.Errorf("failed to get workflow execution: %w", err)
	}

	// Process each transition
	for _, transition := range transitions {
		// Get the target step
		toStep, err := e.repo.GetStepByID(ctx, transition.ToStepID)
		if err != nil {
			e.logger.WithError(err).WithField("to_step_id", transition.ToStepID).Error("Failed to get target step")
			continue
		}

		// Check conditions if any
		if len(transition.Conditions) > 0 {
			// Unmarshal conditions to map
			var conditions map[string]interface{}
			if err := json.Unmarshal(transition.Conditions, &conditions); err == nil {
				// Evaluate conditions (simplified for demonstration)
				// In a real system, this would be more complex
				conditionsMet := e.evaluateConditions(conditions, execution)
				if !conditionsMet {
					e.logger.WithFields(logrus.Fields{
						"from_step_id": currentStep.ID,
						"to_step_id":   toStep.ID,
					}).Info("Transition conditions not met, skipping")
					continue
				}
			} else {
				e.logger.WithError(err).Error("Failed to unmarshal transition conditions")
				continue
			}
		}

		// Create a new step execution for the next step
		now := time.Now()
		metadataMap := map[string]interface{}{
			"previous_step_id": currentStep.ID,
			"transition_id":    transition.ID,
		}
		metadataJSON, _ := json.Marshal(metadataMap)

		nextStepExecution := &WorkflowStepExecution{
			ID:                uuid.New(),
			ExecutionID:       execution.ExecutionID,
			StepID:            toStep.ID,
			Status:            StepStatusPending,
			ExecutionPriority: execution.ExecutionPriority,
			ExecutionMetadata: datatypes.JSON(metadataJSON),
			StartedAt:         now,
			UpdatedAt:         now,
		}

		if err := e.repo.CreateStepExecution(ctx, nextStepExecution); err != nil {
			e.logger.WithError(err).Error("Failed to create next step execution")
			continue
		}

		// If step is auto-advance, execute it immediately
		if toStep.AutoAdvance {
			go func(step *WorkflowStep, stepExec *WorkflowStepExecution) {
				ctx := context.Background() // Use a new context for async execution
				if err := e.ExecuteStep(ctx, step, stepExec); err != nil {
					e.logger.WithError(err).Error("Failed to auto-execute next step")
				}
			}(toStep, nextStepExecution)
		}
	}

	return nil
}

// evaluateConditions checks if transition conditions are met
func (e *DefaultWorkflowExecutor) evaluateConditions(conditions map[string]interface{}, _execution *WorkflowStepExecution) bool {
	// This is a simplified condition evaluation
	// In a real system, this would be more complex with a proper rules engine

	// If no conditions, assume all conditions are met
	if len(conditions) == 0 {
		return true
	}

	// For demonstration, we'll assume conditions are met
	return true
}

// checkWorkflowCompletion checks if all workflow steps are complete
func (e *DefaultWorkflowExecutor) checkWorkflowCompletion(ctx context.Context, executionID uuid.UUID) error {
	// Get all step executions for this workflow execution
	stepExecutions, err := e.repo.ListStepExecutions(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to list step executions: %w", err)
	}

	// If no step executions, something is wrong
	if len(stepExecutions) == 0 {
		return fmt.Errorf("no step executions found for execution %s", executionID)
	}

	// Check if any steps are still pending or active
	for _, execution := range stepExecutions {
		if execution.Status == StepStatusPending || execution.Status == StepStatusActive {
			// Workflow is still in progress
			return nil
		}
	}

	// All steps are complete/failed/skipped, get the workflow execution
	workflowExecution, err := e.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get workflow execution: %w", err)
	}

	// Check if all required steps completed successfully
	allRequiredStepsSucceeded := true
	for _, execution := range stepExecutions {
		// Get the corresponding step to check if it's required
		step, err := e.repo.GetStepByID(ctx, execution.StepID)
		if err != nil {
			e.logger.WithError(err).WithField("step_id", execution.StepID).Error("Failed to get step")
			continue
		}

		if step.IsRequired && execution.Status != StepStatusCompleted {
			allRequiredStepsSucceeded = false
			break
		}
	}

	// Update workflow execution status
	now := time.Now()
	workflowExecution.CompletedAt = &now
	workflowExecution.UpdatedAt = now

	if allRequiredStepsSucceeded {
		workflowExecution.Status = WorkflowStatusCompleted
		resultSuccess := map[string]interface{}{
			"completed_at": now,
			"duration":     now.Sub(workflowExecution.StartedAt).Seconds(),
			"status":       "success",
		}
		resultJSON, _ := json.Marshal(resultSuccess)
		workflowExecution.Result = datatypes.JSON(resultJSON)
	} else {
		workflowExecution.Status = WorkflowStatusFailed
		resultFailed := map[string]interface{}{
			"completed_at": now,
			"duration":     now.Sub(workflowExecution.StartedAt).Seconds(),
			"status":       "failed",
		}
		resultJSON, _ := json.Marshal(resultFailed)
		workflowExecution.Result = datatypes.JSON(resultJSON)
	}

	// Update the workflow execution
	if err := e.repo.UpdateExecution(ctx, workflowExecution); err != nil {
		return fmt.Errorf("failed to update workflow execution: %w", err)
	}

	// Update the workflow status
	workflow, err := e.repo.GetByID(ctx, workflowExecution.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Update workflow with completion information
	workflow.LastExecutedAt = &now
	if allRequiredStepsSucceeded {
		workflow.Status = WorkflowStatusCompleted
	} else {
		workflow.Status = WorkflowStatusFailed
	}

	// Update actual duration if available
	duration := int(now.Sub(workflowExecution.StartedAt).Seconds())
	workflow.ActualDuration = &duration

	// Update workflow success rate
	// This is a simplified calculation
	executionFilter := &WorkflowExecutionFilter{
		WorkflowID: &workflow.ID,
	}
	allExecutions, _, err := e.repo.ListExecutions(ctx, executionFilter)
	if err == nil && len(allExecutions) > 0 {
		successCount := 0
		for _, exec := range allExecutions {
			if exec.Status == WorkflowStatusCompleted {
				successCount++
			}
		}
		workflow.SuccessRate = float64(successCount) / float64(len(allExecutions))
	}

	// Update the workflow
	if err := e.repo.Update(ctx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	return nil
}

func (e *DefaultWorkflowExecutor) notifyAssignees(ctx context.Context, step *WorkflowStep) {
	if e.notifier == nil {
		return
	}

	workflow, err := e.repo.GetByID(ctx, step.WorkflowID)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to get workflow for notification")
		return
	}

	title := fmt.Sprintf("Action Required: %s", step.Name)
	content := fmt.Sprintf("Your action is required for step '%s' in workflow '%s'.", step.Name, workflow.Name)
	data := map[string]string{
		"workflowId": step.WorkflowID.String(),
		"stepId":     step.ID.String(),
	}

	// If assigned to a specific user
	if step.AssignedTo != nil {
		e.notifier.NotifyUser(ctx, *step.AssignedTo, notification.WorkflowActionRequired, title, content, data, "workflow", step.WorkflowID)
		return
	}

	// If assigned to a role
	if step.AssignedToRoleID != nil && e.rolesService != nil {
		userIDs, err := e.rolesService.GetUserIDsByRole(ctx, *step.AssignedToRoleID)
		if err != nil {
			e.logger.WithError(err).WithField("roleId", *step.AssignedToRoleID).Error("Failed to get users by role for notification")
			return
		}
		for _, userID := range userIDs {
			e.notifier.NotifyUser(ctx, userID, notification.WorkflowActionRequired, title, content, data, "workflow", step.WorkflowID)
		}
	}
}

func (e *DefaultWorkflowExecutor) completeWorkflow(ctx context.Context, executionID, workflowID uuid.UUID) {
	// Here you would check if all parallel branches are completed, etc.
	// For now, we assume if we reach the end of a path, it's done.
	execution, err := e.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to get workflow execution for completion notification")
		return
	}

	// Update execution status
	execution.Status = "Completed"
	now := time.Now()
	execution.CompletedAt = &now
	if err := e.repo.UpdateExecution(ctx, execution); err != nil {
		e.logger.WithError(err).Error("Failed to mark workflow execution as completed")
		return
	}

	if e.notifier == nil {
		return
	}

	// Notify initiator
	workflow, err := e.repo.GetByID(ctx, workflowID)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to get workflow for completion notification")
		return
	}

	title := fmt.Sprintf("Workflow '%s' Completed", workflow.Name)
	content := "The workflow has been successfully completed."
	data := map[string]string{
		"workflowId":          workflow.ID.String(),
		"workflowExecutionId": execution.ID.String(),
	}
	e.notifier.NotifyUser(ctx, workflow.CreatedBy, notification.WorkflowCompleted, title, content, data, "workflow", workflow.ID)
}
