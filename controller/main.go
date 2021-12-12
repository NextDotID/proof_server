package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	Engine *gin.Engine
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func Init()  {
	Engine = gin.Default()
	Engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hello": "proof server",
		})
	})
	Engine.POST("/v1/proof/payload", proofPayload)
	Engine.POST("/v1/proof", proofUpload)
}

func errorResp(c *gin.Context, error_code int, err error) {
	c.JSON(error_code, ErrorResponse{
		Message: err.Error(),
	})
}
