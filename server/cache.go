package main

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func sum(path string) ([20]byte, error) {
	return sha1.Sum([]byte(path)), nil
}

func cache(texPath string) ([20]byte, string, error) {
	hash, err := sum(texPath)
	if err != nil {
		return [20]byte{}, "", nil
	}

	outputDir := "cache"

	cachedPath := filepath.Join(outputDir, fmt.Sprintf("%x", hash)+".pdf")
	if _, err := os.Stat(cachedPath); err == nil {
		return hash, cachedPath, nil
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return [20]byte{}, "", err
	}

	cmd := exec.Command("pdflatex",
		"-interaction=nonstopmode",
		"-output-directory="+outputDir,
		"-jobname="+fmt.Sprintf("%x", hash),
		texPath,
	)

	cmd.Dir = filepath.Dir(texPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return [20]byte{}, "", fmt.Errorf("pdflatex failed: %w\n%s", err, string(output))
	}

	return hash, cachedPath, nil
}

type JobProcessor struct {
	jobQueue         <-chan CacheJob
	lastProcessedMap map[string]time.Time
	lastProcessedMu  sync.Mutex
}

func NewJobProcessor(jobQueue <-chan CacheJob) *JobProcessor {
	return &JobProcessor{jobQueue: jobQueue, lastProcessedMap: make(map[string]time.Time)}
}

func (jp *JobProcessor) Start() {
	go func() {
		for job := range jobQueue {
			jp.lastProcessedMu.Lock()
			lastTime, seen := jp.lastProcessedMap[job.TeXPath]
			now := time.Now()
			if seen && now.Sub(lastTime) < 1*time.Second {
				jp.lastProcessedMu.Unlock()
				continue
			}
			jp.lastProcessedMap[job.TeXPath] = now
			jp.lastProcessedMu.Unlock()

			log.Printf("[Processor] Caching: %s", job.TeXPath)
			_, _, err := cache(job.TeXPath)
			if err != nil {
				log.Printf("[Processor] Failed to cache %s: %v", job.TeXPath, err)
			} else {
				log.Printf("[Processor] Successfully cached %s", job.TeXPath)
			}
		}
	}()

}

type CacheJob struct {
	TeXPath    string
	DataSource string
}
type Watcher struct {
	watcher    *fsnotify.Watcher
	jobQueue   chan CacheJob
	dataSource DataSource
}

func NewWatcher(fsnotifyWatcher *fsnotify.Watcher, dataSource DataSource, queue chan CacheJob) *Watcher {
	return &Watcher{
		watcher:    fsnotifyWatcher,
		jobQueue:   queue,
		dataSource: dataSource,
	}
}

func (w *Watcher) AddDirAndSubDirs(root string) {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			err := w.watcher.Add(path)
			if err != nil {
				log.Printf("Failed to watch %s: %v", path, err)
			}
		}
		return nil
	})
}

func (w *Watcher) Start() {
	defer w.watcher.Close()

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				info, err := os.Stat(event.Name)
				if err != nil || info.IsDir() {
					continue
				}
				if strings.HasSuffix(strings.ToLower(info.Name()), ".tex") {
					w.jobQueue <- CacheJob{
						TeXPath: event.Name,
					}
				}
			}

			if event.Op&fsnotify.Remove != 0 {
				if strings.HasSuffix(strings.ToLower(event.Name), ".tex") {
					hash, err := sum(event.Name)
					if err != nil {
						log.Printf("[Watcher] Failed to hash deleted file: %v", err)
						continue
					}
					cachedPath := filepath.Join("cache", fmt.Sprintf("%x", hash)+".pdf")
					cachedPath = strings.ReplaceAll(cachedPath, "\\", "/")
					cachedPath = strings.TrimLeft(cachedPath, "/\\")
					log.Println(cachedPath)
					if err := os.Remove(cachedPath); err != nil && !os.IsNotExist(err) {
						log.Printf("[Watcher] Failed to delete cached PDF: %v", err)
					} else {
						log.Printf("[Watcher] Deleted cached PDF for: %s", event.Name)
					}
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func (w *Watcher) WarmUpExistingFiles() {
	err := filepath.WalkDir(w.dataSource.(*LocalDataSource).Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".tex") {
			hash, err := sum(path)
			if err != nil {
				return nil
			}
			cachedPath := filepath.Join("cache", fmt.Sprintf("%x", hash)+".pdf")
			if _, err := os.Stat(cachedPath); os.IsNotExist(err) {
				w.jobQueue <- CacheJob{TeXPath: path}
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Warm-up failed for %d: %v", w.dataSource.ID(), err)
	}
}
