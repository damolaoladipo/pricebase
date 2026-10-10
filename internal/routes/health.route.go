package routes

import (
	"github.com/damolaoladipo/pricebase/internal/handlers"
	"github.com/gin-gonic/gin"
)

func HealthRoutes(router *gin.RouterGroup) {
	router.GET("/me", handlers.HealthCheck)
	router.GET("/", handlers.Home)
}
