package gowright

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseTesterImpl implements the DatabaseTester interface
type DatabaseTesterImpl struct {
	config      *DatabaseConfig
	connections map[string]*sql.DB
	mutex       sync.RWMutex
	initialized bool
}

// NewDatabaseTester creates a new DatabaseTester instance
func NewDatabaseTester() *DatabaseTesterImpl {
	return &DatabaseTesterImpl{
		connections: make(map[string]*sql.DB),
	}
}

// Initialize sets up the database tester with the provided configuration
func (dt *DatabaseTesterImpl) Initialize(config interface{}) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	dbConfig, ok := config.(*DatabaseConfig)
	if !ok {
		return NewGowrightError(ConfigurationError, "invalid configuration type for DatabaseTester", nil)
	}

	if dbConfig == nil {
		return NewGowrightError(ConfigurationError, "database configuration cannot be nil", nil)
	}

	dt.config = dbConfig
	dt.initialized = true

	return nil
}

// Connect establishes a connection to the specified database
func (dt *DatabaseTesterImpl) Connect(connectionName string) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if !dt.initialized {
		return NewGowrightError(ConfigurationError, "DatabaseTester not initialized", nil)
	}

	// Check if connection already exists
	if _, exists := dt.connections[connectionName]; exists {
		return nil // Connection already established
	}

	// Get connection configuration
	connConfig, exists := dt.config.Connections[connectionName]
	if !exists {
		return NewGowrightError(ConfigurationError, 
			fmt.Sprintf("connection configuration not found: %s", connectionName), nil).
			WithContext("connection_name", connectionName)
	}

	// Validate connection configuration
	if err := connConfig.Validate(); err != nil {
		return NewGowrightError(ConfigurationError, 
			fmt.Sprintf("invalid connection configuration for %s", connectionName), err).
			WithContext("connection_name", connectionName)
	}

	// Open database connection
	db, err := sql.Open(connConfig.Driver, connConfig.DSN)
	if err != nil {
		return NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to open database connection %s", connectionName), err).
			WithContext("connection_name", connectionName).
			WithContext("driver", connConfig.Driver)
	}

	// Configure connection pool
	if connConfig.MaxOpenConns > 0 {
		db.SetMaxOpenConns(connConfig.MaxOpenConns)
	}
	if connConfig.MaxIdleConns > 0 {
		db.SetMaxIdleConns(connConfig.MaxIdleConns)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to ping database connection %s", connectionName), err).
			WithContext("connection_name", connectionName).
			WithContext("driver", connConfig.Driver)
	}

	dt.connections[connectionName] = db
	return nil
}

// Execute executes a SQL query and returns the result
func (dt *DatabaseTesterImpl) Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error) {
	dt.mutex.RLock()
	db, exists := dt.connections[connectionName]
	dt.mutex.RUnlock()

	if !exists {
		// Try to establish connection if it doesn't exist
		if err := dt.Connect(connectionName); err != nil {
			return nil, err
		}
		dt.mutex.RLock()
		db = dt.connections[connectionName]
		dt.mutex.RUnlock()
	}

	startTime := time.Now()
	
	// Determine if this is a SELECT query or a modification query
	if isSelectQuery(query) {
		return dt.executeQuery(db, query, args, startTime)
	} else {
		return dt.executeExec(db, query, args, startTime)
	}
}

// executeQuery handles SELECT queries that return rows
func (dt *DatabaseTesterImpl) executeQuery(db *sql.DB, query string, args []interface{}, startTime time.Time) (*DatabaseResult, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to execute query", err).
			WithContext("query", query).
			WithContext("args", args)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to get column names", err)
	}

	var result []map[string]interface{}
	
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, NewGowrightError(DatabaseError, "failed to scan row", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			
			rowMap[col] = val
		}
		
		result = append(result, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, NewGowrightError(DatabaseError, "error iterating over rows", err)
	}

	duration := time.Since(startTime)
	return &DatabaseResult{
		Rows:         result,
		RowsAffected: int64(len(result)),
		Duration:     duration,
	}, nil
}

// executeExec handles INSERT, UPDATE, DELETE queries
func (dt *DatabaseTesterImpl) executeExec(db *sql.DB, query string, args []interface{}, startTime time.Time) (*DatabaseResult, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to execute statement", err).
			WithContext("query", query).
			WithContext("args", args)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to get rows affected", err)
	}

	duration := time.Since(startTime)
	return &DatabaseResult{
		Rows:         nil, // No rows returned for exec operations
		RowsAffected: rowsAffected,
		Duration:     duration,
	}, nil
}

// BeginTransaction starts a new database transaction
func (dt *DatabaseTesterImpl) BeginTransaction(connectionName string) (Transaction, error) {
	dt.mutex.RLock()
	db, exists := dt.connections[connectionName]
	dt.mutex.RUnlock()

	if !exists {
		// Try to establish connection if it doesn't exist
		if err := dt.Connect(connectionName); err != nil {
			return nil, err
		}
		dt.mutex.RLock()
		db = dt.connections[connectionName]
		dt.mutex.RUnlock()
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to begin transaction on connection %s", connectionName), err).
			WithContext("connection_name", connectionName)
	}

	return &TransactionImpl{
		tx:             tx,
		connectionName: connectionName,
	}, nil
}

// ValidateData validates data against expected results
func (dt *DatabaseTesterImpl) ValidateData(connectionName, query string, expected interface{}) error {
	result, err := dt.Execute(connectionName, query)
	if err != nil {
		return err
	}

	expectedResult, ok := expected.(*DatabaseExpectation)
	if !ok {
		return NewGowrightError(AssertionError, "expected result must be of type *DatabaseExpectation", nil)
	}

	// Validate row count if specified
	if expectedResult.RowCount > 0 {
		if len(result.Rows) != expectedResult.RowCount {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows, got %d", expectedResult.RowCount, len(result.Rows)), nil).
				WithContext("expected_rows", expectedResult.RowCount).
				WithContext("actual_rows", len(result.Rows))
		}
	}

	// Validate rows affected if specified
	if expectedResult.RowsAffected > 0 {
		if result.RowsAffected != expectedResult.RowsAffected {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows affected, got %d", expectedResult.RowsAffected, result.RowsAffected), nil).
				WithContext("expected_rows_affected", expectedResult.RowsAffected).
				WithContext("actual_rows_affected", result.RowsAffected)
		}
	}

	// Validate specific rows if specified
	if expectedResult.Rows != nil && len(expectedResult.Rows) > 0 {
		if len(result.Rows) != len(expectedResult.Rows) {
			return NewGowrightError(AssertionError, 
				fmt.Sprintf("expected %d rows, got %d", len(expectedResult.Rows), len(result.Rows)), nil)
		}

		for i, expectedRow := range expectedResult.Rows {
			if i >= len(result.Rows) {
				return NewGowrightError(AssertionError, 
					fmt.Sprintf("missing row at index %d", i), nil)
			}

			actualRow := result.Rows[i]
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

// Cleanup performs any necessary cleanup operations
func (dt *DatabaseTesterImpl) Cleanup() error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	var errors []error
	
	for name, db := range dt.connections {
		if err := db.Close(); err != nil {
			errors = append(errors, NewGowrightError(DatabaseError, 
				fmt.Sprintf("failed to close connection %s", name), err).
				WithContext("connection_name", name))
		}
	}

	dt.connections = make(map[string]*sql.DB)
	dt.initialized = false

	if len(errors) > 0 {
		return NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to cleanup %d connections", len(errors)), nil).
			WithContext("errors", errors)
	}

	return nil
}

// GetName returns the name of the tester
func (dt *DatabaseTesterImpl) GetName() string {
	return "DatabaseTester"
}

// TransactionImpl implements the Transaction interface
type TransactionImpl struct {
	tx             *sql.Tx
	connectionName string
}

// Commit commits the transaction
func (t *TransactionImpl) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to commit transaction on connection %s", t.connectionName), err).
			WithContext("connection_name", t.connectionName)
	}
	return nil
}

// Rollback rolls back the transaction
func (t *TransactionImpl) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return NewGowrightError(DatabaseError, 
			fmt.Sprintf("failed to rollback transaction on connection %s", t.connectionName), err).
			WithContext("connection_name", t.connectionName)
	}
	return nil
}

// Execute executes a query within the transaction
func (t *TransactionImpl) Execute(query string, args ...interface{}) (*DatabaseResult, error) {
	startTime := time.Now()
	
	if isSelectQuery(query) {
		return t.executeQuery(query, args, startTime)
	} else {
		return t.executeExec(query, args, startTime)
	}
}

// executeQuery handles SELECT queries within transaction
func (t *TransactionImpl) executeQuery(query string, args []interface{}, startTime time.Time) (*DatabaseResult, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to execute query in transaction", err).
			WithContext("query", query).
			WithContext("args", args)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to get column names", err)
	}

	var result []map[string]interface{}
	
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, NewGowrightError(DatabaseError, "failed to scan row", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			
			rowMap[col] = val
		}
		
		result = append(result, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, NewGowrightError(DatabaseError, "error iterating over rows", err)
	}

	duration := time.Since(startTime)
	return &DatabaseResult{
		Rows:         result,
		RowsAffected: int64(len(result)),
		Duration:     duration,
	}, nil
}

// executeExec handles INSERT, UPDATE, DELETE queries within transaction
func (t *TransactionImpl) executeExec(query string, args []interface{}, startTime time.Time) (*DatabaseResult, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to execute statement in transaction", err).
			WithContext("query", query).
			WithContext("args", args)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, NewGowrightError(DatabaseError, "failed to get rows affected", err)
	}

	duration := time.Since(startTime)
	return &DatabaseResult{
		Rows:         nil,
		RowsAffected: rowsAffected,
		Duration:     duration,
	}, nil
}

// Helper functions

// isSelectQuery determines if a query is a SELECT statement
func isSelectQuery(query string) bool {
	// Simple check - look for SELECT at the beginning (case insensitive)
	trimmed := trimLeadingWhitespace(query)
	if len(trimmed) < 6 {
		return false
	}
	
	prefix := trimmed[:6]
	return prefix == "SELECT" || prefix == "select" || prefix == "Select"
}

// trimLeadingWhitespace removes leading whitespace from a string
func trimLeadingWhitespace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	return s[start:]
}

// compareValues compares two values for equality, handling type conversions
func compareValues(expected, actual interface{}) bool {
	if expected == actual {
		return true
	}

	// Handle string comparisons
	if expectedStr, ok := expected.(string); ok {
		if actualStr, ok := actual.(string); ok {
			return expectedStr == actualStr
		}
	}

	// Handle numeric comparisons
	if expectedNum, ok := expected.(float64); ok {
		switch actualVal := actual.(type) {
		case float64:
			return expectedNum == actualVal
		case int64:
			return expectedNum == float64(actualVal)
		case int:
			return expectedNum == float64(actualVal)
		}
	}

	if expectedNum, ok := expected.(int64); ok {
		switch actualVal := actual.(type) {
		case int64:
			return expectedNum == actualVal
		case float64:
			return float64(expectedNum) == actualVal
		case int:
			return expectedNum == int64(actualVal)
		}
	}

	if expectedNum, ok := expected.(int); ok {
		switch actualVal := actual.(type) {
		case int:
			return expectedNum == actualVal
		case int64:
			return int64(expectedNum) == actualVal
		case float64:
			return float64(expectedNum) == actualVal
		}
	}

	// Handle boolean comparisons
	if expectedBool, ok := expected.(bool); ok {
		if actualBool, ok := actual.(bool); ok {
			return expectedBool == actualBool
		}
	}

	return false
}

// ExecuteTest executes a database test and returns the result
func (dt *DatabaseTesterImpl) ExecuteTest(test *DatabaseTest) *TestCaseResult {
	startTime := time.Now()
	result := &TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
		Logs:      make([]string, 0),
		Steps:     make([]AssertionStep, 0),
	}

	// Execute setup queries if any
	for i, setupQuery := range test.Setup {
		stepStart := time.Now()
		_, err := dt.Execute(test.Connection, setupQuery)
		stepEnd := time.Now()

		step := AssertionStep{
			Name:        fmt.Sprintf("Setup %d", i+1),
			Description: fmt.Sprintf("Execute setup query: %s", setupQuery),
			StartTime:   stepStart,
			EndTime:     stepEnd,
			Duration:    stepEnd.Sub(stepStart),
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			result.Error = err
			result.Logs = append(result.Logs, fmt.Sprintf("Setup query %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Setup query %d completed", i+1))
		}

		result.Steps = append(result.Steps, step)

		if result.Status == TestStatusFailed {
			break
		}
	}

	// Execute main test query if setup succeeded
	if result.Status == TestStatusPassed {
		stepStart := time.Now()
		queryResult, err := dt.Execute(test.Connection, test.Query)
		stepEnd := time.Now()

		step := AssertionStep{
			Name:        "Main Query",
			Description: fmt.Sprintf("Execute test query: %s", test.Query),
			StartTime:   stepStart,
			EndTime:     stepEnd,
			Duration:    stepEnd.Sub(stepStart),
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			result.Error = err
			result.Logs = append(result.Logs, fmt.Sprintf("Main query failed: %v", err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Main query completed, returned %d rows", len(queryResult.Rows)))

			// Validate results against expectations if provided
			if test.Expected != nil {
				validationStart := time.Now()
				validationErr := dt.ValidateData(test.Connection, test.Query, test.Expected)
				validationEnd := time.Now()

				validationStep := AssertionStep{
					Name:        "Result Validation",
					Description: "Validate query results against expectations",
					StartTime:   validationStart,
					EndTime:     validationEnd,
					Duration:    validationEnd.Sub(validationStart),
				}

				if validationErr != nil {
					validationStep.Status = TestStatusFailed
					validationStep.Error = validationErr
					result.Status = TestStatusFailed
					if result.Error == nil {
						result.Error = validationErr
					}
					result.Logs = append(result.Logs, fmt.Sprintf("Result validation failed: %v", validationErr))
				} else {
					validationStep.Status = TestStatusPassed
					result.Logs = append(result.Logs, "Result validation passed")
				}

				result.Steps = append(result.Steps, validationStep)
			}
		}

		result.Steps = append(result.Steps, step)
	}

	// Execute teardown queries regardless of test result
	for i, teardownQuery := range test.Teardown {
		stepStart := time.Now()
		_, err := dt.Execute(test.Connection, teardownQuery)
		stepEnd := time.Now()

		step := AssertionStep{
			Name:        fmt.Sprintf("Teardown %d", i+1),
			Description: fmt.Sprintf("Execute teardown query: %s", teardownQuery),
			StartTime:   stepStart,
			EndTime:     stepEnd,
			Duration:    stepEnd.Sub(stepStart),
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			// Don't fail the overall test for teardown errors, but log them
			result.Logs = append(result.Logs, fmt.Sprintf("Teardown query %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Teardown query %d completed", i+1))
		}

		result.Steps = append(result.Steps, step)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	return result
}