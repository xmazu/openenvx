package watch

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const DefaultDebounce = 500 * time.Millisecond

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	files    map[string]bool
	onChange chan struct{}
	mu       sync.Mutex
	timer    *time.Timer
	done     chan struct{}
}

func NewFileWatcher() (*FileWatcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		watcher:  fsw,
		files:    make(map[string]bool),
		onChange: make(chan struct{}, 1),
		done:     make(chan struct{}),
	}, nil
}

func (w *FileWatcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if w.files[absPath] {
		return nil
	}

	if _, err := os.Stat(absPath); err == nil {
		if err := w.watcher.Add(absPath); err != nil {
			return err
		}
		w.files[absPath] = true
	} else {
		dir := filepath.Dir(absPath)
		if err := w.watcher.Add(dir); err != nil {
			return err
		}
		w.files[absPath] = true
	}

	return nil
}

func (w *FileWatcher) Start() <-chan struct{} {
	go w.run()
	return w.onChange
}

func (w *FileWatcher) run() {
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				w.mu.Lock()
				_, isWatched := w.files[event.Name]
				for watched := range w.files {
					if filepath.Base(watched) == filepath.Base(event.Name) {
						isWatched = true
						break
					}
				}
				w.mu.Unlock()

				if isWatched {
					w.trigger()
				}
			}
		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *FileWatcher) trigger() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(DefaultDebounce, func() {
		select {
		case w.onChange <- struct{}{}:
		default:
		}
	})
}

func (w *FileWatcher) Close() error {
	close(w.done)
	if w.timer != nil {
		w.timer.Stop()
	}
	return w.watcher.Close()
}

func (w *FileWatcher) Files() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	files := make([]string, 0, len(w.files))
	for f := range w.files {
		files = append(files, f)
	}
	return files
}
