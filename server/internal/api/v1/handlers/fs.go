package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/torbenconto/TeXBooK/internal/logger"
)

// /api/v1/datasources/:name/fs/list
func (d *DataSourceHandler) ListFiles(c *gin.Context) {
	source := c.Param("name")

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

	json, err := ds.ListFiles()
	if err != nil {
		logger.Log.WithError(err).WithField("data_source", source).Error("failed to list files")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	c.JSON(http.StatusOK, json)
}
