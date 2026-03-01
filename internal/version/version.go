package version

import (
	"fmt"
	"runtime"
)

// These variables are populated at build time using -ldflags
var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// These are typically static constants
const (
	Author  = "Jane Doe <jane@example.com>"
	License = "Apache-2.0"
)

// Info returns a formatted string containing all version metadata
func Info() string {
	return fmt.Sprintf(
		`Version:    %s
Git Commit: %s
Build Time: %s
Go Version: %s
OS/Arch:    %s/%s
Author:     %s
`,
		Version, GitCommit, BuildTime, runtime.Version(), runtime.GOOS, runtime.GOARCH, Author,
	)
}
