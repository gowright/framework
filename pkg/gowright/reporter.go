package gowright

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"sync"
	"time"
)

// ReportManager coordinates all reporting activities
type ReportManager struct {
	config    *ReportConfig
	reporters []Reporter
	mutex     sync.RWMutex
}

// NewReportManager creates a new report manager with the given configuration
func NewReportManager(config *ReportConfig) *ReportManager {
	rm := &ReportManager{
		config:    config,
		reporters: make([]Reporter, 0),
	}
	
	// Initialize reporters based on configuration
	rm.initializeReporters()
	
	return rm
}

// initializeReporters initializes reporters based on configuration
func (rm *ReportManager) initializeReporters() {
	if rm.config == nil {
		return
	}

	// Add local reporters
	if rm.config.LocalReports.JSON {
		rm.AddReporter(&JSONReporter{
			OutputDir: rm.config.LocalReports.OutputDir,
			enabled:   true,
		})
	}

	if rm.config.LocalReports.HTML {
		rm.AddReporter(&HTMLReporter{
			OutputDir: rm.config.LocalReports.OutputDir,
			enabled:   true,
		})
	}

	// Add remote reporters (will be implemented in later tasks)
	// These are placeholder implementations
	if rm.config.RemoteReports.JiraXray != nil {
		rm.AddReporter(&JiraXrayReporter{
			config:  rm.config.RemoteReports.JiraXray,
			enabled: true,
		})
	}

	if rm.config.RemoteReports.AIOTest != nil {
		rm.AddReporter(&AIOTestReporter{
			config:  rm.config.RemoteReports.AIOTest,
			enabled: true,
		})
	}

	if rm.config.RemoteReports.ReportPortal != nil {
		rm.AddReporter(&ReportPortalReporter{
			config:  rm.config.RemoteReports.ReportPortal,
			enabled: true,
		})
	}
}

// AddReporter adds a reporter to the manager
func (rm *ReportManager) AddReporter(reporter Reporter) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.reporters = append(rm.reporters, reporter)
}

// RemoveReporter removes a reporter from the manager
func (rm *ReportManager) RemoveReporter(name string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	for i, reporter := range rm.reporters {
		if reporter.GetName() == name {
			rm.reporters = append(rm.reporters[:i], rm.reporters[i+1:]...)
			break
		}
	}
}

// GenerateReports generates reports using all enabled reporters
func (rm *ReportManager) GenerateReports(results *TestResults) error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var errors []error
	
	for _, reporter := range rm.reporters {
		if reporter.IsEnabled() {
			if err := reporter.GenerateReport(results); err != nil {
				errors = append(errors, fmt.Errorf("reporter %s failed: %w", reporter.GetName(), err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("reporting errors occurred: %v", errors)
	}

	return nil
}

// GetReporters returns a copy of all reporters
func (rm *ReportManager) GetReporters() []Reporter {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	reporters := make([]Reporter, len(rm.reporters))
	copy(reporters, rm.reporters)
	return reporters
}

// JSONReporter generates JSON reports locally
type JSONReporter struct {
	OutputDir string
	enabled   bool
}

// GenerateReport generates a JSON report
func (jr *JSONReporter) GenerateReport(results *TestResults) error {
	if err := jr.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := jr.generateFilename(results.SuiteName)
	filepath := fmt.Sprintf("%s/%s", jr.OutputDir, filename)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test results to JSON: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report to file: %w", err)
	}

	return nil
}

// GetName returns the reporter name
func (jr *JSONReporter) GetName() string {
	return "json"
}

// IsEnabled returns whether the reporter is enabled
func (jr *JSONReporter) IsEnabled() bool {
	return jr.enabled
}

// ensureOutputDir creates the output directory if it doesn't exist
func (jr *JSONReporter) ensureOutputDir() error {
	if jr.OutputDir == "" {
		jr.OutputDir = "./reports"
	}
	return os.MkdirAll(jr.OutputDir, 0755)
}

// generateFilename generates a filename for the JSON report
func (jr *JSONReporter) generateFilename(suiteName string) string {
	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(strings.ToLower(suiteName), " ", "_")
	return fmt.Sprintf("%s_%s.json", safeName, timestamp)
}

// HTMLReporter generates HTML reports locally
type HTMLReporter struct {
	OutputDir string
	enabled   bool
}

// GenerateReport generates an HTML report
func (hr *HTMLReporter) GenerateReport(results *TestResults) error {
	if err := hr.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := hr.generateFilename(results.SuiteName)
	filepath := fmt.Sprintf("%s/%s", hr.OutputDir, filename)

	htmlContent, err := hr.generateHTMLContent(results)
	if err != nil {
		return fmt.Errorf("failed to generate HTML content: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report to file: %w", err)
	}

	return nil
}

// GetName returns the reporter name
func (hr *HTMLReporter) GetName() string {
	return "html"
}

// IsEnabled returns whether the reporter is enabled
func (hr *HTMLReporter) IsEnabled() bool {
	return hr.enabled
}

// ensureOutputDir creates the output directory if it doesn't exist
func (hr *HTMLReporter) ensureOutputDir() error {
	if hr.OutputDir == "" {
		hr.OutputDir = "./reports"
	}
	return os.MkdirAll(hr.OutputDir, 0755)
}

// generateFilename generates a filename for the HTML report
func (hr *HTMLReporter) generateFilename(suiteName string) string {
	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(strings.ToLower(suiteName), " ", "_")
	return fmt.Sprintf("%s_%s.html", safeName, timestamp)
}

// generateHTMLContent generates the HTML content for the report
func (hr *HTMLReporter) generateHTMLContent(results *TestResults) (string, error) {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"formatDuration": func(d time.Duration) string {
			return d.String()
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"statusClass": func(status TestStatus) string {
			switch status {
			case TestStatusPassed:
				return "passed"
			case TestStatusFailed:
				return "failed"
			case TestStatusSkipped:
				return "skipped"
			case TestStatusError:
				return "error"
			default:
				return "unknown"
			}
		},
		"percentage": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
		"sub": func(a, b time.Time) time.Duration {
			return a.Sub(b)
		},
	}).Parse(htmlTemplate)
	
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, results); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}

// JiraXrayReporter sends reports to Jira Xray
type JiraXrayReporter struct {
	config  *JiraXrayConfig
	enabled bool
}

// GenerateReport sends a report to Jira Xray
func (jxr *JiraXrayReporter) GenerateReport(results *TestResults) error {
	// Implementation will be added in later tasks
	return nil
}

// GetName returns the reporter name
func (jxr *JiraXrayReporter) GetName() string {
	return "jira_xray"
}

// IsEnabled returns whether the reporter is enabled
func (jxr *JiraXrayReporter) IsEnabled() bool {
	return jxr.enabled
}

// AIOTestReporter sends reports to AIOTest
type AIOTestReporter struct {
	config  *AIOTestConfig
	enabled bool
}

// GenerateReport sends a report to AIOTest
func (atr *AIOTestReporter) GenerateReport(results *TestResults) error {
	// Implementation will be added in later tasks
	return nil
}

// GetName returns the reporter name
func (atr *AIOTestReporter) GetName() string {
	return "aio_test"
}

// IsEnabled returns whether the reporter is enabled
func (atr *AIOTestReporter) IsEnabled() bool {
	return atr.enabled
}

// ReportPortalReporter sends reports to Report Portal
type ReportPortalReporter struct {
	config  *ReportPortalConfig
	enabled bool
}

// GenerateReport sends a report to Report Portal
func (rpr *ReportPortalReporter) GenerateReport(results *TestResults) error {
	// Implementation will be added in later tasks
	return nil
}

// GetName returns the reporter name
func (rpr *ReportPortalReporter) GetName() string {
	return "report_portal"
}

// IsEnabled returns whether the reporter is enabled
func (rpr *ReportPortalReporter) IsEnabled() bool {
	return rpr.enabled
}

// htmlTemplate is the HTML template for generating reports
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gowright Test Report - {{.SuiteName}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f5f5;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        
        .header .meta {
            opacity: 0.9;
            font-size: 1.1em;
        }
        
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .summary-card {
            background: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            text-align: center;
            border-left: 4px solid #ddd;
        }
        
        .summary-card.total { border-left-color: #3498db; }
        .summary-card.passed { border-left-color: #2ecc71; }
        .summary-card.failed { border-left-color: #e74c3c; }
        .summary-card.skipped { border-left-color: #f39c12; }
        .summary-card.error { border-left-color: #9b59b6; }
        
        .summary-card .number {
            font-size: 2.5em;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .summary-card .label {
            color: #666;
            text-transform: uppercase;
            font-size: 0.9em;
            letter-spacing: 1px;
        }
        
        .summary-card .percentage {
            font-size: 0.9em;
            color: #888;
            margin-top: 5px;
        }
        
        .test-results {
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }
        
        .test-results h2 {
            background: #f8f9fa;
            padding: 20px;
            margin: 0;
            border-bottom: 1px solid #dee2e6;
        }
        
        .test-case {
            border-bottom: 1px solid #eee;
            padding: 20px;
            transition: background-color 0.2s;
        }
        
        .test-case:hover {
            background-color: #f8f9fa;
        }
        
        .test-case:last-child {
            border-bottom: none;
        }
        
        .test-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        
        .test-name {
            font-size: 1.2em;
            font-weight: 600;
        }
        
        .test-status {
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .test-status.passed {
            background-color: #d4edda;
            color: #155724;
        }
        
        .test-status.failed {
            background-color: #f8d7da;
            color: #721c24;
        }
        
        .test-status.skipped {
            background-color: #fff3cd;
            color: #856404;
        }
        
        .test-status.error {
            background-color: #e2e3e5;
            color: #383d41;
        }
        
        .test-meta {
            display: flex;
            gap: 20px;
            color: #666;
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        
        .test-error {
            background-color: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 5px;
            padding: 15px;
            margin-top: 10px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #721c24;
        }
        
        .test-logs {
            margin-top: 15px;
        }
        
        .test-logs h4 {
            margin-bottom: 10px;
            color: #495057;
        }
        
        .log-entry {
            background-color: #f8f9fa;
            border-left: 3px solid #007bff;
            padding: 10px 15px;
            margin-bottom: 5px;
            font-family: 'Courier New', monospace;
            font-size: 0.85em;
        }
        
        .screenshots {
            margin-top: 15px;
        }
        
        .screenshots h4 {
            margin-bottom: 10px;
            color: #495057;
        }
        
        .screenshot-list {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        
        .screenshot-link {
            color: #007bff;
            text-decoration: none;
            padding: 5px 10px;
            background-color: #e9ecef;
            border-radius: 5px;
            font-size: 0.9em;
        }
        
        .screenshot-link:hover {
            background-color: #dee2e6;
        }
        
        .assertion-steps {
            margin-top: 15px;
        }
        
        .assertion-steps h4 {
            margin-bottom: 10px;
            color: #495057;
        }
        
        .assertion-step {
            background-color: #f8f9fa;
            border-left: 3px solid #dee2e6;
            padding: 12px 15px;
            margin-bottom: 8px;
            border-radius: 0 5px 5px 0;
        }
        
        .assertion-step.passed {
            border-left-color: #28a745;
            background-color: #d4edda;
        }
        
        .assertion-step.failed {
            border-left-color: #dc3545;
            background-color: #f8d7da;
        }
        
        .step-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 5px;
        }
        
        .step-name {
            font-weight: 600;
            color: #495057;
        }
        
        .step-status {
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.75em;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .step-status.passed {
            background-color: #28a745;
            color: white;
        }
        
        .step-status.failed {
            background-color: #dc3545;
            color: white;
        }
        
        .step-duration {
            font-size: 0.85em;
            color: #6c757d;
        }
        
        .step-description {
            color: #6c757d;
            font-size: 0.9em;
            margin-bottom: 8px;
        }
        
        .step-values {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
            margin-bottom: 8px;
        }
        
        .expected-value, .actual-value {
            font-family: 'Courier New', monospace;
            font-size: 0.85em;
            padding: 5px 8px;
            border-radius: 3px;
            background-color: #e9ecef;
        }
        
        .expected-value {
            border-left: 3px solid #28a745;
        }
        
        .actual-value {
            border-left: 3px solid #17a2b8;
        }
        
        .step-error {
            font-family: 'Courier New', monospace;
            font-size: 0.85em;
            color: #721c24;
            background-color: #f5c6cb;
            padding: 8px;
            border-radius: 3px;
            border-left: 3px solid #dc3545;
        }
        
        .footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            color: #666;
            font-size: 0.9em;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 10px;
            }
            
            .header h1 {
                font-size: 2em;
            }
            
            .summary {
                grid-template-columns: repeat(2, 1fr);
            }
            
            .test-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 10px;
            }
            
            .test-meta {
                flex-direction: column;
                gap: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.SuiteName}}</h1>
            <div class="meta">
                <div>Started: {{formatTime .StartTime}}</div>
                <div>Completed: {{formatTime .EndTime}}</div>
                <div>Duration: {{formatDuration (sub .EndTime .StartTime)}}</div>
            </div>
        </div>
        
        <div class="summary">
            <div class="summary-card total">
                <div class="number">{{.TotalTests}}</div>
                <div class="label">Total Tests</div>
            </div>
            <div class="summary-card passed">
                <div class="number">{{.PassedTests}}</div>
                <div class="label">Passed</div>
                <div class="percentage">{{printf "%.1f%%" (percentage .PassedTests .TotalTests)}}</div>
            </div>
            <div class="summary-card failed">
                <div class="number">{{.FailedTests}}</div>
                <div class="label">Failed</div>
                <div class="percentage">{{printf "%.1f%%" (percentage .FailedTests .TotalTests)}}</div>
            </div>
            <div class="summary-card skipped">
                <div class="number">{{.SkippedTests}}</div>
                <div class="label">Skipped</div>
                <div class="percentage">{{printf "%.1f%%" (percentage .SkippedTests .TotalTests)}}</div>
            </div>
            {{if gt .ErrorTests 0}}
            <div class="summary-card error">
                <div class="number">{{.ErrorTests}}</div>
                <div class="label">Errors</div>
                <div class="percentage">{{printf "%.1f%%" (percentage .ErrorTests .TotalTests)}}</div>
            </div>
            {{end}}
        </div>
        
        <div class="test-results">
            <h2>Test Results</h2>
            {{range .TestCases}}
            <div class="test-case">
                <div class="test-header">
                    <div class="test-name">{{.Name}}</div>
                    <div class="test-status {{statusClass .Status}}">{{.Status}}</div>
                </div>
                <div class="test-meta">
                    <div>Duration: {{formatDuration .Duration}}</div>
                    <div>Started: {{formatTime .StartTime}}</div>
                    <div>Ended: {{formatTime .EndTime}}</div>
                </div>
                {{if .Error}}
                <div class="test-error">
                    <strong>Error:</strong> {{.Error.Error}}
                </div>
                {{end}}
                {{if .Logs}}
                <div class="test-logs">
                    <h4>Logs</h4>
                    {{range .Logs}}
                    <div class="log-entry">{{.}}</div>
                    {{end}}
                </div>
                {{end}}
                {{if .Screenshots}}
                <div class="screenshots">
                    <h4>Screenshots</h4>
                    <div class="screenshot-list">
                        {{range .Screenshots}}
                        <a href="{{.}}" class="screenshot-link" target="_blank">{{.}}</a>
                        {{end}}
                    </div>
                </div>
                {{end}}
                {{if .Steps}}
                <div class="assertion-steps">
                    <h4>Assertion Steps</h4>
                    {{range .Steps}}
                    <div class="assertion-step {{statusClass .Status}}">
                        <div class="step-header">
                            <span class="step-name">{{.Name}}</span>
                            <span class="step-status {{statusClass .Status}}">{{.Status}}</span>
                            <span class="step-duration">{{formatDuration .Duration}}</span>
                        </div>
                        <div class="step-description">{{.Description}}</div>
                        {{if .Expected}}
                        <div class="step-values">
                            <div class="expected-value">Expected: {{.Expected}}</div>
                            <div class="actual-value">Actual: {{.Actual}}</div>
                        </div>
                        {{end}}
                        {{if .Error}}
                        <div class="step-error">{{.Error.Error}}</div>
                        {{end}}
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <div class="footer">
            Generated by Gowright Testing Framework on {{formatTime .EndTime}}
        </div>
    </div>
</body>
</html>`