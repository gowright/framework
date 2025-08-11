// Package api provides API testing capabilities using HTTP client
package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gowright/framework/pkg/assertions"
	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
)

// APITester implements the APITester interface for HTTP API testing
type APITester struct {
	config      *config.APIConfig
	asserter    *assertions.Asserter
	initialized bool
	client      *resty.Client
}

// NewAPITester creates a new API tester instance
func NewAPITester() *APITester {
	return &APITester{
		asserter: assertions.NewAsserter(),
	}
}

// Initialize sets up the API tester with configuration
func (at *APITester) Initialize(cfg interface{}) error {
	apiConfig, ok := cfg.(*config.APIConfig)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid configuration type for API tester", nil)
	}

	at.config = apiConfig

	// Initialize HTTP client
	at.client = resty.New()
	at.client.SetBaseURL(apiConfig.BaseURL)
	at.client.SetTimeout(apiConfig.Timeout)

	// Set default headers if provided
	if apiConfig.DefaultHeaders != nil {
		at.client.SetHeaders(apiConfig.DefaultHeaders)
	}

	// Configure retry settings
	if apiConfig.RetryCount > 0 {
		at.client.SetRetryCount(apiConfig.RetryCount)
		if apiConfig.RetryDelay > 0 {
			at.client.SetRetryWaitTime(apiConfig.RetryDelay)
		}
	}

	// Set authentication if provided
	if apiConfig.Auth != nil {
		if err := at.setAuthFromConfig(apiConfig.Auth); err != nil {
			return err
		}
	}

	at.initialized = true
	return nil
}

// Cleanup performs cleanup operations
func (at *APITester) Cleanup() error {
	// Close connections, cleanup resources
	at.initialized = false
	return nil
}

// GetName returns the name of the tester
func (at *APITester) GetName() string {
	return "APITester"
}

// Get performs a GET request to the specified endpoint
func (at *APITester) Get(endpoint string, headers map[string]string) (*core.APIResponse, error) {
	if !at.initialized {
		return nil, core.NewGowrightError(core.APIError, "API tester not initialized", nil)
	}

	start := time.Now()

	req := at.client.R()
	if headers != nil {
		req.SetHeaders(headers)
	}

	resp, err := req.Get(endpoint)
	if err != nil {
		return nil, core.NewGowrightError(core.APIError, fmt.Sprintf("GET request failed: %v", err), err)
	}

	duration := time.Since(start)

	return at.buildAPIResponse(resp, duration), nil
}

// Post performs a POST request to the specified endpoint
func (at *APITester) Post(endpoint string, body interface{}, headers map[string]string) (*core.APIResponse, error) {
	if !at.initialized {
		return nil, core.NewGowrightError(core.APIError, "API tester not initialized", nil)
	}

	start := time.Now()

	req := at.client.R()
	if headers != nil {
		req.SetHeaders(headers)
	}
	if body != nil {
		req.SetBody(body)
	}

	resp, err := req.Post(endpoint)
	if err != nil {
		return nil, core.NewGowrightError(core.APIError, fmt.Sprintf("POST request failed: %v", err), err)
	}

	duration := time.Since(start)

	return at.buildAPIResponse(resp, duration), nil
}

// Put performs a PUT request to the specified endpoint
func (at *APITester) Put(endpoint string, body interface{}, headers map[string]string) (*core.APIResponse, error) {
	if !at.initialized {
		return nil, core.NewGowrightError(core.APIError, "API tester not initialized", nil)
	}

	start := time.Now()

	req := at.client.R()
	if headers != nil {
		req.SetHeaders(headers)
	}
	if body != nil {
		req.SetBody(body)
	}

	resp, err := req.Put(endpoint)
	if err != nil {
		return nil, core.NewGowrightError(core.APIError, fmt.Sprintf("PUT request failed: %v", err), err)
	}

	duration := time.Since(start)

	return at.buildAPIResponse(resp, duration), nil
}

// Delete performs a DELETE request to the specified endpoint
func (at *APITester) Delete(endpoint string, headers map[string]string) (*core.APIResponse, error) {
	if !at.initialized {
		return nil, core.NewGowrightError(core.APIError, "API tester not initialized", nil)
	}

	start := time.Now()

	req := at.client.R()
	if headers != nil {
		req.SetHeaders(headers)
	}

	resp, err := req.Delete(endpoint)
	if err != nil {
		return nil, core.NewGowrightError(core.APIError, fmt.Sprintf("DELETE request failed: %v", err), err)
	}

	duration := time.Since(start)

	return at.buildAPIResponse(resp, duration), nil
}

// SetAuth sets authentication for API requests
func (at *APITester) SetAuth(auth *config.AuthConfig) error {
	if !at.initialized {
		return core.NewGowrightError(core.APIError, "API tester not initialized", nil)
	}

	return at.setAuthFromConfig(auth)
}

// ExecuteTest executes an API test and returns the result
func (at *APITester) ExecuteTest(test *core.APITest) *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    core.TestStatusPassed,
	}

	at.asserter.Reset()

	// Execute HTTP request
	var response *core.APIResponse
	var err error

	switch test.Method {
	case "GET":
		response, err = at.Get(test.Endpoint, test.Headers)
	case "POST":
		response, err = at.Post(test.Endpoint, test.Body, test.Headers)
	case "PUT":
		response, err = at.Put(test.Endpoint, test.Body, test.Headers)
	case "DELETE":
		response, err = at.Delete(test.Endpoint, test.Headers)
	default:
		result.Status = core.TestStatusError
		result.Error = core.NewGowrightError(core.APIError, "unsupported HTTP method: "+test.Method, nil)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	if err != nil {
		result.Status = core.TestStatusError
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Validate response against expectations
	if test.Expected != nil {
		at.validateResponse(response, test.Expected)
	}

	// Check for assertion failures
	if at.asserter.HasFailures() {
		result.Status = core.TestStatusFailed
		result.Error = core.NewGowrightError(core.AssertionError, "one or more assertions failed", nil)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Steps = at.asserter.GetSteps()

	return result
}

// validateResponse validates the API response against expectations
func (at *APITester) validateResponse(response *core.APIResponse, expected *core.APIExpectation) {
	// Validate status code
	if expected.StatusCode != 0 {
		at.asserter.Equal(expected.StatusCode, response.StatusCode, "Status code validation")
	}

	// Validate headers
	for key, expectedValue := range expected.Headers {
		if actualValue, exists := response.Headers[key]; exists {
			at.asserter.Equal(expectedValue, actualValue, "Header validation: "+key)
		} else {
			at.asserter.True(false, "Header exists: "+key)
		}
	}

	// Additional body and JSON path validations would go here
}

// setAuthFromConfig configures authentication on the HTTP client
func (at *APITester) setAuthFromConfig(auth *config.AuthConfig) error {
	switch auth.Type {
	case "bearer":
		if auth.Token == "" {
			return core.NewGowrightError(core.ConfigurationError, "bearer token is required", nil)
		}
		at.client.SetAuthToken(auth.Token)
	case "basic":
		if auth.Username == "" || auth.Password == "" {
			return core.NewGowrightError(core.ConfigurationError, "username and password are required for basic auth", nil)
		}
		at.client.SetBasicAuth(auth.Username, auth.Password)
	case "api_key":
		if auth.APIKey == "" {
			return core.NewGowrightError(core.ConfigurationError, "API key is required", nil)
		}
		// Set API key as header (common pattern)
		at.client.SetHeader("X-API-Key", auth.APIKey)
	default:
		return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("unsupported auth type: %s", auth.Type), nil)
	}

	// Set additional headers if provided
	if auth.Headers != nil {
		at.client.SetHeaders(auth.Headers)
	}

	return nil
}

// buildAPIResponse converts a resty response to our APIResponse format
func (at *APITester) buildAPIResponse(resp *resty.Response, duration time.Duration) *core.APIResponse {
	headers := make(map[string]string)
	for key, values := range resp.Header() {
		if len(values) > 0 {
			headers[key] = values[0] // Take first value if multiple
		}
	}

	apiResp := &core.APIResponse{
		StatusCode: resp.StatusCode(),
		Headers:    headers,
		Body:       resp.Body(),
		Duration:   duration,
	}

	// Try to parse JSON if content type is JSON
	contentType := headers["Content-Type"]
	if contentType == "application/json" && len(resp.Body()) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(resp.Body(), &jsonData); err == nil {
			if jsonMap, ok := jsonData.(map[string]interface{}); ok {
				apiResp.JSON = jsonMap
			}
		}
	}

	return apiResp
}
