package server

import (
	"authz/domain/contracts"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// see https://github.com/gin-gonic/gin/issues/57

// GinServer underlying struct
type GinServer struct{}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (g GinServer) Serve(host string, handler http.HandlerFunc, wait *sync.WaitGroup) error {
	router := gin.Default()
	router.GET("/", gin.WrapF(handler))
	err := router.Run(":" + host)
	return err
}

// NewServer creates a new Server object to use.
func (g GinServer) NewServer() contracts.Server {
	return GinServer{}
}
