---
title: Integration Testing
description: Learn how to create end-to-end integration tests with Gowright
---

Integration testing in Gowright allows you to create complex workflows that span multiple systems, combining UI, API, and database testing in a single test scenario.

## Getting Started

### Basic Integration Test

```go
func TestIntegrationBasics(t *testing.T) {
    // Setup all testers
    uiTester := setupUITester(t)
    defer uiTester.Cleanup()
    
    apiTester := setupAPITester(t)
    defer apiTester.Cleanup()
    
    dbTester := setupDatabaseTester(t)
    defer dbTester.Cleanup()
    
    integrationTester := gowright.NewIntegrationTester(uiTester, apiTester, dbTester)
    
    // Your integration tests here
}
```

## Integration Test Builder

### Structured Integration Tests

```go
integrationTest := &gowright.IntegrationTest{
    Name: "User Registration Workflow",
    Steps: []gowright.IntegrationStep{
        {
            Type: gowright.StepTypeAPI,
            Action: gowright.APIStepAction{
                Method:   "POST",
                Endpoint: "/api/users",
                Body: map[string]interface{}{
                    "name":  "Test User",
                    "email": "test@example.com",
                },
            },
            Validation: gowright.APIStepValidation{
                ExpectedStatusCode: http.StatusCreated,
            },
            Name: "Create User via API",
        },
        {
            Type: gowright.StepTypeDatabase,
            Action: gowright.DatabaseStepAction{
                Connection: "main",
                Query:      "SELECT COUNT(*) as count FROM users WHERE email = ?",
                Args:       []interface{}{"test@example.com"},
            },
            Validation: gowright.DatabaseStepValidation{
                ExpectedRowCount: &[]int{1}[0],
            },
            Name: "Verify User in Database",
        },
        {
            Type: gowright.StepTypeUI,
            Action: gowright.UIStepAction{
                Type: "navigate",
                URL:  "https://example.com/users",
            },
            Name: "Navigate to Users Page",
        },
    },
}

result := integrationTester.ExecuteTest(integrationTest)
assert.Equal(t, gowright.TestStatusPassed, result.Status)
```

## Full Stack Workflows

### Complete User Journey

```go
func TestCompleteUserJourney(t *testing.T) {
    // Step 1: Create user via API
    newUser := map[string]interface{}{
        "name":  "Integration Test User",
        "email": "integration@example.com",
    }
    
    apiResponse, err := apiTester.Post("/users", newUser, nil)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, apiResponse.StatusCode)
    
    // Step 2: Verify user in database
    dbResult, err := dbTester.Execute("main", 
        "SELECT id, name, email FROM users WHERE email = ?", 
        "integration@example.com")
    require.NoError(t, err)
    assert.Equal(t, 1, len(dbResult.Rows))
    
    // Step 3: Login via UI
    err = uiTester.Navigate("https://example.com/login")
    require.NoError(t, err)
    
    err = uiTester.Type("#email", "integration@example.com")
    require.NoError(t, err)
    
    err = uiTester.Type("#password", "testpassword123")
    require.NoError(t, err)
    
    err = uiTester.Click("#login-button")
    require.NoError(t, err)
    
    // Step 4: Verify successful login
    err = uiTester.WaitForElement(".dashboard", 10*time.Second)
    require.NoError(t, err)
}
```

## Error Handling and Rollback

### Rollback Mechanisms

```go
integrationTest := &gowright.IntegrationTest{
    Name: "User Registration with Rollback",
    Steps: []gowright.IntegrationStep{
        // ... test steps
    },
    Rollback: []gowright.IntegrationStep{
        {
            Type: gowright.StepTypeDatabase,
            Action: gowright.DatabaseStepAction{
                Connection: "main",
                Query:      "DELETE FROM users WHERE email = ?",
                Args:       []interface{}{"test@example.com"},
            },
            Name: "Cleanup Test User",
        },
    },
}
```

For more examples, see the [Integration Examples](/examples/integration/) section.