package watch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// FileList map[path]isWatched
type FileList map[string]bool

type DirectoryWatcher struct {
	ctx      context.Context
	notify   *fsnotify.Watcher
	fileList map[string]bool

	Recursive bool

	OnCreate EventFunc
	OnWrite  EventFunc
	OnRemove EventFunc
	OnRename EventFunc
	OnChmod  EventFunc

	OnError ErrorFunc
}

func (w *DirectoryWatcher) Go() {
	defer w.notify.Close()

	for path, watching := range w.fileList {
		if watching {
			continue
		}

		if err := filepath.Walk(path, w.WatchFile); err != nil {
			if nil != w.OnError {
				w.OnError(w.ctx, err)
			}

			return
		}

		w.fileList[path] = true
	}

	for {
		select {
		case <-w.ctx.Done():
			fmt.Println("ctx done")
			return

		case err := <-w.notify.Errors:
			if nil != w.OnError {
				w.OnError(w.ctx, err)
			}

		case event := <-w.notify.Events:
			handleEvent(w, event)
		}
	}
}

func (w *DirectoryWatcher) WatchFile(path string, stat os.FileInfo, err error) error {
	if nil != err {
		return err
	}

	if !stat.Mode().IsDir() {
		return nil
	}

	return w.notify.Add(path)
}

// NewDefaultDirecoryWatcher
//
// Setup a new directory watcher with the default options, no event listeners
func NewDefaultDirecoryWatcher(ctx context.Context, path string) (*DirectoryWatcher, error) {
	notify, err := fsnotify.NewWatcher()
	if nil != err {
		return nil, err
	}

	stat, err := os.Stat(path)
	if nil != err {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, errors.New("path is not a directory")
	}

	return &DirectoryWatcher{
		ctx:    ctx,
		notify: notify,
		fileList: FileList{
			path: false,
		},
	}, nil
}

// NewDirectoryWatcher
//
// Setup the directory watcher using a predefined watcher object
//
// This is pretty much just an alias to allow for setting up the watcher in a more declarative way
func NewDirectoryWatcher(ctx context.Context, path string, watcher *DirectoryWatcher) (*DirectoryWatcher, error) {
	notify, err := fsnotify.NewWatcher()
	if nil != err {
		return nil, err
	}

	stat, err := os.Stat(path)
	if nil != err {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, errors.New("path is not a directory")
	}

	watcher.ctx = ctx
	watcher.notify = notify
	watcher.fileList = FileList{
		path: false,
	}

	return watcher, nil
}

// NewMultiDirectoryWatcher
//
// Setup the directory watcher for multiple paths
//
// This also uses a predefined object
func NewMultiDirectoryWatcher(
	ctx context.Context,
	paths []string,
	watcher *DirectoryWatcher,
) (
	*DirectoryWatcher,
	error,
) {
	notify, err := fsnotify.NewWatcher()
	if nil != err {
		return nil, err
	}

	watcher.fileList = make(FileList)

	for _, path := range paths {
		stat, err := os.Stat(path)
		if nil != err {
			return nil, err
		}

		if !stat.IsDir() {
			return nil, errors.New("path is not a directory")
		}

		watcher.fileList[path] = false
	}

	watcher.ctx = ctx
	watcher.notify = notify

	return watcher, nil
}
