package controller

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/validator"
	"github.com/nextdotid/proof-server/common"
)

var (
	Engine *gin.Engine
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func middlewareCors() gin.HandlerFunc {
	// *
	return cors.Default()
}


func Init() {
	if Engine != nil {
		return
	}

	Engine = gin.Default()
	Engine.Use(middlewareCors())

	Engine.GET("/healthz", healthz)
	Engine.POST("/v1/proof/payload", proofPayload)
	Engine.POST("/v1/proof", proofUpload)
	Engine.GET("/v1/proof", proofQuery)
	Engine.POST("/v1/kv/payload", kvPatchPayload)
	Engine.POST("/v1/kv", kvPatch)
}

func errorResp(c *gin.Context, error_code int, err error) {
	c.JSON(error_code, ErrorResponse{
		Message: err.Error(),
	})
}

func healthz(c *gin.Context) {
	platforms := make([]string, 0)
	for p := range validator.PlatformFactories {
		platforms = append(platforms, string(p))
	}

	c.JSON(http.StatusOK, gin.H{
		"hello": "proof server",
		"platforms": platforms,
		"environment": common.Environment,
		"revision": common.Revision,
		"built_at": common.BuildTime,
	})
}
