package watch

import "os"

// Watchable is the common interface for both directory and file watchers
type Watchable interface {
	Go()
	WatchPath(path string, stat os.FileInfo, err error) error
}
