package headless

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/common"
	"github.com/sirupsen/logrus"
)

var (
	Engine         *gin.Engine
	LauncherPath   string
	l              = logrus.WithFields(logrus.Fields{"module": "headless"})
	URLReplacement = map[string]string{}
)

func middlewareCors() gin.HandlerFunc {
	return cors.Default()
}

func Init(launcherPath string, urlReplacementRule string) {
	LauncherPath = launcherPath
	if Engine != nil {
		return
	}

	InitUrlReplacementRule(urlReplacementRule)
	Engine = gin.Default()
	Engine.Use(middlewareCors())

	Engine.GET("/healthz", healthz)
	Engine.POST("/v1/find", validate)
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello":       "proof service",
		"environment": common.Environment,
		"revision":    common.Revision,
		"built_at":    common.BuildTime,
		"runtime":     common.CurrentRuntime,
	})
}

func InitUrlReplacementRule(rule string) {
	if rule == "" {
		return
	}

	for _, r := range strings.Split(rule, ",") {
		parts := strings.Split(r, "=")
		if len(parts) != 2 {
			l.Warnf("invalid url replacement rule: %s", r)
			continue
		}

		URLReplacement[parts[0]] = parts[1]
	}
}
