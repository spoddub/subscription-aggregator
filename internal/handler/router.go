package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	repo   SubscriptionRepository
	logger *slog.Logger
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

func NewRouter(repo SubscriptionRepository, appLogger *slog.Logger) *gin.Engine {
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	if appLogger == nil {
		appLogger = slog.Default()
	}

	handler := &Handler{
		repo:   repo,
		logger: appLogger,
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
