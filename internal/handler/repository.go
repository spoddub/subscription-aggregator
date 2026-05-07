package handler

import (
	"context"

	"github.com/spoddub/subscription-aggregator/internal/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, params model.CreateSubscriptionParams) (model.Subscription, error)

	List(ctx context.Context, filter model.ListSubscriptionFilter) ([]model.Subscription, error)

	GetByID(ctx context.Context, id int64) (model.Subscription, error)

	Update(ctx context.Context, params model.UpdateSubscriptionParams) (model.Subscription, error)

	Delete(ctx context.Context, id int64) error

	GetTotalCost(ctx context.Context, filter model.TotalCostFilter) (int64, error)
}
