package main

import (
	"encoding/gob"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	v1 "github.com/torbenconto/TeXBooK/internal/api/v1"
	"github.com/torbenconto/TeXBooK/internal/datasources"
	"github.com/torbenconto/TeXBooK/internal/logger"
)

func main() {
	logger.Init()
	gob.Register(&datasources.LocalDataSource{})

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		MaxAge:           12 * time.Hour,
	}))

	err := v1.RegisterRoutes(r)
	if err != nil {
		panic(err)
	}

	r.Run(":8080")
}
