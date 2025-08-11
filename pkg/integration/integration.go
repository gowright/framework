// Package integration provides integration testing capabilities that orchestrate UI, API, database, and mobile testing
package integration

import (
	"fmt"
	"time"

	"github.com/gowright/framework/pkg/assertions"
	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
)

// IntegrationTester implements the IntegrationTester interface
type IntegrationTester struct {
	config      *config.Config
	uiTester    core.UITester
	apiTester   core.APITester
	dbTester    core.DatabaseTester
	asserter    *assertions.Asserter
	initialized bool
}

// NewIntegrationTester creates a new integration tester
func NewIntegrationTester() *IntegrationTester {
	return &IntegrationTester{
		asserter: assertions.NewAsserter(),
	}
}

// Initialize initializes the integration tester with configuration
func (it *IntegrationTester) Initialize(cfg interface{}) error {
	config, ok := cfg.(*config.Config)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid configuration type for integration tester", nil)
	}

	it.config = config
	it.initialized = true
	return nil
}

// Cleanup performs cleanup operations
func (it *IntegrationTester) Cleanup() error {
	it.initialized = false
	return nil
}

// GetName returns the name of the tester
func (it *IntegrationTester) GetName() string {
	return "IntegrationTester"
}

// SetUITester sets the UI tester for integration tests
func (it *IntegrationTester) SetUITester(tester core.UITester) {
	it.uiTester = tester
}

// SetAPITester sets the API tester for integration tests
func (it *IntegrationTester) SetAPITester(tester core.APITester) {
	it.apiTester = tester
}

// SetDatabaseTester sets the database tester for integration tests
func (it *IntegrationTester) SetDatabaseTester(tester core.DatabaseTester) {
	it.dbTester = tester
}

// ExecuteStep executes a single integration step
func (it *IntegrationTester) ExecuteStep(step *core.IntegrationStep) error {
	if !it.initialized {
		return core.NewGowrightError(core.ConfigurationError, "integration tester not initialized", nil)
	}

	switch step.Type {
	case core.StepTypeUI:
		return it.executeUIStep(step)
	case core.StepTypeAPI:
		return it.executeAPIStep(step)
	case core.StepTypeDatabase:
		return it.executeDatabaseStep(step)
	default:
		return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("unsupported step type: %s", step.Type), nil)
	}
}

// ExecuteWorkflow executes a complete integration workflow
func (it *IntegrationTester) ExecuteWorkflow(steps []core.IntegrationStep) error {
	for i, step := range steps {
		if err := it.ExecuteStep(&step); err != nil {
			// If a step fails, attempt rollback
			if rollbackErr := it.rollbackSteps(steps[:i]); rollbackErr != nil {
				return core.NewGowrightError(core.ConfigurationError,
					fmt.Sprintf("step %d failed and rollback failed: %v, rollback error: %v", i, err, rollbackErr), err)
			}
			return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("step %d failed: %v", i, err), err)
		}
	}
	return nil
}

// Rollback performs rollback operations for failed tests
func (it *IntegrationTester) Rollback(steps []core.IntegrationStep) error {
	return it.rollbackSteps(steps)
}

// ExecuteTest executes an integration test and returns the result
func (it *IntegrationTester) ExecuteTest(test *core.IntegrationTest) *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    core.TestStatusPassed,
	}

	it.asserter.Reset()

	// Execute main workflow
	if err := it.ExecuteWorkflow(test.Steps); err != nil {
		result.Status = core.TestStatusFailed
		result.Error = err
	}

	// Check for assertion failures
	if it.asserter.HasFailures() && result.Status == core.TestStatusPassed {
		result.Status = core.TestStatusFailed
		result.Error = core.NewGowrightError(core.AssertionError, "one or more assertions failed", nil)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Steps = it.asserter.GetSteps()

	return result
}

// executeUIStep executes a UI step
func (it *IntegrationTester) executeUIStep(step *core.IntegrationStep) error {
	if it.uiTester == nil {
		return core.NewGowrightError(core.ConfigurationError, "UI tester not configured for integration test", nil)
	}

	action, ok := step.Action.(*core.UIStepAction)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid UI step action", nil)
	}

	// Execute UI action based on type
	switch action.Type {
	case "navigate":
		if url, exists := action.Parameters["url"]; exists {
			if urlStr, ok := url.(string); ok {
				return it.uiTester.Navigate(urlStr)
			}
		}
	case "click":
		if selector, exists := action.Parameters["selector"]; exists {
			if selectorStr, ok := selector.(string); ok {
				return it.uiTester.Click(selectorStr)
			}
		}
	case "type":
		if selector, exists := action.Parameters["selector"]; exists {
			if text, exists := action.Parameters["text"]; exists {
				if selectorStr, ok := selector.(string); ok {
					if textStr, ok := text.(string); ok {
						return it.uiTester.Type(selectorStr, textStr)
					}
				}
			}
		}
	}

	return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("unsupported UI action: %s", action.Type), nil)
}

// executeAPIStep executes an API step
func (it *IntegrationTester) executeAPIStep(step *core.IntegrationStep) error {
	if it.apiTester == nil {
		return core.NewGowrightError(core.ConfigurationError, "API tester not configured for integration test", nil)
	}

	action, ok := step.Action.(*core.APIStepAction)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid API step action", nil)
	}

	var response *core.APIResponse
	var err error

	// Execute API action based on method
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
		return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("unsupported HTTP method: %s", action.Method), nil)
	}

	if err != nil {
		return err
	}

	// Validate response if validation is specified
	if step.Validation != nil {
		if validation, ok := step.Validation.(*core.APIStepValidation); ok {
			return it.validateAPIResponse(response, validation)
		}
	}

	return nil
}

// executeDatabaseStep executes a database step
func (it *IntegrationTester) executeDatabaseStep(step *core.IntegrationStep) error {
	if it.dbTester == nil {
		return core.NewGowrightError(core.ConfigurationError, "Database tester not configured for integration test", nil)
	}

	action, ok := step.Action.(*core.DatabaseStepAction)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid database step action", nil)
	}

	result, err := it.dbTester.Execute(action.Connection, action.Query, action.Args...)
	if err != nil {
		return err
	}

	// Validate result if validation is specified
	if step.Validation != nil {
		if validation, ok := step.Validation.(*core.DatabaseStepValidation); ok {
			return it.validateDatabaseResult(result, validation)
		}
	}

	return nil
}

// validateAPIResponse validates an API response against expected criteria
func (it *IntegrationTester) validateAPIResponse(response *core.APIResponse, validation *core.APIStepValidation) error {
	if validation.ExpectedStatusCode != 0 {
		it.asserter.Equal(validation.ExpectedStatusCode, response.StatusCode,
			fmt.Sprintf("Expected status code %d", validation.ExpectedStatusCode))
	}

	for key, expectedValue := range validation.ExpectedHeaders {
		if actualValue, exists := response.Headers[key]; exists {
			it.asserter.Equal(expectedValue, actualValue,
				fmt.Sprintf("Expected header %s to be %s", key, expectedValue))
		} else {
			it.asserter.True(false, fmt.Sprintf("Expected header %s to exist", key))
		}
	}

	// Additional JSON path validations can be added here
	return nil
}

// validateDatabaseResult validates a database result against expected criteria
func (it *IntegrationTester) validateDatabaseResult(result *core.DatabaseResult, validation *core.DatabaseStepValidation) error {
	if validation.ExpectedRowCount != nil {
		it.asserter.Equal(*validation.ExpectedRowCount, result.RowCount,
			fmt.Sprintf("Expected row count %d", *validation.ExpectedRowCount))
	}

	if validation.ExpectedAffected != nil {
		it.asserter.Equal(*validation.ExpectedAffected, result.RowsAffected,
			fmt.Sprintf("Expected affected rows %d", *validation.ExpectedAffected))
	}

	// Additional row content validations can be added here
	return nil
}

// rollbackSteps performs rollback for a set of steps
func (it *IntegrationTester) rollbackSteps(steps []core.IntegrationStep) error {
	// Rollback steps in reverse order
	for i := len(steps) - 1; i >= 0; i-- {
		// Implement rollback logic based on step type
		// This is a simplified implementation
		step := steps[i]
		switch step.Type {
		case core.StepTypeDatabase:
			// For database steps, we might need to execute rollback queries
			// This would require additional rollback information in the step
		case core.StepTypeAPI:
			// For API steps, we might need to call compensating APIs
		case core.StepTypeUI:
			// For UI steps, rollback might involve navigation or state reset
		}
	}
	return nil
}
