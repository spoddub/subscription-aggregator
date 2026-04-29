package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	subscriptions := r.Group("api/subscriptions")
	{
		subscriptions.POST("", CreateSubscription)
		subscriptions.GET("", ListSubscriptions)
		subscriptions.GET("/total", GetTotalCost)
		subscriptions.GET("/:id", GetSubscriptionByID)
		subscriptions.PUT("/:id", UpdateSubscription)
		subscriptions.DELETE("/:id", DeleteSubscription)
	}

	return r

}
