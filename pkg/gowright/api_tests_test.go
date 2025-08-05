package gowright

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPITest(t *testing.T) {
	tester := NewAPITester(nil)

	test := NewAPITest("test-api", "GET", "/users", tester)

	assert.Equal(t, "test-api", test.Name)
	assert.Equal(t, "GET", test.Method)
	assert.Equal(t, "/users", test.Endpoint)
	assert.NotNil(t, test.Headers)
	assert.Equal(t, tester, test.tester)
}

func TestAPITestImpl_Execute(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   []string{"item1", "item2"},
			})
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		case "/users":
			if r.Method == "POST" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"id":   123,
					"name": "John Doe",
				})
			}
		}
	}))
	defer server.Close()

	// Create tester
	config := &APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester(config)
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("successful test without expectations", func(t *testing.T) {
		test := NewAPITest("success-test", "GET", "/success", tester)

		result := test.Execute()

		assert.Equal(t, "success-test", result.Name)
		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
		assert.Greater(t, result.Duration, time.Duration(0))
		assert.NotEmpty(t, result.Logs)
		assert.Contains(t, result.Logs[len(result.Logs)-1], "No expectations defined")
	})

	t.Run("successful test with valid expectations", func(t *testing.T) {
		test := NewAPITest("success-with-expectations", "GET", "/success", tester)
		test.SetExpectedStatus(200)
		test.SetExpectedHeader("Content-Type", "application/json")
		test.SetExpectedJSONPath("$.status", "success")
		test.SetExpectedJSONPath("$.data[0]", "item1")

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
		assert.Contains(t, result.Logs[len(result.Logs)-1], "All validations passed")
	})

	t.Run("test with failed expectations", func(t *testing.T) {
		test := NewAPITest("failed-expectations", "GET", "/success", tester)
		test.SetExpectedStatus(201) // Wrong status code

		result := test.Execute()

		assert.Equal(t, TestStatusFailed, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "status code mismatch")
		assert.Contains(t, result.Logs[len(result.Logs)-1], "Validation failed")
	})

	t.Run("test with network error", func(t *testing.T) {
		// Create tester with invalid URL
		invalidConfig := &APIConfig{
			BaseURL: "http://invalid-url-that-does-not-exist.local",
			Timeout: 1 * time.Second,
		}
		invalidTester := NewAPITester(invalidConfig)
		err := invalidTester.Initialize(invalidConfig)
		require.NoError(t, err)

		test := NewAPITest("network-error", "GET", "/test", invalidTester)

		result := test.Execute()

		assert.Equal(t, TestStatusError, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Logs[len(result.Logs)-1], "Request failed")
	})

	t.Run("POST test with body", func(t *testing.T) {
		test := NewAPITest("post-test", "POST", "/users", tester)
		test.SetBody(map[string]interface{}{
			"name": "John Doe",
		})
		test.SetExpectedStatus(201)
		test.SetExpectedJSONPath("$.id", float64(123)) // JSON numbers are float64
		test.SetExpectedJSONPath("$.name", "John Doe")

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
	})
}

func TestAPITestImpl_ValidationMethods(t *testing.T) {
	// Create test server with various response types
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Custom-Header", "custom-value")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "hello",
				"count":   42,
				"items":   []string{"a", "b", "c"},
				"nested": map[string]interface{}{
					"key": "value",
				},
			})
		case "/xml":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<root><message>hello</message></root>`))
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("plain text response"))
		}
	}))
	defer server.Close()

	config := &APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester(config)
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("JSON response validation", func(t *testing.T) {
		test := NewAPITest("json-test", "GET", "/json", tester)
		test.SetExpectedStatus(200)
		test.SetExpectedHeader("Content-Type", "application/json")
		test.SetExpectedHeader("X-Custom-Header", "custom-value")
		test.SetExpectedJSONPath("$.message", "hello")
		test.SetExpectedJSONPath("$.count", float64(42))
		test.SetExpectedJSONPath("$.items[1]", "b")
		test.SetExpectedJSONPath("$.nested.key", "value")

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("JSON response validation with full body", func(t *testing.T) {
		expectedBody := map[string]interface{}{
			"message": "hello",
			"count":   float64(42),
			"items":   []interface{}{"a", "b", "c"},
			"nested": map[string]interface{}{
				"key": "value",
			},
		}

		test := NewAPITest("json-body-test", "GET", "/json", tester)
		test.SetExpectedStatus(200)
		test.SetExpectedBody(expectedBody)

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("XML response validation", func(t *testing.T) {
		test := NewAPITest("xml-test", "GET", "/xml", tester)
		test.SetExpectedStatus(200)
		test.SetExpectedBody("<root><message>hello</message></root>")

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("text response validation", func(t *testing.T) {
		test := NewAPITest("text-test", "GET", "/text", tester)
		test.SetExpectedStatus(200)
		test.SetExpectedBody("plain text response")

		result := test.Execute()

		assert.Equal(t, TestStatusPassed, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("failed header validation", func(t *testing.T) {
		test := NewAPITest("failed-header", "GET", "/json", tester)
		test.SetExpectedHeader("X-Missing-Header", "value")

		result := test.Execute()

		assert.Equal(t, TestStatusFailed, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "expected header 'X-Missing-Header' not found")
	})

	t.Run("failed JSON path validation", func(t *testing.T) {
		test := NewAPITest("failed-jsonpath", "GET", "/json", tester)
		test.SetExpectedJSONPath("$.message", "wrong-value")

		result := test.Execute()

		assert.Equal(t, TestStatusFailed, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "JSON path value mismatch")
	})

	t.Run("invalid JSON path", func(t *testing.T) {
		test := NewAPITest("invalid-jsonpath", "GET", "/json", tester)
		test.SetExpectedJSONPath("$.nonexistent.deep.path", "value")

		result := test.Execute()

		assert.Equal(t, TestStatusFailed, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "JSON path evaluation failed")
	})
}

func TestAPITestImpl_JSONPathEvaluation(t *testing.T) {
	test := &APITestImpl{}

	testData := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "Anytown",
		},
		"hobbies": []interface{}{"reading", "swimming", "coding"},
		"scores":  []interface{}{85, 92, 78},
	}

	testCases := []struct {
		path     string
		expected interface{}
	}{
		{"$", testData},
		{"$.name", "John"},
		{"$.age", 30},
		{"$.address.street", "123 Main St"},
		{"$.address.city", "Anytown"},
		{"$.hobbies[0]", "reading"},
		{"$.hobbies[2]", "coding"},
		{"$.scores[1]", 92},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path_%s", tc.path), func(t *testing.T) {
			result, err := test.evaluateJSONPath(testData, tc.path)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}

	// Test error cases
	errorCases := []struct {
		path        string
		description string
	}{
		{"$.nonexistent", "nonexistent property"},
		{"$.hobbies[10]", "array index out of bounds"},
		{"$.name[0]", "array access on non-array"},
		{"$.address.nonexistent", "nonexistent nested property"},
	}

	for _, tc := range errorCases {
		t.Run(fmt.Sprintf("error_%s", tc.description), func(t *testing.T) {
			result, err := test.evaluateJSONPath(testData, tc.path)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestAPITestImpl_FluentInterface(t *testing.T) {
	tester := NewAPITester(nil)

	test := NewAPITest("fluent-test", "POST", "/api/users", tester).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer token").
		SetBody(map[string]string{"name": "John"}).
		SetExpectedStatus(201).
		SetExpectedHeader("Location", "/users/123").
		SetExpectedJSONPath("$.id", 123)

	assert.Equal(t, "fluent-test", test.Name)
	assert.Equal(t, "POST", test.Method)
	assert.Equal(t, "/api/users", test.Endpoint)
	assert.Equal(t, "application/json", test.Headers["Content-Type"])
	assert.Equal(t, "Bearer token", test.Headers["Authorization"])
	assert.NotNil(t, test.Body)
	assert.Equal(t, 201, test.Expected.StatusCode)
	assert.Equal(t, "/users/123", test.Expected.Headers["Location"])
	assert.Equal(t, 123, test.Expected.JSONPath["$.id"])
}

func TestAPITestBuilder(t *testing.T) {
	tester := NewAPITester(nil)

	test := NewAPITestBuilder("builder-test", "PUT", "/api/users/1").
		WithTester(tester).
		WithHeader("Content-Type", "application/json").
		WithHeaders(map[string]string{
			"Authorization": "Bearer token",
			"X-Client-ID":   "test-client",
		}).
		WithBody(map[string]interface{}{
			"name":  "Jane Doe",
			"email": "jane@example.com",
		}).
		ExpectStatus(200).
		ExpectHeader("Content-Type", "application/json").
		ExpectJSONPath("$.name", "Jane Doe").
		ExpectJSONPath("$.email", "jane@example.com").
		Build()

	assert.Equal(t, "builder-test", test.Name)
	assert.Equal(t, "PUT", test.Method)
	assert.Equal(t, "/api/users/1", test.Endpoint)
	assert.Equal(t, tester, test.tester)
	assert.Equal(t, "application/json", test.Headers["Content-Type"])
	assert.Equal(t, "Bearer token", test.Headers["Authorization"])
	assert.Equal(t, "test-client", test.Headers["X-Client-ID"])
	assert.NotNil(t, test.Body)
	assert.Equal(t, 200, test.Expected.StatusCode)
	assert.Equal(t, "application/json", test.Expected.Headers["Content-Type"])
	assert.Equal(t, "Jane Doe", test.Expected.JSONPath["$.name"])
	assert.Equal(t, "jane@example.com", test.Expected.JSONPath["$.email"])
}

func TestAPITestImpl_HTTPMethods(t *testing.T) {
	// Create test server that echoes method
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"method": r.Method,
		})
	}))
	defer server.Close()

	config := &APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester(config)
	err := tester.Initialize(config)
	require.NoError(t, err)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			test := NewAPITest(fmt.Sprintf("%s-test", method), method, "/test", tester)
			if method == "POST" || method == "PUT" || method == "PATCH" {
				test.SetBody(map[string]string{"test": "data"})
			}

			if method != "HEAD" { // HEAD responses don't have body
				test.SetExpectedJSONPath("$.method", method)
			}

			result := test.Execute()

			assert.Equal(t, TestStatusPassed, result.Status, "Method %s should pass", method)
			assert.NoError(t, result.Error, "Method %s should not error", method)
		})
	}

	t.Run("unsupported method", func(t *testing.T) {
		test := NewAPITest("unsupported-method", "INVALID", "/test", tester)

		result := test.Execute()

		assert.Equal(t, TestStatusError, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "unsupported HTTP method")
	})
}

func TestAPITestImpl_ErrorHandling(t *testing.T) {
	t.Run("test without tester", func(t *testing.T) {
		test := NewAPITest("no-tester", "GET", "/test", nil)

		result := test.Execute()

		assert.Equal(t, TestStatusError, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "API tester not configured")
	})

	t.Run("invalid JSON in response for JSON path validation", func(t *testing.T) {
		// Create server that returns invalid JSON
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{invalid json"))
		}))
		defer server.Close()

		config := &APIConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		}
		tester := NewAPITester(config)
		err := tester.Initialize(config)
		require.NoError(t, err)

		test := NewAPITest("invalid-json", "GET", "/test", tester)
		test.SetExpectedJSONPath("$.key", "value")

		result := test.Execute()

		assert.Equal(t, TestStatusFailed, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "cannot validate JSON path on non-JSON response")
	})
}

func TestAPITestImpl_CompareValues(t *testing.T) {
	test := &APITestImpl{}

	testCases := []struct {
		name     string
		expected interface{}
		actual   interface{}
		result   bool
	}{
		{"both nil", nil, nil, true},
		{"expected nil", nil, "value", false},
		{"actual nil", "value", nil, false},
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"equal numbers", 42, 42, true},
		{"different numbers", 42, 43, false},
		{"equal floats", 3.14, 3.14, true},
		{"equal arrays", []interface{}{1, 2, 3}, []interface{}{1, 2, 3}, true},
		{"different arrays", []interface{}{1, 2, 3}, []interface{}{1, 2, 4}, false},
		{"equal maps", map[string]interface{}{"key": "value"}, map[string]interface{}{"key": "value"}, true},
		{"different maps", map[string]interface{}{"key": "value1"}, map[string]interface{}{"key": "value2"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := test.compareValues(tc.expected, tc.actual)
			assert.Equal(t, tc.result, result)
		})
	}
}

func TestAPITestImpl_XMLNormalization(t *testing.T) {
	test := &APITestImpl{}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"simple XML",
			"<root><child>value</child></root>",
			"<root><child>value</child></root>",
		},
		{
			"XML with whitespace",
			"<root>\n  <child>value</child>\n</root>",
			"<root><child>value</child></root>",
		},
		{
			"XML with extra spaces",
			"<root>  <child>  value  </child>  </root>",
			"<root><child>  value  </child></root>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := test.normalizeXML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
