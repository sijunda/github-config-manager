// Package version provides build-time version information for GCM.
package version

import (
	"fmt"
	"runtime"
)

// Build-time variables set via ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info holds structured version information.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("gcm %s (%s/%s) built %s commit %s go %s",
		i.Version, i.OS, i.Arch, i.Date, i.Commit, i.GoVersion)
}

// Short returns a short version string.
func (i Info) Short() string {
	return fmt.Sprintf("gcm %s", i.Version)
}
