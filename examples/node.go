//go:build examples
// +build examples

package examples

import (
	"log"
	"net/http"
	"time"

	"github.com/bjornaer/crdt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func router01(dm *DataManager) http.Handler {
	e := gin.Default() // gin.Default gives an engine with logger and recovery already attached
	addRoutes(e, dm)

	return e
}

func router02(dm *DataManager) http.Handler {
	e := gin.Default()
	addRoutes(e, dm)

	return e
}

func main() {

	dataMgmt01 := &DataManager{Store: crdt.NewLWWSet[string](), Peers: {"8081"}, Self: "01"}
	server01 := &http.Server{
		Addr:         ":8080",
		Handler:      router01(dataMgmt01),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	dataMgmt02 := &DataManager{Store: crdt.NewLWWSet[string](), Peers: {"8080"}, Self: "02"}
	server02 := &http.Server{
		Addr:         ":8081",
		Handler:      router02(dataMgmt02),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})

	g.Go(func() error {
		return server02.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
