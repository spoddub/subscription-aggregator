package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Subscription struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateSubscriptionParams struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type UpdateSubscriptionParams struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type TotalCostFilter struct {
	From        time.Time
	To          time.Time
	UserID      *uuid.UUID
	ServiceName *string
}
