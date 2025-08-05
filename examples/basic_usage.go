//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	fmt.Println("Gowright Testing Framework - Basic Usage Example")

	// Create a new Gowright instance with default configuration
	gw := gowright.NewWithDefaults()

	// Get and display the configuration
	config := gw.GetConfig()
	fmt.Printf("Framework initialized with log level: %s\n", config.LogLevel)
	fmt.Printf("Browser headless mode: %t\n", config.BrowserConfig.Headless)
	fmt.Printf("API timeout: %v\n", config.APIConfig.Timeout)

	// Create a simple test suite
	testSuite := &gowright.TestSuite{
		Name:  "Basic Example Suite",
		Tests: make([]gowright.Test, 0),
		SetupFunc: func() error {
			fmt.Println("Setting up test suite...")
			return nil
		},
		TeardownFunc: func() error {
			fmt.Println("Tearing down test suite...")
			return nil
		},
	}

	// Set the test suite
	gw.SetTestSuite(testSuite)

	// Demonstrate configuration loading from environment
	envConfig := gowright.LoadConfigFromEnv()
	fmt.Printf("Environment config loaded with log level: %s\n", envConfig.LogLevel)

	// Demonstrate saving configuration to file
	if err := config.SaveToFile("gowright-config.json"); err != nil {
		log.Printf("Failed to save config: %v", err)
	} else {
		fmt.Println("Configuration saved to gowright-config.json")
	}

	fmt.Println("Basic usage example completed successfully!")
}
