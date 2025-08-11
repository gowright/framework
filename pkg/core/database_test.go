package core

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

// validateResults validates the query results against expected results
func (dt *DatabaseTestImpl) validateResults(actual *DatabaseResult, expected *DatabaseExpectation) error {
	// Validate rows affected if specified
	if expected.RowsAffected >= 0 && actual.RowsAffected != expected.RowsAffected {
		return NewGowrightError(ValidationError,
			fmt.Sprintf("expected %d rows affected, got %d", expected.RowsAffected, actual.RowsAffected),
			nil)
	}

	// Validate result data if specified
	if expected.Rows != nil {
		if actual.Rows == nil {
			return NewGowrightError(ValidationError, "expected result data but got nil", nil)
		}

		// Compare result data (simplified comparison)
		if len(expected.Rows) != len(actual.Rows) {
			return NewGowrightError(ValidationError,
				fmt.Sprintf("expected %d result rows, got %d", len(expected.Rows), len(actual.Rows)),
				nil)
		}

		// Validate each row
		for i, expectedRow := range expected.Rows {
			actualRow := actual.Rows[i]
			if len(expectedRow) != len(actualRow) {
				return NewGowrightError(ValidationError,
					fmt.Sprintf("row %d: expected %d columns, got %d", i, len(expectedRow), len(actualRow)),
					nil)
			}

			// Validate each column (basic comparison)
			for j, expectedValue := range expectedRow {
				actualValue := actualRow[j]
				if expectedValue != actualValue {
					return NewGowrightError(ValidationError,
						fmt.Sprintf("row %d, column %s: expected %v, got %v", i, j, expectedValue, actualValue),
						nil)
				}
			}
		}
	}

	return nil
}

// DatabaseTestBuilder provides a fluent interface for building database tests
type DatabaseTestBuilder struct {
	testCase *DatabaseTest
}

// NewDatabaseTestBuilder creates a new database test builder
func NewDatabaseTestBuilder(name string) *DatabaseTestBuilder {
	return &DatabaseTestBuilder{
		testCase: &DatabaseTest{
			Name:     name,
			Setup:    make([]string, 0),
			Teardown: make([]string, 0),
		},
	}
}

// WithConnection sets the database connection
func (dtb *DatabaseTestBuilder) WithConnection(connection string) *DatabaseTestBuilder {
	dtb.testCase.Connection = connection
	return dtb
}

// WithQuery sets the main query
func (dtb *DatabaseTestBuilder) WithQuery(query string) *DatabaseTestBuilder {
	dtb.testCase.Query = query
	return dtb
}

// WithSetup adds a setup query
func (dtb *DatabaseTestBuilder) WithSetup(query string) *DatabaseTestBuilder {
	dtb.testCase.Setup = append(dtb.testCase.Setup, query)
	return dtb
}

// WithTeardown adds a teardown query
func (dtb *DatabaseTestBuilder) WithTeardown(query string) *DatabaseTestBuilder {
	dtb.testCase.Teardown = append(dtb.testCase.Teardown, query)
	return dtb
}

// WithExpected sets the expected results
func (dtb *DatabaseTestBuilder) WithExpected(expected *DatabaseExpectation) *DatabaseTestBuilder {
	dtb.testCase.Expected = expected
	return dtb
}

// WithExpectedRowsAffected sets the expected number of rows affected
func (dtb *DatabaseTestBuilder) WithExpectedRowsAffected(count int64) *DatabaseTestBuilder {
	if dtb.testCase.Expected == nil {
		dtb.testCase.Expected = &DatabaseExpectation{}
	}
	dtb.testCase.Expected.RowsAffected = count
	return dtb
}

// WithExpectedRows sets the expected result rows
func (dtb *DatabaseTestBuilder) WithExpectedRows(rows []map[string]interface{}) *DatabaseTestBuilder {
	if dtb.testCase.Expected == nil {
		dtb.testCase.Expected = &DatabaseExpectation{}
	}
	dtb.testCase.Expected.Rows = rows
	return dtb
}

// Build creates the database test
func (dtb *DatabaseTestBuilder) Build(tester DatabaseTester) Test {
	return NewDatabaseTest(dtb.testCase, tester)
}
