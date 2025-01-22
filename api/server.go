package api

import (
	db "simple_bank/db/sqlc"

	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service.
type Server struct {
  store *db.Store
  router *gin.Engine
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(store *db.Store) *Server {
  server := &Server{store: store}
  router := gin.Default()

  router.POST("/accounts", server.createAccount)
  router.GET("/accounts/:id", server.getAccount)
  router.GET("/accounts", server.listAccount)

  server.router = router
  return server
}

func (server *Server) Start(address string) error {
  return server.router.Run(address)
}

// this function returns a gin.H map with an error key 
// and the error message as the value
func errorResponse(err error) gin.H {
  return gin.H{"error": err.Error()}
}