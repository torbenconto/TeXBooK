package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func sum(path string) ([20]byte, error) {
	return sha1.Sum([]byte(path)), nil
}

func cache(texPath string, dataSource string) ([20]byte, string, error) {
	hash, err := sum(texPath)
	if err != nil {
		return [20]byte{}, "", nil
	}

	outputDir := filepath.Join("cache", dataSource)

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

type PDFJob struct {
	TeXPath    string
	DataSource string
}

func startPDFWorkerPool(workerCount int, jobChan <-chan PDFJob) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for job := range jobChan {
				_, _, err := cache(job.TeXPath, job.DataSource)
				if err != nil {
					log.Printf("[worker %d] failed to render %s: %v", id, job.TeXPath, err)
				} else {
					log.Printf("[worker %d] rendered thumbnail for %s", id, job.TeXPath)
				}
			}
		}(i)
	}
}
