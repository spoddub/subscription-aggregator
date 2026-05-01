package handler

import (
	"net/http"

	"github.com/spoddub/subscription-aggregator/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *repository.SubscriptionRepository
}

func NewRouter(repo *repository.SubscriptionRepository) *gin.Engine {
	r := gin.Default()

	handler := &Handler{
		repo: repo,
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	subscriptions := r.Group("api/subscriptions")
	{
		subscriptions.POST("", handler.CreateSubscription)
		subscriptions.GET("", handler.ListSubscriptions)
		subscriptions.GET("/total", handler.GetTotalCost)
		subscriptions.GET("/:id", handler.GetSubscriptionByID)
		subscriptions.PUT("/:id", handler.UpdateSubscription)
		subscriptions.DELETE("/:id", handler.DeleteSubscription)
	}

	return r

}
