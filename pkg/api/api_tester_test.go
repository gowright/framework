package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gwconfig "github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPITester(t *testing.T) {
	t.Run("with default constructor", func(t *testing.T) {
		tester := NewAPITester()

		assert.NotNil(t, tester)
		assert.Equal(t, "APITester", tester.GetName())
		assert.False(t, tester.initialized)
	})
}

func TestAPITester_Initialize(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		tester := NewAPITester()
		config := &gwconfig.APIConfig{
			BaseURL: "https://api.example.com",
			Timeout: 15 * time.Second,
			DefaultHeaders: map[string]string{
				"User-Agent": "GoWright/1.0",
			},
			Auth: &gwconfig.AuthConfig{
				Type:  "bearer",
				Token: "test-token",
			},
		}

		err := tester.Initialize(config)

		assert.NoError(t, err)
		assert.True(t, tester.initialized)
		assert.Equal(t, config, tester.config)
	})

	t.Run("with invalid config type", func(t *testing.T) {
		tester := NewAPITester()

		err := tester.Initialize("invalid-config")

		assert.Error(t, err)
		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.ConfigurationError, gowrightErr.Type)
	})
}

func TestAPITester_SetAuth(t *testing.T) {
	tester := NewAPITester()
	config := &gwconfig.APIConfig{
		BaseURL: "https://api.example.com",
		Timeout: 30 * time.Second,
	}
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("bearer auth", func(t *testing.T) {
		auth := &gwconfig.AuthConfig{
			Type:  "bearer",
			Token: "test-bearer-token",
		}

		err := tester.SetAuth(auth)
		assert.NoError(t, err)
	})

	t.Run("basic auth", func(t *testing.T) {
		auth := &gwconfig.AuthConfig{
			Type:     "basic",
			Username: "testuser",
			Password: "testpass",
		}

		err := tester.SetAuth(auth)
		assert.NoError(t, err)
	})

	t.Run("api key auth", func(t *testing.T) {
		auth := &gwconfig.AuthConfig{
			Type:   "api_key",
			APIKey: "test-api-key",
		}

		err := tester.SetAuth(auth)
		assert.NoError(t, err)
	})

	t.Run("not initialized", func(t *testing.T) {
		uninitializedTester := NewAPITester()
		auth := &gwconfig.AuthConfig{
			Type:  "bearer",
			Token: "test-token",
		}

		err := uninitializedTester.SetAuth(auth)
		assert.Error(t, err)
		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.APIError, gowrightErr.Type)
	})
}

func TestAPITester_HTTPMethods(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request information
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": r.Header,
		}

		// Handle request body for POST/PUT
		if r.Method == "POST" || r.Method == "PUT" {
			var body interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				response["body"] = body
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &gwconfig.APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester()
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("GET request", func(t *testing.T) {
		headers := map[string]string{
			"X-Custom-Header": "custom-value",
		}

		response, err := tester.Get("/test", headers)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, response.Headers, "Content-Type")
		assert.Contains(t, response.Headers, "X-Test-Header")
		assert.Greater(t, response.Duration, time.Duration(0))

		// Verify response body
		var responseData map[string]interface{}
		err = json.Unmarshal(response.Body, &responseData)
		assert.NoError(t, err)
		assert.Equal(t, "GET", responseData["method"])
		assert.Equal(t, "/test", responseData["path"])
	})

	t.Run("POST request", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "test",
			"value": 123,
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}

		response, err := tester.Post("/test", body, headers)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		// Verify response body
		var responseData map[string]interface{}
		err = json.Unmarshal(response.Body, &responseData)
		assert.NoError(t, err)
		assert.Equal(t, "POST", responseData["method"])
		assert.Equal(t, "/test", responseData["path"])

		// Verify request body was received
		bodyData, ok := responseData["body"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", bodyData["name"])
		assert.Equal(t, float64(123), bodyData["value"]) // JSON numbers are float64
	})

	t.Run("PUT request", func(t *testing.T) {
		body := map[string]string{
			"update": "value",
		}

		response, err := tester.Put("/test", body, nil)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		// Verify response body
		var responseData map[string]interface{}
		err = json.Unmarshal(response.Body, &responseData)
		assert.NoError(t, err)
		assert.Equal(t, "PUT", responseData["method"])
	})

	t.Run("DELETE request", func(t *testing.T) {
		response, err := tester.Delete("/test", nil)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		// Verify response body
		var responseData map[string]interface{}
		err = json.Unmarshal(response.Body, &responseData)
		assert.NoError(t, err)
		assert.Equal(t, "DELETE", responseData["method"])
	})
}

func TestAPITester_ErrorHandling(t *testing.T) {
	config := &gwconfig.APIConfig{
		BaseURL: "http://invalid-url-that-does-not-exist.local",
		Timeout: 1 * time.Second,
	}
	tester := NewAPITester()
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("network error", func(t *testing.T) {
		response, err := tester.Get("/test", nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		gowrightErr, ok := err.(*core.GowrightError)
		assert.True(t, ok)
		assert.Equal(t, core.APIError, gowrightErr.Type)
	})
}

func TestAPITester_ExecuteTest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	config := &gwconfig.APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester()
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("successful test", func(t *testing.T) {
		test := &core.APITest{
			Name:     "Test API",
			Method:   "GET",
			Endpoint: "/test",
			Expected: &core.APIExpectation{
				StatusCode: 200,
			},
		}

		result := tester.ExecuteTest(test)

		assert.NotNil(t, result)
		assert.Equal(t, "Test API", result.Name)
		assert.Equal(t, core.TestStatusPassed, result.Status)
		assert.NotZero(t, result.Duration)
		assert.Nil(t, result.Error)
	})

	t.Run("unsupported method", func(t *testing.T) {
		test := &core.APITest{
			Name:     "Test Unsupported Method",
			Method:   "UNSUPPORTED",
			Endpoint: "/test",
		}

		result := tester.ExecuteTest(test)

		assert.NotNil(t, result)
		assert.Equal(t, core.TestStatusError, result.Status)
		assert.NotNil(t, result.Error)
	})
}

func TestAPITester_NotInitialized(t *testing.T) {
	tester := NewAPITester()

	testCases := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := tester.Get("/test", nil); return err }},
		{"Post", func() error { _, err := tester.Post("/test", nil, nil); return err }},
		{"Put", func() error { _, err := tester.Put("/test", nil, nil); return err }},
		{"Delete", func() error { _, err := tester.Delete("/test", nil); return err }},
		{"SetAuth", func() error { return tester.SetAuth(&gwconfig.AuthConfig{Type: "bearer", Token: "test"}) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Error(t, err)

			gowrightErr, ok := err.(*core.GowrightError)
			assert.True(t, ok, "Error should be of type GowrightError")
			assert.Equal(t, core.APIError, gowrightErr.Type)
		})
	}
}

func TestAPITester_Cleanup(t *testing.T) {
	tester := NewAPITester()
	config := &gwconfig.APIConfig{
		BaseURL: "https://api.example.com",
		Timeout: 30 * time.Second,
	}

	err := tester.Initialize(config)
	require.NoError(t, err)
	assert.True(t, tester.initialized)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
	assert.False(t, tester.initialized)
}

func TestAPITester_ValidateResponse(t *testing.T) {
	tester := NewAPITester()
	config := &gwconfig.APIConfig{
		BaseURL: "https://api.example.com",
		Timeout: 30 * time.Second,
	}
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("status code validation", func(t *testing.T) {
		response := &core.APIResponse{
			StatusCode: 200,
			Headers:    make(map[string]string),
		}

		expected := &core.APIExpectation{
			StatusCode: 200,
		}

		tester.validateResponse(response, expected)

		// Check if assertion passed
		assert.False(t, tester.asserter.HasFailures())
	})

	t.Run("header validation", func(t *testing.T) {
		response := &core.APIResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Custom":     "test-value",
			},
		}

		expected := &core.APIExpectation{
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Custom":     "test-value",
			},
		}

		tester.asserter.Reset()
		tester.validateResponse(response, expected)

		// Check if assertions passed
		assert.False(t, tester.asserter.HasFailures())
	})

	t.Run("failed validation", func(t *testing.T) {
		response := &core.APIResponse{
			StatusCode: 404,
			Headers:    make(map[string]string),
		}

		expected := &core.APIExpectation{
			StatusCode: 200,
		}

		tester.asserter.Reset()
		tester.validateResponse(response, expected)

		// Check if assertion failed
		assert.True(t, tester.asserter.HasFailures())
	})
}
