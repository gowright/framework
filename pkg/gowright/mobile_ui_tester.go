package gowright

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// MobileDeviceType represents different mobile device types
type MobileDeviceType string

const (
	DeviceIPhone12     MobileDeviceType = "iPhone 12"
	DeviceIPhone12Pro  MobileDeviceType = "iPhone 12 Pro"
	DeviceIPhoneSE     MobileDeviceType = "iPhone SE"
	DevicePixel5       MobileDeviceType = "Pixel 5"
	DeviceGalaxyS21    MobileDeviceType = "Galaxy S21"
	DeviceIPadAir      MobileDeviceType = "iPad Air"
	DeviceCustomMobile MobileDeviceType = "Custom Mobile"
)

// MobileDeviceConfig holds mobile device configuration
type MobileDeviceConfig struct {
	DeviceType    MobileDeviceType `json:"device_type"`
	Width         int              `json:"width"`
	Height        int              `json:"height"`
	PixelRatio    float64          `json:"pixel_ratio"`
	UserAgent     string           `json:"user_agent"`
	TouchEnabled  bool             `json:"touch_enabled"`
	Mobile        bool             `json:"mobile"`
	Orientation   string           `json:"orientation"` // portrait, landscape
}

// MobileUITester extends RodUITester with mobile-specific capabilities
type MobileUITester struct {
	*RodUITester
	mobileConfig *MobileDeviceConfig
}

// NewMobileUITester creates a new MobileUITester instance
func NewMobileUITester(browserConfig *BrowserConfig, mobileConfig *MobileDeviceConfig) *MobileUITester {
	if mobileConfig == nil {
		mobileConfig = GetDefaultMobileConfig(DeviceIPhone12)
	}
	
	rodTester := NewRodUITester(browserConfig)
	
	return &MobileUITester{
		RodUITester:  rodTester,
		mobileConfig: mobileConfig,
	}
}

// Initialize sets up the mobile browser with the provided configuration
func (m *MobileUITester) Initialize(config interface{}) error {
	// First initialize the base RodUITester
	if err := m.RodUITester.Initialize(config); err != nil {
		return err
	}
	
	// Apply mobile device emulation
	if err := m.applyMobileEmulation(); err != nil {
		return NewGowrightError(BrowserError, "failed to apply mobile emulation", err)
	}
	
	return nil
}

// applyMobileEmulation applies mobile device emulation settings
func (m *MobileUITester) applyMobileEmulation() error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}
	
	// Set device metrics
	err := m.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             m.mobileConfig.Width,
		Height:            m.mobileConfig.Height,
		DeviceScaleFactor: m.mobileConfig.PixelRatio,
		Mobile:            m.mobileConfig.Mobile,
	})
	if err != nil {
		return fmt.Errorf("failed to set device metrics: %w", err)
	}
	
	// Set user agent
	if m.mobileConfig.UserAgent != "" {
		err = m.page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: m.mobileConfig.UserAgent,
		})
		if err != nil {
			return fmt.Errorf("failed to set user agent: %w", err)
		}
	}
	
	// Enable touch events if specified
	if m.mobileConfig.TouchEnabled {
		_, err = m.page.Eval(`() => {
			// Enable touch events through JavaScript
			if ('ontouchstart' in window) {
				// Touch events are already supported
			} else {
				// Simulate touch events support
				window.ontouchstart = function() {};
			}
		}`)
		if err != nil {
			return fmt.Errorf("failed to enable touch emulation: %w", err)
		}
	}
	
	return nil
}

// Tap performs a tap action on an element (mobile-specific click)
func (m *MobileUITester) Tap(selector string) error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	element, err := m.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	// Use tap instead of click for mobile
	err = element.Tap()
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to tap element with selector %s", selector), err)
	}

	return nil
}

// Swipe performs a swipe gesture
func (m *MobileUITester) Swipe(startX, startY, endX, endY int, duration time.Duration) error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	// Perform swipe using touch events
	err := m.page.Mouse.MoveTo(proto.Point{X: float64(startX), Y: float64(startY)})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to move to start position", err)
	}

	err = m.page.Mouse.Down(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return NewGowrightError(BrowserError, "failed to start swipe", err)
	}

	// Simulate swipe movement
	steps := 10
	stepX := float64(endX-startX) / float64(steps)
	stepY := float64(endY-startY) / float64(steps)
	stepDuration := duration / time.Duration(steps)

	for i := 1; i <= steps; i++ {
		currentX := float64(startX) + stepX*float64(i)
		currentY := float64(startY) + stepY*float64(i)
		
		err = m.page.Mouse.MoveTo(proto.Point{X: currentX, Y: currentY})
		if err != nil {
			return NewGowrightError(BrowserError, "failed to move during swipe", err)
		}
		
		time.Sleep(stepDuration)
	}

	err = m.page.Mouse.Up(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return NewGowrightError(BrowserError, "failed to end swipe", err)
	}

	return nil
}

// SwipeLeft performs a left swipe gesture
func (m *MobileUITester) SwipeLeft() error {
	width := m.mobileConfig.Width
	height := m.mobileConfig.Height
	
	startX := int(float64(width) * 0.8)  // Start from 80% of width
	endX := int(float64(width) * 0.2)    // End at 20% of width
	y := height / 2                      // Middle of screen
	
	return m.Swipe(startX, y, endX, y, 300*time.Millisecond)
}

// SwipeRight performs a right swipe gesture
func (m *MobileUITester) SwipeRight() error {
	width := m.mobileConfig.Width
	height := m.mobileConfig.Height
	
	startX := int(float64(width) * 0.2)  // Start from 20% of width
	endX := int(float64(width) * 0.8)    // End at 80% of width
	y := height / 2                      // Middle of screen
	
	return m.Swipe(startX, y, endX, y, 300*time.Millisecond)
}

// SwipeUp performs an up swipe gesture
func (m *MobileUITester) SwipeUp() error {
	width := m.mobileConfig.Width
	height := m.mobileConfig.Height
	
	x := width / 2                        // Middle of screen
	startY := int(float64(height) * 0.8)  // Start from 80% of height
	endY := int(float64(height) * 0.2)    // End at 20% of height
	
	return m.Swipe(x, startY, x, endY, 300*time.Millisecond)
}

// SwipeDown performs a down swipe gesture
func (m *MobileUITester) SwipeDown() error {
	width := m.mobileConfig.Width
	height := m.mobileConfig.Height
	
	x := width / 2                        // Middle of screen
	startY := int(float64(height) * 0.2)  // Start from 20% of height
	endY := int(float64(height) * 0.8)    // End at 80% of height
	
	return m.Swipe(x, startY, x, endY, 300*time.Millisecond)
}

// SetOrientation changes the device orientation
func (m *MobileUITester) SetOrientation(orientation string) error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	var width, height int
	
	switch orientation {
	case "portrait":
		if m.mobileConfig.Orientation == "portrait" {
			return nil // Already in portrait
		}
		width = m.mobileConfig.Width
		height = m.mobileConfig.Height
		m.mobileConfig.Orientation = "portrait"
	case "landscape":
		if m.mobileConfig.Orientation == "landscape" {
			return nil // Already in landscape
		}
		// Swap width and height for landscape
		width = m.mobileConfig.Height
		height = m.mobileConfig.Width
		m.mobileConfig.Orientation = "landscape"
	default:
		return NewGowrightError(BrowserError, fmt.Sprintf("unsupported orientation: %s", orientation), nil)
	}

	// Apply new viewport settings
	err := m.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: m.mobileConfig.PixelRatio,
		Mobile:            m.mobileConfig.Mobile,
	})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to set orientation", err)
	}

	return nil
}

// LongPress performs a long press action on an element
func (m *MobileUITester) LongPress(selector string, duration time.Duration) error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	element, err := m.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	// Get element position
	box, err := element.Shape()
	if err != nil {
		return NewGowrightError(BrowserError, "failed to get element position", err)
	}

	centerX := box.Box().X + box.Box().Width/2
	centerY := box.Box().Y + box.Box().Height/2

	// Perform long press
	err = m.page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to move to element", err)
	}

	err = m.page.Mouse.Down(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return NewGowrightError(BrowserError, "failed to start long press", err)
	}

	time.Sleep(duration)

	err = m.page.Mouse.Up(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return NewGowrightError(BrowserError, "failed to end long press", err)
	}

	return nil
}

// Pinch performs a pinch gesture (zoom in/out)
func (m *MobileUITester) Pinch(centerX, centerY int, scale float64) error {
	if m.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	// This is a simplified pinch implementation
	// In a real implementation, you would use multi-touch events
	
	// For now, we'll simulate zoom using the browser's zoom functionality
	currentZoom := 1.0
	targetZoom := currentZoom * scale
	
	err := m.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             m.mobileConfig.Width,
		Height:            m.mobileConfig.Height,
		DeviceScaleFactor: m.mobileConfig.PixelRatio * targetZoom,
		Mobile:            m.mobileConfig.Mobile,
	})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to apply pinch zoom", err)
	}

	return nil
}

// GetMobileConfig returns the current mobile configuration
func (m *MobileUITester) GetMobileConfig() *MobileDeviceConfig {
	return m.mobileConfig
}

// SetMobileConfig updates the mobile configuration
func (m *MobileUITester) SetMobileConfig(config *MobileDeviceConfig) error {
	m.mobileConfig = config
	
	// Re-apply mobile emulation if page is initialized
	if m.page != nil {
		return m.applyMobileEmulation()
	}
	
	return nil
}

// GetDefaultMobileConfig returns default configuration for common mobile devices
func GetDefaultMobileConfig(deviceType MobileDeviceType) *MobileDeviceConfig {
	switch deviceType {
	case DeviceIPhone12:
		return &MobileDeviceConfig{
			DeviceType:   DeviceIPhone12,
			Width:        390,
			Height:       844,
			PixelRatio:   3.0,
			UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	case DeviceIPhone12Pro:
		return &MobileDeviceConfig{
			DeviceType:   DeviceIPhone12Pro,
			Width:        390,
			Height:       844,
			PixelRatio:   3.0,
			UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	case DeviceIPhoneSE:
		return &MobileDeviceConfig{
			DeviceType:   DeviceIPhoneSE,
			Width:        375,
			Height:       667,
			PixelRatio:   2.0,
			UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	case DevicePixel5:
		return &MobileDeviceConfig{
			DeviceType:   DevicePixel5,
			Width:        393,
			Height:       851,
			PixelRatio:   2.75,
			UserAgent:    "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.91 Mobile Safari/537.36",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	case DeviceGalaxyS21:
		return &MobileDeviceConfig{
			DeviceType:   DeviceGalaxyS21,
			Width:        384,
			Height:       854,
			PixelRatio:   2.75,
			UserAgent:    "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.91 Mobile Safari/537.36",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	case DeviceIPadAir:
		return &MobileDeviceConfig{
			DeviceType:   DeviceIPadAir,
			Width:        820,
			Height:       1180,
			PixelRatio:   2.0,
			UserAgent:    "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			TouchEnabled: true,
			Mobile:       true,
			Orientation:  "portrait",
		}
	default:
		// Return iPhone 12 as default
		return GetDefaultMobileConfig(DeviceIPhone12)
	}
}

// GetName returns the name of the mobile tester
func (m *MobileUITester) GetName() string {
	return fmt.Sprintf("MobileUITester (%s)", m.mobileConfig.DeviceType)
}

// IsMobileDevice checks if the current configuration is for a mobile device
func (m *MobileUITester) IsMobileDevice() bool {
	return m.mobileConfig.Mobile
}

// GetDeviceInfo returns information about the current device configuration
func (m *MobileUITester) GetDeviceInfo() map[string]interface{} {
	return map[string]interface{}{
		"device_type":    m.mobileConfig.DeviceType,
		"width":          m.mobileConfig.Width,
		"height":         m.mobileConfig.Height,
		"pixel_ratio":    m.mobileConfig.PixelRatio,
		"orientation":    m.mobileConfig.Orientation,
		"touch_enabled":  m.mobileConfig.TouchEnabled,
		"mobile":         m.mobileConfig.Mobile,
	}
}

// ExecuteTest executes a UI test with mobile-specific capabilities and returns the result
func (m *MobileUITester) ExecuteTest(test *UITest) *TestCaseResult {
	startTime := time.Now()
	result := &TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
		Logs:      make([]string, 0),
		Steps:     make([]AssertionStep, 0),
	}

	// Navigate to the test URL
	if test.URL != "" {
		if err := m.Navigate(test.URL); err != nil {
			result.Status = TestStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(startTime)
			return result
		}
		result.Logs = append(result.Logs, fmt.Sprintf("Navigated to: %s", test.URL))
	}

	// Execute UI actions with mobile-specific handling
	for i, action := range test.Actions {
		actionStart := time.Now()
		var err error

		switch action.Type {
		case "tap":
			err = m.Tap(action.Selector)
		case "click":
			// Use tap for mobile instead of click
			err = m.Tap(action.Selector)
		case "type":
			err = m.Type(action.Selector, action.Value)
		case "navigate":
			err = m.Navigate(action.Value)
		case "swipe_left":
			err = m.SwipeLeft()
		case "swipe_right":
			err = m.SwipeRight()
		case "swipe_up":
			err = m.SwipeUp()
		case "swipe_down":
			err = m.SwipeDown()
		case "long_press":
			duration := 1 * time.Second
			if d, ok := action.Options.(time.Duration); ok {
				duration = d
			}
			err = m.LongPress(action.Selector, duration)
		case "set_orientation":
			err = m.SetOrientation(action.Value)
		case "wait":
			if timeout, ok := action.Options.(time.Duration); ok {
				err = m.WaitForElement(action.Selector, timeout)
			} else {
				err = m.WaitForElement(action.Selector, 10*time.Second)
			}
		default:
			err = NewGowrightError(BrowserError, fmt.Sprintf("unsupported mobile action type: %s", action.Type), nil)
		}

		actionEnd := time.Now()
		step := AssertionStep{
			Name:        fmt.Sprintf("Mobile Action %d: %s", i+1, action.Type),
			Description: fmt.Sprintf("Execute mobile %s action", action.Type),
			StartTime:   actionStart,
			EndTime:     actionEnd,
			Duration:    actionEnd.Sub(actionStart),
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			result.Error = err
			result.Logs = append(result.Logs, fmt.Sprintf("Mobile Action %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Mobile Action %d completed: %s", i+1, action.Type))
		}

		result.Steps = append(result.Steps, step)

		if result.Status == TestStatusFailed {
			break
		}
	}

	// Execute UI assertions (reuse base implementation)
	for i, assertion := range test.Assertions {
		assertionStart := time.Now()
		var success bool
		var err error

		switch assertion.Type {
		case "text_equals":
			text, getErr := m.GetText(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = text == assertion.Expected.(string)
				if !success {
					err = fmt.Errorf("expected text '%s', got '%s'", assertion.Expected, text)
				}
			}
		case "element_present":
			present, getErr := m.IsElementPresent(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = present == assertion.Expected.(bool)
				if !success {
					err = fmt.Errorf("expected element presence %v, got %v", assertion.Expected, present)
				}
			}
		case "element_visible":
			visible, getErr := m.IsElementVisible(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = visible == assertion.Expected.(bool)
				if !success {
					err = fmt.Errorf("expected element visibility %v, got %v", assertion.Expected, visible)
				}
			}
		case "orientation":
			currentOrientation := m.mobileConfig.Orientation
			expectedOrientation := assertion.Expected.(string)
			success = currentOrientation == expectedOrientation
			if !success {
				err = fmt.Errorf("expected orientation '%s', got '%s'", expectedOrientation, currentOrientation)
			}
		default:
			err = NewGowrightError(AssertionError, fmt.Sprintf("unsupported mobile assertion type: %s", assertion.Type), nil)
		}

		assertionEnd := time.Now()
		step := AssertionStep{
			Name:        fmt.Sprintf("Mobile Assertion %d: %s", i+1, assertion.Type),
			Description: fmt.Sprintf("Verify mobile %s", assertion.Type),
			StartTime:   assertionStart,
			EndTime:     assertionEnd,
			Duration:    assertionEnd.Sub(assertionStart),
			Expected:    assertion.Expected,
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			if result.Error == nil {
				result.Error = err
			}
			result.Logs = append(result.Logs, fmt.Sprintf("Mobile Assertion %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Mobile Assertion %d passed: %s", i+1, assertion.Type))
		}

		result.Steps = append(result.Steps, step)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	return result
}