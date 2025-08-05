package gowright

import (
	"fmt"
	"time"
)

// UIActionType represents the type of UI action
type UIActionType string

const (
	ActionClick     UIActionType = "click"
	ActionType      UIActionType = "type"
	ActionNavigate  UIActionType = "navigate"
	ActionWait      UIActionType = "wait"
	ActionScroll    UIActionType = "scroll"
	ActionHover     UIActionType = "hover"
	ActionSelect    UIActionType = "select"
	ActionClear     UIActionType = "clear"
	ActionSubmit    UIActionType = "submit"
	ActionRefresh   UIActionType = "refresh"
	ActionGoBack    UIActionType = "go_back"
	ActionGoForward UIActionType = "go_forward"
	// Mobile-specific actions
	ActionTap            UIActionType = "tap"
	ActionSwipe          UIActionType = "swipe"
	ActionSwipeLeft      UIActionType = "swipe_left"
	ActionSwipeRight     UIActionType = "swipe_right"
	ActionSwipeUp        UIActionType = "swipe_up"
	ActionSwipeDown      UIActionType = "swipe_down"
	ActionLongPress      UIActionType = "long_press"
	ActionPinch          UIActionType = "pinch"
	ActionSetOrientation UIActionType = "set_orientation"
)

// UIActionOptions holds additional options for UI actions
type UIActionOptions struct {
	Timeout       time.Duration          `json:"timeout,omitempty"`
	WaitCondition string                 `json:"wait_condition,omitempty"`
	ScrollOptions *ScrollOptions         `json:"scroll_options,omitempty"`
	SelectOptions *SelectOptions         `json:"select_options,omitempty"`
	ClickOptions  *ClickOptions          `json:"click_options,omitempty"`
	TypeOptions   *TypeOptions           `json:"type_options,omitempty"`
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// ScrollOptions holds options for scroll actions
type ScrollOptions struct {
	Direction string `json:"direction"` // up, down, left, right, top, bottom
	Distance  int    `json:"distance,omitempty"`
}

// SelectOptions holds options for select actions
type SelectOptions struct {
	ByValue bool `json:"by_value"` // true for value, false for text
	Index   int  `json:"index,omitempty"`
}

// ClickOptions holds options for click actions
type ClickOptions struct {
	DoubleClick bool `json:"double_click"`
	RightClick  bool `json:"right_click"`
	Force       bool `json:"force"` // Force click even if element is not visible
}

// TypeOptions holds options for type actions
type TypeOptions struct {
	ClearFirst bool          `json:"clear_first"`
	Delay      time.Duration `json:"delay,omitempty"` // Delay between keystrokes
}

// SwipeOptions holds options for swipe actions
type SwipeOptions struct {
	StartX   int           `json:"start_x,omitempty"`
	StartY   int           `json:"start_y,omitempty"`
	EndX     int           `json:"end_x,omitempty"`
	EndY     int           `json:"end_y,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// LongPressOptions holds options for long press actions
type LongPressOptions struct {
	Duration time.Duration `json:"duration,omitempty"`
}

// PinchOptions holds options for pinch actions
type PinchOptions struct {
	CenterX int     `json:"center_x,omitempty"`
	CenterY int     `json:"center_y,omitempty"`
	Scale   float64 `json:"scale,omitempty"`
}

// UIActionExecutor executes UI actions using the UITester
type UIActionExecutor struct {
	tester UITester
}

// NewUIActionExecutor creates a new UIActionExecutor
func NewUIActionExecutor(tester UITester) *UIActionExecutor {
	return &UIActionExecutor{
		tester: tester,
	}
}

// ExecuteAction executes a UI action
func (e *UIActionExecutor) ExecuteAction(action UIAction) error {
	switch UIActionType(action.Type) {
	case ActionClick:
		return e.executeClick(action)
	case ActionType:
		return e.executeType(action)
	case ActionNavigate:
		return e.executeNavigate(action)
	case ActionWait:
		return e.executeWait(action)
	case ActionScroll:
		return e.executeScroll(action)
	case ActionHover:
		return e.executeHover(action)
	case ActionSelect:
		return e.executeSelect(action)
	case ActionClear:
		return e.executeClear(action)
	case ActionSubmit:
		return e.executeSubmit(action)
	case ActionRefresh:
		return e.executeRefresh(action)
	case ActionGoBack:
		return e.executeGoBack(action)
	case ActionGoForward:
		return e.executeGoForward(action)
	// Mobile-specific actions
	case ActionTap:
		return e.executeTap(action)
	case ActionSwipe:
		return e.executeSwipe(action)
	case ActionSwipeLeft:
		return e.executeSwipeLeft(action)
	case ActionSwipeRight:
		return e.executeSwipeRight(action)
	case ActionSwipeUp:
		return e.executeSwipeUp(action)
	case ActionSwipeDown:
		return e.executeSwipeDown(action)
	case ActionLongPress:
		return e.executeLongPress(action)
	case ActionPinch:
		return e.executePinch(action)
	case ActionSetOrientation:
		return e.executeSetOrientation(action)
	default:
		return NewGowrightError(BrowserError, fmt.Sprintf("unsupported action type: %s", action.Type), nil)
	}
}

// executeClick executes a click action
func (e *UIActionExecutor) executeClick(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for click action", nil)
	}

	// Parse options
	if action.Options != nil {
		if _, ok := action.Options.(*ClickOptions); ok {
			// options = clickOpts - unused for now
		} else if optsMap, ok := action.Options.(map[string]interface{}); ok {
			_ = optsMap // Parse options when needed
			// For now, just use basic click functionality
		}
	}

	// For now, use the basic click method from UITester
	// In a more advanced implementation, we would handle double-click, right-click, etc.
	return e.tester.Click(action.Selector)
}

// executeType executes a type action
func (e *UIActionExecutor) executeType(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for type action", nil)
	}

	return e.tester.Type(action.Selector, action.Value)
}

// executeNavigate executes a navigate action
func (e *UIActionExecutor) executeNavigate(action UIAction) error {
	if action.Value == "" {
		return NewGowrightError(BrowserError, "URL is required for navigate action", nil)
	}

	return e.tester.Navigate(action.Value)
}

// executeWait executes a wait action
func (e *UIActionExecutor) executeWait(action UIAction) error {
	timeout := 30 * time.Second // default timeout

	// Parse timeout from options
	if action.Options != nil {
		if optsMap, ok := action.Options.(map[string]interface{}); ok {
			if timeoutVal, exists := optsMap["timeout"]; exists {
				if timeoutStr, ok := timeoutVal.(string); ok {
					if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
						timeout = parsedTimeout
					}
				} else if timeoutDur, ok := timeoutVal.(time.Duration); ok {
					timeout = timeoutDur
				}
			}
		}
	}

	if action.Selector != "" {
		return e.tester.WaitForElement(action.Selector, timeout)
	}

	// If no selector, just wait for the specified duration
	time.Sleep(timeout)
	return nil
}

// executeScroll executes a scroll action
func (e *UIActionExecutor) executeScroll(action UIAction) error {
	if action.Selector != "" {
		// Scroll to element
		if rodTester, ok := e.tester.(*RodUITester); ok {
			return rodTester.ScrollToElement(action.Selector)
		}
		return NewGowrightError(BrowserError, "scroll to element not supported by this tester", nil)
	}

	// For general page scrolling, we would need to extend the UITester interface
	// For now, return an error indicating this is not implemented
	return NewGowrightError(BrowserError, "general page scrolling not implemented", nil)
}

// executeHover executes a hover action
func (e *UIActionExecutor) executeHover(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for hover action", nil)
	}

	// Hover is not implemented in the basic UITester interface
	// This would require extending the interface or using rod-specific methods
	return NewGowrightError(BrowserError, "hover action not implemented", nil)
}

// executeSelect executes a select action
func (e *UIActionExecutor) executeSelect(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for select action", nil)
	}

	// Select is not implemented in the basic UITester interface
	// This would require extending the interface
	return NewGowrightError(BrowserError, "select action not implemented", nil)
}

// executeClear executes a clear action
func (e *UIActionExecutor) executeClear(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for clear action", nil)
	}

	// Clear by typing empty string (this will select all and replace)
	return e.tester.Type(action.Selector, "")
}

// executeSubmit executes a submit action
func (e *UIActionExecutor) executeSubmit(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for submit action", nil)
	}

	// Submit by clicking the element (assuming it's a submit button or form)
	return e.tester.Click(action.Selector)
}

// executeRefresh executes a refresh action
func (e *UIActionExecutor) executeRefresh(action UIAction) error {
	// Get current URL and navigate to it again
	if rodTester, ok := e.tester.(*RodUITester); ok {
		currentURL, err := rodTester.GetCurrentURL()
		if err != nil {
			return err
		}
		return e.tester.Navigate(currentURL)
	}
	return NewGowrightError(BrowserError, "refresh not supported by this tester", nil)
}

// executeGoBack executes a go back action
func (e *UIActionExecutor) executeGoBack(action UIAction) error {
	// Go back is not implemented in the basic UITester interface
	return NewGowrightError(BrowserError, "go back action not implemented", nil)
}

// executeGoForward executes a go forward action
func (e *UIActionExecutor) executeGoForward(action UIAction) error {
	// Go forward is not implemented in the basic UITester interface
	return NewGowrightError(BrowserError, "go forward action not implemented", nil)
}

// ExecuteActions executes a sequence of UI actions
func (e *UIActionExecutor) ExecuteActions(actions []UIAction) error {
	for i, action := range actions {
		if err := e.ExecuteAction(action); err != nil {
			return NewGowrightError(BrowserError, fmt.Sprintf("failed to execute action %d (%s)", i, action.Type), err)
		}
	}
	return nil
}

// ValidateAction validates that an action has the required fields
func ValidateAction(action UIAction) error {
	if action.Type == "" {
		return NewGowrightError(BrowserError, "action type is required", nil)
	}

	actionType := UIActionType(action.Type)

	switch actionType {
	case ActionClick, ActionHover, ActionClear, ActionSubmit:
		if action.Selector == "" {
			return NewGowrightError(BrowserError, fmt.Sprintf("selector is required for %s action", action.Type), nil)
		}
	case ActionType:
		if action.Selector == "" {
			return NewGowrightError(BrowserError, "selector is required for type action", nil)
		}
		if action.Value == "" {
			return NewGowrightError(BrowserError, "value is required for type action", nil)
		}
	case ActionNavigate:
		if action.Value == "" {
			return NewGowrightError(BrowserError, "URL is required for navigate action", nil)
		}
	case ActionSelect:
		if action.Selector == "" {
			return NewGowrightError(BrowserError, "selector is required for select action", nil)
		}
		if action.Value == "" {
			return NewGowrightError(BrowserError, "value is required for select action", nil)
		}
	case ActionScroll:
		// Scroll can work with or without selector
	case ActionWait:
		// Wait can work with or without selector
	case ActionRefresh, ActionGoBack, ActionGoForward:
		// These actions don't require selector or value
	case ActionTap, ActionLongPress:
		if action.Selector == "" {
			return NewGowrightError(BrowserError, fmt.Sprintf("selector is required for %s action", action.Type), nil)
		}
	case ActionSwipe:
		// Swipe requires options with coordinates
		if action.Options == nil {
			return NewGowrightError(BrowserError, "swipe options are required for swipe action", nil)
		}
	case ActionSwipeLeft, ActionSwipeRight, ActionSwipeUp, ActionSwipeDown:
		// Directional swipes don't require additional parameters
	case ActionPinch:
		if action.Options == nil {
			return NewGowrightError(BrowserError, "pinch options are required for pinch action", nil)
		}
	case ActionSetOrientation:
		if action.Value == "" {
			return NewGowrightError(BrowserError, "orientation value is required for set orientation action", nil)
		}
	default:
		return NewGowrightError(BrowserError, fmt.Sprintf("unsupported action type: %s", action.Type), nil)
	}

	return nil
}

// Mobile-specific action execution methods

// executeTap executes a tap action (mobile-specific click)
func (e *UIActionExecutor) executeTap(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for tap action", nil)
	}

	// Check if tester supports mobile actions
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.Tap(action.Selector)
	}

	// Fallback to regular click for non-mobile testers
	return e.tester.Click(action.Selector)
}

// executeSwipe executes a swipe action
func (e *UIActionExecutor) executeSwipe(action UIAction) error {
	// Parse swipe options
	var options *SwipeOptions
	if action.Options != nil {
		if swipeOpts, ok := action.Options.(*SwipeOptions); ok {
			options = swipeOpts
		} else if optsMap, ok := action.Options.(map[string]interface{}); ok {
			options = &SwipeOptions{}
			if startX, exists := optsMap["start_x"]; exists {
				if x, ok := startX.(int); ok {
					options.StartX = x
				}
			}
			if startY, exists := optsMap["start_y"]; exists {
				if y, ok := startY.(int); ok {
					options.StartY = y
				}
			}
			if endX, exists := optsMap["end_x"]; exists {
				if x, ok := endX.(int); ok {
					options.EndX = x
				}
			}
			if endY, exists := optsMap["end_y"]; exists {
				if y, ok := endY.(int); ok {
					options.EndY = y
				}
			}
			if duration, exists := optsMap["duration"]; exists {
				if dur, ok := duration.(time.Duration); ok {
					options.Duration = dur
				} else if durStr, ok := duration.(string); ok {
					if parsedDur, err := time.ParseDuration(durStr); err == nil {
						options.Duration = parsedDur
					}
				}
			}
		}
	}

	if options == nil {
		return NewGowrightError(BrowserError, "swipe options are required for swipe action", nil)
	}

	// Check if tester supports mobile actions
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		duration := options.Duration
		if duration == 0 {
			duration = 300 * time.Millisecond
		}
		return mobileTester.Swipe(options.StartX, options.StartY, options.EndX, options.EndY, duration)
	}

	return NewGowrightError(BrowserError, "swipe action not supported by this tester", nil)
}

// executeSwipeLeft executes a swipe left action
func (e *UIActionExecutor) executeSwipeLeft(action UIAction) error {
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.SwipeLeft()
	}

	return NewGowrightError(BrowserError, "swipe left action not supported by this tester", nil)
}

// executeSwipeRight executes a swipe right action
func (e *UIActionExecutor) executeSwipeRight(action UIAction) error {
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.SwipeRight()
	}

	return NewGowrightError(BrowserError, "swipe right action not supported by this tester", nil)
}

// executeSwipeUp executes a swipe up action
func (e *UIActionExecutor) executeSwipeUp(action UIAction) error {
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.SwipeUp()
	}

	return NewGowrightError(BrowserError, "swipe up action not supported by this tester", nil)
}

// executeSwipeDown executes a swipe down action
func (e *UIActionExecutor) executeSwipeDown(action UIAction) error {
	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.SwipeDown()
	}

	return NewGowrightError(BrowserError, "swipe down action not supported by this tester", nil)
}

// executeLongPress executes a long press action
func (e *UIActionExecutor) executeLongPress(action UIAction) error {
	if action.Selector == "" {
		return NewGowrightError(BrowserError, "selector is required for long press action", nil)
	}

	// Parse long press options
	duration := 1 * time.Second // default duration
	if action.Options != nil {
		if longPressOpts, ok := action.Options.(*LongPressOptions); ok {
			if longPressOpts.Duration > 0 {
				duration = longPressOpts.Duration
			}
		} else if optsMap, ok := action.Options.(map[string]interface{}); ok {
			if dur, exists := optsMap["duration"]; exists {
				if durVal, ok := dur.(time.Duration); ok {
					duration = durVal
				} else if durStr, ok := dur.(string); ok {
					if parsedDur, err := time.ParseDuration(durStr); err == nil {
						duration = parsedDur
					}
				}
			}
		}
	}

	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.LongPress(action.Selector, duration)
	}

	return NewGowrightError(BrowserError, "long press action not supported by this tester", nil)
}

// executePinch executes a pinch action
func (e *UIActionExecutor) executePinch(action UIAction) error {
	// Parse pinch options
	var options *PinchOptions
	if action.Options != nil {
		if pinchOpts, ok := action.Options.(*PinchOptions); ok {
			options = pinchOpts
		} else if optsMap, ok := action.Options.(map[string]interface{}); ok {
			options = &PinchOptions{}
			if centerX, exists := optsMap["center_x"]; exists {
				if x, ok := centerX.(int); ok {
					options.CenterX = x
				}
			}
			if centerY, exists := optsMap["center_y"]; exists {
				if y, ok := centerY.(int); ok {
					options.CenterY = y
				}
			}
			if scale, exists := optsMap["scale"]; exists {
				if s, ok := scale.(float64); ok {
					options.Scale = s
				}
			}
		}
	}

	if options == nil {
		return NewGowrightError(BrowserError, "pinch options are required for pinch action", nil)
	}

	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.Pinch(options.CenterX, options.CenterY, options.Scale)
	}

	return NewGowrightError(BrowserError, "pinch action not supported by this tester", nil)
}

// executeSetOrientation executes a set orientation action
func (e *UIActionExecutor) executeSetOrientation(action UIAction) error {
	if action.Value == "" {
		return NewGowrightError(BrowserError, "orientation value is required for set orientation action", nil)
	}

	if mobileTester, ok := e.tester.(*MobileUITester); ok {
		return mobileTester.SetOrientation(action.Value)
	}

	return NewGowrightError(BrowserError, "set orientation action not supported by this tester", nil)
}
