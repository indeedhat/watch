package watch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type EventFunc func(context.Context, fsnotify.Event)
type ErrorFunc func(context.Context, error)

type Watcher struct {
    // in
	path   string
	notify *fsnotify.Watcher
	ctx    context.Context

	Recursive bool

	OnCreate EventFunc
	OnWrite  EventFunc
	OnRemove EventFunc
	OnRename EventFunc
	OnChmod  EventFunc

	OnError ErrorFunc
}

func NewWatcher(ctx context.Context, path string) (*Watcher, error) {
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

	return &Watcher{
		path:   path,
		ctx:    ctx,
		notify: notify,
	}, nil
}

func (w *Watcher) Go() {
	go func() {
		defer w.notify.Close()

		if err := filepath.Walk(w.path, w.watchDog); nil != err {
			if nil != w.OnError {
				w.OnError(w.ctx, err)
			}

			fmt.Println("failed to walk path")
			return
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
	}()
}

func (w *Watcher) watchDog(path string, stat os.FileInfo, err error) error {
	if nil != err {
		return err
	}

	if !stat.Mode().IsDir() {
		return nil
	}

	return w.notify.Add(path)
}

func handleEvent(w *Watcher, event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Chmod:
		if nil != w.OnChmod {
			w.OnChmod(w.ctx, event)
		}

	case fsnotify.Create:
		watchPath(w, event.Name)
		if nil != w.OnCreate {
			w.OnCreate(w.ctx, event)
		}

	case fsnotify.Remove:
		if nil != w.OnRemove {
			w.OnRemove(w.ctx, event)
		}

	case fsnotify.Rename:
		if nil != w.OnRename {
			w.OnRename(w.ctx, event)
		}

	case fsnotify.Write:
		if nil != w.OnWrite {
			w.OnWrite(w.ctx, event)
		}
	}
}

func watchPath(w *Watcher, path string) error {
	stat, err := os.Stat(path)
	if nil != err {
		return err
	}

	return w.watchDog(path, stat, nil)
}
