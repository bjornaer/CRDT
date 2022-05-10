//go:build example
// +build example

package main

import (
	"github.com/gin-gonic/gin"
)

func addRoutes(r *gin.Engine, dm *DataManager) {
	r.GET("/", dm.Hello)
	r.GET("/raw", dm.GetRawData)
	r.GET("/item", dm.GetSyncedData)
	r.POST("/item", dm.AddData)
	r.DELETE("/item", dm.RemoveData)
}
