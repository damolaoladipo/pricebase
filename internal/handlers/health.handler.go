package handlers

import (
	"net/http"

	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	env *utils.ENV
}

func NewHealthHandler(env *utils.ENV) *HealthHandler {
	return &HealthHandler{
		env: env,
	}
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":  false,
		"errors": []string{},
		"data": gin.H{
			"name":    "Pricebase API",
			"version": "1.00.00",
		},
		"message": "Pricebase api v1.0.0 is healthy",
		"code":    http.StatusOK,
	})
}

func Home(c *gin.Context) {
	env := utils.Current()
	environment := "development"

	if env != nil {
		if env.IsProd {
			environment = "production"
		} else if env.IsStaging {
			environment = "staging"
		} else if env.IsDev {
			environment = "development"
		}
	}

	c.HTML(http.StatusOK, "health-check", gin.H{
		"error":  false,
		"errors": []string{},
		"data": gin.H{
			"name":    "Pricebase API",
			"version": "01.00.00",
		},
		"message": "pricebase-api is running in " + environment + " mode",
		"status":  http.StatusOK,
	})
}
