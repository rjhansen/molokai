package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

func watchSensor(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			if fsnotify.Write == event.Op&fsnotify.Write {
				updateTable()
			}

		case err := <-watcher.Errors:
			log.Fatal().Msg(fmt.Sprintf("error from file watcher: %v", err))
			panic(err)
		}
	}
}
