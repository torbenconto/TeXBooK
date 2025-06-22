package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/torbenconto/TeXBooK/internal/logger"
)

// /api/v1/datasources/:name/fs/list
func (d *DataSourceHandler) ListFiles(c *gin.Context) {
	source := c.Param("name")
	path := c.Query("path")

	stored, err := d.DataSourceStore.Get()
	if err != nil {
		logger.Log.WithError(err).Error("failed to load data source config")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}

	ds, ok := stored[source]
	if !ok {
		logger.Log.WithField("provided_source", source).Error("invalid data source provided")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid data source"})
		return
	}

	json, err := ds.ListFiles(path)
	if err != nil {
		logger.Log.WithError(err).WithField("data_source", source).Error("failed to list files")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	c.JSON(http.StatusOK, json)
}

// /api/v1/datasources/:name/file?path=x
func (d *DataSourceHandler) File(c *gin.Context) {
	path := c.Query("path")
	source := c.Param("name")

	stored, err := d.DataSourceStore.Get()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to load data source config")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}

	ds, ok := stored[source]
	if !ok {
		logger.Log.WithField("provided_source", source).Error("Invalid data source provided")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid data source"})
		return
	}

	file, err := ds.ReadFile(path)
	if err != nil {
		logger.Log.WithError(err).Error("Error reading file")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(path), ".pdf") {
		contentType = "application/pdf"
	}

	c.Data(http.StatusOK, contentType, file)
}
