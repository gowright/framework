package database

import (
	"testing"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseTester(t *testing.T) {
	tester := NewDatabaseTester()

	assert.NotNil(t, tester)
	assert.False(t, tester.initialized)
	assert.Equal(t, "DatabaseTester", tester.GetName())
}

func TestDatabaseTester_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		expectError bool
		errorType   core.ErrorType
	}{
		{
			name: "valid configuration",
			config: &config.DatabaseConfig{
				Connections: map[string]*config.DatabaseConnection{
					"test": {
						Driver:   "sqlite3",
						Host:     "localhost",
						Database: ":memory:",
					},
				},
			},
			expectError: false,
		},
		{
			name:        "invalid configuration type",
			config:      "invalid",
			expectError: true,
			errorType:   core.ConfigurationError,
		},
		{
			name:        "nil configuration",
			config:      (*config.DatabaseConfig)(nil),
			expectError: true,
			errorType:   core.ConfigurationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewDatabaseTester()
			err := tester.Initialize(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if gowrightErr, ok := err.(*core.GowrightError); ok {
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
		config         *config.DatabaseConfig
		expectError    bool
		errorType      core.ErrorType
	}{
		{
			name:           "connection not found",
			connectionName: "nonexistent",
			config: &config.DatabaseConfig{
				Connections: map[string]*config.DatabaseConnection{},
			},
			expectError: true,
			errorType:   core.ConfigurationError,
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
				if gowrightErr, ok := err.(*core.GowrightError); ok {
					assert.Equal(t, tt.errorType, gowrightErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseTester_Connect_NotInitialized(t *testing.T) {
	tester := NewDatabaseTester()
	err := tester.Connect("test")

	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.DatabaseError, gowrightErr.Type)
	assert.Contains(t, gowrightErr.Message, "database tester not initialized")
}

func TestDatabaseTester_Execute_NotInitialized(t *testing.T) {
	tester := NewDatabaseTester()

	result, err := tester.Execute("test", "SELECT 1")

	assert.Nil(t, result)
	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.DatabaseError, gowrightErr.Type)
}

func TestDatabaseTester_BeginTransaction_NotInitialized(t *testing.T) {
	tester := NewDatabaseTester()

	tx, err := tester.BeginTransaction("test")

	assert.Nil(t, tx)
	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.DatabaseError, gowrightErr.Type)
}

func TestDatabaseTester_ValidateData_NotInitialized(t *testing.T) {
	tester := NewDatabaseTester()

	err := tester.ValidateData("test", "SELECT 1", nil)

	assert.Error(t, err)
	gowrightErr, ok := err.(*core.GowrightError)
	assert.True(t, ok)
	assert.Equal(t, core.DatabaseError, gowrightErr.Type)
}

func TestDatabaseTester_ExecuteTest(t *testing.T) {
	tester := NewDatabaseTester()
	config := &config.DatabaseConfig{
		Connections: map[string]*config.DatabaseConnection{
			"test": {
				Driver:   "sqlite3",
				Database: ":memory:",
			},
		},
	}

	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("successful test", func(t *testing.T) {
		test := &core.DatabaseTest{
			Name:       "Test Query",
			Connection: "test",
			Query:      "SELECT 1 as result",
			Expected: &core.DatabaseExpectation{
				RowCount: 1,
			},
		}

		result := tester.ExecuteTest(test)

		assert.NotNil(t, result)
		assert.Equal(t, "Test Query", result.Name)
		assert.Equal(t, core.TestStatusPassed, result.Status)
		assert.NotZero(t, result.Duration)
		assert.Nil(t, result.Error)
	})

	t.Run("test with setup and teardown", func(t *testing.T) {
		test := &core.DatabaseTest{
			Name:       "Test with Setup",
			Connection: "test",
			Setup:      []string{"CREATE TABLE temp_test (id INTEGER)"},
			Query:      "SELECT COUNT(*) as count FROM temp_test",
			Teardown:   []string{"DROP TABLE temp_test"},
			Expected: &core.DatabaseExpectation{
				RowCount: 1,
			},
		}

		result := tester.ExecuteTest(test)

		assert.NotNil(t, result)
		assert.Equal(t, core.TestStatusPassed, result.Status)
	})
}

func TestDatabaseTester_Cleanup(t *testing.T) {
	tester := NewDatabaseTester()
	config := &config.DatabaseConfig{
		Connections: map[string]*config.DatabaseConnection{
			"test": {
				Driver:   "sqlite3",
				Database: ":memory:",
			},
		},
	}

	err := tester.Initialize(config)
	require.NoError(t, err)
	assert.True(t, tester.initialized)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
	assert.False(t, tester.initialized)
}

func TestDatabaseTester_ValidateResult(t *testing.T) {
	tester := NewDatabaseTester()
	config := &config.DatabaseConfig{
		Connections: map[string]*config.DatabaseConnection{
			"test": {
				Driver:   "sqlite3",
				Database: ":memory:",
			},
		},
	}
	err := tester.Initialize(config)
	require.NoError(t, err)

	t.Run("row count validation", func(t *testing.T) {
		result := &core.DatabaseResult{
			RowCount: 5,
		}

		expected := &core.DatabaseExpectation{
			RowCount: 5,
		}

		tester.asserter.Reset()
		tester.validateResult(result, expected)

		// Check if assertion passed
		assert.False(t, tester.asserter.HasFailures())
	})

	t.Run("rows affected validation", func(t *testing.T) {
		result := &core.DatabaseResult{
			RowsAffected: 3,
		}

		expected := &core.DatabaseExpectation{
			RowsAffected: 3,
		}

		tester.asserter.Reset()
		tester.validateResult(result, expected)

		// Check if assertion passed
		assert.False(t, tester.asserter.HasFailures())
	})

	t.Run("failed validation", func(t *testing.T) {
		result := &core.DatabaseResult{
			RowCount: 3,
		}

		expected := &core.DatabaseExpectation{
			RowCount: 5,
		}

		tester.asserter.Reset()
		tester.validateResult(result, expected)

		// Check if assertion failed
		assert.True(t, tester.asserter.HasFailures())
	})
}

func TestDatabaseTransaction_Methods(t *testing.T) {
	tx := &DatabaseTransaction{}

	// Test Commit
	err := tx.Commit()
	assert.NoError(t, err)

	// Test Rollback
	err = tx.Rollback()
	assert.NoError(t, err)

	// Test Execute
	result, err := tx.Execute("SELECT 1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
