package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"maps"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/torbenconto/TeXBooK/internal/caching"
	"github.com/torbenconto/TeXBooK/internal/datasources"
	"github.com/torbenconto/TeXBooK/internal/logger"
	"github.com/torbenconto/TeXBooK/internal/repository"
)

type DataSourceHandler struct {
	DataSourceStore *repository.Store[repository.StoredDataSources]
}

type AddDataSourceInput struct {
	Type string `json:"type" binding:"required"`
	Name string `json:"name" binding:"required"`
	Path string `json:"path,omitempty"`
}

// POST /api/v1/datasources/add
func (d *DataSourceHandler) AddDataSource(c *gin.Context) {
	var input AddDataSourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.WithError(err).Warn("Invalid input for AddDataSource")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	stored, err := d.DataSourceStore.Get()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to retrieve datasource list from database")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error retrieving datasource list from database"})
		return
	}

	if stored == nil {
		stored = make(repository.StoredDataSources)
	}

	if len(stored) != 0 {
		if _, ok := stored[input.Name]; ok {
			logger.Log.WithField("datasource_name", input.Name).Warn("Data source already exists")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "data source already exists"})
			return
		}
	}

	switch input.Type {
	case "local":
		// Clean input path for windows
		path := filepath.ToSlash(input.Path)

		// Validate that path exists and is a directory
		info, err := os.Stat(path)
		if err != nil {
			logger.Log.WithError(err).WithField("datasource_name", input.Name).Error("Provided path does not exist")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "provided path does not exist"})
			return
		}
		if !info.IsDir() {
			logger.Log.WithField("datasource_name", input.Name).Error("Provided path is not a directory")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "provided path is not a directory"})
			return
		}

		ds := &datasources.LocalDataSource{
			SourcePath:     path,
			BaseDataSource: datasources.BaseDataSource{SourceID: uuid.New().ID()},
		}
		stored[input.Name] = ds

		if err := d.DataSourceStore.Save(stored); err != nil {
			logger.Log.WithError(err).WithField("datasource_name", input.Name).Error("Failed to save data source")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := ds.Connect(); err != nil {
			logger.Log.WithError(err).WithField("datasource_name", input.Name).Error("Failed to connect to data source")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to data source"})
			return
		}
		caching.StartWatchingDataSource(ds)

		logger.Log.WithField("datasource_name", input.Name).Info("Data source added and connected")
		c.JSON(http.StatusOK, gin.H{"status": "data source added"})
	default:
		logger.Log.WithField("datasource_type", input.Type).Warn("Unsupported data source type")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unsupported data source type"})
	}
}

// GET /api/v1/datasources/list
func (d *DataSourceHandler) ListDataSources(c *gin.Context) {
	stored, err := d.DataSourceStore.Get()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to load data sources")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}

	out := make([]gin.H, 0)
	for name, ds := range stored {
		meta := gin.H{
			"type": ds.Type(),
			"id":   ds.ID(),
		}

		maps.Copy(meta, ds.Metadata())

		out = append(out, gin.H{"name": name, "metadata": meta})
	}

	logger.Log.WithField("count", len(out)).Info("Listed data sources")
	c.JSON(http.StatusOK, out)
}
