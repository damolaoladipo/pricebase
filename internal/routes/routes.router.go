package routes

import (
	"github.com/damolaoladipo/pricebase/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	v1 := router.Group("/v1")

	v1.GET("/", handlers.HealthCheck)

	HealthRoutes(v1.Group("/health"))
	AuthRoutes(v1.Group("/auth"))

}
