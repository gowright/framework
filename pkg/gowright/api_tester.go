package gowright

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// APITesterImpl implements the APITester interface using go-resty
type APITesterImpl struct {
	client *resty.Client
	config *APIConfig
	name   string
}

// NewAPITester creates a new APITester instance
func NewAPITester(config *APIConfig) *APITesterImpl {
	if config == nil {
		config = &APIConfig{
			Timeout: 30 * time.Second,
			Headers: make(map[string]string),
		}
	}

	client := resty.New()
	
	return &APITesterImpl{
		client: client,
		config: config,
		name:   "APITester",
	}
}

// Initialize sets up the APITester with the provided configuration
func (at *APITesterImpl) Initialize(config interface{}) error {
	apiConfig, ok := config.(*APIConfig)
	if !ok {
		return NewGowrightError(ConfigurationError, "invalid configuration type for APITester", nil)
	}

	if err := apiConfig.Validate(); err != nil {
		return NewGowrightError(ConfigurationError, "API configuration validation failed", err)
	}

	at.config = apiConfig
	
	// Configure the resty client
	at.client.SetTimeout(apiConfig.Timeout)
	
	if apiConfig.BaseURL != "" {
		at.client.SetBaseURL(apiConfig.BaseURL)
	}
	
	// Set default headers
	if apiConfig.Headers != nil {
		at.client.SetHeaders(apiConfig.Headers)
	}
	
	// Configure authentication if provided
	if apiConfig.AuthConfig != nil {
		if err := at.SetAuth(apiConfig.AuthConfig); err != nil {
			return NewGowrightError(ConfigurationError, "failed to configure authentication", err)
		}
	}
	
	// Configure retry mechanism
	at.client.SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return nil
}

// Cleanup performs any necessary cleanup operations
func (at *APITesterImpl) Cleanup() error {
	// Reset client configuration
	at.client = resty.New()
	return nil
}

// GetName returns the name of the tester
func (at *APITesterImpl) GetName() string {
	return at.name
}

// Get performs a GET request to the specified endpoint
func (at *APITesterImpl) Get(endpoint string, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("GET", endpoint, nil, headers)
}

// Post performs a POST request to the specified endpoint
func (at *APITesterImpl) Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("POST", endpoint, body, headers)
}

// Put performs a PUT request to the specified endpoint
func (at *APITesterImpl) Put(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("PUT", endpoint, body, headers)
}

// Delete performs a DELETE request to the specified endpoint
func (at *APITesterImpl) Delete(endpoint string, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("DELETE", endpoint, nil, headers)
}

// Patch performs a PATCH request to the specified endpoint
func (at *APITesterImpl) Patch(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("PATCH", endpoint, body, headers)
}

// Head performs a HEAD request to the specified endpoint
func (at *APITesterImpl) Head(endpoint string, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("HEAD", endpoint, nil, headers)
}

// Options performs an OPTIONS request to the specified endpoint
func (at *APITesterImpl) Options(endpoint string, headers map[string]string) (*APIResponse, error) {
	return at.executeRequest("OPTIONS", endpoint, nil, headers)
}

// SetAuth sets authentication for API requests
func (at *APITesterImpl) SetAuth(auth *AuthConfig) error {
	if auth == nil {
		return NewGowrightError(ConfigurationError, "authentication configuration cannot be nil", nil)
	}

	if err := auth.Validate(); err != nil {
		return NewGowrightError(ConfigurationError, "authentication configuration validation failed", err)
	}

	switch strings.ToLower(auth.Type) {
	case "bearer":
		at.client.SetAuthToken(auth.Token)
	case "basic":
		at.client.SetBasicAuth(auth.Username, auth.Password)
	case "api_key":
		// API key can be set in headers or as a query parameter
		if auth.Headers != nil {
			for key, value := range auth.Headers {
				at.client.SetHeader(key, value)
			}
		} else {
			// Default to Authorization header if no custom headers specified
			at.client.SetHeader("Authorization", "ApiKey "+auth.Token)
		}
	case "oauth2":
		// For OAuth2, we expect the token to be provided
		if auth.Token != "" {
			at.client.SetAuthToken(auth.Token)
		} else {
			return NewGowrightError(ConfigurationError, "OAuth2 token is required", nil)
		}
	default:
		return NewGowrightError(ConfigurationError, fmt.Sprintf("unsupported authentication type: %s", auth.Type), nil)
	}

	return nil
}

// executeRequest is a helper method that executes HTTP requests
func (at *APITesterImpl) executeRequest(method, endpoint string, body interface{}, headers map[string]string) (*APIResponse, error) {
	startTime := time.Now()
	
	// Create request
	req := at.client.R()
	
	// Set request headers
	if headers != nil {
		req.SetHeaders(headers)
	}
	
	// Set request body if provided
	if body != nil {
		req.SetBody(body)
	}
	
	// Execute request based on method
	var resp *resty.Response
	var err error
	
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = req.Get(endpoint)
	case "POST":
		resp, err = req.Post(endpoint)
	case "PUT":
		resp, err = req.Put(endpoint)
	case "DELETE":
		resp, err = req.Delete(endpoint)
	case "PATCH":
		resp, err = req.Patch(endpoint)
	case "HEAD":
		resp, err = req.Head(endpoint)
	case "OPTIONS":
		resp, err = req.Options(endpoint)
	default:
		return nil, NewGowrightError(APIError, fmt.Sprintf("unsupported HTTP method: %s", method), nil)
	}
	
	duration := time.Since(startTime)
	
	if err != nil {
		return nil, NewGowrightError(APIError, fmt.Sprintf("HTTP request failed: %s %s", method, endpoint), err).
			WithContext("method", method).
			WithContext("endpoint", endpoint).
			WithContext("duration", duration)
	}
	
	// Convert resty response to APIResponse
	apiResponse := &APIResponse{
		StatusCode: resp.StatusCode(),
		Headers:    resp.Header(),
		Body:       resp.Body(),
		Duration:   duration,
	}
	
	return apiResponse, nil
}

// SetBaseURL sets the base URL for all requests
func (at *APITesterImpl) SetBaseURL(baseURL string) {
	at.client.SetBaseURL(baseURL)
	if at.config != nil {
		at.config.BaseURL = baseURL
	}
}

// SetTimeout sets the request timeout
func (at *APITesterImpl) SetTimeout(timeout time.Duration) {
	at.client.SetTimeout(timeout)
	if at.config != nil {
		at.config.Timeout = timeout
	}
}

// SetHeader sets a default header for all requests
func (at *APITesterImpl) SetHeader(key, value string) {
	at.client.SetHeader(key, value)
	if at.config != nil && at.config.Headers != nil {
		at.config.Headers[key] = value
	}
}

// SetHeaders sets multiple default headers for all requests
func (at *APITesterImpl) SetHeaders(headers map[string]string) {
	at.client.SetHeaders(headers)
	if at.config != nil {
		if at.config.Headers == nil {
			at.config.Headers = make(map[string]string)
		}
		for k, v := range headers {
			at.config.Headers[k] = v
		}
	}
}

// GetClient returns the underlying resty client for advanced usage
func (at *APITesterImpl) GetClient() *resty.Client {
	return at.client
}

// GetConfig returns the current API configuration
func (at *APITesterImpl) GetConfig() *APIConfig {
	return at.config
}

// ValidateJSONResponse validates that the response body is valid JSON
func (at *APITesterImpl) ValidateJSONResponse(response *APIResponse) error {
	if response == nil {
		return NewGowrightError(APIError, "response cannot be nil", nil)
	}
	
	var jsonData interface{}
	if err := json.Unmarshal(response.Body, &jsonData); err != nil {
		return NewGowrightError(APIError, "response body is not valid JSON", err).
			WithContext("body", string(response.Body))
	}
	
	return nil
}

// GetJSONResponse parses the response body as JSON and returns the parsed data
func (at *APITesterImpl) GetJSONResponse(response *APIResponse) (interface{}, error) {
	if response == nil {
		return nil, NewGowrightError(APIError, "response cannot be nil", nil)
	}
	
	var jsonData interface{}
	if err := json.Unmarshal(response.Body, &jsonData); err != nil {
		return nil, NewGowrightError(APIError, "failed to parse JSON response", err).
			WithContext("body", string(response.Body))
	}
	
	return jsonData, nil
}

// IsSuccessStatusCode checks if the status code indicates success (2xx)
func (at *APITesterImpl) IsSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsClientErrorStatusCode checks if the status code indicates client error (4xx)
func (at *APITesterImpl) IsClientErrorStatusCode(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerErrorStatusCode checks if the status code indicates server error (5xx)
func (at *APITesterImpl) IsServerErrorStatusCode(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}