#!/bin/bash

# Gowright Testing Framework - Run All Examples
# This script runs all example files in sequence

echo "=== Gowright Testing Framework Examples ==="
echo "Running all examples in sequence..."
echo ""

# Array of example files to run
examples=(
    "basic_usage.go"
    "ui_testing_example.go" 
    "api_testing_example.go"
    "database_testing_example.go"
    "integration_testing_example.go"
    "test_suite_with_assertions.go"
    "assertion_reporting_example.go"
    "reporting_example.go"
)

# Counter for tracking progress
total=${#examples[@]}
current=0

# Run each example
for example in "${examples[@]}"; do
    current=$((current + 1))
    echo "[$current/$total] Running $example..."
    echo "----------------------------------------"
    
    # Check if file exists
    if [ ! -f "examples/$example" ]; then
        echo "❌ Error: $example not found!"
        continue
    fi
    
    # Run the example
    if go run "examples/$example"; then
        echo "✅ $example completed successfully"
    else
        echo "❌ $example failed with exit code $?"
    fi
    
    echo ""
    echo "----------------------------------------"
    echo ""
done

echo "=== All Examples Complete ==="
echo ""
echo "Generated Reports:"
echo "- ./ui-test-reports/ - UI testing reports"
echo "- ./api-test-reports/ - API testing reports"  
echo "- ./database-test-reports/ - Database testing reports"
echo "- ./integration-test-reports/ - Integration testing reports"
echo "- ./comprehensive-reports/ - Comprehensive test suite reports"
echo "- ./assertion-reports/ - Assertion-focused reports"
echo "- ./example-reports/ - Basic reporting examples"
echo ""
echo "Open the HTML reports in your browser to view detailed results!"