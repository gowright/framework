// Package reporting provides test result reporting capabilities
package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
)

// ReportManager manages test result reporting
type ReportManager struct {
	config    *config.ReportConfig
	reporters []core.Reporter
}

// NewReportManager creates a new report manager
func NewReportManager(cfg *config.ReportConfig) *ReportManager {
	rm := &ReportManager{
		config:    cfg,
		reporters: make([]core.Reporter, 0),
	}

	// Handle nil config case
	if cfg == nil {
		return rm
	}

	// Initialize default reporters based on configuration
	for _, format := range cfg.Formats {
		switch format {
		case "json":
			rm.reporters = append(rm.reporters, NewJSONReporter(cfg))
		case "html":
			rm.reporters = append(rm.reporters, NewHTMLReporter(cfg))
		case "xml":
			rm.reporters = append(rm.reporters, NewXMLReporter(cfg))
		case "junit":
			rm.reporters = append(rm.reporters, NewJUnitReporter(cfg))
		}
	}

	return rm
}

// AddReporter adds a custom reporter
func (rm *ReportManager) AddReporter(reporter core.Reporter) {
	rm.reporters = append(rm.reporters, reporter)
}

// GenerateReports generates reports using all configured reporters
func (rm *ReportManager) GenerateReports(results *core.TestResults) error {
	// Handle nil config case
	if rm.config == nil || !rm.config.Enabled {
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(rm.config.OutputDir, 0755); err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to create output directory", err)
	}

	var errors []error
	for _, reporter := range rm.reporters {
		if reporter.IsEnabled() {
			if err := reporter.GenerateReport(results); err != nil {
				errors = append(errors, fmt.Errorf("reporter %s failed: %w", reporter.GetName(), err))
			}
		}
	}

	if len(errors) > 0 {
		return core.NewGowrightError(core.ReportingError, fmt.Sprintf("reporting errors: %v", errors), nil)
	}

	return nil
}

// JSONReporter generates JSON reports
type JSONReporter struct {
	config *config.ReportConfig
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(cfg *config.ReportConfig) *JSONReporter {
	return &JSONReporter{config: cfg}
}

// GenerateReport generates a JSON report
func (jr *JSONReporter) GenerateReport(results *core.TestResults) error {
	filename := filepath.Join(jr.config.OutputDir, fmt.Sprintf("test-results-%s.json",
		time.Now().Format("2006-01-02-15-04-05")))

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to marshal JSON", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to write JSON report", err)
	}

	return nil
}

// GetName returns the reporter name
func (jr *JSONReporter) GetName() string {
	return "JSONReporter"
}

// IsEnabled returns whether the reporter is enabled
func (jr *JSONReporter) IsEnabled() bool {
	return true
}

// HTMLReporter generates HTML reports
type HTMLReporter struct {
	config *config.ReportConfig
}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter(cfg *config.ReportConfig) *HTMLReporter {
	return &HTMLReporter{config: cfg}
}

// GenerateReport generates an HTML report
func (hr *HTMLReporter) GenerateReport(results *core.TestResults) error {
	filename := filepath.Join(hr.config.OutputDir, fmt.Sprintf("test-results-%s.html",
		time.Now().Format("2006-01-02-15-04-05")))

	html := hr.generateHTML(results)

	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to write HTML report", err)
	}

	return nil
}

// GetName returns the reporter name
func (hr *HTMLReporter) GetName() string {
	return "HTMLReporter"
}

// IsEnabled returns whether the reporter is enabled
func (hr *HTMLReporter) IsEnabled() bool {
	return true
}

// generateHTML generates HTML content for the report
func (hr *HTMLReporter) generateHTML(results *core.TestResults) string {
	// This is a simplified HTML template
	// In a real implementation, you'd use a proper template engine
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Test Results - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .summary { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .passed { color: green; }
        .failed { color: red; }
        .error { color: orange; }
        .skipped { color: gray; }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Test Results: %s</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Tests: %d</p>
        <p class="passed">Passed: %d</p>
        <p class="failed">Failed: %d</p>
        <p class="error">Errors: %d</p>
        <p class="skipped">Skipped: %d</p>
        <p>Duration: %v</p>
    </div>
    <h2>Test Cases</h2>
    <table>
        <tr>
            <th>Name</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Error</th>
        </tr>
`, results.SuiteName, results.SuiteName, results.TotalTests, results.PassedTests,
		results.FailedTests, results.ErrorTests, results.SkippedTests,
		results.EndTime.Sub(results.StartTime))

	for _, testCase := range results.TestCases {
		errorMsg := ""
		if testCase.Error != nil {
			errorMsg = testCase.Error.Error()
		}
		html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td class="%s">%s</td>
            <td>%v</td>
            <td>%s</td>
        </tr>
`, testCase.Name, testCase.Status.String(), testCase.Status.String(), testCase.Duration, errorMsg)
	}

	html += `
    </table>
</body>
</html>`

	return html
}

// XMLReporter generates XML reports
type XMLReporter struct {
	config *config.ReportConfig
}

// NewXMLReporter creates a new XML reporter
func NewXMLReporter(cfg *config.ReportConfig) *XMLReporter {
	return &XMLReporter{config: cfg}
}

// GenerateReport generates an XML report
func (xr *XMLReporter) GenerateReport(results *core.TestResults) error {
	filename := filepath.Join(xr.config.OutputDir, fmt.Sprintf("test-results-%s.xml",
		time.Now().Format("2006-01-02-15-04-05")))

	xml := xr.generateXML(results)

	if err := os.WriteFile(filename, []byte(xml), 0644); err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to write XML report", err)
	}

	return nil
}

// GetName returns the reporter name
func (xr *XMLReporter) GetName() string {
	return "XMLReporter"
}

// IsEnabled returns whether the reporter is enabled
func (xr *XMLReporter) IsEnabled() bool {
	return true
}

// generateXML generates XML content for the report
func (xr *XMLReporter) generateXML(results *core.TestResults) string {
	// Simplified XML generation
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
    <testsuite name="%s" tests="%d" failures="%d" errors="%d" skipped="%d" time="%.3f">
`, results.SuiteName, results.TotalTests, results.FailedTests, results.ErrorTests,
		results.SkippedTests, results.EndTime.Sub(results.StartTime).Seconds())

	for _, testCase := range results.TestCases {
		xml += fmt.Sprintf(`        <testcase name="%s" time="%.3f"`,
			testCase.Name, testCase.Duration.Seconds())

		switch testCase.Status {
		case core.TestStatusFailed:
			xml += fmt.Sprintf(`>
            <failure message="%s"></failure>
        </testcase>
`, testCase.Error.Error())
		case core.TestStatusError:
			xml += fmt.Sprintf(`>
            <error message="%s"></error>
        </testcase>
`, testCase.Error.Error())
		case core.TestStatusSkipped:
			xml += `>
            <skipped></skipped>
        </testcase>
`
		default:
			xml += ` />
`
		}
	}

	xml += `    </testsuite>
</testsuites>`

	return xml
}

// JUnitReporter generates JUnit XML reports
type JUnitReporter struct {
	config *config.ReportConfig
}

// NewJUnitReporter creates a new JUnit reporter
func NewJUnitReporter(cfg *config.ReportConfig) *JUnitReporter {
	return &JUnitReporter{config: cfg}
}

// GenerateReport generates a JUnit XML report
func (jr *JUnitReporter) GenerateReport(results *core.TestResults) error {
	filename := filepath.Join(jr.config.OutputDir, fmt.Sprintf("junit-results-%s.xml",
		time.Now().Format("2006-01-02-15-04-05")))

	xml := jr.generateJUnitXML(results)

	if err := os.WriteFile(filename, []byte(xml), 0644); err != nil {
		return core.NewGowrightError(core.ReportingError, "failed to write JUnit report", err)
	}

	return nil
}

// GetName returns the reporter name
func (jr *JUnitReporter) GetName() string {
	return "JUnitReporter"
}

// IsEnabled returns whether the reporter is enabled
func (jr *JUnitReporter) IsEnabled() bool {
	return true
}

// generateJUnitXML generates JUnit XML content
func (jr *JUnitReporter) generateJUnitXML(results *core.TestResults) string {
	// JUnit XML format
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="%s" tests="%d" failures="%d" errors="%d" skipped="%d" time="%.3f" timestamp="%s">
`, results.SuiteName, results.TotalTests, results.FailedTests, results.ErrorTests,
		results.SkippedTests, results.EndTime.Sub(results.StartTime).Seconds(),
		results.StartTime.Format(time.RFC3339))

	for _, testCase := range results.TestCases {
		xml += fmt.Sprintf(`    <testcase classname="%s" name="%s" time="%.3f"`,
			results.SuiteName, testCase.Name, testCase.Duration.Seconds())

		switch testCase.Status {
		case core.TestStatusFailed:
			xml += fmt.Sprintf(`>
        <failure message="%s" type="AssertionError">%s</failure>
    </testcase>
`, testCase.Error.Error(), testCase.Error.Error())
		case core.TestStatusError:
			xml += fmt.Sprintf(`>
        <error message="%s" type="Error">%s</error>
    </testcase>
`, testCase.Error.Error(), testCase.Error.Error())
		case core.TestStatusSkipped:
			xml += `>
        <skipped />
    </testcase>
`
		default:
			xml += ` />
`
		}
	}

	xml += `</testsuite>`

	return xml
}
