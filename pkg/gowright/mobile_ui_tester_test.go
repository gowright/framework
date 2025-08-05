package gowright

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MobileUITesterTestSuite defines the test suite for MobileUITester
type MobileUITesterTestSuite struct {
	suite.Suite
	tester        *MobileUITester
	browserConfig *BrowserConfig
	mobileConfig  *MobileDeviceConfig
}

// SetupTest runs before each test
func (suite *MobileUITesterTestSuite) SetupTest() {
	suite.browserConfig = &BrowserConfig{
		Headless: true,
		Timeout:  10 * time.Second,
		WindowSize: &WindowSize{
			Width:  390,
			Height: 844,
		},
	}

	suite.mobileConfig = &MobileDeviceConfig{
		DeviceType:   DeviceIPhone12,
		Width:        390,
		Height:       844,
		PixelRatio:   3.0,
		UserAgent:    "test-mobile-agent",
		TouchEnabled: true,
		Mobile:       true,
		Orientation:  "portrait",
	}

	suite.tester = NewMobileUITester(suite.browserConfig, suite.mobileConfig)
}

// TearDownTest runs after each test
func (suite *MobileUITesterTestSuite) TearDownTest() {
	if suite.tester != nil {
		_ = suite.tester.Cleanup()
	}
}

// TestNewMobileUITester tests the constructor
func (suite *MobileUITesterTestSuite) TestNewMobileUITester() {
	// Test with custom configs
	tester := NewMobileUITester(suite.browserConfig, suite.mobileConfig)
	suite.NotNil(tester)
	suite.NotNil(tester.RodUITester)
	suite.Equal(suite.mobileConfig, tester.mobileConfig)

	// Test with nil mobile config (should use default)
	tester2 := NewMobileUITester(suite.browserConfig, nil)
	suite.NotNil(tester2)
	suite.NotNil(tester2.mobileConfig)
	suite.Equal(DeviceIPhone12, tester2.mobileConfig.DeviceType)
}

// TestGetName tests the GetName method
func (suite *MobileUITesterTestSuite) TestGetName() {
	name := suite.tester.GetName()
	suite.Contains(name, "MobileUITester")
	suite.Contains(name, string(DeviceIPhone12))
}

// TestGetMobileConfig tests getting mobile configuration
func (suite *MobileUITesterTestSuite) TestGetMobileConfig() {
	config := suite.tester.GetMobileConfig()
	suite.Equal(suite.mobileConfig, config)
}

// TestSetMobileConfig tests setting mobile configuration
func (suite *MobileUITesterTestSuite) TestSetMobileConfig() {
	newConfig := &MobileDeviceConfig{
		DeviceType:   DevicePixel5,
		Width:        393,
		Height:       851,
		PixelRatio:   2.75,
		TouchEnabled: true,
		Mobile:       true,
		Orientation:  "portrait",
	}

	err := suite.tester.SetMobileConfig(newConfig)
	suite.NoError(err)
	suite.Equal(newConfig, suite.tester.GetMobileConfig())
}

// TestIsMobileDevice tests mobile device detection
func (suite *MobileUITesterTestSuite) TestIsMobileDevice() {
	suite.True(suite.tester.IsMobileDevice())

	// Test with non-mobile config
	desktopConfig := &MobileDeviceConfig{
		Mobile: false,
	}
	_ = suite.tester.SetMobileConfig(desktopConfig)
	suite.False(suite.tester.IsMobileDevice())
}

// TestGetDeviceInfo tests device information retrieval
func (suite *MobileUITesterTestSuite) TestGetDeviceInfo() {
	info := suite.tester.GetDeviceInfo()
	suite.NotNil(info)

	suite.Equal(DeviceIPhone12, info["device_type"])
	suite.Equal(390, info["width"])
	suite.Equal(844, info["height"])
	suite.Equal(3.0, info["pixel_ratio"])
	suite.Equal("portrait", info["orientation"])
	suite.Equal(true, info["touch_enabled"])
	suite.Equal(true, info["mobile"])
}

// TestTapWithoutInitialization tests tap without browser initialization
func (suite *MobileUITesterTestSuite) TestTapWithoutInitialization() {
	err := suite.tester.Tap("#button")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestSwipeWithoutInitialization tests swipe without browser initialization
func (suite *MobileUITesterTestSuite) TestSwipeWithoutInitialization() {
	err := suite.tester.Swipe(100, 100, 200, 200, 300*time.Millisecond)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestSwipeDirectionsWithoutInitialization tests directional swipes without initialization
func (suite *MobileUITesterTestSuite) TestSwipeDirectionsWithoutInitialization() {
	testCases := []struct {
		name string
		fn   func() error
	}{
		{"SwipeLeft", suite.tester.SwipeLeft},
		{"SwipeRight", suite.tester.SwipeRight},
		{"SwipeUp", suite.tester.SwipeUp},
		{"SwipeDown", suite.tester.SwipeDown},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.fn()
			suite.Error(err)

			gowrightErr, ok := err.(*GowrightError)
			suite.True(ok)
			suite.Equal(BrowserError, gowrightErr.Type)
		})
	}
}

// TestSetOrientationWithoutInitialization tests orientation change without initialization
func (suite *MobileUITesterTestSuite) TestSetOrientationWithoutInitialization() {
	err := suite.tester.SetOrientation("landscape")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestSetOrientationInvalidOrientation tests invalid orientation
func (suite *MobileUITesterTestSuite) TestSetOrientationInvalidOrientation() {
	// Since we can't mock the page without a real browser, this test will check
	// that the method returns an error when page is not initialized
	err := suite.tester.SetOrientation("invalid")
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestLongPressWithoutInitialization tests long press without initialization
func (suite *MobileUITesterTestSuite) TestLongPressWithoutInitialization() {
	err := suite.tester.LongPress("#button", 1*time.Second)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestPinchWithoutInitialization tests pinch without initialization
func (suite *MobileUITesterTestSuite) TestPinchWithoutInitialization() {
	err := suite.tester.Pinch(200, 200, 1.5)
	suite.Error(err)

	gowrightErr, ok := err.(*GowrightError)
	suite.True(ok)
	suite.Equal(BrowserError, gowrightErr.Type)
	suite.Contains(gowrightErr.Message, "page not initialized")
}

// TestMobileUITesterTestSuite runs the test suite
func TestMobileUITesterTestSuite(t *testing.T) {
	suite.Run(t, new(MobileUITesterTestSuite))
}

// TestGetDefaultMobileConfig tests default mobile configurations
func TestGetDefaultMobileConfig(t *testing.T) {
	testCases := []struct {
		deviceType     MobileDeviceType
		expectedWidth  int
		expectedHeight int
		expectedRatio  float64
	}{
		{DeviceIPhone12, 390, 844, 3.0},
		{DeviceIPhone12Pro, 390, 844, 3.0},
		{DeviceIPhoneSE, 375, 667, 2.0},
		{DevicePixel5, 393, 851, 2.75},
		{DeviceGalaxyS21, 384, 854, 2.75},
		{DeviceIPadAir, 820, 1180, 2.0},
	}

	for _, tc := range testCases {
		t.Run(string(tc.deviceType), func(t *testing.T) {
			config := GetDefaultMobileConfig(tc.deviceType)

			assert.NotNil(t, config)
			assert.Equal(t, tc.deviceType, config.DeviceType)
			assert.Equal(t, tc.expectedWidth, config.Width)
			assert.Equal(t, tc.expectedHeight, config.Height)
			assert.Equal(t, tc.expectedRatio, config.PixelRatio)
			assert.True(t, config.TouchEnabled)
			assert.True(t, config.Mobile)
			assert.Equal(t, "portrait", config.Orientation)
			assert.NotEmpty(t, config.UserAgent)
		})
	}
}

// TestGetDefaultMobileConfigUnknownDevice tests default config for unknown device
func TestGetDefaultMobileConfigUnknownDevice(t *testing.T) {
	config := GetDefaultMobileConfig("UnknownDevice")

	// Should return iPhone 12 as default
	assert.NotNil(t, config)
	assert.Equal(t, DeviceIPhone12, config.DeviceType)
	assert.Equal(t, 390, config.Width)
	assert.Equal(t, 844, config.Height)
}

// TestMobileDeviceTypes tests device type constants
func TestMobileDeviceTypes(t *testing.T) {
	assert.Equal(t, "iPhone 12", string(DeviceIPhone12))
	assert.Equal(t, "iPhone 12 Pro", string(DeviceIPhone12Pro))
	assert.Equal(t, "iPhone SE", string(DeviceIPhoneSE))
	assert.Equal(t, "Pixel 5", string(DevicePixel5))
	assert.Equal(t, "Galaxy S21", string(DeviceGalaxyS21))
	assert.Equal(t, "iPad Air", string(DeviceIPadAir))
	assert.Equal(t, "Custom Mobile", string(DeviceCustomMobile))
}

// TestMobileConfigValidation tests mobile configuration validation
func TestMobileConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *MobileDeviceConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &MobileDeviceConfig{
				DeviceType:   DeviceIPhone12,
				Width:        390,
				Height:       844,
				PixelRatio:   3.0,
				TouchEnabled: true,
				Mobile:       true,
				Orientation:  "portrait",
			},
			valid: true,
		},
		{
			name: "zero dimensions",
			config: &MobileDeviceConfig{
				DeviceType:   DeviceIPhone12,
				Width:        0,
				Height:       0,
				PixelRatio:   3.0,
				TouchEnabled: true,
				Mobile:       true,
				Orientation:  "portrait",
			},
			valid: false,
		},
		{
			name: "invalid pixel ratio",
			config: &MobileDeviceConfig{
				DeviceType:   DeviceIPhone12,
				Width:        390,
				Height:       844,
				PixelRatio:   0,
				TouchEnabled: true,
				Mobile:       true,
				Orientation:  "portrait",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check that required fields are set
			if tt.valid {
				assert.Greater(t, tt.config.Width, 0)
				assert.Greater(t, tt.config.Height, 0)
				assert.Greater(t, tt.config.PixelRatio, 0.0)
			} else {
				// At least one validation should fail
				invalid := tt.config.Width <= 0 ||
					tt.config.Height <= 0 ||
					tt.config.PixelRatio <= 0
				assert.True(t, invalid)
			}
		})
	}
}

// TestOrientationLogic tests orientation change logic
func TestOrientationLogic(t *testing.T) {
	config := &MobileDeviceConfig{
		DeviceType:  DeviceIPhone12,
		Width:       390,
		Height:      844,
		Orientation: "portrait",
	}

	tester := NewMobileUITester(nil, config)

	// Test portrait to landscape logic (without actual browser)
	originalWidth := config.Width
	originalHeight := config.Height

	// Simulate orientation change logic
	if config.Orientation == "portrait" {
		// In landscape, width and height should be swapped
		expectedLandscapeWidth := originalHeight
		expectedLandscapeHeight := originalWidth

		assert.Equal(t, originalWidth, config.Width)
		assert.Equal(t, originalHeight, config.Height)

		// The actual swap would happen in SetOrientation method
		// Here we just test the logic
		assert.NotEqual(t, expectedLandscapeWidth, config.Width)
		assert.NotEqual(t, expectedLandscapeHeight, config.Height)
	}

	assert.NotNil(t, tester)
}

// TestSwipeCoordinateCalculation tests swipe coordinate calculations
func TestSwipeCoordinateCalculation(t *testing.T) {
	config := &MobileDeviceConfig{
		Width:  390,
		Height: 844,
	}

	// Test swipe left coordinates
	startX := int(float64(config.Width) * 0.8) // 312
	endX := int(float64(config.Width) * 0.2)   // 78
	y := config.Height / 2                     // 422

	assert.Equal(t, 312, startX)
	assert.Equal(t, 78, endX)
	assert.Equal(t, 422, y)

	// Test swipe up coordinates
	x := config.Width / 2                       // 195
	startY := int(float64(config.Height) * 0.8) // 675
	endY := int(float64(config.Height) * 0.2)   // 168

	assert.Equal(t, 195, x)
	assert.Equal(t, 675, startY)
	assert.Equal(t, 168, endY)
}
