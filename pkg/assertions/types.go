// Package assertions provides common assertion utilities and types for all testing modules
package assertions

import (
	"time"
)

// TestStatus represents the status of a test
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusError   TestStatus = "error"
)

// String returns the string representation of TestStatus
func (ts TestStatus) String() string {
	switch ts {
	case TestStatusPassed:
		return "passed"
	case TestStatusFailed:
		return "failed"
	case TestStatusSkipped:
		return "skipped"
	case TestStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// AssertionStep represents a single assertion step in a test
type AssertionStep struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      TestStatus    `json:"status"`
	Error       error         `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Expected    interface{}   `json:"expected,omitempty"`
	Actual      interface{}   `json:"actual,omitempty"`
}
