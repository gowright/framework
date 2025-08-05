package gowright

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPITester(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		tester := NewAPITester(nil)

		assert.NotNil(t, tester)
		assert.NotNil(t, tester.client)
		assert.NotNil(t, tester.config)
		assert.Equal(t, "APITester", tester.GetName())
		assert.Equal(t, 30*time.Second, tester.config.Timeout)
		assert.NotNil(t, tester.config.Headers)
	})

	t.Run("with valid config", func(t *testing.T) {
		config := &APIConfig{
			BaseURL: "https://api.example.com",
			Timeout: 10 * time.Second,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		tester := NewAPITester(config)

		assert.NotNil(t, tester)
		assert.Equal(t, config, tester.config)
	})
}

func TestAPITesterImpl_Initialize(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		tester := NewAPITester(nil)
		config := &APIConfig{
			BaseURL: "https://api.example.com",
			Timeout: 15 * time.Second,
			Headers: map[string]string{
				"User-Agent": "GoWright/1.0",
			},
			AuthConfig: &AuthConfig{
				Type:  "bearer",
				Token: "test-token",
			},
		}

		err := tester.Initialize(config)

		assert.NoError(t, err)
		assert.Equal(t, config, tester.config)
	})

	t.Run("with invalid config type", func(t *testing.T) {
		tester := NewAPITester(nil)

		err := tester.Initialize("invalid-config")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration type")
	})

	t.Run("with invalid config", func(t *testing.T) {
		tester := NewAPITester(nil)
		config := &APIConfig{
			Timeout: -1 * time.Second, // Invalid timeout
		}

		err := tester.Initialize(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestAPITesterImpl_SetAuth(t *testing.T) {
	tester := NewAPITester(nil)

	t.Run("bearer auth", func(t *testing.T) {
		auth := &AuthConfig{
			Type:  "bearer",
			Token: "test-bearer-token",
		}

		err := tester.SetAuth(auth)

		assert.NoError(t, err)
	})

	t.Run("basic auth", func(t *testing.T) {
		auth := &AuthConfig{
			Type:     "basic",
			Username: "testuser",
			Password: "testpass",
		}

		err := tester.SetAuth(auth)

		assert.NoError(t, err)
	})

	t.Run("api key auth with headers", func(t *testing.T) {
		auth := &AuthConfig{
			Type:  "api_key",
			Token: "test-api-key",
			Headers: map[string]string{
				"X-API-Key": "test-api-key",
			},
		}

		err := tester.SetAuth(auth)

		assert.NoError(t, err)
	})

	t.Run("api key auth without headers", func(t *testing.T) {
		auth := &AuthConfig{
			Type:  "api_key",
			Token: "test-api-key",
		}

		err := tester.SetAuth(auth)

		assert.NoError(t, err)
	})

	t.Run("oauth2 auth", func(t *testing.T) {
		auth := &AuthConfig{
			Type:  "oauth2",
			Token: "oauth2-token",
		}

		err := tester.SetAuth(auth)

		assert.NoError(t, err)
	})

	t.Run("oauth2 auth without token", func(t *testing.T) {
		auth := &AuthConfig{
			Type: "oauth2",
		}

		err := tester.SetAuth(auth)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OAuth2 token is required")
	})

	t.Run("unsupported auth type", func(t *testing.T) {
		auth := &AuthConfig{
			Type: "unsupported",
		}

		err := tester.SetAuth(auth)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication configuration validation failed")
	})

	t.Run("nil auth config", func(t *testing.T) {
		err := tester.SetAuth(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication configuration cannot be nil")
	})
}

func TestAPITesterImpl_HTTPMethods(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request information
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": r.Header,
		}

		// Handle request body for POST/PUT/PATCH
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
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

	config := &APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	tester := NewAPITester(config)
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

	t.Run("PATCH request", func(t *testing.T) {
		body := map[string]string{
			"patch": "value",
		}

		response, err := tester.Patch("/test", body, nil)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		// Verify response body
		var responseData map[string]interface{}
		err = json.Unmarshal(response.Body, &responseData)
		assert.NoError(t, err)
		assert.Equal(t, "PATCH", responseData["method"])
	})

	t.Run("HEAD request", func(t *testing.T) {
		response, err := tester.Head("/test", nil)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, response.Headers, "Content-Type")
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		response, err := tester.Options("/test", nil)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

func TestAPITesterImpl_ErrorHandling(t *testing.T) {
	config := &APIConfig{
		BaseURL: "http://invalid-url-that-does-not-exist.local",
		Timeout: 1 * time.Second,
	}
	tester := NewAPITester(config)
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("network error", func(t *testing.T) {
		response, err := tester.Get("/test", nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, APIError, gowrightErr.Type)
		assert.Contains(t, gowrightErr.Context, "method")
		assert.Contains(t, gowrightErr.Context, "endpoint")
	})
}

func TestAPITesterImpl_UtilityMethods(t *testing.T) {
	tester := NewAPITester(nil)

	t.Run("SetBaseURL", func(t *testing.T) {
		baseURL := "https://api.example.com"
		tester.SetBaseURL(baseURL)

		assert.Equal(t, baseURL, tester.config.BaseURL)
	})

	t.Run("SetTimeout", func(t *testing.T) {
		timeout := 20 * time.Second
		tester.SetTimeout(timeout)

		assert.Equal(t, timeout, tester.config.Timeout)
	})

	t.Run("SetHeader", func(t *testing.T) {
		tester.SetHeader("X-Custom", "value")

		assert.Equal(t, "value", tester.config.Headers["X-Custom"])
	})

	t.Run("SetHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Header1": "value1",
			"Header2": "value2",
		}
		tester.SetHeaders(headers)

		assert.Equal(t, "value1", tester.config.Headers["Header1"])
		assert.Equal(t, "value2", tester.config.Headers["Header2"])
	})

	t.Run("GetClient", func(t *testing.T) {
		client := tester.GetClient()

		assert.NotNil(t, client)
		assert.Equal(t, tester.client, client)
	})

	t.Run("GetConfig", func(t *testing.T) {
		config := tester.GetConfig()

		assert.NotNil(t, config)
		assert.Equal(t, tester.config, config)
	})
}

func TestAPITesterImpl_JSONValidation(t *testing.T) {
	tester := NewAPITester(nil)

	t.Run("valid JSON response", func(t *testing.T) {
		response := &APIResponse{
			StatusCode: 200,
			Body:       []byte(`{"key": "value", "number": 123}`),
		}

		err := tester.ValidateJSONResponse(response)
		assert.NoError(t, err)

		data, err := tester.GetJSONResponse(response)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		jsonMap, ok := data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", jsonMap["key"])
		assert.Equal(t, float64(123), jsonMap["number"])
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		response := &APIResponse{
			StatusCode: 200,
			Body:       []byte(`{invalid json`),
		}

		err := tester.ValidateJSONResponse(response)
		assert.Error(t, err)

		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, APIError, gowrightErr.Type)
		assert.Contains(t, gowrightErr.Context, "body")

		data, err := tester.GetJSONResponse(response)
		assert.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("nil response", func(t *testing.T) {
		err := tester.ValidateJSONResponse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response cannot be nil")

		data, err := tester.GetJSONResponse(nil)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "response cannot be nil")
	})
}

func TestAPITesterImpl_StatusCodeCheckers(t *testing.T) {
	tester := NewAPITester(nil)

	t.Run("success status codes", func(t *testing.T) {
		assert.True(t, tester.IsSuccessStatusCode(200))
		assert.True(t, tester.IsSuccessStatusCode(201))
		assert.True(t, tester.IsSuccessStatusCode(204))
		assert.True(t, tester.IsSuccessStatusCode(299))

		assert.False(t, tester.IsSuccessStatusCode(199))
		assert.False(t, tester.IsSuccessStatusCode(300))
		assert.False(t, tester.IsSuccessStatusCode(400))
		assert.False(t, tester.IsSuccessStatusCode(500))
	})

	t.Run("client error status codes", func(t *testing.T) {
		assert.True(t, tester.IsClientErrorStatusCode(400))
		assert.True(t, tester.IsClientErrorStatusCode(401))
		assert.True(t, tester.IsClientErrorStatusCode(404))
		assert.True(t, tester.IsClientErrorStatusCode(499))

		assert.False(t, tester.IsClientErrorStatusCode(399))
		assert.False(t, tester.IsClientErrorStatusCode(500))
		assert.False(t, tester.IsClientErrorStatusCode(200))
	})

	t.Run("server error status codes", func(t *testing.T) {
		assert.True(t, tester.IsServerErrorStatusCode(500))
		assert.True(t, tester.IsServerErrorStatusCode(501))
		assert.True(t, tester.IsServerErrorStatusCode(503))
		assert.True(t, tester.IsServerErrorStatusCode(599))

		assert.False(t, tester.IsServerErrorStatusCode(499))
		assert.False(t, tester.IsServerErrorStatusCode(600))
		assert.False(t, tester.IsServerErrorStatusCode(200))
		assert.False(t, tester.IsServerErrorStatusCode(400))
	})
}

func TestAPITesterImpl_Cleanup(t *testing.T) {
	tester := NewAPITester(&APIConfig{
		BaseURL: "https://api.example.com",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	})

	// Verify initial state
	assert.NotNil(t, tester.client)

	// Cleanup
	err := tester.Cleanup()

	assert.NoError(t, err)
	assert.NotNil(t, tester.client) // Should have a new clean client
}

func TestAPITesterImpl_Integration(t *testing.T) {
	// Create a more complex test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		// Handle different endpoints
		switch r.URL.Path {
		case "/users":
			switch r.Method {
			case "GET":
				users := []map[string]interface{}{
					{"id": 1, "name": "John Doe"},
					{"id": 2, "name": "Jane Smith"},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users)
			case "POST":
				var user map[string]interface{}
				_ = json.NewDecoder(r.Body).Decode(&user)
				user["id"] = 3
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(user)
			}
		case "/users/1":
			switch r.Method {
			case "PUT":
				var user map[string]interface{}
				_ = json.NewDecoder(r.Body).Decode(&user)
				user["id"] = 1
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(user)
			case "DELETE":
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create tester with authentication
	config := &APIConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		AuthConfig: &AuthConfig{
			Type:  "bearer",
			Token: "test-token",
		},
	}

	tester := NewAPITester(config)
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("complete user management workflow", func(t *testing.T) {
		// Get all users
		response, err := tester.Get("/users", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		users, err := tester.GetJSONResponse(response)
		assert.NoError(t, err)
		userList, ok := users.([]interface{})
		assert.True(t, ok)
		assert.Len(t, userList, 2)

		// Create a new user
		newUser := map[string]interface{}{
			"name":  "Bob Johnson",
			"email": "bob@example.com",
		}

		response, err = tester.Post("/users", newUser, map[string]string{
			"Content-Type": "application/json",
		})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)

		createdUser, err := tester.GetJSONResponse(response)
		assert.NoError(t, err)
		userMap, ok := createdUser.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(3), userMap["id"])
		assert.Equal(t, "Bob Johnson", userMap["name"])

		// Update user
		updateData := map[string]interface{}{
			"name":  "John Updated",
			"email": "john.updated@example.com",
		}

		response, err = tester.Put("/users/1", updateData, map[string]string{
			"Content-Type": "application/json",
		})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		// Delete user
		response, err = tester.Delete("/users/1", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, response.StatusCode)
	})

	t.Run("unauthorized request", func(t *testing.T) {
		// Create tester without authentication
		unauthorizedConfig := &APIConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		}
		unauthorizedTester := NewAPITester(unauthorizedConfig)
		err := unauthorizedTester.Initialize(unauthorizedConfig)
		require.NoError(t, err)

		response, err := unauthorizedTester.Get("/users", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)

		errorData, err := unauthorizedTester.GetJSONResponse(response)
		assert.NoError(t, err)
		errorMap, ok := errorData.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "unauthorized", errorMap["error"])
	})
}
