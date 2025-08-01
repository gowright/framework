package gowright

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Note: MockUITester is defined in gowright_test.go

// UIActionsTestSuite defines the test suite for UI actions
type UIActionsTestSuite struct {
	suite.Suite
	mockTester *MockUITester
	executor   *UIActionExecutor
}

// SetupTest runs before each test
func (suite *UIActionsTestSuite) SetupTest() {
	suite.mockTester = new(MockUITester)
	suite.executor = NewUIActionExecutor(suite.mockTester)
}

// TearDownTest runs after each test
func (suite *UIActionsTestSuite) TearDownTest() {
	suite.mockTester.AssertExpectations(suite.T())
}

// TestNewUIActionExecutor tests the constructor
func (suite *UIActionsTestSuite) TestNewUIActionExecutor() {
	executor := NewUIActionExecutor(suite.mockTester)
	suite.NotNil(executor)
	suite.Equal(suite.mockTester, executor.tester)
}

// TestExecuteClickAction tests click action execution
func (suite *UIActionsTestSuite) TestExecuteClickAction() {
	action := UIAction{
		Type:     string(ActionClick),
		Selector: "#button",
	}
	
	suite.mockTester.On("Click", "#button").Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteClickActionWithoutSelector tests click action without selector
func (suite *UIActionsTestSuite) TestExecuteClickActionWithoutSelector() {
	action := UIAction{
		Type: string(ActionClick),
	}
	
	err := suite.executor.ExecuteAction(action)
	suite.Error(err)
	
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "selector is required")
}

// TestExecuteTypeAction tests type action execution
func (suite *UIActionsTestSuite) TestExecuteTypeAction() {
	action := UIAction{
		Type:     string(ActionType),
		Selector: "#input",
		Value:    "test text",
	}
	
	suite.mockTester.On("Type", "#input", "test text").Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteTypeActionWithoutSelector tests type action without selector
func (suite *UIActionsTestSuite) TestExecuteTypeActionWithoutSelector() {
	action := UIAction{
		Type:  string(ActionType),
		Value: "test text",
	}
	
	err := suite.executor.ExecuteAction(action)
	suite.Error(err)
	
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "selector is required")
}

// TestExecuteNavigateAction tests navigate action execution
func (suite *UIActionsTestSuite) TestExecuteNavigateAction() {
	action := UIAction{
		Type:  string(ActionNavigate),
		Value: "https://example.com",
	}
	
	suite.mockTester.On("Navigate", "https://example.com").Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteNavigateActionWithoutURL tests navigate action without URL
func (suite *UIActionsTestSuite) TestExecuteNavigateActionWithoutURL() {
	action := UIAction{
		Type: string(ActionNavigate),
	}
	
	err := suite.executor.ExecuteAction(action)
	suite.Error(err)
	
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "URL is required")
}

// TestExecuteWaitActionWithSelector tests wait action with selector
func (suite *UIActionsTestSuite) TestExecuteWaitActionWithSelector() {
	action := UIAction{
		Type:     string(ActionWait),
		Selector: "#element",
	}
	
	suite.mockTester.On("WaitForElement", "#element", 30*time.Second).Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteWaitActionWithoutSelector tests wait action without selector
func (suite *UIActionsTestSuite) TestExecuteWaitActionWithoutSelector() {
	action := UIAction{
		Type: string(ActionWait),
	}
	
	// This should not call any mock methods, just sleep
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteClearAction tests clear action execution
func (suite *UIActionsTestSuite) TestExecuteClearAction() {
	action := UIAction{
		Type:     string(ActionClear),
		Selector: "#input",
	}
	
	suite.mockTester.On("Type", "#input", "").Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteSubmitAction tests submit action execution
func (suite *UIActionsTestSuite) TestExecuteSubmitAction() {
	action := UIAction{
		Type:     string(ActionSubmit),
		Selector: "#submit-button",
	}
	
	suite.mockTester.On("Click", "#submit-button").Return(nil)
	
	err := suite.executor.ExecuteAction(action)
	suite.NoError(err)
}

// TestExecuteUnsupportedAction tests unsupported action type
func (suite *UIActionsTestSuite) TestExecuteUnsupportedAction() {
	action := UIAction{
		Type: "unsupported_action",
	}
	
	err := suite.executor.ExecuteAction(action)
	suite.Error(err)
	
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "unsupported action type")
}

// TestExecuteActions tests executing multiple actions
func (suite *UIActionsTestSuite) TestExecuteActions() {
	actions := []UIAction{
		{
			Type:  string(ActionNavigate),
			Value: "https://example.com",
		},
		{
			Type:     string(ActionClick),
			Selector: "#button",
		},
		{
			Type:     string(ActionType),
			Selector: "#input",
			Value:    "test",
		},
	}
	
	suite.mockTester.On("Navigate", "https://example.com").Return(nil)
	suite.mockTester.On("Click", "#button").Return(nil)
	suite.mockTester.On("Type", "#input", "test").Return(nil)
	
	err := suite.executor.ExecuteActions(actions)
	suite.NoError(err)
}

// TestExecuteActionsWithFailure tests executing actions with one failing
func (suite *UIActionsTestSuite) TestExecuteActionsWithFailure() {
	actions := []UIAction{
		{
			Type:  string(ActionNavigate),
			Value: "https://example.com",
		},
		{
			Type:     string(ActionClick),
			Selector: "#button",
		},
	}
	
	suite.mockTester.On("Navigate", "https://example.com").Return(nil)
	suite.mockTester.On("Click", "#button").Return(NewGowrightError(BrowserError, "element not found", nil))
	
	err := suite.executor.ExecuteActions(actions)
	suite.Error(err)
	
	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "failed to execute action 1")
}

// TestUIActionsTestSuite runs the test suite
func TestUIActionsTestSuite(t *testing.T) {
	suite.Run(t, new(UIActionsTestSuite))
}

// TestValidateAction tests action validation
func TestValidateAction(t *testing.T) {
	tests := []struct {
		name        string
		action      UIAction
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid click action",
			action: UIAction{
				Type:     string(ActionClick),
				Selector: "#button",
			},
			expectError: false,
		},
		{
			name: "click action without selector",
			action: UIAction{
				Type: string(ActionClick),
			},
			expectError: true,
			errorMsg:    "selector is required",
		},
		{
			name: "valid type action",
			action: UIAction{
				Type:     string(ActionType),
				Selector: "#input",
				Value:    "text",
			},
			expectError: false,
		},
		{
			name: "type action without selector",
			action: UIAction{
				Type:  string(ActionType),
				Value: "text",
			},
			expectError: true,
			errorMsg:    "selector is required",
		},
		{
			name: "type action without value",
			action: UIAction{
				Type:     string(ActionType),
				Selector: "#input",
			},
			expectError: true,
			errorMsg:    "value is required",
		},
		{
			name: "valid navigate action",
			action: UIAction{
				Type:  string(ActionNavigate),
				Value: "https://example.com",
			},
			expectError: false,
		},
		{
			name: "navigate action without URL",
			action: UIAction{
				Type: string(ActionNavigate),
			},
			expectError: true,
			errorMsg:    "URL is required",
		},
		{
			name: "valid wait action with selector",
			action: UIAction{
				Type:     string(ActionWait),
				Selector: "#element",
			},
			expectError: false,
		},
		{
			name: "valid wait action without selector",
			action: UIAction{
				Type: string(ActionWait),
			},
			expectError: false,
		},
		{
			name: "action without type",
			action: UIAction{
				Selector: "#element",
			},
			expectError: true,
			errorMsg:    "action type is required",
		},
		{
			name: "unsupported action type",
			action: UIAction{
				Type: "invalid_action",
			},
			expectError: true,
			errorMsg:    "unsupported action type",
		},
		{
			name: "valid refresh action",
			action: UIAction{
				Type: string(ActionRefresh),
			},
			expectError: false,
		},
		{
			name: "valid scroll action with selector",
			action: UIAction{
				Type:     string(ActionScroll),
				Selector: "#element",
			},
			expectError: false,
		},
		{
			name: "valid scroll action without selector",
			action: UIAction{
				Type: string(ActionScroll),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(tt.action)
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

// TestActionTypes tests action type constants
func TestActionTypes(t *testing.T) {
	assert.Equal(t, "click", string(ActionClick))
	assert.Equal(t, "type", string(ActionType))
	assert.Equal(t, "navigate", string(ActionNavigate))
	assert.Equal(t, "wait", string(ActionWait))
	assert.Equal(t, "scroll", string(ActionScroll))
	assert.Equal(t, "hover", string(ActionHover))
	assert.Equal(t, "select", string(ActionSelect))
	assert.Equal(t, "clear", string(ActionClear))
	assert.Equal(t, "submit", string(ActionSubmit))
	assert.Equal(t, "refresh", string(ActionRefresh))
	assert.Equal(t, "go_back", string(ActionGoBack))
	assert.Equal(t, "go_forward", string(ActionGoForward))
}