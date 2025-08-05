//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	fmt.Println("=== Gowright Database Testing Example ===\n")

	// Create database configuration
	config := &gowright.DatabaseConfig{
		Connections: map[string]*gowright.DBConnection{
			"primary": {
				Driver:       "sqlite3",
				DSN:          ":memory:",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			"secondary": {
				Driver:       "sqlite3",
				DSN:          "./test.db",
				MaxOpenConns: 5,
				MaxIdleConns: 2,
			},
		},
	}

	// Create and initialize database tester
	tester := gowright.NewDatabaseTester(config)
	if err := tester.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize database tester: %v", err)
	}
	defer tester.Cleanup()

	// Example 1: Basic table creation and data insertion
	fmt.Println("1. Testing table creation and data insertion")
	setupTest := gowright.NewDatabaseTest("Database Setup Test", "primary")

	// Setup schema
	setupTest.AddSetupQuery(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			active BOOLEAN DEFAULT 1
		)
	`)

	setupTest.AddSetupQuery(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			product_name VARCHAR(100),
			quantity INTEGER,
			price DECIMAL(10,2),
			order_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)

	// Insert test data
	setupTest.AddSetupQuery(`
		INSERT INTO users (username, email) VALUES 
		('john_doe', 'john@example.com'),
		('jane_smith', 'jane@example.com'),
		('bob_wilson', 'bob@example.com')
	`)

	setupTest.AddSetupQuery(`
		INSERT INTO orders (user_id, product_name, quantity, price) VALUES 
		(1, 'Laptop', 1, 999.99),
		(1, 'Mouse', 2, 25.50),
		(2, 'Keyboard', 1, 75.00),
		(3, 'Monitor', 1, 299.99)
	`)

	// Test query
	setupTest.SetQuery("SELECT COUNT(*) as user_count FROM users")
	setupTest.SetExpectedRowCount(1)
	setupTest.SetExpectedColumnValue("user_count", 3)

	result := setupTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 2: Transaction testing with rollback
	fmt.Println("\n2. Testing transaction management and rollback")
	transactionTest := gowright.NewDatabaseTest("Transaction Test", "primary")

	// Start transaction and insert data
	transactionTest.SetQuery(`
		BEGIN TRANSACTION;
		INSERT INTO users (username, email) VALUES ('test_user', 'test@example.com');
		INSERT INTO orders (user_id, product_name, quantity, price) VALUES (4, 'Test Product', 1, 50.00);
		ROLLBACK;
	`)

	// Verify rollback worked - user count should still be 3
	transactionTest.AddTeardownQuery("SELECT COUNT(*) as user_count FROM users")
	transactionTest.SetExpectedColumnValue("user_count", 3)

	result = transactionTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 3: Complex JOIN queries and data validation
	fmt.Println("\n3. Testing complex JOIN queries and data validation")
	joinTest := gowright.NewDatabaseTest("JOIN Query Test", "primary")

	joinTest.SetQuery(`
		SELECT 
			u.username,
			u.email,
			COUNT(o.id) as order_count,
			SUM(o.price * o.quantity) as total_spent
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		GROUP BY u.id, u.username, u.email
		ORDER BY total_spent DESC
	`)

	joinTest.SetExpectedRowCount(3)

	// Validate specific user data
	joinTest.AddCustomAssertion(func(rows []map[string]interface{}) error {
		// Find john_doe's record
		for _, row := range rows {
			if row["username"] == "john_doe" {
				orderCount := row["order_count"].(int64)
				totalSpent := row["total_spent"].(float64)

				if orderCount != 2 {
					return fmt.Errorf("expected john_doe to have 2 orders, got %d", orderCount)
				}

				expectedTotal := 999.99 + (25.50 * 2) // Laptop + 2 mice
				if totalSpent != expectedTotal {
					return fmt.Errorf("expected john_doe total spent to be %.2f, got %.2f", expectedTotal, totalSpent)
				}

				return nil
			}
		}
		return fmt.Errorf("john_doe not found in results")
	})

	result = joinTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 4: Data integrity and constraint testing
	fmt.Println("\n4. Testing data integrity and constraints")
	constraintTest := gowright.NewDatabaseTest("Constraint Test", "primary")

	// Try to insert duplicate username (should fail)
	constraintTest.SetQuery("INSERT INTO users (username, email) VALUES ('john_doe', 'duplicate@example.com')")
	constraintTest.SetExpectedError(true) // We expect this to fail due to UNIQUE constraint

	result = constraintTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 5: Performance testing with large dataset
	fmt.Println("\n5. Testing database performance with larger dataset")
	performanceTest := gowright.NewDatabaseTest("Performance Test", "primary")

	// Insert many records for performance testing
	performanceTest.AddSetupQuery(`
		CREATE TABLE performance_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data VARCHAR(100),
			number_field INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// Insert 1000 records
	for i := 0; i < 1000; i++ {
		performanceTest.AddSetupQuery(fmt.Sprintf(
			"INSERT INTO performance_test (data, number_field) VALUES ('test_data_%d', %d)",
			i, i*10,
		))
	}

	// Test query performance
	startTime := time.Now()
	performanceTest.SetQuery(`
		SELECT COUNT(*) as total_count, 
			   AVG(number_field) as avg_number,
			   MAX(number_field) as max_number,
			   MIN(number_field) as min_number
		FROM performance_test 
		WHERE number_field > 5000
	`)

	performanceTest.SetExpectedRowCount(1)
	performanceTest.AddCustomAssertion(func(rows []map[string]interface{}) error {
		duration := time.Since(startTime)
		if duration > 5*time.Second {
			return fmt.Errorf("query took too long: %v", duration)
		}

		row := rows[0]
		totalCount := row["total_count"].(int64)
		if totalCount != 500 { // Records with number_field > 5000
			return fmt.Errorf("expected 500 records, got %d", totalCount)
		}

		return nil
	})

	result = performanceTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 6: Multi-database testing
	fmt.Println("\n6. Testing multi-database operations")
	multiDbTest := gowright.NewDatabaseTest("Multi-Database Test", "secondary")

	// Setup secondary database
	multiDbTest.AddSetupQuery(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action VARCHAR(50),
			table_name VARCHAR(50),
			record_id INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	multiDbTest.AddSetupQuery(`
		INSERT INTO audit_log (action, table_name, record_id) VALUES 
		('INSERT', 'users', 1),
		('INSERT', 'users', 2),
		('UPDATE', 'users', 1),
		('DELETE', 'orders', 5)
	`)

	multiDbTest.SetQuery("SELECT COUNT(*) as log_count FROM audit_log WHERE action = 'INSERT'")
	multiDbTest.SetExpectedColumnValue("log_count", 2)

	result = multiDbTest.Execute(tester)
	printDatabaseTestResult(result)

	// Example 7: Database migration testing
	fmt.Println("\n7. Testing database schema migrations")
	migrationTest := gowright.NewDatabaseTest("Migration Test", "primary")

	// Add new column to existing table
	migrationTest.AddSetupQuery("ALTER TABLE users ADD COLUMN last_login DATETIME")
	migrationTest.AddSetupQuery("UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = 1")

	// Test that migration worked
	migrationTest.SetQuery(`
		SELECT username, last_login 
		FROM users 
		WHERE last_login IS NOT NULL
	`)

	migrationTest.SetExpectedRowCount(1)
	migrationTest.AddCustomAssertion(func(rows []map[string]interface{}) error {
		if len(rows) == 0 {
			return fmt.Errorf("no rows returned")
		}

		row := rows[0]
		if row["username"] != "john_doe" {
			return fmt.Errorf("expected john_doe, got %v", row["username"])
		}

		if row["last_login"] == nil {
			return fmt.Errorf("last_login should not be null")
		}

		return nil
	})

	result = migrationTest.Execute(tester)
	printDatabaseTestResult(result)

	// Generate comprehensive database test report
	fmt.Println("\nGenerating database test reports...")

	testResults := &gowright.TestResults{
		SuiteName:    "Database Testing Example Suite",
		StartTime:    time.Now().Add(-5 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   7,
		PassedTests:  6,
		FailedTests:  0,
		SkippedTests: 0,
		ErrorTests:   1,                           // The constraint test that expected an error
		TestCases:    []gowright.TestCaseResult{}, // Would contain all results
	}

	reportConfig := &gowright.ReportConfig{
		LocalReports: gowright.LocalReportConfig{
			JSON:      true,
			HTML:      true,
			OutputDir: "./database-test-reports",
		},
	}

	reportManager := gowright.NewReportManager(reportConfig)
	if err := reportManager.GenerateReports(testResults); err != nil {
		log.Printf("Failed to generate reports: %v", err)
	} else {
		fmt.Printf("Database test reports generated in: %s\n", config.LocalReports.OutputDir)
	}

	fmt.Println("\n=== Database Testing Complete ===")
}

func printDatabaseTestResult(result *gowright.TestCaseResult) {
	fmt.Printf("Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status.String())
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	if len(result.Logs) > 0 {
		fmt.Println("Logs:")
		for _, logEntry := range result.Logs {
			fmt.Printf("  - %s\n", logEntry)
		}
	}

	fmt.Println("---")
}
