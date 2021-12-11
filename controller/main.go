package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	Engine *gin.Engine
)


func Init()  {
	Engine = gin.Default()
	Engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hello": "proof server",
		})
	})
}
