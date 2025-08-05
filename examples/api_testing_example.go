//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	// Create API configuration
	config := &gowright.APIConfig{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Timeout: 30 * time.Second,
		Headers: map[string]string{
			"User-Agent": "Gowright-API-Tester/1.0",
		},
	}

	// Create and initialize API tester
	tester := gowright.NewAPITester(config)
	if err := tester.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize API tester: %v", err)
	}
	defer tester.Cleanup()

	fmt.Println("=== Gowright API Testing Example ===\n")

	// Example 1: Simple GET request
	fmt.Println("1. Testing GET /posts/1")
	getTest := gowright.NewAPITest("Get Post", "GET", "/posts/1", tester).
		SetExpectedStatus(200).
		SetExpectedHeader("Content-Type", "application/json; charset=utf-8").
		SetExpectedJSONPath("$.id", float64(1)).
		SetExpectedJSONPath("$.userId", float64(1))

	result := getTest.Execute()
	printTestResult(result)

	// Example 2: POST request with body
	fmt.Println("\n2. Testing POST /posts")
	postBody := map[string]interface{}{
		"title":  "Gowright Test Post",
		"body":   "This is a test post created by Gowright",
		"userId": 1,
	}

	postTest := gowright.NewAPITestBuilder("Create Post", "POST", "/posts").
		WithTester(tester).
		WithHeader("Content-Type", "application/json").
		WithBody(postBody).
		ExpectStatus(201).
		ExpectJSONPath("$.title", "Gowright Test Post").
		ExpectJSONPath("$.body", "This is a test post created by Gowright").
		ExpectJSONPath("$.userId", float64(1)).
		Build()

	result = postTest.Execute()
	printTestResult(result)

	// Example 3: GET request with query validation
	fmt.Println("\n3. Testing GET /posts (multiple posts)")
	postsTest := gowright.NewAPITest("Get All Posts", "GET", "/posts", tester).
		SetExpectedStatus(200).
		SetExpectedHeader("Content-Type", "application/json; charset=utf-8")

	result = postsTest.Execute()
	printTestResult(result)

	// Example 4: Testing error scenario
	fmt.Println("\n4. Testing GET /posts/999 (non-existent)")
	notFoundTest := gowright.NewAPITest("Get Non-existent Post", "GET", "/posts/999", tester).
		SetExpectedStatus(404)

	result = notFoundTest.Execute()
	printTestResult(result)

	// Example 5: PUT request
	fmt.Println("\n5. Testing PUT /posts/1")
	putBody := map[string]interface{}{
		"id":     1,
		"title":  "Updated Post Title",
		"body":   "Updated post body",
		"userId": 1,
	}

	putTest := gowright.NewAPITest("Update Post", "PUT", "/posts/1", tester).
		SetHeader("Content-Type", "application/json").
		SetBody(putBody).
		SetExpectedStatus(200).
		SetExpectedJSONPath("$.id", float64(1)).
		SetExpectedJSONPath("$.title", "Updated Post Title")

	result = putTest.Execute()
	printTestResult(result)

	// Example 6: DELETE request
	fmt.Println("\n6. Testing DELETE /posts/1")
	deleteTest := gowright.NewAPITest("Delete Post", "DELETE", "/posts/1", tester).
		SetExpectedStatus(200)

	result = deleteTest.Execute()
	printTestResult(result)

	fmt.Println("\n=== API Testing Complete ===")
}

func printTestResult(result *gowright.TestCaseResult) {
	fmt.Printf("Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status.String())
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	if len(result.Logs) > 0 {
		fmt.Println("Logs:")
		for _, log := range result.Logs {
			fmt.Printf("  - %s\n", log)
		}
	}

	fmt.Println("---")
}
