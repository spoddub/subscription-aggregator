package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/spoddub/subscription-aggregator/internal/repository"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	repo *repository.SubscriptionRepository
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

func NewRouter(repo *repository.SubscriptionRepository) *gin.Engine {
	r := gin.Default()

	handler := &Handler{
		repo: repo,
	}

	r.GET("/ping", Ping)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

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
