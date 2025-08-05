package gowright

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UIAssertionsTestSuite defines the test suite for UI assertions
type UIAssertionsTestSuite struct {
	suite.Suite
	mockTester *MockUITester
	executor   *UIAssertionExecutor
}

// SetupTest runs before each test
func (suite *UIAssertionsTestSuite) SetupTest() {
	suite.mockTester = new(MockUITester)
	suite.executor = NewUIAssertionExecutor(suite.mockTester)
}

// TearDownTest runs after each test
func (suite *UIAssertionsTestSuite) TearDownTest() {
	suite.mockTester.AssertExpectations(suite.T())
}

// TestNewUIAssertionExecutor tests the constructor
func (suite *UIAssertionsTestSuite) TestNewUIAssertionExecutor() {
	executor := NewUIAssertionExecutor(suite.mockTester)
	suite.NotNil(executor)
	suite.Equal(suite.mockTester, executor.tester)
}

// TestExecuteElementPresentAssertion tests element present assertion
func (suite *UIAssertionsTestSuite) TestExecuteElementPresentAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertElementPresent),
		Selector: "#element",
	}

	suite.mockTester.On("WaitForElement", "#element", 30*time.Second).Return(nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecuteElementPresentAssertionWithoutSelector tests element present assertion without selector
func (suite *UIAssertionsTestSuite) TestExecuteElementPresentAssertionWithoutSelector() {
	assertion := UIAssertion{
		Type: string(AssertElementPresent),
	}

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "selector is required")
}

// TestExecuteTextEqualsAssertion tests text equals assertion
func (suite *UIAssertionsTestSuite) TestExecuteTextEqualsAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertTextEquals),
		Selector: "#element",
		Expected: "expected text",
	}

	suite.mockTester.On("GetText", "#element").Return("expected text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecuteTextEqualsAssertionMismatch tests text equals assertion with mismatch
func (suite *UIAssertionsTestSuite) TestExecuteTextEqualsAssertionMismatch() {
	assertion := UIAssertion{
		Type:     string(AssertTextEquals),
		Selector: "#element",
		Expected: "expected text",
	}

	suite.mockTester.On("GetText", "#element").Return("actual text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "text mismatch")
}

// TestExecuteTextContainsAssertion tests text contains assertion
func (suite *UIAssertionsTestSuite) TestExecuteTextContainsAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertTextContains),
		Selector: "#element",
		Expected: "partial",
	}

	suite.mockTester.On("GetText", "#element").Return("this is partial text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecuteTextContainsAssertionNotFound tests text contains assertion when text not found
func (suite *UIAssertionsTestSuite) TestExecuteTextContainsAssertionNotFound() {
	assertion := UIAssertion{
		Type:     string(AssertTextContains),
		Selector: "#element",
		Expected: "missing",
	}

	suite.mockTester.On("GetText", "#element").Return("this is some text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "does not contain expected value")
}

// TestExecuteTextNotContainsAssertion tests text not contains assertion
func (suite *UIAssertionsTestSuite) TestExecuteTextNotContainsAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertTextNotContains),
		Selector: "#element",
		Expected: "forbidden",
	}

	suite.mockTester.On("GetText", "#element").Return("this is allowed text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecuteTextNotContainsAssertionFound tests text not contains assertion when text is found
func (suite *UIAssertionsTestSuite) TestExecuteTextNotContainsAssertionFound() {
	assertion := UIAssertion{
		Type:     string(AssertTextNotContains),
		Selector: "#element",
		Expected: "forbidden",
	}

	suite.mockTester.On("GetText", "#element").Return("this contains forbidden text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "should not contain expected value")
}

// TestExecuteTextMatchesAssertion tests text matches regex assertion
func (suite *UIAssertionsTestSuite) TestExecuteTextMatchesAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertTextMatches),
		Selector: "#element",
		Expected: `^\d{3}-\d{3}-\d{4}$`, // Phone number pattern
	}

	suite.mockTester.On("GetText", "#element").Return("123-456-7890", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecuteTextMatchesAssertionNoMatch tests text matches assertion with no match
func (suite *UIAssertionsTestSuite) TestExecuteTextMatchesAssertionNoMatch() {
	assertion := UIAssertion{
		Type:     string(AssertTextMatches),
		Selector: "#element",
		Expected: `^\d{3}-\d{3}-\d{4}$`, // Phone number pattern
	}

	suite.mockTester.On("GetText", "#element").Return("not a phone number", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "does not match pattern")
}

// TestExecuteTextMatchesAssertionInvalidRegex tests text matches assertion with invalid regex
func (suite *UIAssertionsTestSuite) TestExecuteTextMatchesAssertionInvalidRegex() {
	assertion := UIAssertion{
		Type:     string(AssertTextMatches),
		Selector: "#element",
		Expected: `[invalid regex`, // Invalid regex
	}

	suite.mockTester.On("GetText", "#element").Return("some text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "invalid regular expression pattern")
}

// TestExecutePageSourceContainsAssertion tests page source contains assertion
func (suite *UIAssertionsTestSuite) TestExecutePageSourceContainsAssertion() {
	assertion := UIAssertion{
		Type:     string(AssertPageSourceContains),
		Expected: "<title>Test Page</title>",
	}

	suite.mockTester.On("GetPageSource").Return("<html><head><title>Test Page</title></head><body></body></html>", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestExecutePageSourceContainsAssertionNotFound tests page source contains assertion when not found
func (suite *UIAssertionsTestSuite) TestExecutePageSourceContainsAssertionNotFound() {
	assertion := UIAssertion{
		Type:     string(AssertPageSourceContains),
		Expected: "<title>Missing Page</title>",
	}

	suite.mockTester.On("GetPageSource").Return("<html><head><title>Test Page</title></head><body></body></html>", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "does not contain expected value")
}

// TestExecuteUnsupportedAssertion tests unsupported assertion type
func (suite *UIAssertionsTestSuite) TestExecuteUnsupportedAssertion() {
	assertion := UIAssertion{
		Type: "unsupported_assertion",
	}

	err := suite.executor.ExecuteAssertion(assertion)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "unsupported assertion type")
}

// TestExecuteAssertions tests executing multiple assertions
func (suite *UIAssertionsTestSuite) TestExecuteAssertions() {
	assertions := []UIAssertion{
		{
			Type:     string(AssertElementPresent),
			Selector: "#element1",
		},
		{
			Type:     string(AssertTextEquals),
			Selector: "#element2",
			Expected: "test text",
		},
	}

	suite.mockTester.On("WaitForElement", "#element1", 30*time.Second).Return(nil)
	suite.mockTester.On("GetText", "#element2").Return("test text", nil)

	err := suite.executor.ExecuteAssertions(assertions)
	suite.NoError(err)
}

// TestExecuteAssertionsWithFailure tests executing assertions with one failing
func (suite *UIAssertionsTestSuite) TestExecuteAssertionsWithFailure() {
	assertions := []UIAssertion{
		{
			Type:     string(AssertElementPresent),
			Selector: "#element1",
		},
		{
			Type:     string(AssertTextEquals),
			Selector: "#element2",
			Expected: "expected text",
		},
	}

	suite.mockTester.On("WaitForElement", "#element1", 30*time.Second).Return(nil)
	suite.mockTester.On("GetText", "#element2").Return("actual text", nil)

	err := suite.executor.ExecuteAssertions(assertions)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(AssertionError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "assertion 1")
}

// TestCaseSensitiveOption tests case sensitive option
func (suite *UIAssertionsTestSuite) TestCaseSensitiveOption() {
	// Test case insensitive (default behavior when case_sensitive is false)
	assertion := UIAssertion{
		Type:     string(AssertTextEquals),
		Selector: "#element",
		Expected: "EXPECTED TEXT",
		Options: map[string]interface{}{
			"case_sensitive": false,
		},
	}

	suite.mockTester.On("GetText", "#element").Return("expected text", nil)

	err := suite.executor.ExecuteAssertion(assertion)
	suite.NoError(err)
}

// TestUIAssertionsTestSuite runs the test suite
func TestUIAssertionsTestSuite(t *testing.T) {
	suite.Run(t, new(UIAssertionsTestSuite))
}

// TestValidateAssertion tests assertion validation
func TestValidateAssertion(t *testing.T) {
	tests := []struct {
		name        string
		assertion   UIAssertion
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid element present assertion",
			assertion: UIAssertion{
				Type:     string(AssertElementPresent),
				Selector: "#element",
			},
			expectError: false,
		},
		{
			name: "element present assertion without selector",
			assertion: UIAssertion{
				Type: string(AssertElementPresent),
			},
			expectError: true,
			errorMsg:    "selector is required",
		},
		{
			name: "valid text equals assertion",
			assertion: UIAssertion{
				Type:     string(AssertTextEquals),
				Selector: "#element",
				Expected: "text",
			},
			expectError: false,
		},
		{
			name: "text equals assertion without selector",
			assertion: UIAssertion{
				Type:     string(AssertTextEquals),
				Expected: "text",
			},
			expectError: true,
			errorMsg:    "selector is required",
		},
		{
			name: "text equals assertion without expected value",
			assertion: UIAssertion{
				Type:     string(AssertTextEquals),
				Selector: "#element",
			},
			expectError: true,
			errorMsg:    "expected value is required",
		},
		{
			name: "valid URL equals assertion",
			assertion: UIAssertion{
				Type:     string(AssertURLEquals),
				Expected: "https://example.com",
			},
			expectError: false,
		},
		{
			name: "URL equals assertion without expected value",
			assertion: UIAssertion{
				Type: string(AssertURLEquals),
			},
			expectError: true,
			errorMsg:    "expected value is required",
		},
		{
			name: "assertion without type",
			assertion: UIAssertion{
				Selector: "#element",
			},
			expectError: true,
			errorMsg:    "assertion type is required",
		},
		{
			name: "unsupported assertion type",
			assertion: UIAssertion{
				Type: "invalid_assertion",
			},
			expectError: true,
			errorMsg:    "unsupported assertion type",
		},
		{
			name: "valid element count assertion",
			assertion: UIAssertion{
				Type:     string(AssertElementCount),
				Selector: ".items",
				Expected: 5,
			},
			expectError: false,
		},
		{
			name: "element count assertion without selector",
			assertion: UIAssertion{
				Type:     string(AssertElementCount),
				Expected: 5,
			},
			expectError: true,
			errorMsg:    "selector is required",
		},
		{
			name: "element count assertion without expected value",
			assertion: UIAssertion{
				Type:     string(AssertElementCount),
				Selector: ".items",
			},
			expectError: true,
			errorMsg:    "expected count is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssertion(tt.assertion)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAssertionTypes tests assertion type constants
func TestAssertionTypes(t *testing.T) {
	assert.Equal(t, "element_present", string(AssertElementPresent))
	assert.Equal(t, "element_not_present", string(AssertElementNotPresent))
	assert.Equal(t, "element_visible", string(AssertElementVisible))
	assert.Equal(t, "element_not_visible", string(AssertElementNotVisible))
	assert.Equal(t, "text_equals", string(AssertTextEquals))
	assert.Equal(t, "text_contains", string(AssertTextContains))
	assert.Equal(t, "text_not_contains", string(AssertTextNotContains))
	assert.Equal(t, "text_matches", string(AssertTextMatches))
	assert.Equal(t, "attribute_equals", string(AssertAttributeEquals))
	assert.Equal(t, "attribute_contains", string(AssertAttributeContains))
	assert.Equal(t, "url_equals", string(AssertURLEquals))
	assert.Equal(t, "url_contains", string(AssertURLContains))
	assert.Equal(t, "title_equals", string(AssertTitleEquals))
	assert.Equal(t, "title_contains", string(AssertTitleContains))
	assert.Equal(t, "element_count", string(AssertElementCount))
	assert.Equal(t, "page_source_contains", string(AssertPageSourceContains))
}

// TestHelperMethods tests helper methods
func TestHelperMethods(t *testing.T) {
	mockTester := new(MockUITester)
	executor := NewUIAssertionExecutor(mockTester)

	// Test getExpectedString
	assertion1 := UIAssertion{Expected: "test string"}
	result1 := executor.getExpectedString(assertion1)
	assert.Equal(t, "test string", result1)

	assertion2 := UIAssertion{Expected: 123}
	result2 := executor.getExpectedString(assertion2)
	assert.Equal(t, "123", result2)

	// Test isCaseSensitive
	assertion3 := UIAssertion{
		Options: map[string]interface{}{
			"case_sensitive": false,
		},
	}
	result3 := executor.isCaseSensitive(assertion3)
	assert.False(t, result3)

	assertion4 := UIAssertion{} // No options, should default to true
	result4 := executor.isCaseSensitive(assertion4)
	assert.True(t, result4)

	// Test getTimeout
	assertion5 := UIAssertion{
		Options: map[string]interface{}{
			"timeout": "10s",
		},
	}
	result5 := executor.getTimeout(assertion5)
	assert.Equal(t, 10*time.Second, result5)

	assertion6 := UIAssertion{} // No options, should use default
	result6 := executor.getTimeout(assertion6)
	assert.Equal(t, 30*time.Second, result6)
}
