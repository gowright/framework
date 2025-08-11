package core

// TestExecutionError represents an error that occurred during test execution
type TestExecutionError struct {
	*GowrightError
	TestName string
}

// NewTestExecutionError creates a new test execution error
func NewTestExecutionError(testName, message string, cause error) *TestExecutionError {
	return &TestExecutionError{
		GowrightError: NewGowrightError(TestExecutionErrorType, message, cause),
		TestName:      testName,
	}
}

// TestSetupError represents an error that occurred during test setup
type TestSetupError struct {
	*GowrightError
	TestName string
}

// NewTestSetupError creates a new test setup error
func NewTestSetupError(testName, message string, cause error) *TestSetupError {
	return &TestSetupError{
		GowrightError: NewGowrightError(TestSetupErrorType, message, cause),
		TestName:      testName,
	}
}
