package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/indeedhat/watch"
)

func main() {
	defer func() {
		err := recover()
		if nil != err {
			log.Printf("recover: %s", err)
		}
	}()

	filePath, err := parseInput()
	if nil != err {
		log.Fatalf("Failed to start: %s", err)
	}

	log.Printf("path: %s", filePath)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	watcher, err := watch.NewWatcher(ctx, filePath)
	if nil != err {
		panic(err)
	}

	watcher.OnCreate = func(ctx context.Context, event fsnotify.Event) {
		log.Printf("%s: %s\n", event.Name, event.Op.String())
	}

	watcher.Go()

	<-ctx.Done()
}

func parseInput() (string, error) {
	log.Printf("%s\n", os.Args)
	if 1 == len(os.Args) {
		return "", errors.New("Not enough args")
	}

	return os.Args[1], nil
}
