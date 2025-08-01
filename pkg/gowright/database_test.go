package gowright

import (
	"fmt"
	"time"
)

// DatabaseTestImpl implements the Test interface for database testing
type DatabaseTestImpl struct {
	testCase *DatabaseTest
	tester   DatabaseTester
}

// NewDatabaseTest creates a new database test instance
func NewDatabaseTest(testCase *DatabaseTest, tester DatabaseTester) *DatabaseTestImpl {
	return &DatabaseTestImpl{
		testCase: testCase,
		tester:   tester,
	}
}

// GetName returns the name of the test
func (dt *DatabaseTestImpl) GetName() string {
	return dt.testCase.Name
}

// Execute executes the database test and returns the result
func (dt *DatabaseTestImpl) Execute() *TestCaseResult {
	startTime := time.Now()
	result := &TestCaseResult{
		Name:      dt.testCase.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
		Logs:      []string{},
	}

	// Execute setup queries if any
	if len(dt.testCase.Setup) > 0 {
		for i, setupQuery := range dt.testCase.Setup {
			result.Logs = append(result.Logs, fmt.Sprintf("Executing setup query %d: %s", i+1, setupQuery))
			if _, err := dt.tester.Execute(dt.testCase.Connection, setupQuery); err != nil {
				result.Status = TestStatusError
				result.Error = NewGowrightError(DatabaseError, 
					fmt.Sprintf("setup query %d failed", i+1), err).
					WithContext("setup_query", setupQuery).
					WithContext("connection", dt.testCase.Connection)
				result.Logs = append(result.Logs, fmt.Sprintf("Setup query %d failed: %v", i+1, err))
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(startTime)
				return result
			}
			result.Logs = append(result.Logs, fmt.Sprintf("Setup query %d executed successfully", i+1))
		}
	}

	// Execute main query
	result.Logs = append(result.Logs, fmt.Sprintf("Executing main query: %s", dt.testCase.Query))
	queryResult, err := dt.tester.Execute(dt.testCase.Connection, dt.testCase.Query)
	if err != nil {
		result.Status = TestStatusError
		result.Error = NewGowrightError(DatabaseError, "main query execution failed", err).
			WithContext("query", dt.testCase.Query).
			WithContext("connection", dt.testCase.Connection)
		result.Logs = append(result.Logs, fmt.Sprintf("Main query failed: %v", err))
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result
	}

	result.Logs = append(result.Logs, fmt.Sprintf("Main query executed successfully, %d rows affected", queryResult.RowsAffected))

	// Validate results if expected results are provided
	if dt.testCase.Expected != nil {
		if err := dt.validateResults(queryResult, dt.testCase.Expected); err != nil {
			result.Status = TestStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(startTime)
			return result
		}
		result.Logs = append(result.Logs, "Result validation passed")
	}

	// Execute teardown queries if any
	if len(dt.testCase.Teardown) > 0 {
		for i, teardownQuery := range dt.testCase.Teardown {
			if _, err := dt.tester.Execute(dt.testCase.Connection, teardownQuery); err != nil {
				// Teardown failures are logged but don't fail the test
				result.Logs = append(result.Logs, fmt.Sprintf("Warning: teardown query %d failed: %v", i+1, err))
			} else {
				result.Logs = append(result.Logs, fmt.Sprintf("Teardown query %d executed successfully", i+1))
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	return result
}

// validateResults validates the query results against expected values
func (dt *DatabaseTestImpl) validateResults(actual *DatabaseResult, expected *DatabaseExpectation) error {
	// Validate row count if specified
	if expected.RowCount > 0 {
		if len(actual.Rows) != expected.RowCount {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows, got %d", expected.RowCount, len(actual.Rows)), nil).
				WithContext("expected_rows", expected.RowCount).
				WithContext("actual_rows", len(actual.Rows))
		}
	}

	// Validate rows affected if specified
	if expected.RowsAffected > 0 {
		if actual.RowsAffected != expected.RowsAffected {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows affected, got %d", expected.RowsAffected, actual.RowsAffected), nil).
				WithContext("expected_rows_affected", expected.RowsAffected).
				WithContext("actual_rows_affected", actual.RowsAffected)
		}
	}

	// Validate specific rows if specified
	if expected.Rows != nil && len(expected.Rows) > 0 {
		if len(actual.Rows) != len(expected.Rows) {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows, got %d", len(expected.Rows), len(actual.Rows)), nil)
		}

		for i, expectedRow := range expected.Rows {
			if i >= len(actual.Rows) {
				return NewGowrightError(AssertionError, 
					fmt.Sprintf("missing row at index %d", i), nil)
			}

			actualRow := actual.Rows[i]
			for key, expectedValue := range expectedRow {
				actualValue, exists := actualRow[key]
				if !exists {
					return NewGowrightError(AssertionError, 
						fmt.Sprintf("missing column '%s' in row %d", key, i), nil).
						WithContext("row_index", i).
						WithContext("column", key)
				}

				if !compareValues(expectedValue, actualValue) {
					return NewGowrightError(AssertionError, 
						fmt.Sprintf("value mismatch in row %d, column '%s': expected %v, got %v", 
							i, key, expectedValue, actualValue), nil).
						WithContext("row_index", i).
						WithContext("column", key).
						WithContext("expected", expectedValue).
						WithContext("actual", actualValue)
				}
			}
		}
	}

	return nil
}

// DatabaseAssertions provides database-specific assertion methods
type DatabaseAssertions struct {
	tester DatabaseTester
}

// NewDatabaseAssertions creates a new DatabaseAssertions instance
func NewDatabaseAssertions(tester DatabaseTester) *DatabaseAssertions {
	return &DatabaseAssertions{
		tester: tester,
	}
}

// AssertRowCount asserts that a query returns the expected number of rows
func (da *DatabaseAssertions) AssertRowCount(connectionName, query string, expectedCount int, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for row count assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	actualCount := len(result.Rows)
	if actualCount != expectedCount {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("expected %d rows, got %d", expectedCount, actualCount), nil).
			WithContext("expected_count", expectedCount).
			WithContext("actual_count", actualCount).
			WithContext("query", query)
	}

	return nil
}

// AssertRowsAffected asserts that a query affects the expected number of rows
func (da *DatabaseAssertions) AssertRowsAffected(connectionName, query string, expectedAffected int64, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for rows affected assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	if result.RowsAffected != expectedAffected {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("expected %d rows affected, got %d", expectedAffected, result.RowsAffected), nil).
			WithContext("expected_affected", expectedAffected).
			WithContext("actual_affected", result.RowsAffected).
			WithContext("query", query)
	}

	return nil
}

// AssertRowExists asserts that at least one row exists matching the query
func (da *DatabaseAssertions) AssertRowExists(connectionName, query string, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for row existence assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) == 0 {
		return NewGowrightError(AssertionError, "expected at least one row, got none", nil).
			WithContext("query", query)
	}

	return nil
}

// AssertRowNotExists asserts that no rows exist matching the query
func (da *DatabaseAssertions) AssertRowNotExists(connectionName, query string, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for row non-existence assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) > 0 {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("expected no rows, got %d", len(result.Rows)), nil).
			WithContext("actual_count", len(result.Rows)).
			WithContext("query", query)
	}

	return nil
}

// AssertColumnValue asserts that a specific column has the expected value in the first row
func (da *DatabaseAssertions) AssertColumnValue(connectionName, query, columnName string, expectedValue interface{}, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for column value assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) == 0 {
		return NewGowrightError(AssertionError, "no rows returned for column value assertion", nil).
			WithContext("query", query).
			WithContext("column", columnName)
	}

	firstRow := result.Rows[0]
	actualValue, exists := firstRow[columnName]
	if !exists {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("column '%s' not found in result", columnName), nil).
			WithContext("column", columnName).
			WithContext("available_columns", getColumnNames(firstRow))
	}

	if !compareValues(expectedValue, actualValue) {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("column '%s' value mismatch: expected %v, got %v", 
				columnName, expectedValue, actualValue), nil).
			WithContext("column", columnName).
			WithContext("expected", expectedValue).
			WithContext("actual", actualValue)
	}

	return nil
}

// AssertColumnContains asserts that a column value contains the expected substring
func (da *DatabaseAssertions) AssertColumnContains(connectionName, query, columnName, expectedSubstring string, args ...interface{}) error {
	result, err := da.tester.Execute(connectionName, query, args...)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to execute query for column contains assertion", err).
			WithContext("query", query).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) == 0 {
		return NewGowrightError(AssertionError, "no rows returned for column contains assertion", nil).
			WithContext("query", query).
			WithContext("column", columnName)
	}

	firstRow := result.Rows[0]
	actualValue, exists := firstRow[columnName]
	if !exists {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("column '%s' not found in result", columnName), nil).
			WithContext("column", columnName)
	}

	actualStr, ok := actualValue.(string)
	if !ok {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("column '%s' is not a string: %T", columnName, actualValue), nil).
			WithContext("column", columnName).
			WithContext("actual_type", fmt.Sprintf("%T", actualValue))
	}

	if !contains(actualStr, expectedSubstring) {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("column '%s' value '%s' does not contain '%s'", 
				columnName, actualStr, expectedSubstring), nil).
			WithContext("column", columnName).
			WithContext("actual_value", actualStr).
			WithContext("expected_substring", expectedSubstring)
	}

	return nil
}

// AssertTableExists asserts that a table exists in the database
func (da *DatabaseAssertions) AssertTableExists(connectionName, tableName string) error {
	// This is a generic implementation - specific databases might need different queries
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	result, err := da.tester.Execute(connectionName, query, tableName)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to check table existence", err).
			WithContext("table", tableName).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) == 0 {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("table '%s' does not exist", tableName), nil).
			WithContext("table", tableName)
	}

	return nil
}

// AssertTableNotExists asserts that a table does not exist in the database
func (da *DatabaseAssertions) AssertTableNotExists(connectionName, tableName string) error {
	// This is a generic implementation - specific databases might need different queries
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	result, err := da.tester.Execute(connectionName, query, tableName)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to check table non-existence", err).
			WithContext("table", tableName).
			WithContext("connection", connectionName)
	}

	if len(result.Rows) > 0 {
		return NewGowrightError(AssertionError, 
			fmt.Sprintf("table '%s' exists but should not", tableName), nil).
			WithContext("table", tableName)
	}

	return nil
}

// TransactionTestRunner provides utilities for running tests within transactions
type TransactionTestRunner struct {
	tester DatabaseTester
}

// NewTransactionTestRunner creates a new TransactionTestRunner
func NewTransactionTestRunner(tester DatabaseTester) *TransactionTestRunner {
	return &TransactionTestRunner{
		tester: tester,
	}
}

// RunInTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tr *TransactionTestRunner) RunInTransaction(connectionName string, testFunc func(tx Transaction) error) error {
	tx, err := tr.tester.BeginTransaction(connectionName)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to begin transaction", err).
			WithContext("connection", connectionName)
	}

	defer func() {
		// Always attempt rollback in case of panic or if commit wasn't called
		tx.Rollback()
	}()

	if err := testFunc(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return NewGowrightError(DatabaseError, "failed to rollback transaction after error", rollbackErr).
				WithContext("original_error", err).
				WithContext("connection", connectionName)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return NewGowrightError(DatabaseError, "failed to commit transaction", err).
			WithContext("connection", connectionName)
	}

	return nil
}

// RunWithRollback executes a function within a transaction and always rolls back
// This is useful for tests that need to verify behavior without persisting changes
func (tr *TransactionTestRunner) RunWithRollback(connectionName string, testFunc func(tx Transaction) error) error {
	tx, err := tr.tester.BeginTransaction(connectionName)
	if err != nil {
		return NewGowrightError(DatabaseError, "failed to begin transaction", err).
			WithContext("connection", connectionName)
	}

	defer func() {
		tx.Rollback()
	}()

	return testFunc(tx)
}

// Helper functions

// getColumnNames extracts column names from a row map
func getColumnNames(row map[string]interface{}) []string {
	columns := make([]string, 0, len(row))
	for col := range row {
		columns = append(columns, col)
	}
	return columns
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstring(s, substr) >= 0)
}

// findSubstring finds the index of a substring in a string
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}