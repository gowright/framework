package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/gowright/framework/pkg/ui"
)

func main() {
	// Create UI tester
	tester := ui.NewUITester()

	// Configure browser with cookie notice disabling
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       false,
		WindowSize:     "1920x1080",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",
	}

	// Initialize tester
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	// Create a simple UI test
	test := &core.UITest{
		Name: "StackOverflow Test",
		URL:  "https://stackoverflow.com/",
		Actions: []core.UIAction{
			{Type: "scroll_page", Options: ui.ScrollOptions{
				Direction: "bottom",
				Distance:  0,
				Speed:     "fast",
			}},
		},
		Assertions: []core.UIAssertion{
			{Type: "element_exists", Selector: "div[class='site-footer--logo']"},
		},
	}

	// Execute the test
	fmt.Println("Executing UI test...")
	result := tester.ExecuteTest(test)

	// Print results
	fmt.Printf("Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	if len(result.Steps) > 0 {
		fmt.Println("Steps:")
		for _, step := range result.Steps {
			fmt.Printf("  - %s\n", step)
		}
	}

	fmt.Println("UI test completed!")
}
