package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// UIAssertionType represents the type of UI assertion
type UIAssertionType string

const (
	AssertElementPresent     UIAssertionType = "element_present"
	AssertElementNotPresent  UIAssertionType = "element_not_present"
	AssertElementVisible     UIAssertionType = "element_visible"
	AssertElementNotVisible  UIAssertionType = "element_not_visible"
	AssertTextEquals         UIAssertionType = "text_equals"
	AssertTextContains       UIAssertionType = "text_contains"
	AssertTextNotContains    UIAssertionType = "text_not_contains"
	AssertTextMatches        UIAssertionType = "text_matches"
	AssertAttributeEquals    UIAssertionType = "attribute_equals"
	AssertAttributeContains  UIAssertionType = "attribute_contains"
	AssertURLEquals          UIAssertionType = "url_equals"
	AssertURLContains        UIAssertionType = "url_contains"
	AssertTitleEquals        UIAssertionType = "title_equals"
	AssertTitleContains      UIAssertionType = "title_contains"
	AssertElementCount       UIAssertionType = "element_count"
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
	tester core.UITester
}

// NewUIAssertionExecutor creates a new UIAssertionExecutor
func NewUIAssertionExecutor(tester core.UITester) *UIAssertionExecutor {
	return &UIAssertionExecutor{
		tester: tester,
	}
}

// ExecuteAssertion executes a UI assertion
func (uae *UIAssertionExecutor) ExecuteAssertion(assertion *core.UIAssertion) error {
	switch UIAssertionType(assertion.Type) {
	case AssertElementPresent:
		return uae.assertElementPresent(assertion.Selector, true)
	case AssertElementNotPresent:
		return uae.assertElementPresent(assertion.Selector, false)
	case AssertElementVisible:
		return uae.assertElementVisible(assertion.Selector, true)
	case AssertElementNotVisible:
		return uae.assertElementVisible(assertion.Selector, false)
	case AssertTextEquals:
		return uae.assertTextEquals(assertion.Selector, assertion.Expected.(string), false)
	case AssertTextContains:
		return uae.assertTextContains(assertion.Selector, assertion.Expected.(string), true)
	case AssertTextNotContains:
		return uae.assertTextContains(assertion.Selector, assertion.Expected.(string), false)
	case AssertTextMatches:
		return uae.assertTextMatches(assertion.Selector, assertion.Expected.(string))
	default:
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("unsupported assertion type: %s", assertion.Type), nil)
	}
}

// assertElementPresent checks if an element is present
func (uae *UIAssertionExecutor) assertElementPresent(selector string, shouldBePresent bool) error {
	// Implementation would check element presence
	// For now, this is a placeholder
	return nil
}

// assertElementVisible checks if an element is visible
func (uae *UIAssertionExecutor) assertElementVisible(selector string, shouldBeVisible bool) error {
	// Implementation would check element visibility
	// For now, this is a placeholder
	return nil
}

// assertTextEquals checks if element text equals expected value
func (uae *UIAssertionExecutor) assertTextEquals(selector, expected string, caseSensitive bool) error {
	text, err := uae.tester.GetText(selector)
	if err != nil {
		return err
	}

	if !caseSensitive {
		text = strings.ToLower(text)
		expected = strings.ToLower(expected)
	}

	if text != expected {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("expected text '%s', got '%s'", expected, text), nil)
	}

	return nil
}

// assertTextContains checks if element text contains expected value
func (uae *UIAssertionExecutor) assertTextContains(selector, expected string, shouldContain bool) error {
	text, err := uae.tester.GetText(selector)
	if err != nil {
		return err
	}

	contains := strings.Contains(text, expected)
	if shouldContain && !contains {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("expected text to contain '%s', got '%s'", expected, text), nil)
	}
	if !shouldContain && contains {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("expected text to not contain '%s', got '%s'", expected, text), nil)
	}

	return nil
}

// assertTextMatches checks if element text matches a regex pattern
func (uae *UIAssertionExecutor) assertTextMatches(selector, pattern string) error {
	text, err := uae.tester.GetText(selector)
	if err != nil {
		return err
	}

	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("invalid regex pattern '%s': %v", pattern, err), err)
	}

	if !matched {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("text '%s' does not match pattern '%s'", text, pattern), nil)
	}

	return nil
}
