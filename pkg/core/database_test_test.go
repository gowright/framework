package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseTestImpl(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "test_query",
		Connection: "test_conn",
		Query:      "SELECT 1",
	}

	tester := &MockDatabaseTester{}
	dbTest := NewDatabaseTest(testCase, tester)

	assert.NotNil(t, dbTest)
	assert.Equal(t, "test_query", dbTest.GetName())
	assert.Equal(t, testCase, dbTest.testCase)
	assert.Equal(t, tester, dbTest.tester)
}

func TestDatabaseTestImpl_Execute_Success(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "successful_test",
		Connection: "test_conn",
		Query:      "SELECT 1",
		Setup:      []string{"CREATE TABLE test (id INT)"},
		Teardown:   []string{"DROP TABLE test"},
		Expected: &DatabaseExpectation{
			RowsAffected: 1,
			Rows:         []map[string]interface{}{{"col1": 1}},
		},
	}

	tester := &MockDatabaseTester{}

	// Mock setup queries
	tester.On("Execute", "test_conn", "CREATE TABLE test (id INT)", []interface{}(nil)).Return(&DatabaseResult{RowsAffected: 0}, nil)

	// Mock main query
	tester.On("Execute", "test_conn", "SELECT 1", []interface{}(nil)).Return(&DatabaseResult{
		RowsAffected: 1,
		Rows:         []map[string]interface{}{{"col1": 1}},
	}, nil)

	// Mock teardown queries
	tester.On("Execute", "test_conn", "DROP TABLE test", []interface{}(nil)).Return(&DatabaseResult{RowsAffected: 0}, nil)

	dbTest := NewDatabaseTest(testCase, tester)
	result := dbTest.Execute()

	assert.Equal(t, "successful_test", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)
	assert.NotZero(t, result.Duration)
	assert.NotEmpty(t, result.Logs)

	tester.AssertExpectations(t)
}

func TestDatabaseTestImpl_Execute_SetupFailure(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "setup_failure_test",
		Connection: "test_conn",
		Query:      "SELECT 1",
		Setup:      []string{"INVALID SQL"},
	}

	tester := &MockDatabaseTester{}
	tester.On("Execute", "test_conn", "INVALID SQL", []interface{}(nil)).Return((*DatabaseResult)(nil), NewGowrightError(DatabaseError, "syntax error", nil))

	dbTest := NewDatabaseTest(testCase, tester)
	result := dbTest.Execute()

	assert.Equal(t, "setup_failure_test", result.Name)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "setup query 1 failed")

	tester.AssertExpectations(t)
}

func TestDatabaseTestImpl_Execute_MainQueryFailure(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "main_query_failure_test",
		Connection: "test_conn",
		Query:      "INVALID SQL",
	}

	tester := &MockDatabaseTester{}
	tester.On("Execute", "test_conn", "INVALID SQL", []interface{}(nil)).Return((*DatabaseResult)(nil), NewGowrightError(DatabaseError, "syntax error", nil))

	dbTest := NewDatabaseTest(testCase, tester)
	result := dbTest.Execute()

	assert.Equal(t, "main_query_failure_test", result.Name)
	assert.Equal(t, TestStatusError, result.Status)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "main query execution failed")

	tester.AssertExpectations(t)
}

func TestDatabaseTestImpl_Execute_ValidationFailure(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "validation_failure_test",
		Connection: "test_conn",
		Query:      "SELECT 1",
		Expected: &DatabaseExpectation{
			RowsAffected: 2, // Wrong expectation
		},
	}

	tester := &MockDatabaseTester{}
	tester.On("Execute", "test_conn", "SELECT 1", []interface{}(nil)).Return(&DatabaseResult{
		RowsAffected: 1, // Actual result
		Rows:         []map[string]interface{}{{"col1": 1}},
	}, nil)

	dbTest := NewDatabaseTest(testCase, tester)
	result := dbTest.Execute()

	assert.Equal(t, "validation_failure_test", result.Name)
	assert.Equal(t, TestStatusFailed, result.Status)
	assert.Error(t, result.Error)

	tester.AssertExpectations(t)
}

func TestDatabaseTestImpl_Execute_TeardownFailure(t *testing.T) {
	testCase := &DatabaseTest{
		Name:       "teardown_failure_test",
		Connection: "test_conn",
		Query:      "SELECT 1",
		Teardown:   []string{"INVALID TEARDOWN SQL"},
	}

	tester := &MockDatabaseTester{}
	tester.On("Execute", "test_conn", "SELECT 1", []interface{}(nil)).Return(&DatabaseResult{
		RowsAffected: 1,
		Rows:         []map[string]interface{}{{"col1": 1}},
	}, nil)
	tester.On("Execute", "test_conn", "INVALID TEARDOWN SQL", []interface{}(nil)).Return((*DatabaseResult)(nil), NewGowrightError(DatabaseError, "syntax error", nil))

	dbTest := NewDatabaseTest(testCase, tester)
	result := dbTest.Execute()

	// Teardown failures should not fail the test, just log warnings
	assert.Equal(t, "teardown_failure_test", result.Name)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.NoError(t, result.Error)
	assert.Contains(t, result.Logs[len(result.Logs)-1], "Warning: teardown query 1 failed")

	tester.AssertExpectations(t)
}

func TestDatabaseTestBuilder(t *testing.T) {
	builder := NewDatabaseTestBuilder("test_builder")

	expected := &DatabaseExpectation{
		RowsAffected: 1,
		Rows:         []map[string]interface{}{{"col1": "test"}},
	}

	tester := &MockDatabaseTester{}

	test := builder.
		WithConnection("test_conn").
		WithQuery("SELECT 'test'").
		WithSetup("CREATE TABLE test (name VARCHAR(50))").
		WithTeardown("DROP TABLE test").
		WithExpected(expected).
		Build(tester)

	assert.NotNil(t, test)
	assert.Equal(t, "test_builder", test.GetName())

	// Verify the test case was built correctly
	dbTest := test.(*DatabaseTestImpl)
	assert.Equal(t, "test_conn", dbTest.testCase.Connection)
	assert.Equal(t, "SELECT 'test'", dbTest.testCase.Query)
	assert.Len(t, dbTest.testCase.Setup, 1)
	assert.Equal(t, "CREATE TABLE test (name VARCHAR(50))", dbTest.testCase.Setup[0])
	assert.Len(t, dbTest.testCase.Teardown, 1)
	assert.Equal(t, "DROP TABLE test", dbTest.testCase.Teardown[0])
	assert.Equal(t, expected, dbTest.testCase.Expected)
}

func TestDatabaseTestBuilder_WithExpectedRowsAffected(t *testing.T) {
	builder := NewDatabaseTestBuilder("test_rows_affected")
	tester := &MockDatabaseTester{}

	test := builder.
		WithConnection("test_conn").
		WithQuery("INSERT INTO test VALUES (1)").
		WithExpectedRowsAffected(1).
		Build(tester)

	dbTest := test.(*DatabaseTestImpl)
	assert.NotNil(t, dbTest.testCase.Expected)
	assert.Equal(t, int64(1), dbTest.testCase.Expected.RowsAffected)
}

func TestDatabaseTestBuilder_WithExpectedData(t *testing.T) {
	builder := NewDatabaseTestBuilder("test_data")
	tester := &MockDatabaseTester{}

	expectedRows := []map[string]interface{}{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
	}

	test := builder.
		WithConnection("test_conn").
		WithQuery("SELECT name, age FROM users").
		WithExpectedRows(expectedRows).
		Build(tester)

	dbTest := test.(*DatabaseTestImpl)
	assert.NotNil(t, dbTest.testCase.Expected)
	assert.Equal(t, expectedRows, dbTest.testCase.Expected.Rows)
}
