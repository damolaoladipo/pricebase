package routes

import (
	"github.com/damolaoladipo/pricebase/internal/handlers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup) {
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)
	router.POST("/verify-otp", handlers.VerifyOTP)
	router.POST("/resend-otp", handlers.ResendOTP)
	router.POST("/activate", handlers.ActivateAccount)
	router.POST("/forgot-password", handlers.ForgotPassword)
	router.POST("/reset-password", handlers.ResetPassword)

}
