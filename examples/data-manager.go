//go:build examples
// +build examples

package examples

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bjornaer/crdt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type DataManager struct {
	Store *crdt.LastWriterWinsSet[string]
	Peers []string
	Self  string
}

type RequestBody struct {
	Item string
}

func SendRequest(url string) (http.Response, error) {
	if url == "" {
		return http.Response{}, errors.New("empty url provided")
	}

	client := http.Client{
		Timeout: time.Duration(5 * 60 * time.Second),
	}

	response, err := client.Get(url)
	if err != nil {
		return http.Response{}, err
	}

	return *response, nil
}

func SendCRDTRequest(peer string) (crdt.LastWriterWinsSet[string], error) {
	var _lwwSet crdt.LastWriterWinsSet[string]

	// Return an empty LWW Set followed by an error if the peer is nil
	if peer == "" {
		return _lwwSet, errors.New("empty peer provided")
	}

	// generate the request URL
	url := fmt.Sprintf("http://localhost:%s/raw", peer)
	response, err := SendRequest(url)
	if err != nil {
		return _lwwSet, err
	}

	// Return an empty LWW Set followed by an error
	// if the peer's response is not HTTP 200 OK
	if response.StatusCode != http.StatusOK {
		return _lwwSet, errors.New("received invalid http response status:" + fmt.Sprint(response.StatusCode))
	}

	// Decode the peer's LWW Set to be usable by our local LWW Set
	var lwwSet crdt.LastWriterWinsSet[string]
	err = json.NewDecoder(response.Body).Decode(&lwwSet)
	if err != nil {
		return _lwwSet, err
	}

	// Return the decoded peer's LWW Set
	_lwwSet = lwwSet
	return _lwwSet, nil
}

func (dm *DataManager) SyncWithPeers() {
	// Iterate over the peer list and send a /raw GET request
	// to each peer to obtain its LWW Set
	for _, peer := range dm.Peers {
		peerSet, err := SendCRDTRequest(peer)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "peer": peer}).Error("failed sending raw set values request")
			continue
		}

		// Merge the peer's Set with our local Set
		dm.Store.Merge(peerSet)
	}
}

func (dm *DataManager) Hello(context *gin.Context) {
	context.JSON(
		http.StatusOK,
		gin.H{
			"code":  http.StatusOK,
			"error": fmt.Sprintf("Welcome server %s", dm.Self),
		},
	)
}

func (dm *DataManager) GetRawData(context *gin.Context) {
	context.JSON(200, dm.Store)
}

func (dm *DataManager) GetSyncedData(context *gin.Context) {
	dm.SyncWithPeers()
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
