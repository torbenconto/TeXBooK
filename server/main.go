package main

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	allowedExts = map[string]bool{".tex": true}
	watchersMap = make(map[uint32]*Watcher)
	jobQueue    = make(chan CacheJob, 200)
)

func collectFiles(fsys fs.FS, dir string) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && allowedExts[strings.ToLower(filepath.Ext(d.Name()))] {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

type DataSource interface {
	Connect() error
	Disconnect() error
	Type() string
	Describe() map[string]any
	ID() uint32
}

type FileSystemAccessible interface {
	FS() (fs.FS, error)
}

func GetFileSystem(ds DataSource) (fs.FS, error) {
	if fsa, ok := ds.(FileSystemAccessible); ok {
		return fsa.FS()
	}
	return nil, fmt.Errorf("data source of type %s does not provide file system access", ds.Type())
}

type LocalDataSource struct {
	Path    string
	IDField uint32
}

func (l *LocalDataSource) FS() (fs.FS, error) {
	if err := l.Connect(); err != nil {
		return nil, err
	}
	return os.DirFS(l.Path), nil
}

func (l *LocalDataSource) Type() string { return "local" }
func (l *LocalDataSource) ID() uint32   { return l.IDField }

func (l *LocalDataSource) Describe() map[string]any {
	return map[string]any{"type": "local", "path": l.Path}
}
func (l *LocalDataSource) Connect() error {
	if l.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if _, err := os.Stat(l.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %v", l.Path)
	} else if err != nil {
		return fmt.Errorf("error checking path: %w", err)
	}
	return nil
}
func (l *LocalDataSource) Disconnect() error { return nil }

type DataSourceStored map[string]DataSource

func startWatchingDataSource(fsWatcher *fsnotify.Watcher, ds DataSource) {
	local, ok := ds.(*LocalDataSource)
	if !ok {
		log.Printf("Skipping non-local data source: %d", ds.ID())
		return
	}

	watcher := NewWatcher(fsWatcher, ds, jobQueue)
	watchersMap[ds.ID()] = watcher

	go watcher.Start()
	watcher.WarmUpExistingFiles()

	watcher.AddDirAndSubDirs(local.Path)
}

func main() {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	processor := NewJobProcessor(jobQueue)
	processor.Start()

	router := gin.Default()
	dataSourceStore, _ := New[DataSourceStored]("TeXBooK.db", "DataSourceStore")
	gob.Register(&LocalDataSource{})

	stored, err := dataSourceStore.Get()
	if err != nil {
		log.Fatalf("failed to load data source store: %v", err)
	}
	for _, ds := range stored {
		if err := ds.Connect(); err != nil {
			log.Printf("Failed to connect to data source %d: %v", ds.ID(), err)
			continue
		}
		startWatchingDataSource(fsWatcher, ds)

		if localDS, ok := ds.(*LocalDataSource); ok {
			w := NewWatcher(fsWatcher, localDS, jobQueue)
			watchersMap[ds.ID()] = w
			w.AddDirAndSubDirs(localDS.Path)
			go w.Start()
			go w.WarmUpExistingFiles()
		}
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://bambu-portal-v2.vercel.app", "https://www.bambu-portal-v2.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		dataSources := api.Group("/data")
		{
			type dataSourcesAddInput struct {
				Type string `json:"type" binding:"required"`
				Name string `json:"name" binding:"required"`
				Path string `json:"path,omitempty"`
			}

			dataSources.POST("/add", func(ctx *gin.Context) {
				var input dataSourcesAddInput
				if err := ctx.ShouldBindJSON(&input); err != nil {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
					return
				}

				stored, err := dataSourceStore.Get()
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
					return
				}

				if stored == nil {
					stored = make(DataSourceStored)
				}

				if _, exists := stored[input.Name]; exists {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "data source already exists"})
					return
				}

				switch input.Type {
				case "local":
					ds := &LocalDataSource{Path: input.Path, IDField: uuid.New().ID()}
					stored[input.Name] = ds
					if err := dataSourceStore.Save(stored); err != nil {
						ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					if err := ds.Connect(); err != nil {
						ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to data source"})
						return
					}
					startWatchingDataSource(fsWatcher, ds)

					ctx.JSON(http.StatusOK, gin.H{"status": "data source added"})
				default:
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unsupported data source type"})
				}
			})

			dataSources.GET("/list", func(ctx *gin.Context) {
				stored, err := dataSourceStore.Get()
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
					return
				}

				out := make([]gin.H, 0, len(stored))
				for name, ds := range stored {
					out = append(out, gin.H{"name": name, "metadata": ds.Describe()})
				}
				ctx.JSON(http.StatusOK, out)
			})
		}

		fsGroup := api.Group("/fs")
		{
			fsGroup.GET("/file", func(ctx *gin.Context) {
				source := ctx.Query("source")
				filePath := ctx.Query("path")

				if source == "" {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no data source provided"})
					return
				}
				if filePath == "" {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no file path provided"})
					return
				}

				stored, err := dataSourceStore.Get()
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
					return
				}

				ds, ok := stored[source]
				if !ok {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid data source"})
					return
				}

				if err := ds.Connect(); err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to data source"})
					return
				}
				defer ds.Disconnect()

				dsFS, err := GetFileSystem(ds)
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file system"})
					return
				}

				unesc, _ := url.QueryUnescape(filePath)
				cleanPath := filepath.Clean(unesc)
				cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
				cleanPath = strings.TrimLeft(cleanPath, "/\\")
				if cleanPath == "." || cleanPath == ".." || filepath.IsAbs(cleanPath) {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
					return
				}

				fmt.Println("Serving file:", cleanPath)

				content, err := fs.ReadFile(dsFS, cleanPath)
				if err != nil {
					fmt.Println(err.Error())
					ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "file not found or unable to read: " + err.Error()})
					return
				}

				contentType := "application/octet-stream"
				if strings.HasSuffix(strings.ToLower(cleanPath), ".pdf") {
					contentType = "application/pdf"
				}

				ctx.Data(http.StatusOK, contentType, content)
			})

			fsGroup.GET("/list", func(ctx *gin.Context) {
				source := ctx.Query("source")
				if source == "" {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no data source provided"})
					return
				}

				stored, err := dataSourceStore.Get()
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
					return
				}

				ds, ok := stored[source]
				if !ok {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid data source"})
					return
				}

				if err := ds.Connect(); err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to data source"})
					return
				}
				defer ds.Disconnect()

				dsFS, err := GetFileSystem(ds)
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file system"})
					return
				}

				files, err := collectFiles(dsFS, ".")
				if err != nil {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list files: " + err.Error()})
					return
				}

				out := make([]gin.H, 0, len(files))
				for _, file := range files {
					absPath := filepath.Join(stored[source].(*LocalDataSource).Path, file)

					hash := sha1.Sum([]byte(absPath))
					thumbPath := fmt.Sprintf("/cache/%x.pdf", hash)
					out = append(out, gin.H{"path": file, "thumbnail": thumbPath})
				}
				ctx.JSON(http.StatusOK, gin.H{"files": out})
			})

		}

	}

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start TeXBooK server: %v", err)
	}
}
