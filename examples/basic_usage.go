package main

import (
	"fmt"
	"log"

	"github/gowright/framework/pkg/gowright"
)

func main() {
	// Create a new Gowright instance with default configuration
	gw, err := gowright.NewWithDefaults()
	if err != nil {
		log.Fatalf("Failed to create Gowright instance: %v", err)
	}

	fmt.Printf("Gowright Framework v%s initialized successfully!\n", gowright.Version())
	fmt.Printf("Configuration: %+v\n", gw.GetConfig())

	// Example of loading configuration from file
	// gw, err := gowright.NewFromFile("config.json")
	// if err != nil {
	//     log.Fatalf("Failed to load config: %v", err)
	// }

	// Example of loading configuration from environment
	// gw := gowright.NewFromEnv()
}