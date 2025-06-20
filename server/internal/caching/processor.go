package caching

import (
	"crypto/sha1"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/torbenconto/TeXBooK/internal/logger"
)

func cache(path string) error {
	hash := sha1.Sum([]byte(path))

	cmd := exec.Command("pdflatex",
		"-interaction=nonstopmode",
		"-output-directory=cache",
		"-jobname="+fmt.Sprintf("%x", hash),
		path,
	)

	cmd.Dir = filepath.Dir(path)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
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
	for job := range jp.jobQueue {
		jp.lastProcessedMu.Lock()
		lastTime, seen := jp.lastProcessedMap[job.Path]
		if seen && time.Since(lastTime) < 1*time.Second {
			jp.lastProcessedMu.Unlock()
			continue
		}
		jp.lastProcessedMap[job.Path] = time.Now()
		jp.lastProcessedMu.Unlock()

		logger.Log.WithField("job", job).Info("Processing job")

		err := cache(job.Path)
		if err != nil {
			logger.Log.WithError(err).WithField("job", job).Error("Failed to cache file")
			continue
		}
	}
}
