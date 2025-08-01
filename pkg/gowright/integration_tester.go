package gowright

import (
	"fmt"
	"sync"
	"time"
)

// IntegrationTesterImpl implements the IntegrationTester interface
type IntegrationTesterImpl struct {
	uiTester  UITester
	apiTester APITester
	dbTester  DatabaseTester
	config    *IntegrationConfig
	name      string
	mutex     sync.RWMutex
}

// IntegrationConfig holds configuration for integration testing
type IntegrationConfig struct {
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	RollbackOnError bool          `json:"rollback_on_error"`
	ParallelSteps   bool          `json:"parallel_steps"`
}

// NewIntegrationTester creates a new IntegrationTester instance
func NewIntegrationTester(uiTester UITester, apiTester APITester, dbTester DatabaseTester) *IntegrationTesterImpl {
	return &IntegrationTesterImpl{
		uiTester:  uiTester,
		apiTester: apiTester,
		dbTester:  dbTester,
		name:      "IntegrationTester",
		config: &IntegrationConfig{
			MaxRetries:      3,
			RetryDelay:      1 * time.Second,
			RollbackOnError: true,
			ParallelSteps:   false,
		},
	}
}

// Initialize sets up the integration tester with the provided configuration
func (it *IntegrationTesterImpl) Initialize(config interface{}) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	if config != nil {
		if integrationConfig, ok := config.(*IntegrationConfig); ok {
			it.config = integrationConfig
		}
	}

	// Initialize all sub-testers if they haven't been initialized
	if it.uiTester != nil {
		if err := it.uiTester.Initialize(nil); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize UI tester", err)
		}
	}

	if it.apiTester != nil {
		if err := it.apiTester.Initialize(nil); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize API tester", err)
		}
	}

	if it.dbTester != nil {
		if err := it.dbTester.Initialize(nil); err != nil {
			return NewGowrightError(ConfigurationError, "failed to initialize database tester", err)
		}
	}

	return nil
}

// Cleanup performs cleanup operations for all testers
func (it *IntegrationTesterImpl) Cleanup() error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	var errors []error

	if it.uiTester != nil {
		if err := it.uiTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("UI tester cleanup failed: %w", err))
		}
	}

	if it.apiTester != nil {
		if err := it.apiTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("API tester cleanup failed: %w", err))
		}
	}

	if it.dbTester != nil {
		if err := it.dbTester.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("database tester cleanup failed: %w", err))
		}
	}

	if len(errors) > 0 {
		return NewGowrightError(ConfigurationError, "cleanup failed for one or more testers", fmt.Errorf("%v", errors))
	}

	return nil
}

// GetName returns the name of the integration tester
func (it *IntegrationTesterImpl) GetName() string {
	return it.name
}

// ExecuteStep executes a single integration step
func (it *IntegrationTesterImpl) ExecuteStep(step *IntegrationStep) error {
	if step == nil {
		return NewGowrightError(ConfigurationError, "integration step cannot be nil", nil)
	}

	it.mutex.RLock()
	defer it.mutex.RUnlock()

	var err error
	maxRetries := it.config.MaxRetries
	retryDelay := it.config.RetryDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		switch step.Type {
		case StepTypeUI:
			err = it.executeUIStep(step)
		case StepTypeAPI:
			err = it.executeAPIStep(step)
		case StepTypeDatabase:
			err = it.executeDatabaseStep(step)
		default:
			return NewGowrightError(ConfigurationError, fmt.Sprintf("unsupported step type: %s", step.Type.String()), nil)
		}

		if err == nil {
			return nil
		}

		// If this is the last attempt, return the error
		if attempt == maxRetries {
			return NewGowrightError(AssertionError, fmt.Sprintf("step '%s' failed after %d attempts", step.Name, maxRetries+1), err)
		}
	}

	return err
}

// ExecuteWorkflow executes a complete integration workflow
func (it *IntegrationTesterImpl) ExecuteWorkflow(steps []IntegrationStep) error {
	if len(steps) == 0 {
		return NewGowrightError(ConfigurationError, "workflow steps cannot be empty", nil)
	}

	executedSteps := make([]IntegrationStep, 0, len(steps))

	for i, step := range steps {
		if err := it.ExecuteStep(&step); err != nil {
			// If rollback is enabled, attempt to rollback executed steps
			if it.config.RollbackOnError && len(executedSteps) > 0 {
				rollbackErr := it.Rollback(executedSteps)
				if rollbackErr != nil {
					return NewGowrightError(AssertionError, 
						fmt.Sprintf("step %d failed and rollback also failed", i+1), 
						fmt.Errorf("original error: %w, rollback error: %w", err, rollbackErr))
				}
			}
			return NewGowrightError(AssertionError, fmt.Sprintf("workflow failed at step %d: %s", i+1, step.Name), err)
		}
		executedSteps = append(executedSteps, step)
	}

	return nil
}

// Rollback performs rollback operations for failed tests
func (it *IntegrationTesterImpl) Rollback(steps []IntegrationStep) error {
	if len(steps) == 0 {
		return nil
	}

	var errors []error

	// Execute rollback steps in reverse order
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		if err := it.executeRollbackStep(&step); err != nil {
			errors = append(errors, fmt.Errorf("rollback failed for step '%s': %w", step.Name, err))
		}
	}

	if len(errors) > 0 {
		return NewGowrightError(AssertionError, "rollback operations failed", fmt.Errorf("%v", errors))
	}

	return nil
}

// executeUIStep executes a UI-specific integration step
func (it *IntegrationTesterImpl) executeUIStep(step *IntegrationStep) error {
	if it.uiTester == nil {
		return NewGowrightError(ConfigurationError, "UI tester not available", nil)
	}

	action, ok := step.Action.(*UIStepAction)
	if !ok {
		return NewGowrightError(ConfigurationError, "invalid UI step action type", nil)
	}

	switch action.Type {
	case "navigate":
		if url, ok := action.Parameters["url"].(string); ok {
			return it.uiTester.Navigate(url)
		}
		return NewGowrightError(ConfigurationError, "navigate action requires 'url' parameter", nil)

	case "click":
		if selector, ok := action.Parameters["selector"].(string); ok {
			return it.uiTester.Click(selector)
		}
		return NewGowrightError(ConfigurationError, "click action requires 'selector' parameter", nil)

	case "type":
		selector, hasSelector := action.Parameters["selector"].(string)
		text, hasText := action.Parameters["text"].(string)
		if hasSelector && hasText {
			return it.uiTester.Type(selector, text)
		}
		return NewGowrightError(ConfigurationError, "type action requires 'selector' and 'text' parameters", nil)

	case "wait":
		selector, hasSelector := action.Parameters["selector"].(string)
		timeout, hasTimeout := action.Parameters["timeout"].(time.Duration)
		if hasSelector && hasTimeout {
			return it.uiTester.WaitForElement(selector, timeout)
		}
		return NewGowrightError(ConfigurationError, "wait action requires 'selector' and 'timeout' parameters", nil)

	default:
		return NewGowrightError(ConfigurationError, fmt.Sprintf("unsupported UI action type: %s", action.Type), nil)
	}
}

// executeAPIStep executes an API-specific integration step
func (it *IntegrationTesterImpl) executeAPIStep(step *IntegrationStep) error {
	if it.apiTester == nil {
		return NewGowrightError(ConfigurationError, "API tester not available", nil)
	}

	action, ok := step.Action.(*APIStepAction)
	if !ok {
		return NewGowrightError(ConfigurationError, "invalid API step action type", nil)
	}

	var response *APIResponse
	var err error

	switch action.Method {
	case "GET":
		response, err = it.apiTester.Get(action.Endpoint, action.Headers)
	case "POST":
		response, err = it.apiTester.Post(action.Endpoint, action.Body, action.Headers)
	case "PUT":
		response, err = it.apiTester.Put(action.Endpoint, action.Body, action.Headers)
	case "DELETE":
		response, err = it.apiTester.Delete(action.Endpoint, action.Headers)
	default:
		return NewGowrightError(ConfigurationError, fmt.Sprintf("unsupported HTTP method: %s", action.Method), nil)
	}

	if err != nil {
		return err
	}

	// Validate response if validation is specified
	if step.Validation != nil {
		if validation, ok := step.Validation.(*APIStepValidation); ok {
			return it.validateAPIResponse(response, validation)
		}
	}

	return nil
}

// executeDatabaseStep executes a database-specific integration step
func (it *IntegrationTesterImpl) executeDatabaseStep(step *IntegrationStep) error {
	if it.dbTester == nil {
		return NewGowrightError(ConfigurationError, "database tester not available", nil)
	}

	action, ok := step.Action.(*DatabaseStepAction)
	if !ok {
		return NewGowrightError(ConfigurationError, "invalid database step action type", nil)
	}

	result, err := it.dbTester.Execute(action.Connection, action.Query, action.Args...)
	if err != nil {
		return err
	}

	// Validate result if validation is specified
	if step.Validation != nil {
		if validation, ok := step.Validation.(*DatabaseStepValidation); ok {
			return it.validateDatabaseResult(result, validation)
		}
	}

	return nil
}

// executeRollbackStep executes rollback for a specific step
func (it *IntegrationTesterImpl) executeRollbackStep(step *IntegrationStep) error {
	switch step.Type {
	case StepTypeUI:
		// UI rollback might involve navigating back or resetting state
		return it.executeUIRollback(step)
	case StepTypeAPI:
		// API rollback might involve calling cleanup endpoints
		return it.executeAPIRollback(step)
	case StepTypeDatabase:
		// Database rollback might involve running cleanup queries
		return it.executeDatabaseRollback(step)
	default:
		return NewGowrightError(ConfigurationError, fmt.Sprintf("unsupported rollback step type: %s", step.Type.String()), nil)
	}
}

// executeUIRollback performs UI-specific rollback operations
func (it *IntegrationTesterImpl) executeUIRollback(step *IntegrationStep) error {
	// UI rollback implementation - could involve taking screenshots, navigating back, etc.
	// For now, we'll just take a screenshot for debugging purposes
	if it.uiTester != nil {
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("rollback_%s_%s.png", step.Name, timestamp)
		_, err := it.uiTester.TakeScreenshot(filename)
		return err
	}
	return nil
}

// executeAPIRollback performs API-specific rollback operations
func (it *IntegrationTesterImpl) executeAPIRollback(step *IntegrationStep) error {
	// API rollback implementation - could involve calling cleanup endpoints
	// This would be implemented based on specific rollback requirements
	return nil
}

// executeDatabaseRollback performs database-specific rollback operations
func (it *IntegrationTesterImpl) executeDatabaseRollback(step *IntegrationStep) error {
	// Database rollback implementation - could involve running cleanup queries
	// This would be implemented based on specific rollback requirements
	return nil
}

// validateAPIResponse validates an API response against expected criteria
func (it *IntegrationTesterImpl) validateAPIResponse(response *APIResponse, validation *APIStepValidation) error {
	if validation.ExpectedStatusCode != 0 && response.StatusCode != validation.ExpectedStatusCode {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("expected status code %d, got %d", validation.ExpectedStatusCode, response.StatusCode), nil)
	}

	// Additional validation logic can be added here
	return nil
}

// validateDatabaseResult validates a database result against expected criteria
func (it *IntegrationTesterImpl) validateDatabaseResult(result *DatabaseResult, validation *DatabaseStepValidation) error {
	if validation.ExpectedRowCount != nil && len(result.Rows) != *validation.ExpectedRowCount {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("expected %d rows, got %d", *validation.ExpectedRowCount, len(result.Rows)), nil)
	}

	// Additional validation logic can be added here
	return nil
}