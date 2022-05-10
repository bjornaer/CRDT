//go:build examples
// +build examples

package examples

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func addElm(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func removeElm(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func getElms(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func hello(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"code":  http.StatusOK,
			"error": "Welcome server 01",
		},
	)
}

func addRoutes(r *gin.Engine) {
	r.GET("/", hello)
	r.GET("/item", getElms)
	r.POST("/item", addElm)
	r.DELETE("/item", removeElm)
}
