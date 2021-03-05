package syncer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ProximaCentauri1989/businesscard-syncer/models"
	"github.com/radovskyb/watcher"
)

type Handler interface {
	Handle(ctx context.Context, event watcher.Event, wg *sync.WaitGroup)
}

type Syncer struct {
	root            string
	watch           *watcher.Watcher
	pollingInterval time.Duration
	eventHandlers   map[string]Handler
	wg              *sync.WaitGroup
	cancelations    []context.CancelFunc
}

// Creates a new syncer and starts event listening. Event polling should be started by caller manually using 'Start' method
func NewSyncer(root string) (*Syncer, error) {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(
		watcher.Rename,
		watcher.Move,
		watcher.Create,
		watcher.Remove,
		watcher.Chmod,
		watcher.Write)

	syncer := &Syncer{
		root:         root,
		watch:        w,
		wg:           new(sync.WaitGroup),
		cancelations: make([]context.CancelFunc, 0),
	}
	if err := w.AddRecursive(root); err != nil {
		return nil, err
	}

	return syncer, nil
}

func (s *Syncer) listen() {
	go func() {
		for {
			select {
			case event := <-s.watch.Event:
				ctx, cancel := context.WithCancel(context.Background())
				for _, handle := range s.eventHandlers {
					s.cancelations = append(s.cancelations, cancel)
					s.wg.Add(1)
					go handle(ctx, event, s.wg)
				}
			case err := <-s.watch.Error:
				log.Fatalln(err)
			case <-s.watch.Closed:
				return
			}
		}
	}()
}

func (s *Syncer) Start(pollingInterval time.Duration) error {
	s.pollingInterval = pollingInterval
	// detach event listening
	s.listen()
	return s.watch.Start(s.pollingInterval)
}

func (s *Syncer) Add(name string, handle Handler) {
	s.eventHandlers[name] = handle
}

func (s *Syncer) Stop() {
	for _, cancel := range s.cancelations {
		cancel()
	}
	s.watch.Close()
	s.wg.Wait()
}

func (s *Syncer) ShowWatchContext() []models.Object {
	objects := make([]models.Object, 0)
	for path, f := range s.watch.WatchedFiles() {
		objects = append(objects, models.Object{
			path: path,
			obj:  f,
		})
	}
	return objects
}

func (s *Syncer) ListHandlers() []string {
	handlers := make([]string, 0)
	for name, _ := range s.eventHandlers {
		handlers = append(handlers, name)
	}
	return handlers
}
