package headless

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/common"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
)

var (
	Engine *gin.Engine
	l      = logrus.WithFields(logrus.Fields{"module": "headless"})
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func middlewareCors() gin.HandlerFunc {
	return cors.Default()
}

func Init() {
	if Engine != nil {
		return
	}

	Engine = gin.Default()
	Engine.Use(middlewareCors())

	Engine.GET("/healthz", healthz)
	Engine.POST("/v1/validate", validate)
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
		"hello":       "proof service",
		"environment": common.Environment,
		"revision":    common.Revision,
		"built_at":    common.BuildTime,
	})
}
