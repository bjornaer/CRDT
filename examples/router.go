//go:build examples
// +build examples

package examples

import (
	"github.com/gin-gonic/gin"
)

func addRoutes(r *gin.Engine, dm *DataManager) {
	r.GET("/", hello)
	r.GET("/raw", dm.GetRawData)
	r.GET("/item", dm.GetSyncedData)
	r.POST("/item", dm.AddData)
	r.DELETE("/item", dm.RemoveData)
}
