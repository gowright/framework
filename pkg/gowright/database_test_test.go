package gowright

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseTest(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "test_query",
		Connection: "test_conn",
		Query:      "SELECT 1",
	}
	
	tester := NewDatabaseTester()
	dbTest := NewDatabaseTest(testCase, tester)
	
	assert.NotNil(t, dbTest)
	assert.Equal(t, "test_query", dbTest.GetName())
	assert.Equal(t, testCase, dbTest.testCase)
	assert.Equal(t, tester, dbTest.tester)
}

func TestDatabaseTestImpl_Execute(t *testing.T) {
	// Setup database tester
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

	// Create test table
	_, err = tester.Execute("test", "CREATE TABLE test_users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	require.NoError(t, err)

	tests := []struct {
		name           string
		testCase       *DatabaseTest
		expectedStatus TestStatus
		setupData      bool
	}{
		{
			name: "successful select query",
			testCase: &DatabaseTest{
				Name:       "select_test",
				Connection: "test",
				Setup:      []string{"INSERT INTO test_users (name, age) VALUES ('Alice', 30)"},
				Query:      "SELECT name, age FROM test_users WHERE name = 'Alice'",
				Expected: &DatabaseExpectation{
					RowCount: 1,
					Rows: []map[string]interface{}{
						{"name": "Alice", "age": int64(30)},
					},
				},
				Teardown: []string{"DELETE FROM test_users WHERE name = 'Alice'"},
			},
			expectedStatus: TestStatusPassed,
		},
		{
			name: "successful insert query",
			testCase: &DatabaseTest{
				Name:       "insert_test",
				Connection: "test",
				Query:      "INSERT INTO test_users (name, age) VALUES ('Bob', 25)",
				Expected: &DatabaseExpectation{
					RowsAffected: 1,
				},
				Teardown: []string{"DELETE FROM test_users WHERE name = 'Bob'"},
			},
			expectedStatus: TestStatusPassed,
		},
		{
			name: "failed assertion",
			testCase: &DatabaseTest{
				Name:       "failed_assertion_test",
				Connection: "test",
				Setup:      []string{"INSERT INTO test_users (name, age) VALUES ('Charlie', 35)"},
				Query:      "SELECT name, age FROM test_users WHERE name = 'Charlie'",
				Expected: &DatabaseExpectation{
					RowCount: 2, // Wrong count - should be 1
				},
				Teardown: []string{"DELETE FROM test_users WHERE name = 'Charlie'"},
			},
			expectedStatus: TestStatusFailed,
		},
		{
			name: "setup query failure",
			testCase: &DatabaseTest{
				Name:       "setup_failure_test",
				Connection: "test",
				Setup:      []string{"INSERT INTO nonexistent_table (name) VALUES ('test')"},
				Query:      "SELECT 1",
			},
			expectedStatus: TestStatusError,
		},
		{
			name: "main query failure",
			testCase: &DatabaseTest{
				Name:       "query_failure_test",
				Connection: "test",
				Query:      "SELECT * FROM nonexistent_table",
			},
			expectedStatus: TestStatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbTest := NewDatabaseTest(tt.testCase, tester)
			result := dbTest.Execute()
			
			assert.Equal(t, tt.testCase.Name, result.Name)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.NotZero(t, result.Duration)
			assert.NotEmpty(t, result.Logs)
			
			if tt.expectedStatus == TestStatusError || tt.expectedStatus == TestStatusFailed {
				assert.NotNil(t, result.Error)
			} else {
				assert.Nil(t, result.Error)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertRowCount(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE count_test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO count_test (value) VALUES ('a'), ('b'), ('c')")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "correct count",
			query:         "SELECT * FROM count_test",
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "wrong count",
			query:         "SELECT * FROM count_test",
			expectedCount: 2,
			expectError:   true,
		},
		{
			name:          "filtered count",
			query:         "SELECT * FROM count_test WHERE value = 'a'",
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertRowCount("test", tt.query, tt.expectedCount)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertRowsAffected(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE affected_test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO affected_test (value) VALUES ('a'), ('b'), ('c')")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name             string
		query            string
		expectedAffected int64
		expectError      bool
	}{
		{
			name:             "single insert",
			query:            "INSERT INTO affected_test (value) VALUES ('d')",
			expectedAffected: 1,
			expectError:      false,
		},
		{
			name:             "update multiple rows",
			query:            "UPDATE affected_test SET value = 'updated' WHERE value IN ('a', 'b')",
			expectedAffected: 2,
			expectError:      false,
		},
		{
			name:             "wrong affected count",
			query:            "DELETE FROM affected_test WHERE value = 'c'",
			expectedAffected: 2, // Should be 1
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertRowsAffected("test", tt.query, tt.expectedAffected)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertRowExists(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE exists_test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO exists_test (name) VALUES ('Alice')")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "row exists",
			query:       "SELECT * FROM exists_test WHERE name = 'Alice'",
			expectError: false,
		},
		{
			name:        "row does not exist",
			query:       "SELECT * FROM exists_test WHERE name = 'Bob'",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertRowExists("test", tt.query)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertRowNotExists(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE not_exists_test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO not_exists_test (name) VALUES ('Alice')")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "row does not exist",
			query:       "SELECT * FROM not_exists_test WHERE name = 'Bob'",
			expectError: false,
		},
		{
			name:        "row exists",
			query:       "SELECT * FROM not_exists_test WHERE name = 'Alice'",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertRowNotExists("test", tt.query)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertColumnValue(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE column_test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO column_test (name, age) VALUES ('Alice', 30)")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name          string
		query         string
		columnName    string
		expectedValue interface{}
		expectError   bool
	}{
		{
			name:          "correct string value",
			query:         "SELECT name, age FROM column_test WHERE name = 'Alice'",
			columnName:    "name",
			expectedValue: "Alice",
			expectError:   false,
		},
		{
			name:          "correct integer value",
			query:         "SELECT name, age FROM column_test WHERE name = 'Alice'",
			columnName:    "age",
			expectedValue: int64(30),
			expectError:   false,
		},
		{
			name:          "wrong value",
			query:         "SELECT name, age FROM column_test WHERE name = 'Alice'",
			columnName:    "age",
			expectedValue: int64(25),
			expectError:   true,
		},
		{
			name:          "column not found",
			query:         "SELECT name FROM column_test WHERE name = 'Alice'",
			columnName:    "age",
			expectedValue: int64(30),
			expectError:   true,
		},
		{
			name:          "no rows returned",
			query:         "SELECT name, age FROM column_test WHERE name = 'Bob'",
			columnName:    "name",
			expectedValue: "Bob",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertColumnValue("test", tt.query, tt.columnName, tt.expectedValue)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertColumnContains(t *testing.T) {
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

	// Setup test data
	_, err = tester.Execute("test", "CREATE TABLE contains_test (id INTEGER PRIMARY KEY, description TEXT)")
	require.NoError(t, err)
	
	_, err = tester.Execute("test", "INSERT INTO contains_test (description) VALUES ('This is a test description')")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name               string
		query              string
		columnName         string
		expectedSubstring  string
		expectError        bool
	}{
		{
			name:              "contains substring",
			query:             "SELECT description FROM contains_test",
			columnName:        "description",
			expectedSubstring: "test",
			expectError:       false,
		},
		{
			name:              "does not contain substring",
			query:             "SELECT description FROM contains_test",
			columnName:        "description",
			expectedSubstring: "missing",
			expectError:       true,
		},
		{
			name:              "column not found",
			query:             "SELECT description FROM contains_test",
			columnName:        "nonexistent",
			expectedSubstring: "test",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertColumnContains("test", tt.query, tt.columnName, tt.expectedSubstring)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestDatabaseAssertions_AssertTableExists(t *testing.T) {
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

	// Create test table
	_, err = tester.Execute("test", "CREATE TABLE existing_table (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	assertions := NewDatabaseAssertions(tester)

	tests := []struct {
		name        string
		tableName   string
		expectError bool
	}{
		{
			name:        "table exists",
			tableName:   "existing_table",
			expectError: false,
		},
		{
			name:        "table does not exist",
			tableName:   "nonexistent_table",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertions.AssertTableExists("test", tt.tableName)
			
			if tt.expectError {
				assert.Error(t, err)
				gowrightErr, ok := err.(*GowrightError)
				assert.True(t, ok)
				assert.Equal(t, AssertionError, gowrightErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestTransactionTestRunner_RunInTransaction(t *testing.T) {
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

	// Setup test table
	_, err = tester.Execute("test", "CREATE TABLE tx_test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	runner := NewTransactionTestRunner(tester)

	t.Run("successful transaction", func(t *testing.T) {
		err := runner.RunInTransaction("test", func(tx Transaction) error {
			_, err := tx.Execute("INSERT INTO tx_test (value) VALUES (?)", "test_value")
			return err
		})
		
		assert.NoError(t, err)
		
		// Verify data was committed
		result, err := tester.Execute("test", "SELECT COUNT(*) as count FROM tx_test")
		assert.NoError(t, err)
		assert.Len(t, result.Rows, 1)
	})

	t.Run("failed transaction", func(t *testing.T) {
		err := runner.RunInTransaction("test", func(tx Transaction) error {
			_, err := tx.Execute("INSERT INTO tx_test (value) VALUES (?)", "rollback_value")
			if err != nil {
				return err
			}
			// Return an error to trigger rollback
			return NewGowrightError(DatabaseError, "intentional error", nil)
		})
		
		assert.Error(t, err)
		
		// Verify data was rolled back - count should still be 1
		result, err := tester.Execute("test", "SELECT COUNT(*) as count FROM tx_test")
		assert.NoError(t, err)
		assert.Len(t, result.Rows, 1)
		count := result.Rows[0]["count"]
		assert.Equal(t, int64(1), count)
	})

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

func TestTransactionTestRunner_RunWithRollback(t *testing.T) {
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

	// Setup test table
	_, err = tester.Execute("test", "CREATE TABLE rollback_test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	runner := NewTransactionTestRunner(tester)

	err = runner.RunWithRollback("test", func(tx Transaction) error {
		_, err := tx.Execute("INSERT INTO rollback_test (value) VALUES (?)", "temp_value")
		assert.NoError(t, err)
		
		// Verify data exists within transaction
		result, err := tx.Execute("SELECT COUNT(*) as count FROM rollback_test")
		assert.NoError(t, err)
		assert.Len(t, result.Rows, 1)
		count := result.Rows[0]["count"]
		assert.Equal(t, int64(1), count)
		
		return nil
	})
	
	assert.NoError(t, err)
	
	// Verify data was rolled back
	result, err := tester.Execute("test", "SELECT COUNT(*) as count FROM rollback_test")
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 1)
	count := result.Rows[0]["count"]
	assert.Equal(t, int64(0), count)

	// Cleanup
	err = tester.Cleanup()
	assert.NoError(t, err)
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("getColumnNames", func(t *testing.T) {
		row := map[string]interface{}{
			"id":   1,
			"name": "test",
			"age":  30,
		}
		
		columns := getColumnNames(row)
		assert.Len(t, columns, 3)
		assert.Contains(t, columns, "id")
		assert.Contains(t, columns, "name")
		assert.Contains(t, columns, "age")
	})

	t.Run("contains", func(t *testing.T) {
		tests := []struct {
			s      string
			substr string
			result bool
		}{
			{"hello world", "world", true},
			{"hello world", "foo", false},
			{"test", "", true},
			{"", "test", false},
			{"", "", true},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.result, contains(tt.s, tt.substr))
		}
	})

	t.Run("findSubstring", func(t *testing.T) {
		tests := []struct {
			s      string
			substr string
			result int
		}{
			{"hello world", "world", 6},
			{"hello world", "foo", -1},
			{"test", "", 0},
			{"", "test", -1},
			{"", "", 0},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.result, findSubstring(tt.s, tt.substr))
		}
	})
}