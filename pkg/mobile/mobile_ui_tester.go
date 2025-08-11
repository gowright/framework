package mobile

import (
	"fmt"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
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
	DeviceType   MobileDeviceType `json:"device_type"`
	Width        int              `json:"width"`
	Height       int              `json:"height"`
	PixelRatio   float64          `json:"pixel_ratio"`
	UserAgent    string           `json:"user_agent"`
	TouchEnabled bool             `json:"touch_enabled"`
	Mobile       bool             `json:"mobile"`
	Orientation  string           `json:"orientation"` // portrait, landscape
}

// MobileUITester provides mobile-specific UI testing capabilities
type MobileUITester struct {
	config       *config.MobileConfig
	deviceConfig *MobileDeviceConfig
	initialized  bool
}

// NewMobileUITester creates a new MobileUITester instance
func NewMobileUITester() *MobileUITester {
	return &MobileUITester{
		initialized: false,
	}
}

// Initialize sets up the mobile tester with configuration
func (m *MobileUITester) Initialize(cfg interface{}) error {
	mobileConfig, ok := cfg.(*config.MobileConfig)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid configuration type for mobile tester", nil)
	}

	m.config = mobileConfig
	m.initialized = true

	// Set default device config if not provided
	if m.deviceConfig == nil {
		m.deviceConfig = GetDefaultMobileConfig(DeviceIPhone12)
	}

	// Initialize Appium connection here
	// This would involve setting up Appium WebDriver connection

	return nil
}

// Cleanup performs cleanup operations
func (m *MobileUITester) Cleanup() error {
	// Close Appium sessions, cleanup resources
	m.initialized = false
	return nil
}

// GetName returns the name of the tester
func (m *MobileUITester) GetName() string {
	if m.deviceConfig != nil {
		return fmt.Sprintf("MobileUITester (%s)", m.deviceConfig.DeviceType)
	}
	return "MobileUITester"
}

// Tap taps on an element identified by the selector
func (m *MobileUITester) Tap(selector string) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return nil
}

// Type types text into an element identified by the selector
func (m *MobileUITester) Type(selector, text string) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return nil
}

// Swipe performs a swipe gesture
func (m *MobileUITester) Swipe(startX, startY, endX, endY int, duration time.Duration) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return nil
}

// SwipeLeft performs a left swipe gesture
func (m *MobileUITester) SwipeLeft() error {
	if m.deviceConfig == nil {
		return core.NewGowrightError(core.BrowserError, "device config not set", nil)
	}

	width := m.deviceConfig.Width
	height := m.deviceConfig.Height

	startX := int(float64(width) * 0.8) // Start from 80% of width
	endX := int(float64(width) * 0.2)   // End at 20% of width
	y := height / 2                     // Middle of screen

	return m.Swipe(startX, y, endX, y, 300*time.Millisecond)
}

// SwipeRight performs a right swipe gesture
func (m *MobileUITester) SwipeRight() error {
	if m.deviceConfig == nil {
		return core.NewGowrightError(core.BrowserError, "device config not set", nil)
	}

	width := m.deviceConfig.Width
	height := m.deviceConfig.Height

	startX := int(float64(width) * 0.2) // Start from 20% of width
	endX := int(float64(width) * 0.8)   // End at 80% of width
	y := height / 2                     // Middle of screen

	return m.Swipe(startX, y, endX, y, 300*time.Millisecond)
}

// SwipeUp performs an up swipe gesture
func (m *MobileUITester) SwipeUp() error {
	if m.deviceConfig == nil {
		return core.NewGowrightError(core.BrowserError, "device config not set", nil)
	}

	width := m.deviceConfig.Width
	height := m.deviceConfig.Height

	x := width / 2                       // Middle of screen
	startY := int(float64(height) * 0.8) // Start from 80% of height
	endY := int(float64(height) * 0.2)   // End at 20% of height

	return m.Swipe(x, startY, x, endY, 300*time.Millisecond)
}

// SwipeDown performs a down swipe gesture
func (m *MobileUITester) SwipeDown() error {
	if m.deviceConfig == nil {
		return core.NewGowrightError(core.BrowserError, "device config not set", nil)
	}

	width := m.deviceConfig.Width
	height := m.deviceConfig.Height

	x := width / 2                       // Middle of screen
	startY := int(float64(height) * 0.2) // Start from 20% of height
	endY := int(float64(height) * 0.8)   // End at 80% of height

	return m.Swipe(x, startY, x, endY, 300*time.Millisecond)
}

// SetOrientation changes the device orientation
func (m *MobileUITester) SetOrientation(orientation string) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	if m.deviceConfig == nil {
		return core.NewGowrightError(core.BrowserError, "device config not set", nil)
	}

	switch orientation {
	case "portrait", "landscape":
		m.deviceConfig.Orientation = orientation
		// Implementation would use Appium WebDriver to change orientation
		return nil
	default:
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("unsupported orientation: %s", orientation), nil)
	}
}

// LongPress performs a long press action on an element
func (m *MobileUITester) LongPress(selector string, duration time.Duration) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return nil
}

// GetText retrieves text from an element identified by the selector
func (m *MobileUITester) GetText(selector string) (string, error) {
	if !m.initialized {
		return "", core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return "", nil
}

// WaitForElement waits for an element to be present
func (m *MobileUITester) WaitForElement(selector string, timeout time.Duration) error {
	if !m.initialized {
		return core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return nil
}

// TakeScreenshot captures a screenshot and returns the file path
func (m *MobileUITester) TakeScreenshot(filename string) (string, error) {
	if !m.initialized {
		return "", core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	// Implementation would use Appium WebDriver
	return "", nil
}

// GetMobileConfig returns the current mobile configuration
func (m *MobileUITester) GetMobileConfig() *MobileDeviceConfig {
	return m.deviceConfig
}

// SetMobileConfig updates the mobile configuration
func (m *MobileUITester) SetMobileConfig(config *MobileDeviceConfig) error {
	m.deviceConfig = config
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

// IsMobileDevice checks if the current configuration is for a mobile device
func (m *MobileUITester) IsMobileDevice() bool {
	return m.deviceConfig != nil && m.deviceConfig.Mobile
}

// GetDeviceInfo returns information about the current device configuration
func (m *MobileUITester) GetDeviceInfo() (map[string]interface{}, error) {
	if !m.initialized {
		return nil, core.NewGowrightError(core.BrowserError, "mobile tester not initialized", nil)
	}

	if m.deviceConfig == nil {
		return make(map[string]interface{}), nil
	}

	return map[string]interface{}{
		"device_type":   m.deviceConfig.DeviceType,
		"width":         m.deviceConfig.Width,
		"height":        m.deviceConfig.Height,
		"pixel_ratio":   m.deviceConfig.PixelRatio,
		"orientation":   m.deviceConfig.Orientation,
		"touch_enabled": m.deviceConfig.TouchEnabled,
		"mobile":        m.deviceConfig.Mobile,
	}, nil
}
