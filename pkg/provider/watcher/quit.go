package watcher

import (
	"context"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

func QuitOnChange(ctx context.Context, paths ...string) error {
	return quitOnChange(ctx, func(event fsnotify.Event) {
		logrus.Fatalf("Detected change in %s, closing to force restart", event.Name)
	}, paths...)
}

func quitOnChange(ctx context.Context, cb func(fsnotify.Event), paths ...string) error {
	if utils.IsMemFs() {
		logrus.Infof("Skipping registering a watch on certificate paths")
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				cb(event)
			}
		}
	}()
	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			return err
		}
		logrus.Infof("Starting to watch %s for restarts", path)
	}
	return nil
}
