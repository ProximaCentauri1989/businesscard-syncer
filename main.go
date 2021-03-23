package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ProximaCentauri1989/businesscard-syncer/config"
	"github.com/ProximaCentauri1989/businesscard-syncer/core/handlers"
	"github.com/ProximaCentauri1989/businesscard-syncer/core/watcher"
)

var fatalErr error = nil

func setFatal(err error) {
	log.Printf("An error occured during startup: %s", err)
	flag.PrintDefaults()
	fatalErr = err
}

func main() {
	// error catcher
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	// config
	cfg := config.ReadConfig()

	// event catcher
	watcher, err := watcher.NewWatcher(cfg.GetRoot())
	if err != nil {
		setFatal(err)
		return
	}

	// event handlers
	s3syncer := handlers.NewS3Syncer(cfg.GetRegion(), cfg.GetBucketName(), cfg.GetRoot())
	fakeHandler := handlers.NewFakeHandler()
	watcher.Add("s3syncer", s3syncer)
	watcher.Add("stub", fakeHandler)

	err = watcher.Start(time.Second * 1)
	if err != nil {
		setFatal(err)
		return
	}

	// gracefull stop
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	watcher.Stop()
}
