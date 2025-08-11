package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// APITestImpl implements the Test interface for API testing
type APITestImpl struct {
	Name     string               `json:"name"`
	Method   string               `json:"method"`
	Endpoint string               `json:"endpoint"`
	Headers  map[string]string    `json:"headers,omitempty"`
	Body     interface{}          `json:"body,omitempty"`
	Expected *core.APIExpectation `json:"expected"`
	tester   *APITester
}

// NewAPITest creates a new API test instance
func NewAPITest(name, method, endpoint string, tester *APITester) *APITestImpl {
	return &APITestImpl{
		Name:     name,
		Method:   strings.ToUpper(method),
		Endpoint: endpoint,
		Headers:  make(map[string]string),
		tester:   tester,
	}
}

// GetName returns the name of the test
func (at *APITestImpl) GetName() string {
	return at.Name
}

// Execute runs the API test and returns the result
func (at *APITestImpl) Execute() *core.TestCaseResult {
	startTime := time.Now()

	result := &core.TestCaseResult{
		Name:      at.Name,
		StartTime: startTime,
		Status:    core.TestStatusPassed,
		Logs:      []string{},
	}

	// Execute the HTTP request
	var response *core.APIResponse
	var err error

	switch at.Method {
	case "GET":
		response, err = at.tester.Get(at.Endpoint, at.Headers)
	case "POST":
		response, err = at.tester.Post(at.Endpoint, at.Body, at.Headers)
	case "PUT":
		response, err = at.tester.Put(at.Endpoint, at.Body, at.Headers)
	case "DELETE":
		response, err = at.tester.Delete(at.Endpoint, at.Headers)
	default:
		result.Status = core.TestStatusError
		result.Error = core.NewGowrightError(core.APIError, fmt.Sprintf("unsupported HTTP method: %s", at.Method), nil)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result
	}

	if err != nil {
		result.Status = core.TestStatusError
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result
	}

	// Validate response if expectations are set
	if at.Expected != nil {
		if err := at.validateResponse(response); err != nil {
			result.Status = core.TestStatusFailed
			result.Error = err
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Logs = append(result.Logs, fmt.Sprintf("API %s request to %s completed", at.Method, at.Endpoint))

	return result
}

// validateResponse validates the API response against expectations
func (at *APITestImpl) validateResponse(response *core.APIResponse) error {
	// Validate status code
	if at.Expected.StatusCode != 0 && response.StatusCode != at.Expected.StatusCode {
		return core.NewGowrightError(core.AssertionError,
			fmt.Sprintf("expected status code %d, got %d", at.Expected.StatusCode, response.StatusCode), nil)
	}

	// Validate headers
	for key, expectedValue := range at.Expected.Headers {
		if actualValue, exists := response.Headers[key]; !exists {
			return core.NewGowrightError(core.AssertionError,
				fmt.Sprintf("expected header '%s' not found", key), nil)
		} else if actualValue != expectedValue {
			return core.NewGowrightError(core.AssertionError,
				fmt.Sprintf("expected header '%s' to be '%s', got '%s'", key, expectedValue, actualValue), nil)
		}
	}

	// Validate JSON body if expected
	if at.Expected.Body != nil {
		var actualBody interface{}
		if err := json.Unmarshal(response.Body, &actualBody); err != nil {
			return core.NewGowrightError(core.AssertionError, "failed to parse response body as JSON", err)
		}

		if !reflect.DeepEqual(at.Expected.Body, actualBody) {
			return core.NewGowrightError(core.AssertionError,
				fmt.Sprintf("response body mismatch: expected %v, got %v", at.Expected.Body, actualBody), nil)
		}
	}

	// Validate JSON path expressions
	for path, expectedValue := range at.Expected.JSONPath {
		actualValue, err := at.extractJSONPath(response.Body, path)
		if err != nil {
			return core.NewGowrightError(core.AssertionError,
				fmt.Sprintf("failed to extract JSON path '%s': %v", path, err), err)
		}

		if !reflect.DeepEqual(expectedValue, actualValue) {
			return core.NewGowrightError(core.AssertionError,
				fmt.Sprintf("JSON path '%s' mismatch: expected %v, got %v", path, expectedValue, actualValue), nil)
		}
	}

	return nil
}

// extractJSONPath extracts a value from JSON using a simple path expression
func (at *APITestImpl) extractJSONPath(jsonData []byte, path string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Simple path parsing (e.g., "data.users[0].name")
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		// Handle array indexing
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			arrayPart := part[:strings.Index(part, "[")]
			indexPart := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			// Navigate to array
			if arrayPart != "" {
				if obj, ok := current.(map[string]interface{}); ok {
					current = obj[arrayPart]
				} else {
					return nil, fmt.Errorf("expected object at path '%s'", arrayPart)
				}
			}

			// Parse index
			index, err := strconv.Atoi(indexPart)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s'", indexPart)
			}

			// Navigate to array element
			if arr, ok := current.([]interface{}); ok {
				if index < 0 || index >= len(arr) {
					return nil, fmt.Errorf("array index %d out of bounds", index)
				}
				current = arr[index]
			} else {
				return nil, fmt.Errorf("expected array at path")
			}
		} else {
			// Navigate to object property
			if obj, ok := current.(map[string]interface{}); ok {
				current = obj[part]
			} else {
				return nil, fmt.Errorf("expected object at path '%s'", part)
			}
		}
	}

	return current, nil
}

// SetHeader sets a header for the API request
func (at *APITestImpl) SetHeader(key, value string) *APITestImpl {
	if at.Headers == nil {
		at.Headers = make(map[string]string)
	}
	at.Headers[key] = value
	return at
}

// SetHeaders sets multiple headers for the API request
func (at *APITestImpl) SetHeaders(headers map[string]string) *APITestImpl {
	if at.Headers == nil {
		at.Headers = make(map[string]string)
	}
	for key, value := range headers {
		at.Headers[key] = value
	}
	return at
}

// SetBody sets the request body
func (at *APITestImpl) SetBody(body interface{}) *APITestImpl {
	at.Body = body
	return at
}

// SetExpectedStatusCode sets the expected status code
func (at *APITestImpl) SetExpectedStatusCode(statusCode int) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &core.APIExpectation{}
	}
	at.Expected.StatusCode = statusCode
	return at
}

// SetExpectedHeader sets an expected response header
func (at *APITestImpl) SetExpectedHeader(key, value string) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &core.APIExpectation{}
	}
	if at.Expected.Headers == nil {
		at.Expected.Headers = make(map[string]string)
	}
	at.Expected.Headers[key] = value
	return at
}

// SetExpectedBody sets the expected response body
func (at *APITestImpl) SetExpectedBody(body interface{}) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &core.APIExpectation{}
	}
	at.Expected.Body = body
	return at
}

// SetExpectedJSONPath sets an expected JSON path value
func (at *APITestImpl) SetExpectedJSONPath(path string, value interface{}) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &core.APIExpectation{}
	}
	if at.Expected.JSONPath == nil {
		at.Expected.JSONPath = make(map[string]interface{})
	}
	at.Expected.JSONPath[path] = value
	return at
}

// ValidateJSONSchema validates the response against a JSON schema
func (at *APITestImpl) ValidateJSONSchema(schema string) error {
	// This would implement JSON schema validation
	// For now, it's a placeholder
	return nil
}

// ValidateResponseTime validates that the response time is within acceptable limits
func (at *APITestImpl) ValidateResponseTime(maxDuration time.Duration) error {
	// This would be implemented by tracking response time during execution
	// For now, it's a placeholder
	return nil
}

// ValidateRegex validates response body against a regex pattern
func (at *APITestImpl) ValidateRegex(pattern string) error {
	// This would validate response body against regex
	// For now, it's a placeholder
	return nil
}

// Clone creates a copy of the API test
func (at *APITestImpl) Clone() *APITestImpl {
	clone := &APITestImpl{
		Name:     at.Name,
		Method:   at.Method,
		Endpoint: at.Endpoint,
		Headers:  make(map[string]string),
		Body:     at.Body,
		tester:   at.tester,
	}

	// Copy headers
	for k, v := range at.Headers {
		clone.Headers[k] = v
	}

	// Copy expectations
	if at.Expected != nil {
		clone.Expected = &core.APIExpectation{
			StatusCode: at.Expected.StatusCode,
			Body:       at.Expected.Body,
		}

		if at.Expected.Headers != nil {
			clone.Expected.Headers = make(map[string]string)
			for k, v := range at.Expected.Headers {
				clone.Expected.Headers[k] = v
			}
		}

		if at.Expected.JSONPath != nil {
			clone.Expected.JSONPath = make(map[string]interface{})
			for k, v := range at.Expected.JSONPath {
				clone.Expected.JSONPath[k] = v
			}
		}
	}

	return clone
}
