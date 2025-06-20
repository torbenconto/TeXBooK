package main

import (
	"encoding/gob"

	"github.com/gin-gonic/gin"
	v1 "github.com/torbenconto/TeXBooK/internal/api/v1"
	"github.com/torbenconto/TeXBooK/internal/datasources"
	"github.com/torbenconto/TeXBooK/internal/logger"
)

func main() {
	logger.Init()
	gob.Register(&datasources.LocalDataSource{})

	r := gin.Default()

	err := v1.RegisterRoutes(r)
	if err != nil {
		panic(err)
	}

	r.Run(":8080")
}
