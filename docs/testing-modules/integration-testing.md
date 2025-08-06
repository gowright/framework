# Integration Testing

Gowright's integration testing module orchestrates complex workflows that span multiple systems, combining UI automation, API testing, and database validation into cohesive end-to-end test scenarios.

## Overview

Integration testing in Gowright provides:

- Multi-module workflow orchestration
- Cross-system test coordination
- Data flow validation across layers
- Error handling and rollback mechanisms
- Performance testing across integrated systems
- Real-world scenario simulation
- Comprehensive reporting across all test steps

## Basic Concepts

### Integration Test Structure

An integration test consists of multiple steps that can involve:
- **UI Steps**: Browser interactions and validations
- **API Steps**: HTTP requests and response validations
- **Database Steps**: Data queries and validations
- **Custom Steps**: Business logic and external system interactions

### Step Dependencies

Steps can depend on previous steps' results:
- Data from API responses can be used in subsequent UI interactions
- Database IDs can be used in API requests
- UI form submissions can trigger API calls that update databases

## Basic Usage

### Simple Integration Test

```go
package main

import (
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestUserRegistrationWorkflow(t *testing.T) {
    // Initialize framework
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    err := framework.Initialize()
    assert.NoError(t, err)
    
    // Create integration tester
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    // Define integration test
    integrationTest := &gowright.IntegrationTest{
        Name: "User Registration End-to-End Workflow",
        Description: "Test complete user registration from UI to database persistence",
        Steps: []gowright.IntegrationStep{
            {
                Name: "Navigate to Registration Page",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Navigate: "https://example.com/register",
                },
                Validation: gowright.UIStepValidation{
                    ExpectedTitle: "Register - Example App",
                    ElementExists: []string{"form#registration-form"},
                },
            },
            {
                Name: "Fill Registration Form",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Interactions: []gowright.UIInteraction{
                        {Type: "type", Selector: "input[name='firstName']", Value: "John"},
                        {Type: "type", Selector: "input[name='lastName']", Value: "Doe"},
                        {Type: "type", Selector: "input[name='email']", Value: "john.doe@example.com"},
                        {Type: "type", Selector: "input[name='password']", Value: "SecurePass123!"},
                        {Type: "click", Selector: "button[type='submit']"},
                    },
                },
                Validation: gowright.UIStepValidation{
                    WaitForElement: "div.success-message",
                    ExpectedText:   []string{"Registration successful"},
                },
            },
            {
                Name: "Verify User via API",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "GET",
                    Endpoint: "/api/users?email=john.doe@example.com",
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.data[0].email":     "john.doe@example.com",
                        "$.data[0].firstName": "John",
                        "$.data[0].lastName":  "Doe",
                        "$.data[0].id":        gowright.NotNil,
                    },
                },
                OutputVariables: map[string]string{
                    "userId": "$.data[0].id",
                },
            },
            {
                Name: "Confirm User in Database",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT id, email, first_name, last_name, created_at FROM users WHERE email = ?",
                    Args:       []interface{}{"john.doe@example.com"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                    ColumnValues: map[string]interface{}{
                        "email":      "john.doe@example.com",
                        "first_name": "John",
                        "last_name":  "Doe",
                    },
                    CustomValidations: []gowright.CustomValidation{
                        {
                            Column: "created_at",
                            Validator: func(value interface{}) bool {
                                // Verify user was created recently (within last 5 minutes)
                                createdAt, ok := value.(time.Time)
                                if !ok {
                                    return false
                                }
                                return time.Since(createdAt) < 5*time.Minute
                            },
                        },
                    },
                },
            },
        },
        Cleanup: []gowright.CleanupStep{
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM users WHERE email = ?",
                    Args:       []interface{}{"john.doe@example.com"},
                },
            },
        },
    }
    
    // Execute integration test
    result := integrationTester.ExecuteTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    assert.Equal(t, 4, len(result.StepResults))
    
    // Verify all steps passed
    for _, stepResult := range result.StepResults {
        assert.Equal(t, gowright.TestStatusPassed, stepResult.Status)
    }
}
```

## Advanced Integration Patterns

### E-commerce Order Flow

```go
func TestEcommerceOrderFlow(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    integrationTest := &gowright.IntegrationTest{
        Name: "Complete E-commerce Order Flow",
        Steps: []gowright.IntegrationStep{
            // Step 1: Create product via API
            {
                Name: "Create Test Product",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/products",
                    Body: map[string]interface{}{
                        "name":        "Test Product",
                        "price":       29.99,
                        "description": "A test product for integration testing",
                        "stock":       100,
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 201,
                    JSONPath: map[string]interface{}{
                        "$.id":    gowright.NotNil,
                        "$.name":  "Test Product",
                        "$.price": 29.99,
                    },
                },
                OutputVariables: map[string]string{
                    "productId": "$.id",
                },
            },
            
            // Step 2: Navigate to product page
            {
                Name: "Navigate to Product Page",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Navigate: "https://shop.example.com/products/{{.productId}}",
                },
                Validation: gowright.UIStepValidation{
                    ElementExists: []string{"button.add-to-cart"},
                    ExpectedText:  []string{"Test Product", "$29.99"},
                },
            },
            
            // Step 3: Add to cart
            {
                Name: "Add Product to Cart",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Interactions: []gowright.UIInteraction{
                        {Type: "click", Selector: "button.add-to-cart"},
                        {Type: "wait", Selector: "div.cart-notification"},
                    },
                },
                Validation: gowright.UIStepValidation{
                    ExpectedText: []string{"Added to cart"},
                },
            },
            
            // Step 4: Proceed to checkout
            {
                Name: "Navigate to Checkout",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Interactions: []gowright.UIInteraction{
                        {Type: "click", Selector: "a.cart-link"},
                        {Type: "click", Selector: "button.checkout"},
                    },
                },
                Validation: gowright.UIStepValidation{
                    ElementExists: []string{"form.checkout-form"},
                },
            },
            
            // Step 5: Fill checkout form
            {
                Name: "Complete Checkout Form",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Interactions: []gowright.UIInteraction{
                        {Type: "type", Selector: "input[name='email']", Value: "customer@example.com"},
                        {Type: "type", Selector: "input[name='firstName']", Value: "Jane"},
                        {Type: "type", Selector: "input[name='lastName']", Value: "Smith"},
                        {Type: "type", Selector: "input[name='address']", Value: "123 Main St"},
                        {Type: "type", Selector: "input[name='city']", Value: "Anytown"},
                        {Type: "type", Selector: "input[name='zipCode']", Value: "12345"},
                        {Type: "type", Selector: "input[name='cardNumber']", Value: "4111111111111111"},
                        {Type: "type", Selector: "input[name='expiryDate']", Value: "12/25"},
                        {Type: "type", Selector: "input[name='cvv']", Value: "123"},
                        {Type: "click", Selector: "button.place-order"},
                    },
                },
                Validation: gowright.UIStepValidation{
                    WaitForElement: "div.order-confirmation",
                    ExpectedText:   []string{"Order placed successfully"},
                },
                OutputVariables: map[string]string{
                    "orderNumber": "div.order-number",
                },
            },
            
            // Step 6: Verify order via API
            {
                Name: "Verify Order via API",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "GET",
                    Endpoint: "/api/orders/{{.orderNumber}}",
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.status":           "pending",
                        "$.customerEmail":    "customer@example.com",
                        "$.items[0].productId": "{{.productId}}",
                        "$.total":            29.99,
                    },
                },
                OutputVariables: map[string]string{
                    "orderId": "$.id",
                },
            },
            
            // Step 7: Verify database records
            {
                Name: "Verify Order in Database",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query: `
                        SELECT o.id, o.status, o.total, oi.product_id, oi.quantity, oi.price
                        FROM orders o
                        JOIN order_items oi ON o.id = oi.order_id
                        WHERE o.order_number = ?
                    `,
                    Args: []interface{}{"{{.orderNumber}}"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                    ColumnValues: map[string]interface{}{
                        "status":     "pending",
                        "total":      29.99,
                        "product_id": "{{.productId}}",
                        "quantity":   1,
                        "price":      29.99,
                    },
                },
            },
            
            // Step 8: Update inventory
            {
                Name: "Verify Inventory Update",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT stock FROM products WHERE id = ?",
                    Args:       []interface{}{"{{.productId}}"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                    ColumnValues: map[string]interface{}{
                        "stock": 99, // Should be reduced by 1
                    },
                },
            },
        },
        Cleanup: []gowright.CleanupStep{
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM order_items WHERE order_id = (SELECT id FROM orders WHERE order_number = ?)",
                    Args:       []interface{}{"{{.orderNumber}}"},
                },
            },
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM orders WHERE order_number = ?",
                    Args:       []interface{}{"{{.orderNumber}}"},
                },
            },
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM products WHERE id = ?",
                    Args:       []interface{}{"{{.productId}}"},
                },
            },
        },
    }
    
    result := integrationTester.ExecuteTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### API-First Integration Testing

```go
func TestAPIFirstIntegration(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    integrationTest := &gowright.IntegrationTest{
        Name: "API-First User Management Flow",
        Steps: []gowright.IntegrationStep{
            // Create user via API
            {
                Name: "Create User via API",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/users",
                    Body: map[string]interface{}{
                        "username": "apiuser",
                        "email":    "apiuser@example.com",
                        "password": "SecurePass123!",
                        "role":     "user",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 201,
                    JSONPath: map[string]interface{}{
                        "$.id":       gowright.NotNil,
                        "$.username": "apiuser",
                        "$.email":    "apiuser@example.com",
                        "$.role":     "user",
                    },
                },
                OutputVariables: map[string]string{
                    "userId": "$.id",
                },
            },
            
            // Verify user can login via API
            {
                Name: "Login via API",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/auth/login",
                    Body: map[string]interface{}{
                        "email":    "apiuser@example.com",
                        "password": "SecurePass123!",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.token":    gowright.NotEmpty,
                        "$.user.id":  "{{.userId}}",
                        "$.expires": gowright.NotNil,
                    },
                },
                OutputVariables: map[string]string{
                    "authToken": "$.token",
                },
            },
            
            // Use token to access protected resource
            {
                Name: "Access Protected Resource",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "GET",
                    Endpoint: "/api/users/{{.userId}}/profile",
                    Headers: map[string]string{
                        "Authorization": "Bearer {{.authToken}}",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.id":       "{{.userId}}",
                        "$.username": "apiuser",
                        "$.email":    "apiuser@example.com",
                    },
                },
            },
            
            // Update user profile via API
            {
                Name: "Update User Profile",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "PUT",
                    Endpoint: "/api/users/{{.userId}}",
                    Headers: map[string]string{
                        "Authorization": "Bearer {{.authToken}}",
                    },
                    Body: map[string]interface{}{
                        "firstName": "API",
                        "lastName":  "User",
                        "bio":       "Created via API integration test",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.firstName": "API",
                        "$.lastName":  "User",
                        "$.bio":       "Created via API integration test",
                    },
                },
            },
            
            // Verify UI reflects API changes
            {
                Name: "Verify Profile in UI",
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Navigate: "https://example.com/login",
                    Interactions: []gowright.UIInteraction{
                        {Type: "type", Selector: "input[name='email']", Value: "apiuser@example.com"},
                        {Type: "type", Selector: "input[name='password']", Value: "SecurePass123!"},
                        {Type: "click", Selector: "button[type='submit']"},
                        {Type: "wait", Selector: "a[href='/profile']"},
                        {Type: "click", Selector: "a[href='/profile']"},
                    },
                },
                Validation: gowright.UIStepValidation{
                    ExpectedText: []string{"API User", "Created via API integration test"},
                    ElementExists: []string{"div.profile-info"},
                },
            },
            
            // Verify database consistency
            {
                Name: "Verify Database State",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query: `
                        SELECT username, email, first_name, last_name, bio, role
                        FROM users 
                        WHERE id = ?
                    `,
                    Args: []interface{}{"{{.userId}}"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                    ColumnValues: map[string]interface{}{
                        "username":   "apiuser",
                        "email":      "apiuser@example.com",
                        "first_name": "API",
                        "last_name":  "User",
                        "bio":        "Created via API integration test",
                        "role":       "user",
                    },
                },
            },
        },
        Cleanup: []gowright.CleanupStep{
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM users WHERE id = ?",
                    Args:       []interface{}{"{{.userId}}"},
                },
            },
        },
    }
    
    result := integrationTester.ExecuteTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Error Handling and Recovery

### Graceful Error Handling

```go
func TestErrorHandlingIntegration(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    integrationTest := &gowright.IntegrationTest{
        Name: "Error Handling and Recovery",
        ErrorHandling: gowright.ErrorHandlingConfig{
            ContinueOnError: true,
            MaxRetries:      3,
            RetryDelay:      2 * time.Second,
        },
        Steps: []gowright.IntegrationStep{
            {
                Name: "Attempt Invalid API Call",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/invalid-endpoint",
                    Body:     map[string]interface{}{"test": "data"},
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 404, // Expect this to fail
                },
                ErrorHandling: gowright.StepErrorHandling{
                    ExpectError:     true,
                    ContinueOnError: true,
                },
            },
            {
                Name: "Fallback to Valid API Call",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "GET",
                    Endpoint: "/api/health",
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                },
                Condition: gowright.StepCondition{
                    RunIf: "previous_step_failed",
                },
            },
            {
                Name: "Test Database Recovery",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT 1 as health_check",
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                },
                ErrorHandling: gowright.StepErrorHandling{
                    RetryOnError: true,
                    MaxRetries:   3,
                    RetryDelay:   1 * time.Second,
                },
            },
        },
    }
    
    result := integrationTester.ExecuteTest(integrationTest)
    // Test should pass despite the first step failing
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Rollback Mechanisms

```go
func TestRollbackMechanism(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    integrationTest := &gowright.IntegrationTest{
        Name: "Transaction Rollback Test",
        Steps: []gowright.IntegrationStep{
            {
                Name: "Create User",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/users",
                    Body: map[string]interface{}{
                        "username": "rollbackuser",
                        "email":    "rollback@example.com",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 201,
                },
                OutputVariables: map[string]string{
                    "userId": "$.id",
                },
                Rollback: gowright.RollbackAction{
                    Type: gowright.StepTypeAPI,
                    Action: gowright.APIStepAction{
                        Method:   "DELETE",
                        Endpoint: "/api/users/{{.userId}}",
                    },
                },
            },
            {
                Name: "Create User Profile",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/users/{{.userId}}/profile",
                    Body: map[string]interface{}{
                        "firstName": "Rollback",
                        "lastName":  "User",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 201,
                },
                Rollback: gowright.RollbackAction{
                    Type: gowright.StepTypeAPI,
                    Action: gowright.APIStepAction{
                        Method:   "DELETE",
                        Endpoint: "/api/users/{{.userId}}/profile",
                    },
                },
            },
            {
                Name: "Simulate Failure",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/simulate-error",
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 500, // This will fail
                },
                ErrorHandling: gowright.StepErrorHandling{
                    ExpectError:      true,
                    TriggerRollback: true, // Trigger rollback of previous steps
                },
            },
        },
    }
    
    result := integrationTester.ExecuteTest(integrationTest)
    
    // Verify rollback was executed
    assert.Equal(t, gowright.TestStatusFailed, result.Status)
    assert.True(t, result.RollbackExecuted)
    
    // Verify user was cleaned up
    apiTester := framework.GetAPITester()
    response, err := apiTester.Get("/api/users/{{.userId}}", nil)
    assert.NoError(t, err)
    assert.Equal(t, 404, response.StatusCode) // User should not exist
}
```

## Performance Integration Testing

### Load Testing Across Systems

```go
func TestPerformanceIntegration(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    integrationTest := &gowright.IntegrationTest{
        Name: "Performance Integration Test",
        Performance: gowright.PerformanceConfig{
            ConcurrentUsers: 10,
            Duration:        30 * time.Second,
            RampUpTime:      5 * time.Second,
        },
        Steps: []gowright.IntegrationStep{
            {
                Name: "Load Test User Creation",
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/users",
                    Body: map[string]interface{}{
                        "username": "loadtest_{{.iteration}}",
                        "email":    "loadtest{{.iteration}}@example.com",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 201,
                    MaxResponseTime:    500 * time.Millisecond,
                },
                Performance: gowright.StepPerformanceConfig{
                    MaxExecutionTime: 1 * time.Second,
                    SuccessRate:      0.95, // 95% success rate required
                },
            },
            {
                Name: "Verify Database Performance",
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT COUNT(*) as user_count FROM users WHERE username LIKE 'loadtest_%'",
                },
                Validation: gowright.DatabaseStepValidation{
                    CustomValidations: []gowright.CustomValidation{
                        {
                            Column: "user_count",
                            Validator: func(value interface{}) bool {
                                count, ok := value.(int)
                                return ok && count > 0
                            },
                        },
                    },
                },
                Performance: gowright.StepPerformanceConfig{
                    MaxExecutionTime: 100 * time.Millisecond,
                },
            },
        },
        Cleanup: []gowright.CleanupStep{
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "DELETE FROM users WHERE username LIKE 'loadtest_%'",
                },
            },
        },
    }
    
    result := integrationTester.ExecutePerformanceTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    assert.GreaterOrEqual(t, result.PerformanceMetrics.SuccessRate, 0.95)
    assert.LessOrEqual(t, result.PerformanceMetrics.AverageResponseTime, 500*time.Millisecond)
}
```

## Configuration Examples

### Complete Integration Configuration

```json
{
  "integration_config": {
    "default_timeout": "30s",
    "max_retries": 3,
    "retry_delay": "2s",
    "parallel_execution": false,
    "error_handling": {
      "continue_on_error": false,
      "capture_screenshots_on_error": true,
      "capture_logs_on_error": true
    },
    "performance": {
      "enable_metrics": true,
      "max_execution_time": "5m",
      "resource_limits": {
        "max_memory_mb": 512,
        "max_cpu_percent": 80
      }
    },
    "reporting": {
      "detailed_step_reports": true,
      "include_variable_values": true,
      "screenshot_on_failure": true
    }
  }
}
```

## Best Practices

### 1. Design for Maintainability

```go
// Good - Use descriptive step names and clear structure
integrationTest := &gowright.IntegrationTest{
    Name: "User Registration with Email Verification",
    Steps: []gowright.IntegrationStep{
        {
            Name: "Submit Registration Form",
            // Clear, specific step name
        },
        {
            Name: "Verify Email Sent via API",
            // Describes what and how
        },
    },
}
```

### 2. Use Variable Passing

```go
// Good - Pass data between steps
{
    Name: "Create User",
    OutputVariables: map[string]string{
        "userId": "$.id",
        "email":  "$.email",
    },
},
{
    Name: "Verify User",
    Action: gowright.APIStepAction{
        Endpoint: "/api/users/{{.userId}}", // Use variable from previous step
    },
}
```

### 3. Implement Proper Cleanup

```go
// Always include cleanup steps
integrationTest := &gowright.IntegrationTest{
    // ... test steps
    Cleanup: []gowright.CleanupStep{
        {
            Type: gowright.StepTypeDatabase,
            Action: gowright.DatabaseStepAction{
                Query: "DELETE FROM test_data WHERE created_by = 'integration_test'",
            },
        },
    },
}
```

### 4. Handle Timing Issues

```go
// Good - Use explicit waits
{
    Name: "Wait for Async Process",
    Type: gowright.StepTypeCustom,
    Action: gowright.CustomStepAction{
        Function: func(ctx context.Context) error {
            return waitForCondition(func() bool {
                // Check if async process completed
                return checkAsyncProcessStatus()
            }, 30*time.Second)
        },
    },
}
```

### 5. Test Error Scenarios

```go
// Include negative test cases
integrationTest := &gowright.IntegrationTest{
    Name: "Error Handling Integration",
    Steps: []gowright.IntegrationStep{
        {
            Name: "Test Invalid Input",
            ErrorHandling: gowright.StepErrorHandling{
                ExpectError: true, // This step should fail
            },
        },
        {
            Name: "Verify Error Response",
            // Validate error handling worked correctly
        },
    },
}
```

## Troubleshooting

### Common Issues

**Variable substitution not working:**
```go
// Ensure variables are properly captured
OutputVariables: map[string]string{
    "userId": "$.id", // JSONPath for API responses
    "userId": "id",   // Column name for database results
}

// Use correct template syntax
Endpoint: "/api/users/{{.userId}}" // Correct
Endpoint: "/api/users/${userId}"   // Incorrect
```

**Timing issues between steps:**
```go
// Add explicit waits between steps
{
    Name: "Wait for Processing",
    Type: gowright.StepTypeCustom,
    Action: gowright.CustomStepAction{
        Function: func(ctx context.Context) error {
            time.Sleep(2 * time.Second)
            return nil
        },
    },
}
```

**Resource cleanup failures:**
```go
// Use defer-like cleanup that always runs
integrationTest := &gowright.IntegrationTest{
    // ... steps
    Cleanup: []gowright.CleanupStep{
        // These run even if test fails
    },
}
```

## Next Steps

- [Advanced Features](../advanced/test-suites.md) - Complex test orchestration
- [Examples](../examples/integration-testing.md) - More integration examples
- [Best Practices](../reference/best-practices.md) - Integration testing best practices
- [Performance Testing](../advanced/parallel-execution.md) - Scale integration tests