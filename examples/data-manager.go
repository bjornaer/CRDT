//go:build examples
// +build examples

package examples

import (
	"time"

	"github.com/bjornaer/crdt"
	"github.com/gin-gonic/gin"
)

type DataManager struct {
	Store *crdt.LastWriterWinsSet[string]
	Peers []string
}

type RequestBody struct {
	Item string
}

func (dm *DataManager) SyncWithPeers() {
	otherSet := &crdt.LastWriterWinsSet{} // fetch it from peer
	return dm.Store.Merge(otherSet)
}

func (dm *DataManager) GetRawData(context *gin.Context) {
	context.JSON(200, dm.Store)
}

func (dm *DataManager) GetSyncedData(context *gin.Context) {
	items, err := dm.Store.Get()
	if err != nil {
		context.String(500, err)
	}
	context.JSON(200, items)
}

func (dm *DataManager) AddData(context *gin.Context) {
	var reqBod RequestBody

	if err := context.BindJSON(&reqBod); err != nil {
		context.String(400, "Incorrect payload")
	}
	votePack, err := dm.Store.Add(reqBod.Item, time.Now())
	if err != nil {
		context.String(500, err)
	}
	context.JSON(200, votePack)
}

func (dm *DataManager) RemoveData(context *gin.Context) {
	var reqBod RequestBody

	if err := context.BindJSON(&reqBod); err != nil {
		context.String(400, "Incorrect payload")
	}
	err := dm.Store.Remove(reqBod.Item, time.Now())
	if err != nil {
		context.String(500, "Something went wrong")
	}
	context.JSON(200)
}
