package core

import (
	"fmt"
	"time"
)

// IntegrationTestImpl implements the Test interface for integration tests
type IntegrationTestImpl struct {
	integrationTest *IntegrationTest
	tester          IntegrationTester
	setupData       map[string]interface{}
	teardownData    map[string]interface{}
	failureContext  *FailureContext
}

// FailureContext holds detailed information about test failures across systems
type FailureContext struct {
	FailedStep      *IntegrationStep       `json:"failed_step"`
	StepIndex       int                    `json:"step_index"`
	Error           error                  `json:"error"`
	SystemStates    map[string]interface{} `json:"system_states"`
	Screenshots     []string               `json:"screenshots"`
	APIResponses    []*APIResponse         `json:"api_responses"`
	DatabaseResults []*DatabaseResult      `json:"database_results"`
	Logs            []string               `json:"logs"`
	Timestamp       time.Time              `json:"timestamp"`
}

// TestDataSetup represents setup operations for test data across systems
type TestDataSetup struct {
	UISetup       []UISetupAction       `json:"ui_setup,omitempty"`
	APISetup      []APISetupAction      `json:"api_setup,omitempty"`
	DatabaseSetup []DatabaseSetupAction `json:"database_setup,omitempty"`
}

// TestDataTeardown represents teardown operations for test data across systems
type TestDataTeardown struct {
	UITeardown       []UITeardownAction       `json:"ui_teardown,omitempty"`
	APITeardown      []APITeardownAction      `json:"api_teardown,omitempty"`
	DatabaseTeardown []DatabaseTeardownAction `json:"database_teardown,omitempty"`
}

// UISetupAction represents a UI setup operation
type UISetupAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// APISetupAction represents an API setup operation
type APISetupAction struct {
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	StoreAs  string            `json:"store_as,omitempty"` // Store response data for later use
}

// DatabaseSetupAction represents a database setup operation
type DatabaseSetupAction struct {
	Connection string        `json:"connection"`
	Query      string        `json:"query"`
	Args       []interface{} `json:"args,omitempty"`
	StoreAs    string        `json:"store_as,omitempty"` // Store result data for later use
}

// UITeardownAction represents a UI teardown operation
type UITeardownAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// APITeardownAction represents an API teardown operation
type APITeardownAction struct {
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
}

// DatabaseTeardownAction represents a database teardown operation
type DatabaseTeardownAction struct {
	Connection string        `json:"connection"`
	Query      string        `json:"query"`
	Args       []interface{} `json:"args,omitempty"`
}

// NewIntegrationTestImpl creates a new integration test implementation
func NewIntegrationTestImpl(integrationTest *IntegrationTest, tester IntegrationTester) *IntegrationTestImpl {
	return &IntegrationTestImpl{
		integrationTest: integrationTest,
		tester:          tester,
		setupData:       make(map[string]interface{}),
		teardownData:    make(map[string]interface{}),
		failureContext: &FailureContext{
			SystemStates:    make(map[string]interface{}),
			Screenshots:     make([]string, 0),
			APIResponses:    make([]*APIResponse, 0),
			DatabaseResults: make([]*DatabaseResult, 0),
			Logs:            make([]string, 0),
		},
	}
}

// GetName returns the name of the integration test
func (it *IntegrationTestImpl) GetName() string {
	return it.integrationTest.Name
}

// Execute runs the complete integration test workflow
func (it *IntegrationTestImpl) Execute() *TestCaseResult {
	startTime := time.Now()
	result := &TestCaseResult{
		Name:      it.integrationTest.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
	}

	// Execute setup phase
	if err := it.executeSetup(); err != nil {
		result.Status = TestStatusError
		result.Error = NewGowrightError(ConfigurationError, "setup phase failed", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Execute main test steps
	if err := it.executeSteps(); err != nil {
		result.Status = TestStatusFailed
		result.Error = err
		result.Screenshots = it.failureContext.Screenshots
		result.Logs = it.failureContext.Logs

		// Attempt teardown even if test failed
		if teardownErr := it.executeTeardown(); teardownErr != nil {
			result.Error = NewGowrightError(AssertionError,
				"test failed and teardown also failed",
				fmt.Errorf("test error: %w, teardown error: %w", err, teardownErr))
		}

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Execute teardown phase
	if err := it.executeTeardown(); err != nil {
		result.Status = TestStatusError
		result.Error = NewGowrightError(ConfigurationError, "teardown phase failed", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}

// executeSetup runs setup operations across all systems
func (it *IntegrationTestImpl) executeSetup() error {
	// Setup operations would be defined in the integration test configuration
	// For now, we'll implement a basic setup that initializes the tester
	if err := it.tester.Initialize(nil); err != nil {
		return NewGowrightError(ConfigurationError, "failed to initialize integration tester", err)
	}

	it.addLog("Setup phase completed successfully")
	return nil
}

// executeSteps runs the main integration test steps
func (it *IntegrationTestImpl) executeSteps() error {
	for i, step := range it.integrationTest.Steps {
		it.addLog(fmt.Sprintf("Executing step %d: %s", i+1, step.Name))

		if err := it.tester.ExecuteStep(&step); err != nil {
			// Collect failure context
			it.failureContext.FailedStep = &step
			it.failureContext.StepIndex = i
			it.failureContext.Error = err
			it.failureContext.Timestamp = time.Now()

			// Collect system states for debugging
			it.collectSystemStates(&step)

			return NewGowrightError(AssertionError,
				fmt.Sprintf("integration test failed at step %d: %s", i+1, step.Name), err)
		}

		// Collect successful step data for context
		it.collectStepData(&step)
	}

	it.addLog("All integration test steps completed successfully")
	return nil
}

// executeTeardown runs teardown operations across all systems
func (it *IntegrationTestImpl) executeTeardown() error {
	// Execute rollback steps if defined
	if len(it.integrationTest.Rollback) > 0 {
		it.addLog("Executing rollback steps")
		if err := it.tester.Rollback(it.integrationTest.Rollback); err != nil {
			return NewGowrightError(ConfigurationError, "rollback execution failed", err)
		}
	}

	// Cleanup the tester
	if err := it.tester.Cleanup(); err != nil {
		return NewGowrightError(ConfigurationError, "failed to cleanup integration tester", err)
	}

	it.addLog("Teardown phase completed successfully")
	return nil
}

// collectSystemStates gathers current state information from all systems
func (it *IntegrationTestImpl) collectSystemStates(failedStep *IntegrationStep) {
	switch failedStep.Type {
	case StepTypeUI:
		it.collectUIState()
	case StepTypeAPI:
		it.collectAPIState(failedStep)
	case StepTypeDatabase:
		it.collectDatabaseState(failedStep)
	}
}

// collectUIState captures UI-specific failure context
func (it *IntegrationTestImpl) collectUIState() {
	// Try to capture screenshot for debugging
	// Try to get UI tester from integration tester (simplified approach)
	if it.tester != nil {
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("failure_%s_%s.png", it.integrationTest.Name, timestamp)

		// For now, just log that we would capture a screenshot
		it.addLog(fmt.Sprintf("Would capture screenshot: %s", filename))

		// Add placeholder for system state
		it.failureContext.SystemStates["ui_state"] = "captured"
	}
}

// collectAPIState captures API-specific failure context
func (it *IntegrationTestImpl) collectAPIState(failedStep *IntegrationStep) {
	// Store the failed API action details
	if action, ok := failedStep.Action.(*APIStepAction); ok {
		it.failureContext.SystemStates["failed_api_action"] = map[string]interface{}{
			"method":   action.Method,
			"endpoint": action.Endpoint,
			"headers":  action.Headers,
			"body":     action.Body,
		}
	}
}

// collectDatabaseState captures database-specific failure context
func (it *IntegrationTestImpl) collectDatabaseState(failedStep *IntegrationStep) {
	// Store the failed database action details
	if action, ok := failedStep.Action.(*DatabaseStepAction); ok {
		it.failureContext.SystemStates["failed_database_action"] = map[string]interface{}{
			"connection": action.Connection,
			"query":      action.Query,
			"args":       action.Args,
		}
	}
}

// collectStepData collects data from successful steps for context
func (it *IntegrationTestImpl) collectStepData(step *IntegrationStep) {
	switch step.Type {
	case StepTypeAPI:
		// For API steps, we could store response data if needed
		it.addLog(fmt.Sprintf("API step '%s' completed successfully", step.Name))
	case StepTypeDatabase:
		// For database steps, we could store result data if needed
		it.addLog(fmt.Sprintf("Database step '%s' completed successfully", step.Name))
	case StepTypeUI:
		// For UI steps, we could store element states if needed
		it.addLog(fmt.Sprintf("UI step '%s' completed successfully", step.Name))
	}
}

// addLog adds a log entry to the failure context
func (it *IntegrationTestImpl) addLog(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	it.failureContext.Logs = append(it.failureContext.Logs, logEntry)
}

// GetFailureContext returns the failure context for debugging
func (it *IntegrationTestImpl) GetFailureContext() *FailureContext {
	return it.failureContext
}

// SetSetupData stores data from setup operations for use in test steps
func (it *IntegrationTestImpl) SetSetupData(key string, value interface{}) {
	it.setupData[key] = value
}

// GetSetupData retrieves data stored during setup operations
func (it *IntegrationTestImpl) GetSetupData(key string) (interface{}, bool) {
	value, exists := it.setupData[key]
	return value, exists
}

// SetTeardownData stores data for use in teardown operations
func (it *IntegrationTestImpl) SetTeardownData(key string, value interface{}) {
	it.teardownData[key] = value
}

// GetTeardownData retrieves data for teardown operations
func (it *IntegrationTestImpl) GetTeardownData(key string) (interface{}, bool) {
	value, exists := it.teardownData[key]
	return value, exists
}
