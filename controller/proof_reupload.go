package controller

import "github.com/gin-gonic/gin"

// proofReupload revalidate and save a proof which was sent as a
// proof post but haven't been been called in `POST /v1/proof`.
func proofReupload(c *gin.Context) {
	c.JSON(200, gin.H{"TODO": "implement me"})
}
