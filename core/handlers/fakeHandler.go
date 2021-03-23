package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/radovskyb/watcher"
)

type fakeHandler struct {
}

func (s *fakeHandler) Handle(ctx context.Context, event watcher.Event, wg *sync.WaitGroup) {
	log.Printf("Event '%s' at path '%s' received at %s", event.Op, event.Path, event.ModTime().String())
}

func NewFakeHandler() *fakeHandler {
	return &fakeHandler{}
}
