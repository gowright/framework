package core

import (
	"time"
)

// NewIntegrationTesterFromComponents creates a new integration tester from components
func NewIntegrationTesterFromComponents(uiTester UITester, apiTester APITester, dbTester DatabaseTester) IntegrationTester {
	// This function should be implemented in the integration package to avoid circular imports
	// For now, return nil as a placeholder
	return nil
}

// FunctionTest implements the Test interface for function-based tests
type FunctionTest struct {
	name     string
	testFunc func(*TestContext)
}

// NewFunctionTest creates a new function-based test
func NewFunctionTest(name string, testFunc func(*TestContext)) Test {
	return &FunctionTest{
		name:     name,
		testFunc: testFunc,
	}
}

// GetName returns the test name
func (ft *FunctionTest) GetName() string {
	return ft.name
}

// Execute runs the function test
func (ft *FunctionTest) Execute() *TestCaseResult {
	startTime := time.Now()
	context := NewTestContext(ft.name)
	defer context.Close()

	var result *TestCaseResult

	// Execute the test function with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = NewTestExecutionError(ft.name, "test panicked", nil)
				}
				result = &TestCaseResult{
					Name:      ft.name,
					Status:    TestStatusError,
					StartTime: startTime,
					EndTime:   time.Now(),
					Error:     err,
				}
			}
		}()

		// Execute the test function
		ft.testFunc(context)

		// Create result based on context state
		result = context.ToTestCaseResult()

		// Check if context indicates failure
		if context.Err() != nil {
			result.Status = TestStatusFailed
			result.Error = context.Err()
		}
	}()

	if result == nil {
		result = context.ToTestCaseResult()
	}

	result.Duration = result.EndTime.Sub(startTime)
	return result
}

// TestBuilder provides a fluent interface for building tests
type FunctionTestBuilder struct {
	name     string
	testFunc func(*TestContext)
	setup    func() error
	teardown func() error
	timeout  time.Duration
	retries  int
}

// NewFunctionTestBuilder creates a new function test builder
func NewFunctionTestBuilder(name string) *FunctionTestBuilder {
	return &FunctionTestBuilder{
		name:    name,
		timeout: 5 * time.Minute,
		retries: 0,
	}
}

// WithTestFunc sets the test function
func (tb *FunctionTestBuilder) WithTestFunc(testFunc func(*TestContext)) *FunctionTestBuilder {
	tb.testFunc = testFunc
	return tb
}

// WithSetup sets the setup function
func (tb *FunctionTestBuilder) WithSetup(setup func() error) *FunctionTestBuilder {
	tb.setup = setup
	return tb
}

// WithTeardown sets the teardown function
func (tb *FunctionTestBuilder) WithTeardown(teardown func() error) *FunctionTestBuilder {
	tb.teardown = teardown
	return tb
}

// WithTimeout sets the test timeout
func (tb *FunctionTestBuilder) WithTimeout(timeout time.Duration) *FunctionTestBuilder {
	tb.timeout = timeout
	return tb
}

// WithRetries sets the number of retries
func (tb *FunctionTestBuilder) WithRetries(retries int) *FunctionTestBuilder {
	tb.retries = retries
	return tb
}

// Build creates the test
func (tb *FunctionTestBuilder) Build() Test {
	if tb.testFunc == nil {
		panic("test function is required")
	}

	return &EnhancedFunctionTest{
		name:     tb.name,
		testFunc: tb.testFunc,
		setup:    tb.setup,
		teardown: tb.teardown,
		timeout:  tb.timeout,
		retries:  tb.retries,
	}
}

// EnhancedFunctionTest is an enhanced version of FunctionTest with setup/teardown
type EnhancedFunctionTest struct {
	name     string
	testFunc func(*TestContext)
	setup    func() error
	teardown func() error
	timeout  time.Duration
	retries  int
}

// GetName returns the test name
func (eft *EnhancedFunctionTest) GetName() string {
	return eft.name
}

// Execute runs the enhanced function test
func (eft *EnhancedFunctionTest) Execute() *TestCaseResult {
	startTime := time.Now()

	// Try the test with retries
	var lastResult *TestCaseResult
	for attempt := 0; attempt <= eft.retries; attempt++ {
		result := eft.executeOnce(startTime)
		lastResult = result

		// If test passed or errored (not failed), don't retry
		if result.Status == TestStatusPassed || result.Status == TestStatusError {
			break
		}

		// If this was the last attempt, break
		if attempt == eft.retries {
			break
		}

		// Wait a bit before retrying
		time.Sleep(time.Second)
	}

	return lastResult
}

// executeOnce executes the test once
func (eft *EnhancedFunctionTest) executeOnce(startTime time.Time) *TestCaseResult {
	context := NewTestContextWithTimeout(eft.name, eft.timeout)
	defer context.Close()

	// Execute setup if defined
	if eft.setup != nil {
		if err := eft.setup(); err != nil {
			return &TestCaseResult{
				Name:      eft.name,
				Status:    TestStatusError,
				StartTime: startTime,
				EndTime:   time.Now(),
				Error:     NewTestSetupError(eft.name, "setup failed", err),
			}
		}
	}

	// Add teardown to context cleanup
	if eft.teardown != nil {
		context.AddCleanup(eft.teardown)
	}

	var result *TestCaseResult

	// Execute the test function with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = NewTestExecutionError(eft.name, "test panicked", nil)
				}
				result = &TestCaseResult{
					Name:      eft.name,
					Status:    TestStatusError,
					StartTime: startTime,
					EndTime:   time.Now(),
					Error:     err,
				}
			}
		}()

		// Execute the test function
		eft.testFunc(context)

		// Create result based on context state
		status := TestStatusPassed
		var err error

		if context.Err() != nil {
			status = TestStatusFailed
			err = context.Err()
		}

		result = context.ToTestCaseResult()
		if err != nil {
			result.Status = status
			result.Error = err
		}
	}()

	if result == nil {
		result = context.ToTestCaseResult()
	}

	result.Duration = result.EndTime.Sub(startTime)
	return result
}
