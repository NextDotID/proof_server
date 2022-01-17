package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/validator"
)

var (
	Engine *gin.Engine
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func Init() {
	Engine = gin.Default()
	Engine.GET("/healthz", healthz)
	Engine.POST("/v1/proof/payload", proofPayload)
	Engine.POST("/v1/proof", proofUpload)
	Engine.GET("/v1/proof", proofQuery)
}

func errorResp(c *gin.Context, error_code int, err error) {
	c.JSON(error_code, ErrorResponse{
		Message: err.Error(),
	})
}

func healthz(c *gin.Context) {
	supported_platforms := make([]string, 0)
	for p, _ := range validator.PlatformFactories {
		supported_platforms = append(supported_platforms, string(p))
	}

	c.JSON(http.StatusOK, gin.H{
		"hello": "proof server",
		"supported_platforms": supported_platforms,
	})
}
