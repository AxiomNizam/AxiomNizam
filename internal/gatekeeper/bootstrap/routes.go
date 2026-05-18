package bootstrap

import (
	"example.com/axiomnizam/internal/gatekeeper/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all Gatekeeper HTTP routes on the given router group.
func RegisterRoutes(api *gin.RouterGroup, httpHandler *handlers.HTTPHandler) {
	httpHandler.RegisterRoutes(api)
}
