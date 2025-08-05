package gowright

import (
	"fmt"
	"runtime"
)

// Version information for the Gowright framework
const (
	// Version is the current version of the Gowright framework
	Version = "1.0.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "unknown"

	// BuildDate is the build date (set during build)
	BuildDate = "unknown"
)

// VersionInfo contains detailed version information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	info := GetVersionInfo()
	if info.GitCommit != "unknown" && info.BuildDate != "unknown" {
		return fmt.Sprintf("Gowright %s (commit: %s, built: %s, %s, %s)",
			info.Version, info.GitCommit, info.BuildDate, info.GoVersion, info.Platform)
	}
	return fmt.Sprintf("Gowright %s (%s, %s)", info.Version, info.GoVersion, info.Platform)
}
