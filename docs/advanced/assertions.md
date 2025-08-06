# Assertions

Gowright provides a comprehensive assertion system that goes beyond basic equality checks, offering rich validation capabilities for API responses, database results, UI elements, and custom business logic.

## Overview

The assertion system in Gowright provides:

- **Rich Assertion Methods**: Beyond simple equality checks
- **Custom Assertions**: Define your own validation logic
- **Detailed Error Messages**: Clear feedback when assertions fail
- **Chained Assertions**: Combine multiple validations
- **Type-Safe Assertions**: Compile-time type checking
- **Performance Assertions**: Validate timing and resource usage

## Basic Assertions

### Simple Value Assertions

```go
package main

import (
    "testing"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestBasicAssertions(t *testing.T) {
    // Create assertion instance
    assertion := gowright.NewTestAssertion("Basic Value Tests")
    
    // Equality assertions
    assertion.Equal(42, 42, "Numbers should be equal")
    assertion.Equal("hello", "hello", "Strings should be equal")
    assertion.NotEqual(42, 43, "Numbers should not be equal")
    
    // Nil assertions
    var nilValue *string
    assertion.Nil(nilValue, "Value should be nil")
    assertion.NotNil("not nil", "Value should not be nil")
    
    // Boolean assertions
    assertion.True(true, "Value should be true")
    assertion.False(false, "Value should be false")
    
    // Numeric assertions
    assertion.Greater(10, 5, "10 should be greater than 5")
    assertion.GreaterOrEqual(10, 10, "10 should be greater or equal to 10")
    assertion.Less(5, 10, "5 should be less than 10")
    assertion.LessOrEqual(5, 5, "5 should be less or equal to 5")
    
    // String assertions
    assertion.Contains("hello world", "world", "String should contain substring")
    assertion.NotContains("hello", "xyz", "String should not contain substring")
    assertion.StartsWith("hello world", "hello", "String should start with prefix")
    assertion.EndsWith("hello world", "world", "String should end with suffix")
    assertion.Matches("hello123", `^hello\d+$`, "String should match regex")
    
    // Collection assertions
    slice := []int{1, 2, 3, 4, 5}
    assertion.Len(slice, 5, "Slice should have 5 elements")
    assertion.Empty([]int{}, "Empty slice should be empty")
    assertion.NotEmpty(slice, "Slice should not be empty")
    assertion.Contains(slice, 3, "Slice should contain element")
    
    // Get assertion result
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    assert.Empty(t, result.Failures)
}
```

### Advanced Matchers

```go
func TestAdvancedMatchers(t *testing.T) {
    assertion := gowright.NewTestAssertion("Advanced Matcher Tests")
    
    // Range assertions
    assertion.InRange(5, 1, 10, "Value should be in range")
    assertion.NotInRange(15, 1, 10, "Value should not be in range")
    
    // Type assertions
    assertion.IsType(42, int(0), "Value should be int type")
    assertion.IsType("hello", "", "Value should be string type")
    
    // Collection matchers
    numbers := []int{1, 2, 3, 4, 5}
    assertion.All(numbers, func(item interface{}) bool {
        return item.(int) > 0
    }, "All numbers should be positive")
    
    assertion.Any(numbers, func(item interface{}) bool {
        return item.(int) > 4
    }, "At least one number should be greater than 4")
    
    assertion.None(numbers, func(item interface{}) bool {
        return item.(int) < 0
    }, "No numbers should be negative")
    
    // Custom matchers
    assertion.That(42, gowright.IsEven(), "Number should be even")
    assertion.That("hello@example.com", gowright.IsValidEmail(), "Should be valid email")
    assertion.That(time.Now(), gowright.IsRecent(5*time.Minute), "Time should be recent")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}

// Custom matcher examples
func IsEven() gowright.Matcher {
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        if num, ok := actual.(int); ok {
            if num%2 == 0 {
                return true, ""
            }
            return false, fmt.Sprintf("expected %d to be even", num)
        }
        return false, "expected integer value"
    })
}

func IsValidEmail() gowright.Matcher {
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        if email, ok := actual.(string); ok {
            if matched, _ := regexp.MatchString(`^[^@]+@[^@]+\.[^@]+$`, email); matched {
                return true, ""
            }
            return false, fmt.Sprintf("'%s' is not a valid email address", email)
        }
        return false, "expected string value"
    })
}

func IsRecent(duration time.Duration) gowright.Matcher {
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        if t, ok := actual.(time.Time); ok {
            if time.Since(t) <= duration {
                return true, ""
            }
            return false, fmt.Sprintf("time %v is not within %v of now", t, duration)
        }
        return false, "expected time.Time value"
    })
}
```

## API Response Assertions

### JSON Response Validation

```go
func TestJSONResponseAssertions(t *testing.T) {
    // Mock API response
    responseBody := `{
        "id": 123,
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30,
        "active": true,
        "tags": ["user", "premium"],
        "profile": {
            "bio": "Software developer",
            "location": "New York"
        },
        "created_at": "2024-01-15T10:30:00Z"
    }`
    
    assertion := gowright.NewAPIResponseAssertion("JSON Response Test", []byte(responseBody))
    
    // JSONPath assertions
    assertion.JSONPath("$.id", 123, "User ID should match")
    assertion.JSONPath("$.name", "John Doe", "User name should match")
    assertion.JSONPath("$.email", gowright.IsValidEmail(), "Email should be valid")
    assertion.JSONPath("$.age", gowright.GreaterThan(18), "User should be adult")
    assertion.JSONPath("$.active", true, "User should be active")
    
    // Nested object assertions
    assertion.JSONPath("$.profile.bio", gowright.NotEmpty(), "Bio should not be empty")
    assertion.JSONPath("$.profile.location", gowright.Contains("New"), "Location should contain 'New'")
    
    // Array assertions
    assertion.JSONPath("$.tags", gowright.HasLength(2), "Should have 2 tags")
    assertion.JSONPath("$.tags[0]", "user", "First tag should be 'user'")
    assertion.JSONPath("$.tags", gowright.Contains("premium"), "Should contain 'premium' tag")
    
    // Date/time assertions
    assertion.JSONPath("$.created_at", gowright.IsISO8601(), "Should be valid ISO8601 date")
    assertion.JSONPath("$.created_at", gowright.IsRecent(24*time.Hour), "Should be recent")
    
    // Schema validation
    schema := `{
        "type": "object",
        "required": ["id", "name", "email"],
        "properties": {
            "id": {"type": "number"},
            "name": {"type": "string"},
            "email": {"type": "string", "format": "email"},
            "age": {"type": "number", "minimum": 0},
            "active": {"type": "boolean"}
        }
    }`
    assertion.JSONSchema(schema, "Response should match schema")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### HTTP Response Assertions

```go
func TestHTTPResponseAssertions(t *testing.T) {
    // Mock HTTP response
    response := &gowright.APIResponse{
        StatusCode: 200,
        Headers: map[string][]string{
            "Content-Type":   {"application/json; charset=utf-8"},
            "Cache-Control":  {"no-cache"},
            "X-Rate-Limit":   {"1000"},
            "X-Rate-Remaining": {"999"},
        },
        Body:         []byte(`{"message": "success"}`),
        ResponseTime: 150 * time.Millisecond,
        Size:         25,
    }
    
    assertion := gowright.NewHTTPResponseAssertion("HTTP Response Test", response)
    
    // Status code assertions
    assertion.StatusCode(200, "Should return OK status")
    assertion.StatusCodeIn([]int{200, 201, 202}, "Should return success status")
    assertion.IsSuccess("Should be successful response")
    assertion.IsNotError("Should not be error response")
    
    // Header assertions
    assertion.Header("Content-Type", "application/json; charset=utf-8", "Content type should match")
    assertion.HeaderContains("Content-Type", "application/json", "Should be JSON response")
    assertion.HeaderExists("X-Rate-Limit", "Rate limit header should exist")
    assertion.HeaderMatches("X-Rate-Limit", `^\d+$`, "Rate limit should be numeric")
    
    // Performance assertions
    assertion.ResponseTime(gowright.LessThan(200*time.Millisecond), "Response should be fast")
    assertion.ResponseSize(gowright.LessThan(1024), "Response should be small")
    
    // Body assertions
    assertion.BodyContains("success", "Body should contain success message")
    assertion.BodyNotContains("error", "Body should not contain error")
    assertion.BodyMatches(`"message":\s*"success"`, "Body should match pattern")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Database Assertions

### Query Result Validation

```go
func TestDatabaseAssertions(t *testing.T) {
    // Mock database results
    queryResults := []map[string]interface{}{
        {
            "id":         1,
            "username":   "john_doe",
            "email":      "john@example.com",
            "age":        30,
            "created_at": time.Now().Add(-24 * time.Hour),
            "balance":    1250.50,
            "active":     true,
        },
        {
            "id":         2,
            "username":   "jane_smith",
            "email":      "jane@example.com",
            "age":        25,
            "created_at": time.Now().Add(-48 * time.Hour),
            "balance":    750.25,
            "active":     true,
        },
    }
    
    assertion := gowright.NewDatabaseAssertion("Database Query Test", queryResults)
    
    // Row count assertions
    assertion.RowCount(2, "Should return 2 rows")
    assertion.RowCountGreaterThan(1, "Should return more than 1 row")
    assertion.RowCountLessThan(5, "Should return less than 5 rows")
    assertion.NotEmpty("Results should not be empty")
    
    // Column existence assertions
    assertion.HasColumn("id", "Should have id column")
    assertion.HasColumn("username", "Should have username column")
    assertion.HasColumns([]string{"id", "username", "email"}, "Should have required columns")
    
    // Value assertions for specific rows
    assertion.RowValue(0, "username", "john_doe", "First user should be john_doe")
    assertion.RowValue(1, "email", "jane@example.com", "Second user email should match")
    
    // Column value assertions
    assertion.ColumnValues("active", gowright.All(true), "All users should be active")
    assertion.ColumnValues("age", gowright.AllGreaterThan(18), "All users should be adults")
    assertion.ColumnValues("balance", gowright.AllGreaterThan(0), "All balances should be positive")
    
    // Aggregate assertions
    assertion.Sum("balance", 2000.75, "Total balance should be correct")
    assertion.Average("age", 27.5, "Average age should be correct")
    assertion.Min("age", 25, "Minimum age should be 25")
    assertion.Max("age", 30, "Maximum age should be 30")
    
    // Custom validations
    assertion.CustomValidation(func(rows []map[string]interface{}) (bool, string) {
        for _, row := range rows {
            if email, ok := row["email"].(string); ok {
                if !strings.Contains(email, "@") {
                    return false, fmt.Sprintf("Invalid email format: %s", email)
                }
            }
        }
        return true, ""
    }, "All emails should be valid")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Schema Validation

```go
func TestDatabaseSchemaAssertions(t *testing.T) {
    // Mock table schema
    tableSchema := &gowright.TableSchema{
        Name: "users",
        Columns: []gowright.ColumnInfo{
            {Name: "id", Type: "INTEGER", PrimaryKey: true, NotNull: true},
            {Name: "username", Type: "VARCHAR(50)", Unique: true, NotNull: true},
            {Name: "email", Type: "VARCHAR(255)", Unique: true, NotNull: true},
            {Name: "age", Type: "INTEGER", NotNull: false},
            {Name: "created_at", Type: "TIMESTAMP", Default: "CURRENT_TIMESTAMP"},
        },
        Indexes: []gowright.IndexInfo{
            {Name: "idx_username", Columns: []string{"username"}, Unique: true},
            {Name: "idx_email", Columns: []string{"email"}, Unique: true},
        },
    }
    
    assertion := gowright.NewSchemaAssertion("Table Schema Test", tableSchema)
    
    // Table existence
    assertion.TableExists("users", "Users table should exist")
    
    // Column assertions
    assertion.ColumnExists("id", "ID column should exist")
    assertion.ColumnType("id", "INTEGER", "ID should be integer type")
    assertion.ColumnIsPrimaryKey("id", "ID should be primary key")
    assertion.ColumnIsNotNull("username", "Username should be not null")
    assertion.ColumnIsUnique("email", "Email should be unique")
    assertion.ColumnHasDefault("created_at", "Created at should have default")
    
    // Index assertions
    assertion.IndexExists("idx_username", "Username index should exist")
    assertion.IndexIsUnique("idx_email", "Email index should be unique")
    assertion.IndexCovers("idx_username", []string{"username"}, "Index should cover username")
    
    // Constraint assertions
    assertion.HasPrimaryKey([]string{"id"}, "Should have primary key on id")
    assertion.HasUniqueConstraint("username", "Should have unique constraint on username")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## UI Element Assertions

### Element State Validation

```go
func TestUIElementAssertions(t *testing.T) {
    // Mock UI element
    element := &gowright.UIElement{
        Selector:    "button#submit",
        Text:        "Submit Form",
        Value:       "",
        Visible:     true,
        Enabled:     true,
        Selected:    false,
        TagName:     "button",
        Attributes: map[string]string{
            "id":    "submit",
            "type":  "submit",
            "class": "btn btn-primary",
        },
        CSSProperties: map[string]string{
            "color":            "#ffffff",
            "background-color": "#007bff",
            "display":          "block",
        },
        Position: gowright.ElementPosition{X: 100, Y: 200},
        Size:     gowright.ElementSize{Width: 120, Height: 40},
    }
    
    assertion := gowright.NewUIElementAssertion("UI Element Test", element)
    
    // Visibility assertions
    assertion.IsVisible("Element should be visible")
    assertion.IsNotHidden("Element should not be hidden")
    assertion.IsDisplayed("Element should be displayed")
    
    // State assertions
    assertion.IsEnabled("Element should be enabled")
    assertion.IsNotDisabled("Element should not be disabled")
    assertion.IsClickable("Element should be clickable")
    
    // Text assertions
    assertion.HasText("Submit Form", "Element should have correct text")
    assertion.TextContains("Submit", "Text should contain 'Submit'")
    assertion.TextMatches(`^Submit\s+Form$`, "Text should match pattern")
    
    // Attribute assertions
    assertion.HasAttribute("id", "Element should have id attribute")
    assertion.AttributeEquals("type", "submit", "Type should be submit")
    assertion.AttributeContains("class", "btn-primary", "Should have primary button class")
    
    // CSS assertions
    assertion.CSSProperty("color", "#ffffff", "Text color should be white")
    assertion.CSSProperty("display", "block", "Should be block display")
    assertion.CSSPropertyContains("background-color", "#007bff", "Should have blue background")
    
    // Position and size assertions
    assertion.Position(100, 200, "Element should be at correct position")
    assertion.Size(120, 40, "Element should have correct size")
    assertion.IsInViewport("Element should be in viewport")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Page State Validation

```go
func TestPageStateAssertions(t *testing.T) {
    // Mock page state
    pageState := &gowright.PageState{
        URL:           "https://example.com/dashboard",
        Title:         "Dashboard - Example App",
        ReadyState:    "complete",
        LoadTime:      1200 * time.Millisecond,
        ElementCount:  45,
        ErrorCount:    0,
        ConsoleErrors: []string{},
        NetworkRequests: []gowright.NetworkRequest{
            {URL: "/api/user", Status: 200, Duration: 150 * time.Millisecond},
            {URL: "/api/dashboard", Status: 200, Duration: 200 * time.Millisecond},
        },
    }
    
    assertion := gowright.NewPageStateAssertion("Page State Test", pageState)
    
    // URL assertions
    assertion.URL("https://example.com/dashboard", "Should be on dashboard page")
    assertion.URLContains("/dashboard", "URL should contain dashboard")
    assertion.URLMatches(`^https://example\.com/`, "URL should match domain pattern")
    
    // Title assertions
    assertion.Title("Dashboard - Example App", "Page title should match")
    assertion.TitleContains("Dashboard", "Title should contain Dashboard")
    
    // Page state assertions
    assertion.IsLoaded("Page should be fully loaded")
    assertion.LoadTime(gowright.LessThan(2*time.Second), "Page should load quickly")
    assertion.ElementCount(gowright.GreaterThan(40), "Should have many elements")
    
    // Error assertions
    assertion.NoConsoleErrors("Should have no console errors")
    assertion.NoNetworkErrors("Should have no network errors")
    assertion.ErrorCount(0, "Should have zero errors")
    
    // Network assertions
    assertion.NetworkRequestCount(2, "Should have made 2 network requests")
    assertion.AllNetworkRequestsSuccessful("All requests should be successful")
    assertion.NetworkRequestExists("/api/user", "Should have made user API request")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Custom Assertions

### Business Logic Assertions

```go
func TestBusinessLogicAssertions(t *testing.T) {
    // Mock business objects
    order := &Order{
        ID:          "ORD-123",
        CustomerID:  "CUST-456",
        Items:       []OrderItem{{ProductID: "PROD-789", Quantity: 2, Price: 29.99}},
        Total:       59.98,
        Status:      "pending",
        CreatedAt:   time.Now().Add(-1 * time.Hour),
        ShippingAddress: Address{
            Street: "123 Main St",
            City:   "Anytown",
            State:  "CA",
            ZIP:    "12345",
        },
    }
    
    assertion := gowright.NewCustomAssertion("Business Logic Test")
    
    // Order validation assertions
    assertion.That(order, IsValidOrder(), "Order should be valid")
    assertion.That(order.Total, EqualsItemsTotal(order.Items), "Total should match items")
    assertion.That(order.Status, IsValidOrderStatus(), "Status should be valid")
    assertion.That(order.CreatedAt, IsRecent(24*time.Hour), "Order should be recent")
    
    // Address validation
    assertion.That(order.ShippingAddress, IsValidAddress(), "Shipping address should be valid")
    assertion.That(order.ShippingAddress.ZIP, IsValidZIPCode(), "ZIP code should be valid")
    
    // Business rule assertions
    assertion.That(order, HasValidCustomer(), "Order should have valid customer")
    assertion.That(order, HasInStockItems(), "All items should be in stock")
    assertion.That(order, MeetsMinimumOrderValue(25.00), "Order should meet minimum value")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}

// Custom business logic matchers
func IsValidOrder() gowright.Matcher {
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        order, ok := actual.(*Order)
        if !ok {
            return false, "expected Order object"
        }
        
        if order.ID == "" {
            return false, "order ID is required"
        }
        if order.CustomerID == "" {
            return false, "customer ID is required"
        }
        if len(order.Items) == 0 {
            return false, "order must have at least one item"
        }
        if order.Total <= 0 {
            return false, "order total must be positive"
        }
        
        return true, ""
    })
}

func EqualsItemsTotal(items []OrderItem) gowright.Matcher {
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        total, ok := actual.(float64)
        if !ok {
            return false, "expected float64 total"
        }
        
        expectedTotal := 0.0
        for _, item := range items {
            expectedTotal += float64(item.Quantity) * item.Price
        }
        
        if math.Abs(total-expectedTotal) > 0.01 {
            return false, fmt.Sprintf("expected total %.2f, got %.2f", expectedTotal, total)
        }
        
        return true, ""
    })
}

func IsValidOrderStatus() gowright.Matcher {
    validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
    return gowright.MatcherFunc(func(actual interface{}) (bool, string) {
        status, ok := actual.(string)
        if !ok {
            return false, "expected string status"
        }
        
        for _, validStatus := range validStatuses {
            if status == validStatus {
                return true, ""
            }
        }
        
        return false, fmt.Sprintf("invalid status '%s', must be one of: %v", status, validStatuses)
    })
}
```

### Performance Assertions

```go
func TestPerformanceAssertions(t *testing.T) {
    // Mock performance metrics
    metrics := &gowright.PerformanceMetrics{
        ResponseTime:     150 * time.Millisecond,
        DNSLookupTime:    10 * time.Millisecond,
        TCPConnectTime:   20 * time.Millisecond,
        TLSHandshakeTime: 30 * time.Millisecond,
        FirstByteTime:    80 * time.Millisecond,
        ContentLoadTime:  40 * time.Millisecond,
        TotalSize:        1024 * 50, // 50KB
        CompressedSize:   1024 * 15, // 15KB
        RequestCount:     5,
        ErrorCount:       0,
        MemoryUsage:      1024 * 1024 * 25, // 25MB
        CPUUsage:         15.5, // 15.5%
    }
    
    assertion := gowright.NewPerformanceAssertion("Performance Test", metrics)
    
    // Response time assertions
    assertion.ResponseTime(gowright.LessThan(200*time.Millisecond), "Response should be fast")
    assertion.DNSLookupTime(gowright.LessThan(50*time.Millisecond), "DNS lookup should be fast")
    assertion.FirstByteTime(gowright.LessThan(100*time.Millisecond), "TTFB should be fast")
    
    // Size assertions
    assertion.TotalSize(gowright.LessThan(100*1024), "Total size should be under 100KB")
    assertion.CompressionRatio(gowright.GreaterThan(0.6), "Should have good compression")
    
    // Resource usage assertions
    assertion.MemoryUsage(gowright.LessThan(50*1024*1024), "Memory usage should be reasonable")
    assertion.CPUUsage(gowright.LessThan(20.0), "CPU usage should be low")
    
    // Request assertions
    assertion.RequestCount(5, "Should have made 5 requests")
    assertion.ErrorCount(0, "Should have no errors")
    assertion.ErrorRate(0.0, "Error rate should be zero")
    
    // Throughput assertions (for load tests)
    assertion.Throughput(gowright.GreaterThan(100.0), "Should handle 100+ requests/sec")
    assertion.ConcurrentUsers(gowright.LessThan(1000), "Should support under 1000 concurrent users")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Assertion Chaining and Composition

### Fluent Assertion API

```go
func TestFluentAssertions(t *testing.T) {
    user := map[string]interface{}{
        "id":       123,
        "username": "john_doe",
        "email":    "john@example.com",
        "age":      30,
        "active":   true,
        "tags":     []string{"premium", "verified"},
    }
    
    // Fluent assertion chaining
    result := gowright.Assert(user).
        Field("id").IsNotNil().IsGreaterThan(0).
        Field("username").IsNotEmpty().Matches(`^[a-z_]+$`).
        Field("email").IsValidEmail().
        Field("age").IsInRange(18, 120).
        Field("active").IsTrue().
        Field("tags").IsNotEmpty().Contains("premium").
        Execute()
    
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Conditional Assertions

```go
func TestConditionalAssertions(t *testing.T) {
    user := map[string]interface{}{
        "id":       123,
        "username": "john_doe",
        "email":    "john@example.com",
        "age":      30,
        "role":     "admin",
        "active":   true,
    }
    
    assertion := gowright.NewTestAssertion("Conditional Assertions")
    
    // Basic assertions
    assertion.Equal(user["id"], 123, "ID should match")
    
    // Conditional assertions based on role
    if role, ok := user["role"].(string); ok && role == "admin" {
        assertion.True(user["active"].(bool), "Admin users must be active")
        assertion.NotEmpty(user["email"], "Admin users must have email")
    }
    
    // Conditional assertions with helper
    assertion.When(user["age"].(int) >= 18, func(a *gowright.TestAssertion) {
        a.True(user["active"].(bool), "Adult users should be active")
    })
    
    assertion.Unless(user["role"] == "guest", func(a *gowright.TestAssertion) {
        a.NotEmpty(user["email"], "Non-guest users must have email")
    })
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Error Handling and Reporting

### Detailed Error Messages

```go
func TestDetailedErrorMessages(t *testing.T) {
    assertion := gowright.NewTestAssertion("Detailed Error Test")
    
    // This will fail to demonstrate error reporting
    assertion.Equal(42, 43, "Numbers should be equal")
    assertion.Contains("hello", "xyz", "String should contain substring")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusFailed, result.Status)
    assert.Len(t, result.Failures, 2)
    
    // Check detailed error messages
    firstFailure := result.Failures[0]
    assert.Contains(t, firstFailure.Message, "Numbers should be equal")
    assert.Contains(t, firstFailure.Details, "Expected: 43")
    assert.Contains(t, firstFailure.Details, "Actual: 42")
    
    secondFailure := result.Failures[1]
    assert.Contains(t, secondFailure.Message, "String should contain substring")
    assert.Contains(t, secondFailure.Details, "Expected to contain: xyz")
    assert.Contains(t, secondFailure.Details, "Actual string: hello")
}
```

### Soft Assertions

```go
func TestSoftAssertions(t *testing.T) {
    // Soft assertions continue even after failures
    assertion := gowright.NewSoftAssertion("Soft Assertion Test")
    
    assertion.Equal(1, 2, "This will fail but continue")
    assertion.Equal("hello", "world", "This will also fail but continue")
    assertion.Equal(true, true, "This will pass")
    assertion.Equal(42, 42, "This will also pass")
    
    result := assertion.GetResult()
    
    // Should have 2 failures but still executed all assertions
    assert.Equal(t, gowright.TestStatusFailed, result.Status)
    assert.Len(t, result.Failures, 2)
    assert.Equal(t, 4, result.TotalAssertions)
    assert.Equal(t, 2, result.PassedAssertions)
    assert.Equal(t, 2, result.FailedAssertions)
}
```

## Best Practices

### 1. Use Descriptive Messages

```go
// Good - Clear, descriptive messages
assertion.Equal(user.Age, 25, "User age should be 25 for test scenario")
assertion.Contains(response.Body, "success", "API should return success message")

// Avoid - Generic or missing messages
assertion.Equal(user.Age, 25, "")
assertion.Contains(response.Body, "success", "check response")
```

### 2. Choose Appropriate Assertion Types

```go
// Good - Specific assertions
assertion.IsValidEmail(user.Email, "Email should be valid format")
assertion.IsInRange(user.Age, 18, 120, "Age should be reasonable")

// Avoid - Generic assertions
assertion.True(strings.Contains(user.Email, "@"), "Email check")
assertion.True(user.Age >= 18 && user.Age <= 120, "Age check")
```

### 3. Group Related Assertions

```go
// Good - Logical grouping
userAssertion := gowright.NewTestAssertion("User Validation")
userAssertion.NotNil(user, "User should exist")
userAssertion.Equal(user.ID, expectedID, "User ID should match")
userAssertion.Equal(user.Name, expectedName, "User name should match")

profileAssertion := gowright.NewTestAssertion("Profile Validation")
profileAssertion.NotNil(user.Profile, "Profile should exist")
profileAssertion.NotEmpty(user.Profile.Bio, "Bio should not be empty")
```

### 4. Use Custom Matchers for Complex Logic

```go
// Good - Custom matcher for business logic
assertion.That(order, IsValidOrder(), "Order should meet business rules")

// Avoid - Complex inline validation
assertion.True(
    order.ID != "" && order.Total > 0 && len(order.Items) > 0,
    "Order validation"
)
```

### 5. Handle Edge Cases

```go
// Good - Handle nil and edge cases
assertion.When(user != nil, func(a *gowright.TestAssertion) {
    a.NotEmpty(user.Name, "User name should not be empty")
    a.IsValidEmail(user.Email, "User email should be valid")
})

// Good - Validate collections safely
if len(users) > 0 {
    assertion.All(users, IsValidUser(), "All users should be valid")
}
```

## Next Steps

- [Reporting](reporting.md) - Comprehensive test reporting
- [Parallel Execution](parallel-execution.md) - Optimize assertion performance
- [Examples](../examples/basic-usage.md) - Assertion examples
- [Best Practices](../reference/best-practices.md) - Assertion best practices