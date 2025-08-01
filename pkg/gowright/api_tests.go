package gowright

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// APITestImpl implements the Test interface for API testing
type APITestImpl struct {
	Name     string                 `json:"name"`
	Method   string                 `json:"method"`
	Endpoint string                 `json:"endpoint"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Body     interface{}            `json:"body,omitempty"`
	Expected *APIExpectation        `json:"expected"`
	tester   *APITesterImpl
}

// NewAPITest creates a new API test instance
func NewAPITest(name, method, endpoint string, tester *APITesterImpl) *APITestImpl {
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
func (at *APITestImpl) Execute() *TestCaseResult {
	startTime := time.Now()
	
	result := &TestCaseResult{
		Name:      at.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
		Logs:      []string{},
	}
	
	// Log test execution start
	result.Logs = append(result.Logs, fmt.Sprintf("Starting API test: %s %s", at.Method, at.Endpoint))
	
	// Execute the HTTP request
	response, err := at.executeRequest()
	if err != nil {
		result.Status = TestStatusError
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		result.Logs = append(result.Logs, fmt.Sprintf("Request failed: %v", err))
		return result
	}
	
	result.Logs = append(result.Logs, fmt.Sprintf("Request completed with status: %d", response.StatusCode))
	
	// Validate response if expectations are defined
	if at.Expected != nil {
		if err := at.validateResponse(response); err != nil {
			result.Status = TestStatusFailed
			result.Error = err
			result.Logs = append(result.Logs, fmt.Sprintf("Validation failed: %v", err))
		} else {
			result.Logs = append(result.Logs, "All validations passed")
		}
	} else {
		result.Logs = append(result.Logs, "No expectations defined, test passed")
	}
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	
	return result
}

// executeRequest performs the HTTP request based on the test configuration
func (at *APITestImpl) executeRequest() (*APIResponse, error) {
	if at.tester == nil {
		return nil, NewGowrightError(APIError, "API tester not configured", nil)
	}
	
	switch at.Method {
	case "GET":
		return at.tester.Get(at.Endpoint, at.Headers)
	case "POST":
		return at.tester.Post(at.Endpoint, at.Body, at.Headers)
	case "PUT":
		return at.tester.Put(at.Endpoint, at.Body, at.Headers)
	case "DELETE":
		return at.tester.Delete(at.Endpoint, at.Headers)
	case "PATCH":
		return at.tester.Patch(at.Endpoint, at.Body, at.Headers)
	case "HEAD":
		return at.tester.Head(at.Endpoint, at.Headers)
	case "OPTIONS":
		return at.tester.Options(at.Endpoint, at.Headers)
	default:
		return nil, NewGowrightError(APIError, fmt.Sprintf("unsupported HTTP method: %s", at.Method), nil)
	}
}

// validateResponse validates the API response against expectations
func (at *APITestImpl) validateResponse(response *APIResponse) error {
	if at.Expected == nil {
		return nil
	}
	
	// Validate status code
	if at.Expected.StatusCode != 0 && response.StatusCode != at.Expected.StatusCode {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("status code mismatch: expected %d, got %d", at.Expected.StatusCode, response.StatusCode), nil).
			WithContext("expected_status", at.Expected.StatusCode).
			WithContext("actual_status", response.StatusCode)
	}
	
	// Validate headers
	if err := at.validateHeaders(response); err != nil {
		return err
	}
	
	// Validate body
	if err := at.validateBody(response); err != nil {
		return err
	}
	
	// Validate JSON path expressions
	if err := at.validateJSONPath(response); err != nil {
		return err
	}
	
	return nil
}

// validateHeaders validates response headers
func (at *APITestImpl) validateHeaders(response *APIResponse) error {
	if at.Expected.Headers == nil {
		return nil
	}
	
	for expectedKey, expectedValue := range at.Expected.Headers {
		actualValues, exists := response.Headers[expectedKey]
		if !exists {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected header '%s' not found", expectedKey), nil).
				WithContext("expected_header", expectedKey)
		}
		
		// Check if any of the actual values match the expected value
		found := false
		for _, actualValue := range actualValues {
			if actualValue == expectedValue {
				found = true
				break
			}
		}
		
		if !found {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("header '%s' value mismatch: expected '%s', got %v", expectedKey, expectedValue, actualValues), nil).
				WithContext("header_key", expectedKey).
				WithContext("expected_value", expectedValue).
				WithContext("actual_values", actualValues)
		}
	}
	
	return nil
}

// validateBody validates response body
func (at *APITestImpl) validateBody(response *APIResponse) error {
	if at.Expected.Body == nil {
		return nil
	}
	
	// Determine content type for appropriate validation
	contentType := at.getContentType(response)
	
	switch {
	case strings.Contains(contentType, "application/json"):
		return at.validateJSONBody(response)
	case strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml"):
		return at.validateXMLBody(response)
	default:
		return at.validateTextBody(response)
	}
}

// validateJSONBody validates JSON response body
func (at *APITestImpl) validateJSONBody(response *APIResponse) error {
	var actualData interface{}
	if err := json.Unmarshal(response.Body, &actualData); err != nil {
		return NewGowrightError(AssertionError, "response body is not valid JSON", err).
			WithContext("body", string(response.Body))
	}
	
	// Compare with expected data
	if !at.compareValues(at.Expected.Body, actualData) {
		return NewGowrightError(AssertionError, "JSON body mismatch", nil).
			WithContext("expected", at.Expected.Body).
			WithContext("actual", actualData)
	}
	
	return nil
}

// validateXMLBody validates XML response body
func (at *APITestImpl) validateXMLBody(response *APIResponse) error {
	// For XML validation, we'll do a simple string comparison for now
	// In a more sophisticated implementation, we could parse and compare XML structures
	expectedStr, ok := at.Expected.Body.(string)
	if !ok {
		return NewGowrightError(AssertionError, "expected XML body must be a string", nil)
	}
	
	actualStr := string(response.Body)
	
	// Normalize whitespace for comparison
	expectedNormalized := at.normalizeXML(expectedStr)
	actualNormalized := at.normalizeXML(actualStr)
	
	if expectedNormalized != actualNormalized {
		return NewGowrightError(AssertionError, "XML body mismatch", nil).
			WithContext("expected", expectedStr).
			WithContext("actual", actualStr)
	}
	
	return nil
}

// validateTextBody validates plain text response body
func (at *APITestImpl) validateTextBody(response *APIResponse) error {
	expectedStr := fmt.Sprintf("%v", at.Expected.Body)
	actualStr := string(response.Body)
	
	if expectedStr != actualStr {
		return NewGowrightError(AssertionError, "text body mismatch", nil).
			WithContext("expected", expectedStr).
			WithContext("actual", actualStr)
	}
	
	return nil
}

// validateJSONPath validates JSON path expressions
func (at *APITestImpl) validateJSONPath(response *APIResponse) error {
	if at.Expected.JSONPath == nil {
		return nil
	}
	
	// Parse response as JSON
	var jsonData interface{}
	if err := json.Unmarshal(response.Body, &jsonData); err != nil {
		return NewGowrightError(AssertionError, "cannot validate JSON path on non-JSON response", err)
	}
	
	// Validate each JSON path expression
	for path, expectedValue := range at.Expected.JSONPath {
		actualValue, err := at.evaluateJSONPath(jsonData, path)
		if err != nil {
			return NewGowrightError(AssertionError, fmt.Sprintf("JSON path evaluation failed: %s", path), err).
				WithContext("json_path", path)
		}
		
		if !at.compareValues(expectedValue, actualValue) {
			return NewGowrightError(AssertionError, fmt.Sprintf("JSON path value mismatch at '%s'", path), nil).
				WithContext("json_path", path).
				WithContext("expected", expectedValue).
				WithContext("actual", actualValue)
		}
	}
	
	return nil
}

// evaluateJSONPath evaluates a simple JSON path expression
// This is a basic implementation - in production, you might want to use a proper JSON path library
func (at *APITestImpl) evaluateJSONPath(data interface{}, path string) (interface{}, error) {
	if path == "" || path == "$" {
		return data, nil
	}
	
	// Remove leading $ if present
	if strings.HasPrefix(path, "$.") {
		path = path[2:]
	} else if strings.HasPrefix(path, "$") {
		path = path[1:]
	}
	
	// Split path into segments
	segments := strings.Split(path, ".")
	current := data
	
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		
		// Handle array indexing
		if strings.Contains(segment, "[") && strings.Contains(segment, "]") {
			current = at.handleArrayAccess(current, segment)
			if current == nil {
				return nil, fmt.Errorf("array access failed for segment: %s", segment)
			}
		} else {
			// Handle object property access
			current = at.handleObjectAccess(current, segment)
			if current == nil {
				return nil, fmt.Errorf("property access failed for segment: %s", segment)
			}
		}
	}
	
	return current, nil
}

// handleArrayAccess handles array indexing in JSON path
func (at *APITestImpl) handleArrayAccess(data interface{}, segment string) interface{} {
	// Extract property name and index
	parts := strings.Split(segment, "[")
	if len(parts) != 2 {
		return nil
	}
	
	propertyName := parts[0]
	indexStr := strings.TrimSuffix(parts[1], "]")
	
	// Get the property first if it exists
	if propertyName != "" {
		data = at.handleObjectAccess(data, propertyName)
		if data == nil {
			return nil
		}
	}
	
	// Convert to slice
	slice, ok := data.([]interface{})
	if !ok {
		return nil
	}
	
	// Parse index
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil
	}
	
	// Check bounds
	if index < 0 || index >= len(slice) {
		return nil
	}
	
	return slice[index]
}

// handleObjectAccess handles object property access in JSON path
func (at *APITestImpl) handleObjectAccess(data interface{}, property string) interface{} {
	switch obj := data.(type) {
	case map[string]interface{}:
		return obj[property]
	case map[interface{}]interface{}:
		return obj[property]
	default:
		return nil
	}
}

// compareValues compares two values for equality, handling different types appropriately
func (at *APITestImpl) compareValues(expected, actual interface{}) bool {
	// Handle nil cases
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}
	
	// Use reflection for deep comparison
	return reflect.DeepEqual(expected, actual)
}

// getContentType extracts content type from response headers
func (at *APITestImpl) getContentType(response *APIResponse) string {
	contentTypes, exists := response.Headers["Content-Type"]
	if !exists || len(contentTypes) == 0 {
		return ""
	}
	return strings.ToLower(contentTypes[0])
}

// normalizeXML normalizes XML string for comparison by removing extra whitespace
func (at *APITestImpl) normalizeXML(xmlStr string) string {
	// Remove extra whitespace between tags
	re := regexp.MustCompile(`>\s+<`)
	normalized := re.ReplaceAllString(xmlStr, "><")
	
	// Trim leading/trailing whitespace
	return strings.TrimSpace(normalized)
}

// SetHeader sets a request header
func (at *APITestImpl) SetHeader(key, value string) *APITestImpl {
	if at.Headers == nil {
		at.Headers = make(map[string]string)
	}
	at.Headers[key] = value
	return at
}

// SetHeaders sets multiple request headers
func (at *APITestImpl) SetHeaders(headers map[string]string) *APITestImpl {
	if at.Headers == nil {
		at.Headers = make(map[string]string)
	}
	for k, v := range headers {
		at.Headers[k] = v
	}
	return at
}

// SetBody sets the request body
func (at *APITestImpl) SetBody(body interface{}) *APITestImpl {
	at.Body = body
	return at
}

// SetExpectedStatus sets the expected status code
func (at *APITestImpl) SetExpectedStatus(statusCode int) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &APIExpectation{}
	}
	at.Expected.StatusCode = statusCode
	return at
}

// SetExpectedHeader sets an expected response header
func (at *APITestImpl) SetExpectedHeader(key, value string) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &APIExpectation{}
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
		at.Expected = &APIExpectation{}
	}
	at.Expected.Body = body
	return at
}

// SetExpectedJSONPath sets an expected JSON path value
func (at *APITestImpl) SetExpectedJSONPath(path string, value interface{}) *APITestImpl {
	if at.Expected == nil {
		at.Expected = &APIExpectation{}
	}
	if at.Expected.JSONPath == nil {
		at.Expected.JSONPath = make(map[string]interface{})
	}
	at.Expected.JSONPath[path] = value
	return at
}

// WithTester sets the API tester for this test
func (at *APITestImpl) WithTester(tester *APITesterImpl) *APITestImpl {
	at.tester = tester
	return at
}

// APITestBuilder provides a fluent interface for building API tests
type APITestBuilder struct {
	test *APITestImpl
}

// NewAPITestBuilder creates a new API test builder
func NewAPITestBuilder(name, method, endpoint string) *APITestBuilder {
	return &APITestBuilder{
		test: &APITestImpl{
			Name:     name,
			Method:   strings.ToUpper(method),
			Endpoint: endpoint,
			Headers:  make(map[string]string),
		},
	}
}

// WithTester sets the API tester
func (b *APITestBuilder) WithTester(tester *APITesterImpl) *APITestBuilder {
	b.test.tester = tester
	return b
}

// WithHeader adds a request header
func (b *APITestBuilder) WithHeader(key, value string) *APITestBuilder {
	b.test.SetHeader(key, value)
	return b
}

// WithHeaders adds multiple request headers
func (b *APITestBuilder) WithHeaders(headers map[string]string) *APITestBuilder {
	b.test.SetHeaders(headers)
	return b
}

// WithBody sets the request body
func (b *APITestBuilder) WithBody(body interface{}) *APITestBuilder {
	b.test.SetBody(body)
	return b
}

// ExpectStatus sets the expected status code
func (b *APITestBuilder) ExpectStatus(statusCode int) *APITestBuilder {
	b.test.SetExpectedStatus(statusCode)
	return b
}

// ExpectHeader sets an expected response header
func (b *APITestBuilder) ExpectHeader(key, value string) *APITestBuilder {
	b.test.SetExpectedHeader(key, value)
	return b
}

// ExpectBody sets the expected response body
func (b *APITestBuilder) ExpectBody(body interface{}) *APITestBuilder {
	b.test.SetExpectedBody(body)
	return b
}

// ExpectJSONPath sets an expected JSON path value
func (b *APITestBuilder) ExpectJSONPath(path string, value interface{}) *APITestBuilder {
	b.test.SetExpectedJSONPath(path, value)
	return b
}

// Build returns the constructed API test
func (b *APITestBuilder) Build() *APITestImpl {
	return b.test
}