package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/torbenconto/TeXBooK/internal/api/v1/handlers"
	"github.com/torbenconto/TeXBooK/internal/caching"
	"github.com/torbenconto/TeXBooK/internal/logger"
	"github.com/torbenconto/TeXBooK/internal/repository"
)

func RegisterRoutes(r *gin.Engine) error {
	store, err := repository.New[repository.StoredDataSources]("TeXBooK.db", "StoredDataSources")
	if err != nil {
		return err
	}

	// Connect to all stored data sources
	stored, err := store.Get()
	if err != nil {
		return err
	}
	for _, ds := range stored {
		if err := ds.Connect(); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"data_source_id": ds.ID(),
				"error":          err,
			}).Error("Failed to connect to data source")

			continue
		}

		caching.StartWatchingDataSource(ds)
	}

	processor := caching.NewJobProcessor(caching.JobQueue)
	go processor.Start()

	api := r.Group("/api/v1")
	// /api/v1
	{
		api.GET("/ping", handlers.Ping)

		dsHandler := &handlers.DataSourceHandler{
			DataSourceStore: store,
		}

		datasources := api.Group("/datasources")
		// /api/v1/datasources
		{
			datasources.POST("/add", dsHandler.AddDataSource)
			datasources.GET("/list", dsHandler.ListDataSources)

			onSource := datasources.Group("/:name")
			// /api/v1/datasources/:name
			{
				fs := onSource.Group("/fs")
				// /api/v1/datasources/:name/fs
				{
					fs.GET("/list", dsHandler.ListFiles)
					fs.GET("/file", dsHandler.File)
				}
			}
		}
	}

	return nil
}
