package gowright

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// UIAssertionType represents the type of UI assertion
type UIAssertionType string

const (
	AssertElementPresent    UIAssertionType = "element_present"
	AssertElementNotPresent UIAssertionType = "element_not_present"
	AssertElementVisible    UIAssertionType = "element_visible"
	AssertElementNotVisible UIAssertionType = "element_not_visible"
	AssertTextEquals        UIAssertionType = "text_equals"
	AssertTextContains      UIAssertionType = "text_contains"
	AssertTextNotContains   UIAssertionType = "text_not_contains"
	AssertTextMatches       UIAssertionType = "text_matches"
	AssertAttributeEquals   UIAssertionType = "attribute_equals"
	AssertAttributeContains UIAssertionType = "attribute_contains"
	AssertURLEquals         UIAssertionType = "url_equals"
	AssertURLContains       UIAssertionType = "url_contains"
	AssertTitleEquals       UIAssertionType = "title_equals"
	AssertTitleContains     UIAssertionType = "title_contains"
	AssertElementCount      UIAssertionType = "element_count"
	AssertPageSourceContains UIAssertionType = "page_source_contains"
)

// UIAssertionOptions holds additional options for UI assertions
type UIAssertionOptions struct {
	Timeout       time.Duration          `json:"timeout,omitempty"`
	CaseSensitive bool                   `json:"case_sensitive,omitempty"`
	Attribute     string                 `json:"attribute,omitempty"`
	Regex         bool                   `json:"regex,omitempty"`
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// UIAssertionExecutor executes UI assertions using the UITester
type UIAssertionExecutor struct {
	tester UITester
}

// NewUIAssertionExecutor creates a new UIAssertionExecutor
func NewUIAssertionExecutor(tester UITester) *UIAssertionExecutor {
	return &UIAssertionExecutor{
		tester: tester,
	}
}

// ExecuteAssertion executes a UI assertion
func (e *UIAssertionExecutor) ExecuteAssertion(assertion UIAssertion) error {
	assertionType := UIAssertionType(assertion.Type)
	
	switch assertionType {
	case AssertElementPresent:
		return e.assertElementPresent(assertion)
	case AssertElementNotPresent:
		return e.assertElementNotPresent(assertion)
	case AssertElementVisible:
		return e.assertElementVisible(assertion)
	case AssertElementNotVisible:
		return e.assertElementNotVisible(assertion)
	case AssertTextEquals:
		return e.assertTextEquals(assertion)
	case AssertTextContains:
		return e.assertTextContains(assertion)
	case AssertTextNotContains:
		return e.assertTextNotContains(assertion)
	case AssertTextMatches:
		return e.assertTextMatches(assertion)
	case AssertAttributeEquals:
		return e.assertAttributeEquals(assertion)
	case AssertAttributeContains:
		return e.assertAttributeContains(assertion)
	case AssertURLEquals:
		return e.assertURLEquals(assertion)
	case AssertURLContains:
		return e.assertURLContains(assertion)
	case AssertTitleEquals:
		return e.assertTitleEquals(assertion)
	case AssertTitleContains:
		return e.assertTitleContains(assertion)
	case AssertElementCount:
		return e.assertElementCount(assertion)
	case AssertPageSourceContains:
		return e.assertPageSourceContains(assertion)
	default:
		return NewGowrightError(AssertionError, fmt.Sprintf("unsupported assertion type: %s", assertion.Type), nil)
	}
}

// assertElementPresent asserts that an element is present on the page
func (e *UIAssertionExecutor) assertElementPresent(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for element_present assertion", nil)
	}

	timeout := e.getTimeout(assertion)
	err := e.tester.WaitForElement(assertion.Selector, timeout)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' is not present", assertion.Selector), err)
	}
	return nil
}

// assertElementNotPresent asserts that an element is not present on the page
func (e *UIAssertionExecutor) assertElementNotPresent(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for element_not_present assertion", nil)
	}

	// Check if element is present using a short timeout
	if rodTester, ok := e.tester.(*RodUITester); ok {
		present, err := rodTester.IsElementPresent(assertion.Selector)
		if err != nil {
			return NewGowrightError(AssertionError, "failed to check element presence", err)
		}
		if present {
			return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' should not be present but was found", assertion.Selector), nil)
		}
	} else {
		// Fallback: try to wait for element with very short timeout
		err := e.tester.WaitForElement(assertion.Selector, 100*time.Millisecond)
		if err == nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' should not be present but was found", assertion.Selector), nil)
		}
	}
	return nil
}

// assertElementVisible asserts that an element is visible on the page
func (e *UIAssertionExecutor) assertElementVisible(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for element_visible assertion", nil)
	}

	if rodTester, ok := e.tester.(*RodUITester); ok {
		visible, err := rodTester.IsElementVisible(assertion.Selector)
		if err != nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("failed to check visibility of element '%s'", assertion.Selector), err)
		}
		if !visible {
			return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' is not visible", assertion.Selector), nil)
		}
	} else {
		// Fallback: if element is present, assume it's visible
		err := e.assertElementPresent(assertion)
		if err != nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' is not visible (not present)", assertion.Selector), err)
		}
	}
	return nil
}

// assertElementNotVisible asserts that an element is not visible on the page
func (e *UIAssertionExecutor) assertElementNotVisible(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for element_not_visible assertion", nil)
	}

	if rodTester, ok := e.tester.(*RodUITester); ok {
		visible, err := rodTester.IsElementVisible(assertion.Selector)
		if err != nil {
			// If we can't check visibility, element might not be present, which is fine
			return nil
		}
		if visible {
			return NewGowrightError(AssertionError, fmt.Sprintf("element with selector '%s' should not be visible but is visible", assertion.Selector), nil)
		}
	}
	return nil
}

// assertTextEquals asserts that an element's text equals the expected value
func (e *UIAssertionExecutor) assertTextEquals(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for text_equals assertion", nil)
	}

	actualText, err := e.tester.GetText(assertion.Selector)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("failed to get text from element '%s'", assertion.Selector), err)
	}

	expectedText := e.getExpectedString(assertion)
	caseSensitive := e.isCaseSensitive(assertion)

	if !caseSensitive {
		actualText = strings.ToLower(actualText)
		expectedText = strings.ToLower(expectedText)
	}

	if actualText != expectedText {
		return NewGowrightError(AssertionError, fmt.Sprintf("text mismatch for element '%s': expected '%s', got '%s'", assertion.Selector, expectedText, actualText), nil)
	}
	return nil
}

// assertTextContains asserts that an element's text contains the expected value
func (e *UIAssertionExecutor) assertTextContains(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for text_contains assertion", nil)
	}

	actualText, err := e.tester.GetText(assertion.Selector)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("failed to get text from element '%s'", assertion.Selector), err)
	}

	expectedText := e.getExpectedString(assertion)
	caseSensitive := e.isCaseSensitive(assertion)

	if !caseSensitive {
		actualText = strings.ToLower(actualText)
		expectedText = strings.ToLower(expectedText)
	}

	if !strings.Contains(actualText, expectedText) {
		return NewGowrightError(AssertionError, fmt.Sprintf("text does not contain expected value for element '%s': expected to contain '%s', got '%s'", assertion.Selector, expectedText, actualText), nil)
	}
	return nil
}

// assertTextNotContains asserts that an element's text does not contain the expected value
func (e *UIAssertionExecutor) assertTextNotContains(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for text_not_contains assertion", nil)
	}

	actualText, err := e.tester.GetText(assertion.Selector)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("failed to get text from element '%s'", assertion.Selector), err)
	}

	expectedText := e.getExpectedString(assertion)
	caseSensitive := e.isCaseSensitive(assertion)

	if !caseSensitive {
		actualText = strings.ToLower(actualText)
		expectedText = strings.ToLower(expectedText)
	}

	if strings.Contains(actualText, expectedText) {
		return NewGowrightError(AssertionError, fmt.Sprintf("text should not contain expected value for element '%s': should not contain '%s', but got '%s'", assertion.Selector, expectedText, actualText), nil)
	}
	return nil
}

// assertTextMatches asserts that an element's text matches a regular expression
func (e *UIAssertionExecutor) assertTextMatches(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for text_matches assertion", nil)
	}

	actualText, err := e.tester.GetText(assertion.Selector)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("failed to get text from element '%s'", assertion.Selector), err)
	}

	pattern := e.getExpectedString(assertion)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return NewGowrightError(AssertionError, fmt.Sprintf("invalid regular expression pattern: %s", pattern), err)
	}

	if !regex.MatchString(actualText) {
		return NewGowrightError(AssertionError, fmt.Sprintf("text does not match pattern for element '%s': pattern '%s', got '%s'", assertion.Selector, pattern, actualText), nil)
	}
	return nil
}

// assertAttributeEquals asserts that an element's attribute equals the expected value
func (e *UIAssertionExecutor) assertAttributeEquals(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for attribute_equals assertion", nil)
	}

	attribute := e.getAttribute(assertion)
	if attribute == "" {
		return NewGowrightError(AssertionError, "attribute name is required for attribute_equals assertion", nil)
	}

	// This would require extending the UITester interface to support attribute retrieval
	return NewGowrightError(AssertionError, "attribute assertions not implemented - requires UITester interface extension", nil)
}

// assertAttributeContains asserts that an element's attribute contains the expected value
func (e *UIAssertionExecutor) assertAttributeContains(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for attribute_contains assertion", nil)
	}

	attribute := e.getAttribute(assertion)
	if attribute == "" {
		return NewGowrightError(AssertionError, "attribute name is required for attribute_contains assertion", nil)
	}

	// This would require extending the UITester interface to support attribute retrieval
	return NewGowrightError(AssertionError, "attribute assertions not implemented - requires UITester interface extension", nil)
}

// assertURLEquals asserts that the current URL equals the expected value
func (e *UIAssertionExecutor) assertURLEquals(assertion UIAssertion) error {
	if rodTester, ok := e.tester.(*RodUITester); ok {
		currentURL, err := rodTester.GetCurrentURL()
		if err != nil {
			return NewGowrightError(AssertionError, "failed to get current URL", err)
		}

		expectedURL := e.getExpectedString(assertion)
		if currentURL != expectedURL {
			return NewGowrightError(AssertionError, fmt.Sprintf("URL mismatch: expected '%s', got '%s'", expectedURL, currentURL), nil)
		}
	} else {
		return NewGowrightError(AssertionError, "URL assertions not supported by this tester", nil)
	}
	return nil
}

// assertURLContains asserts that the current URL contains the expected value
func (e *UIAssertionExecutor) assertURLContains(assertion UIAssertion) error {
	if rodTester, ok := e.tester.(*RodUITester); ok {
		currentURL, err := rodTester.GetCurrentURL()
		if err != nil {
			return NewGowrightError(AssertionError, "failed to get current URL", err)
		}

		expectedText := e.getExpectedString(assertion)
		caseSensitive := e.isCaseSensitive(assertion)

		if !caseSensitive {
			currentURL = strings.ToLower(currentURL)
			expectedText = strings.ToLower(expectedText)
		}

		if !strings.Contains(currentURL, expectedText) {
			return NewGowrightError(AssertionError, fmt.Sprintf("URL does not contain expected value: expected to contain '%s', got '%s'", expectedText, currentURL), nil)
		}
	} else {
		return NewGowrightError(AssertionError, "URL assertions not supported by this tester", nil)
	}
	return nil
}

// assertTitleEquals asserts that the page title equals the expected value
func (e *UIAssertionExecutor) assertTitleEquals(assertion UIAssertion) error {
	// This would require extending the UITester interface to support title retrieval
	return NewGowrightError(AssertionError, "title assertions not implemented - requires UITester interface extension", nil)
}

// assertTitleContains asserts that the page title contains the expected value
func (e *UIAssertionExecutor) assertTitleContains(assertion UIAssertion) error {
	// This would require extending the UITester interface to support title retrieval
	return NewGowrightError(AssertionError, "title assertions not implemented - requires UITester interface extension", nil)
}

// assertElementCount asserts that the number of elements matching the selector equals the expected count
func (e *UIAssertionExecutor) assertElementCount(assertion UIAssertion) error {
	if assertion.Selector == "" {
		return NewGowrightError(AssertionError, "selector is required for element_count assertion", nil)
	}

	// This would require extending the UITester interface to support element counting
	return NewGowrightError(AssertionError, "element count assertions not implemented - requires UITester interface extension", nil)
}

// assertPageSourceContains asserts that the page source contains the expected value
func (e *UIAssertionExecutor) assertPageSourceContains(assertion UIAssertion) error {
	pageSource, err := e.tester.GetPageSource()
	if err != nil {
		return NewGowrightError(AssertionError, "failed to get page source", err)
	}

	expectedText := e.getExpectedString(assertion)
	caseSensitive := e.isCaseSensitive(assertion)

	if !caseSensitive {
		pageSource = strings.ToLower(pageSource)
		expectedText = strings.ToLower(expectedText)
	}

	if !strings.Contains(pageSource, expectedText) {
		return NewGowrightError(AssertionError, fmt.Sprintf("page source does not contain expected value: expected to contain '%s'", expectedText), nil)
	}
	return nil
}

// Helper methods

// getTimeout extracts timeout from assertion options
func (e *UIAssertionExecutor) getTimeout(assertion UIAssertion) time.Duration {
	defaultTimeout := 30 * time.Second
	
	if assertion.Options != nil {
		if optsMap, ok := assertion.Options.(map[string]interface{}); ok {
			if timeoutVal, exists := optsMap["timeout"]; exists {
				if timeoutStr, ok := timeoutVal.(string); ok {
					if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
						return parsedTimeout
					}
				} else if timeoutDur, ok := timeoutVal.(time.Duration); ok {
					return timeoutDur
				}
			}
		}
	}
	
	return defaultTimeout
}

// getExpectedString extracts expected string value from assertion
func (e *UIAssertionExecutor) getExpectedString(assertion UIAssertion) string {
	if str, ok := assertion.Expected.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", assertion.Expected)
}

// isCaseSensitive checks if the assertion should be case sensitive
func (e *UIAssertionExecutor) isCaseSensitive(assertion UIAssertion) bool {
	if assertion.Options != nil {
		if optsMap, ok := assertion.Options.(map[string]interface{}); ok {
			if caseSensitive, exists := optsMap["case_sensitive"]; exists {
				if cs, ok := caseSensitive.(bool); ok {
					return cs
				}
			}
		}
	}
	return true // default to case sensitive
}

// getAttribute extracts attribute name from assertion options
func (e *UIAssertionExecutor) getAttribute(assertion UIAssertion) string {
	if assertion.Options != nil {
		if optsMap, ok := assertion.Options.(map[string]interface{}); ok {
			if attribute, exists := optsMap["attribute"]; exists {
				if attr, ok := attribute.(string); ok {
					return attr
				}
			}
		}
	}
	return ""
}

// getExpectedInt extracts expected integer value from assertion
func (e *UIAssertionExecutor) getExpectedInt(assertion UIAssertion) (int, error) {
	if intVal, ok := assertion.Expected.(int); ok {
		return intVal, nil
	}
	if strVal, ok := assertion.Expected.(string); ok {
		return strconv.Atoi(strVal)
	}
	if floatVal, ok := assertion.Expected.(float64); ok {
		return int(floatVal), nil
	}
	return 0, fmt.Errorf("expected value is not a valid integer: %v", assertion.Expected)
}

// ExecuteAssertions executes a sequence of UI assertions
func (e *UIAssertionExecutor) ExecuteAssertions(assertions []UIAssertion) error {
	for i, assertion := range assertions {
		if err := e.ExecuteAssertion(assertion); err != nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("assertion %d (%s) failed", i, assertion.Type), err)
		}
	}
	return nil
}

// ValidateAssertion validates that an assertion has the required fields
func ValidateAssertion(assertion UIAssertion) error {
	if assertion.Type == "" {
		return NewGowrightError(AssertionError, "assertion type is required", nil)
	}

	assertionType := UIAssertionType(assertion.Type)
	
	switch assertionType {
	case AssertElementPresent, AssertElementNotPresent, AssertElementVisible, AssertElementNotVisible:
		if assertion.Selector == "" {
			return NewGowrightError(AssertionError, fmt.Sprintf("selector is required for %s assertion", assertion.Type), nil)
		}
	case AssertTextEquals, AssertTextContains, AssertTextNotContains, AssertTextMatches:
		if assertion.Selector == "" {
			return NewGowrightError(AssertionError, fmt.Sprintf("selector is required for %s assertion", assertion.Type), nil)
		}
		if assertion.Expected == nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("expected value is required for %s assertion", assertion.Type), nil)
		}
	case AssertAttributeEquals, AssertAttributeContains:
		if assertion.Selector == "" {
			return NewGowrightError(AssertionError, fmt.Sprintf("selector is required for %s assertion", assertion.Type), nil)
		}
		if assertion.Expected == nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("expected value is required for %s assertion", assertion.Type), nil)
		}
		// Should also validate that attribute is specified in options
	case AssertURLEquals, AssertURLContains, AssertTitleEquals, AssertTitleContains, AssertPageSourceContains:
		if assertion.Expected == nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("expected value is required for %s assertion", assertion.Type), nil)
		}
	case AssertElementCount:
		if assertion.Selector == "" {
			return NewGowrightError(AssertionError, "selector is required for element_count assertion", nil)
		}
		if assertion.Expected == nil {
			return NewGowrightError(AssertionError, "expected count is required for element_count assertion", nil)
		}
	default:
		return NewGowrightError(AssertionError, fmt.Sprintf("unsupported assertion type: %s", assertion.Type), nil)
	}

	return nil
}