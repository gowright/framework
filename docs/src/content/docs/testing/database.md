---
title: Database Testing
description: Learn how to test database operations with Gowright
---

Gowright provides comprehensive database testing capabilities with support for multiple database types, transactions, and connection management.

## Getting Started

### Basic Setup

```go
func TestDatabaseBasics(t *testing.T) {
    dbTester := gowright.NewDatabaseTester()
    config := &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "test": {
                Driver: "sqlite3",
                DSN:    ":memory:",
            },
        },
    }
    
    err := dbTester.Initialize(config)
    require.NoError(t, err)
    defer dbTester.Cleanup()
    
    // Your database tests here
}
```

### Configuration

The `DatabaseConfig` supports multiple database connections:

```go
type DatabaseConfig struct {
    Connections map[string]*DBConnection `json:"connections"`
}

type DBConnection struct {
    Driver          string        `json:"driver"`
    DSN             string        `json:"dsn"`
    MaxOpenConns    int           `json:"max_open_conns,omitempty"`
    MaxIdleConns    int           `json:"max_idle_conns,omitempty"`
    ConnMaxLifetime time.Duration `json:"conn_max_lifetime,omitempty"`
}
```

## Supported Databases

### PostgreSQL

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "postgres": {
            Driver: "postgres",
            DSN:    "postgres://user:password@localhost:5432/testdb?sslmode=disable",
        },
    },
}
```

### MySQL

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "mysql": {
            Driver: "mysql",
            DSN:    "user:password@tcp(localhost:3306)/testdb",
        },
    },
}
```

### SQLite

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "sqlite": {
            Driver: "sqlite3",
            DSN:    "./test.db", // or ":memory:" for in-memory
        },
    },
}
```

## Basic Operations

### Executing Queries

```go
// Create table
_, err := dbTester.Execute("test", `
    CREATE TABLE users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL
    )
`)
require.NoError(t, err)

// Insert data
result, err := dbTester.Execute("test", 
    "INSERT INTO users (name, email) VALUES (?, ?)", 
    "John Doe", "john@example.com")
require.NoError(t, err)
assert.Equal(t, int64(1), result.RowsAffected)

// Query data
queryResult, err := dbTester.Execute("test", 
    "SELECT id, name, email FROM users WHERE email = ?", 
    "john@example.com")
require.NoError(t, err)
assert.Equal(t, 1, len(queryResult.Rows))
```

## Transaction Management

### Basic Transactions

```go
// Begin transaction
tx, err := dbTester.BeginTransaction("main")
require.NoError(t, err)

// Execute operations within transaction
_, err = tx.Execute("INSERT INTO users (name, email) VALUES (?, ?)", 
    "Jane Doe", "jane@example.com")
require.NoError(t, err)

// Commit or rollback
err = tx.Commit()
require.NoError(t, err)
```

### Transaction Rollback

```go
tx, err := dbTester.BeginTransaction("main")
require.NoError(t, err)

// Insert test data
_, err = tx.Execute("INSERT INTO users (name, email) VALUES (?, ?)", 
    "Test User", "test@example.com")
require.NoError(t, err)

// Rollback transaction
err = tx.Rollback()
require.NoError(t, err)

// Verify data was rolled back
result, err := dbTester.Execute("main", 
    "SELECT COUNT(*) as count FROM users WHERE email = ?", 
    "test@example.com")
require.NoError(t, err)
assert.Equal(t, 0, result.Rows[0]["count"])
```

## Test Builder Pattern

### Database Test Structure

```go
dbTest := &gowright.DatabaseTest{
    Name:       "User Creation Test",
    Connection: "main",
    Setup: []string{
        "DELETE FROM users WHERE email = 'test@example.com'",
    },
    Query: "INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')",
    Expected: &gowright.DatabaseExpectation{
        RowsAffected: 1,
    },
    Teardown: []string{
        "DELETE FROM users WHERE email = 'test@example.com'",
    },
}

result := dbTester.ExecuteTest(dbTest)
assert.Equal(t, gowright.TestStatusPassed, result.Status)
```

## Data Validation

### Row Count Validation

```go
result, err := dbTester.Execute("test", "SELECT COUNT(*) as count FROM users")
require.NoError(t, err)
assert.Equal(t, 1, result.Rows[0]["count"])
```

### Data Content Validation

```go
result, err := dbTester.Execute("test", 
    "SELECT name, email FROM users WHERE id = ?", 1)
require.NoError(t, err)

user := result.Rows[0]
assert.Equal(t, "John Doe", user["name"])
assert.Equal(t, "john@example.com", user["email"])
```

## Best Practices

### Test Data Management

```go
func setupTestData(dbTester *gowright.DatabaseTester) error {
    queries := []string{
        "DELETE FROM orders",
        "DELETE FROM users",
        "INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')",
    }
    
    for _, query := range queries {
        if _, err := dbTester.Execute("test", query); err != nil {
            return err
        }
    }
    return nil
}
```

### Connection Pooling

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "main": {
            Driver:          "postgres",
            DSN:             "postgres://user:pass@localhost/testdb",
            MaxOpenConns:    25,
            MaxIdleConns:    5,
            ConnMaxLifetime: 5 * time.Minute,
        },
    },
}
```

For more examples, see the [Database Examples](/examples/database/) section.