package gowright

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestNewDatabaseTester(t *testing.T) {
	tester := NewDatabaseTester()
	
	assert.NotNil(t, tester)
	assert.NotNil(t, tester.connections)
	assert.False(t, tester.initialized)
	assert.Equal(t, "DatabaseTester", tester.GetName())
}

func TestDatabaseTester_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		expectError bool
		errorType   ErrorType
	}{
		{
			name: "valid configuration",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{
					"test": {
						Driver:       "sqlite3",
						DSN:          ":memory:",
						MaxOpenConns: 10,
						MaxIdleConns: 5,
					},
				},
			},
			expectError: false,
		},
		{
			name:        "invalid configuration type",
			config:      "invalid",
			expectError: true,
			errorType:   ConfigurationError,
		},
		{
			name:        "nil configuration",
			config:      (*DatabaseConfig)(nil),
			expectError: true,
			errorType:   ConfigurationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewDatabaseTester()
			err := tester.Initialize(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if gowrightErr, ok := err.(*GowrightError); ok {
					assert.Equal(t, tt.errorType, gowrightErr.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, tester.initialized)
				assert.NotNil(t, tester.config)
			}
		})
	}
}

func TestDatabaseTester_Connect(t *testing.T) {
	tests := []struct {
		name           string
		connectionName string
		config         *DatabaseConfig
		expectError    bool
		errorType      ErrorType
	}{
		{
			name:           "valid sqlite connection",
			connectionName: "test",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{
					"test": {
						Driver:       "sqlite3",
						DSN:          ":memory:",
						MaxOpenConns: 10,
						MaxIdleConns: 5,
					},
				},
			},
			expectError: false,
		},
		{
			name:           "connection not found",
			connectionName: "nonexistent",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{},
			},
			expectError: true,
			errorType:   ConfigurationError,
		},
		{
			name:           "invalid driver",
			connectionName: "test",
			config: &DatabaseConfig{
				Connections: map[string]*DBConnection{
					"test": {
						Driver:       "invalid_driver",
						DSN:          "invalid_dsn",
						MaxOpenConns: 10,
						MaxIdleConns: 5,
					},
				},
			},
			expectError: true,
			errorType:   DatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewDatabaseTester()
			err := tester.Initialize(tt.config)
			require.NoError(t, err)

			err = tester.Connect(tt.connectionName)

			if tt.expectError {
				assert.Error(t, err)
				if gowrightErr, ok := err.(*GowrightError); ok {
					assert.Equal(t, tt.errorType, gowrightErr.Type)
				}
			} else {
				assert.NoError(t, err)
				// Verify connection exists
				tester.mutex.RLock()
				_, exists := tester.connections[tt.connectionName]
				tester.mutex.RUnlock()
				assert.True(t, exists)
			}
		})
	}
}

func TestDatabaseTester_Connect_NotInitialized(t *testing.T) {
	tester := NewDatabaseTester()
	err := tester.Connect("test")
	
	assert.Error(t, err)
	gowrightErr, ok := err.(*GowrightError)
	assert.True(t, ok)
	assert.Equal(t, ConfigurationError, gowrightErr.Type)
	assert.Contains(t, gowrightErr.Message, "not initialized")
}

func TestDatabaseTester_Execute_SQLite(t *testing.T) {
	// Use real SQLite for integration testing
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(t, err)

	// Test table creation (exec operation)
	result, err := tester.Execute("test", "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.RowsAffected) // CREATE TABLE doesn't affect rows

	// Test insert (exec operation)
	result, err = tester.Execute("test", "INSERT INTO users (name, email) VALUES (?, ?)", "John Doe", "john@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Test select (query operation)
	result, err = tester.Execute("test", "SELECT id, name, email FROM users WHERE name = ?", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, "John Doe", result.Rows[0]["name"])
	assert.Equal(t, "john@example.com", result.Rows[0]["email"])

	// Test update (exec operation)
	result, err = tester.Execute("test", "UPDATE users SET email = ? WHERE name = ?", "john.doe@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Test delete (exec operation)
	result, err = tester.Execute("test", "DELETE FROM users WHERE name = ?", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseTester_BeginTransaction(t *testing.T) {
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(t, err)

	// Create test table
	_, err = tester.Execute("test", "CREATE TABLE test_tx (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	// Begin transaction
	tx, err := tester.BeginTransaction("test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Execute within transaction
	result, err := tx.Execute("INSERT INTO test_tx (value) VALUES (?)", "test_value")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Commit transaction
	err = tx.Commit()
	assert.NoError(t, err)

	// Verify data was committed
	result, err = tester.Execute("test", "SELECT COUNT(*) as count FROM test_tx")
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseTester_TransactionRollback(t *testing.T) {
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(t, err)

	// Create test table
	_, err = tester.Execute("test", "CREATE TABLE test_rollback (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	// Begin transaction
	tx, err := tester.BeginTransaction("test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Execute within transaction
	result, err := tx.Execute("INSERT INTO test_rollback (value) VALUES (?)", "test_value")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Rollback transaction
	err = tx.Rollback()
	assert.NoError(t, err)

	// Verify data was rolled back
	result, err = tester.Execute("test", "SELECT COUNT(*) as count FROM test_rollback")
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 1)
	// The count should be 0 since we rolled back
	count := result.Rows[0]["count"]
	assert.Equal(t, int64(0), count)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseTester_ValidateData(t *testing.T) {
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(t, err)

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE validation_test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO validation_test (name, age) VALUES (?, ?)", "Alice", 30)
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO validation_test (name, age) VALUES (?, ?)", "Bob", 25)
	require.NoError(t, err)

	tests := []struct {
		name        string
		query       string
		expected    *DatabaseExpectation
		expectError bool
	}{
		{
			name:  "validate row count",
			query: "SELECT * FROM validation_test",
			expected: &DatabaseExpectation{
				RowCount: 2,
			},
			expectError: false,
		},
		{
			name:  "validate specific rows",
			query: "SELECT name, age FROM validation_test WHERE name = 'Alice'",
			expected: &DatabaseExpectation{
				Rows: []map[string]interface{}{
					{"name": "Alice", "age": int64(30)},
				},
			},
			expectError: false,
		},
		{
			name:  "validate row count mismatch",
			query: "SELECT * FROM validation_test",
			expected: &DatabaseExpectation{
				RowCount: 3, // Wrong count
			},
			expectError: true,
		},
		{
			name:  "validate wrong row data",
			query: "SELECT name, age FROM validation_test WHERE name = 'Alice'",
			expected: &DatabaseExpectation{
				Rows: []map[string]interface{}{
					{"name": "Alice", "age": int64(25)}, // Wrong age
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tester.ValidateData("test", tt.query, tt.expected)
			
			if tt.expectError {
				assert.Error(t, err)
				if gowrightErr, ok := err.(*GowrightError); ok {
					assert.Equal(t, AssertionError, gowrightErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseTester_Cleanup(t *testing.T) {
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test1": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			"test2": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(t, err)

	// Connect to both databases
	err = tester.Connect("test1")
	require.NoError(t, err)
	
	err = tester.Connect("test2")
	require.NoError(t, err)

	// Verify connections exist
	tester.mutex.RLock()
	assert.Len(t, tester.connections, 2)
	tester.mutex.RUnlock()

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)

	// Verify connections are cleaned up
	tester.mutex.RLock()
	assert.Len(t, tester.connections, 0)
	tester.mutex.RUnlock()
	assert.False(t, tester.initialized)
}

func TestIsSelectQuery(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"SELECT * FROM users", true},
		{"select id from users", true},
		{"Select name from users", true},
		{"  SELECT * FROM users", true},
		{"\n\tSELECT * FROM users", true},
		{"INSERT INTO users VALUES (1, 'test')", false},
		{"UPDATE users SET name = 'test'", false},
		{"DELETE FROM users", false},
		{"CREATE TABLE users (id INT)", false},
		{"DROP TABLE users", false},
		{"", false},
		{"SEL", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := isSelectQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		result   bool
	}{
		{"exact match strings", "test", "test", true},
		{"different strings", "test", "other", false},
		{"exact match int", 42, 42, true},
		{"int to int64", 42, int64(42), true},
		{"int64 to int", int64(42), 42, true},
		{"int to float64", 42, 42.0, true},
		{"float64 to int", 42.0, 42, true},
		{"different numbers", 42, 43, false},
		{"exact match bool", true, true, true},
		{"different bool", true, false, false},
		{"nil values", nil, nil, true},
		{"different types", "42", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.expected, tt.actual)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestTrimLeadingWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SELECT", "SELECT"},
		{"  SELECT", "SELECT"},
		{"\t\nSELECT", "SELECT"},
		{"\r\n  \tSELECT * FROM users", "SELECT * FROM users"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimLeadingWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test error scenarios
func TestDatabaseTester_ErrorScenarios(t *testing.T) {
	t.Run("execute on non-existent connection", func(t *testing.T) {
		tester := NewDatabaseTester()
		config := &DatabaseConfig{
			Connections: map[string]*DBConnection{},
		}
		
		err := tester.Initialize(config)
		require.NoError(t, err)

		_, err = tester.Execute("nonexistent", "SELECT 1")
		assert.Error(t, err)
		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, ConfigurationError, gowrightErr.Type)
	})

	t.Run("begin transaction on non-existent connection", func(t *testing.T) {
		tester := NewDatabaseTester()
		config := &DatabaseConfig{
			Connections: map[string]*DBConnection{},
		}
		
		err := tester.Initialize(config)
		require.NoError(t, err)

		_, err = tester.BeginTransaction("nonexistent")
		assert.Error(t, err)
		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, ConfigurationError, gowrightErr.Type)
	})

	t.Run("validate data with wrong expected type", func(t *testing.T) {
		tester := NewDatabaseTester()
		config := &DatabaseConfig{
			Connections: map[string]*DBConnection{
				"test": {
					Driver: "sqlite3",
					DSN:    ":memory:",
				},
			},
		}
		
		err := tester.Initialize(config)
		require.NoError(t, err)

		err = tester.ValidateData("test", "SELECT 1", "wrong_type")
		assert.Error(t, err)
		gowrightErr, ok := err.(*GowrightError)
		assert.True(t, ok)
		assert.Equal(t, AssertionError, gowrightErr.Type)
	})
}

// Benchmark tests
func BenchmarkDatabaseTester_Execute(b *testing.B) {
	tester := NewDatabaseTester()
	config := &DatabaseConfig{
		Connections: map[string]*DBConnection{
			"test": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}
	
	err := tester.Initialize(config)
	require.NoError(b, err)

	// Setup test table
	_, err = tester.Execute("test", "CREATE TABLE bench_test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(b, err)

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tester.Execute("test", "INSERT INTO bench_test (value) VALUES (?)", fmt.Sprintf("value_%d", i))
		if err != nil {
			b.Fatal(err)
		}
	}

	// Cleanup
	tester.Cleanup()
}