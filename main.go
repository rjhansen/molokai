/*
   Copyright 2021, Rob Hansen.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

func init() {
	initLogging()
	loadConfig()
	initDatabase()
	sensorQueue = make(map[string][]Reading)
}

func main() {
	var watcher *fsnotify.Watcher
	var err error
	done := make(chan bool)

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		log.Fatal().Msg("could not initialize file watcher!")
		panic("could not initialize file watcher")
	}
	defer func() { _ = watcher.Close() }()
	if err = watcher.Add("/tmp/kure.json"); err != nil {
		log.Fatal().Msgf("could not add file to watcher: %v", err)
	}

	go watchSensor(watcher)

	<-done
}
