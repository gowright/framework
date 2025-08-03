package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gowright/framework/pkg/gowright"
)

var (
	versionFlag = flag.Bool("version", false, "Show version information")
	configFlag  = flag.String("config", "", "Path to configuration file")
	helpFlag    = flag.Bool("help", false, "Show help information")
)

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *versionFlag {
		showVersion()
		return
	}

	// Default behavior - show help
	showHelp()
}

func showVersion() {
	info := gowright.GetVersionInfo()
	
	fmt.Printf("Gowright Testing Framework\n")
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Git Commit: %s\n", info.GitCommit)
	fmt.Printf("Build Date: %s\n", info.BuildDate)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
	fmt.Printf("Platform: %s\n", info.Platform)
	
	// Also output as JSON if requested
	if len(os.Args) > 2 && os.Args[2] == "--json" {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Printf("\nJSON:\n%s\n", string(jsonData))
	}
}

func showHelp() {
	fmt.Printf(`Gowright Testing Framework v%s

A comprehensive testing framework for Go that supports UI, API, database, and integration testing.

USAGE:
    gowright [OPTIONS]

OPTIONS:
    --version           Show version information
    --version --json    Show version information in JSON format
    --config <file>     Specify configuration file path
    --help              Show this help message

EXAMPLES:
    gowright --version                    # Show version
    gowright --version --json             # Show version as JSON
    gowright --config ./config.json       # Use specific config file

DOCUMENTATION:
    For detailed documentation and examples, visit:
    https://github.com/your-org/gowright

    API Documentation: https://pkg.go.dev/github.com/gowright/framework
    
GETTING STARTED:
    1. Import the framework in your Go test files:
       import "github.com/gowright/framework/pkg/gowright"
    
    2. Create a basic test:
       func TestExample(t *testing.T) {
           framework := gowright.NewWithDefaults()
           defer framework.Close()
           // Your test code here
       }
    
    3. Run your tests:
       go test ./...

SUPPORT:
    - GitHub Issues: https://github.com/your-org/gowright/issues
    - Discussions: https://github.com/your-org/gowright/discussions
    - Email: support@gowright.dev

`, gowright.GetVersion())
}