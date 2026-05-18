package bootstrap

import (
	"example.com/axiomnizam/internal/gatekeeper/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all Gatekeeper HTTP routes on the given router.
func RegisterRoutes(router *gin.Engine, httpHandler *handlers.HTTPHandler) {
	httpHandler.RegisterRoutes(router)
}
