package watcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

type Handler interface {
	Handle(ctx context.Context, event watcher.Event, wg *sync.WaitGroup)
}

type Watcher struct {
	root            string
	watch           *watcher.Watcher
	pollingInterval time.Duration
	eventHandlers   map[string]Handler
	wg              *sync.WaitGroup
	cancelations    []context.CancelFunc
}

// Creates a new syncer and starts event listening. Event polling should be started by caller manually using 'Start' method
func NewWatcher(root string) (*Watcher, error) {
	if root == "" {
		return nil, fmt.Errorf("root can not be empty")
	}
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(
		watcher.Rename,
		watcher.Move,
		watcher.Create,
		watcher.Remove,
		watcher.Chmod,
		watcher.Write)

	syncer := &Watcher{
		root:          root,
		watch:         w,
		wg:            new(sync.WaitGroup),
		cancelations:  make([]context.CancelFunc, 0),
		eventHandlers: make(map[string]Handler, 0),
	}
	if err := w.AddRecursive(root); err != nil {
		return nil, err
	}

	return syncer, nil
}

func (s *Watcher) listen() {
	go func() {
		for {
			select {
			case event := <-s.watch.Event:
				ctx, cancel := context.WithCancel(context.Background())
				for _, h := range s.eventHandlers {
					s.cancelations = append(s.cancelations, cancel)
					s.wg.Add(1)
					go h.Handle(ctx, event, s.wg)
				}
			case err := <-s.watch.Error:
				log.Fatalln(err)
			case <-s.watch.Closed:
				return
			}
		}
	}()
}

func (s *Watcher) Start(pollingInterval time.Duration) {
	s.pollingInterval = pollingInterval
	// detach event listening
	s.listen()
	err := s.watch.Start(s.pollingInterval)
	if err != nil {
		log.Fatalf("Failed to start event listener")
	}
}

func (s *Watcher) Add(name string, handle Handler) {
	s.eventHandlers[name] = handle
}

func (s *Watcher) Stop() {
	for _, cancel := range s.cancelations {
		cancel()
	}
	s.watch.Close()
	log.Println("Before wait")
	s.wg.Wait()
	log.Println("After wait")
}

func (s *Watcher) ShowWatchContext() []Object {
	objects := make([]Object, 0)
	for path, f := range s.watch.WatchedFiles() {
		objects = append(objects, Object{
			path: path,
			obj:  f,
		})
	}
	return objects
}

func (s *Watcher) ListHandlers() []string {
	handlers := make([]string, 0)
	for name, _ := range s.eventHandlers {
		handlers = append(handlers, name)
	}
	return handlers
}
