package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spoddub/subscription-aggregator/internal/model"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, params model.CreateSubscriptionParams) (model.Subscription, error) {
	const query = `
		INSERT INTO subscriptions (
			service_name,
            price,
            user_id,
            start_date,
            end_date
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING 
			id, 
			service_name,
			price,
			user_id, 
			start_date, 
			end_date, 
			created_at, 
			updated_at
	`

	row := r.pool.QueryRow(ctx, query, params.ServiceName, params.Price, params.UserID, params.StartDate, nullableTime(params.EndDate))

	subscription, err := scanSubscription(row)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("could not create subscription: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	const query = `
		SELECT
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)

	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("could not scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id int64) (model.Subscription, error) {
	const query = `
		SELECT
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	subscription, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Subscription{}, model.ErrNotFound
		}

		return model.Subscription{}, fmt.Errorf("get subscription by id: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, params model.UpdateSubscriptionParams) (model.Subscription, error) {
	const query = `
		UPDATE subscriptions
		SET
			service_name = $2,
			price = $3,
			user_id = $4,
			start_date = $5,
			end_date = $6,
			updated_at = now()
		WHERE id = $1
			RETURNING
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		`

	row := r.pool.QueryRow(ctx, query, params.ID, params.ServiceName, params.Price, params.UserID, params.StartDate, nullableTime(params.EndDate))

	subscription, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Subscription{}, model.ErrNotFound
		}

		return model.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM subscriptions
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("could not delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("not found: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, filter model.TotalCostFilter) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(
			price * (
				EXTRACT(YEAR FROM age(
					LEAST(COALESCE(end_date, $2), $2),
					GREATEST(start_date, $1)
				))::int * 12
				+
				EXTRACT(MONTH FROM age(
					LEAST(COALESCE(end_date, $2), $2),
					GREATEST(start_date, $1)
				))::int
				+ 1
			)
		), 0)::bigint
		FROM subscriptions
		WHERE start_date <= $2
		  AND COALESCE(end_date, $2) >= $1
		  AND ($3::uuid IS NULL OR user_id = $3)
		  AND ($4::text IS NULL OR service_name = $4)
	`

	var userIDArg any
	if filter.UserID != nil {
		userIDArg = filter.UserID
	}

	var serviceNameArg any
	if filter.ServiceName != nil {
		serviceNameArg = filter.ServiceName
	}

	var total int64

	err := r.pool.QueryRow(
		ctx,
		query,
		filter.From,
		filter.To,
		userIDArg,
		serviceNameArg).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("could not get total cost: %w", err)
	}

	return total, nil
}

func scanSubscription(row pgx.Row) (model.Subscription, error) {
	var subscription model.Subscription
	var endDate *time.Time

	err := row.Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&subscription.StartDate,
		&endDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)
	if err != nil {
		return model.Subscription{}, err
	}

	subscription.EndDate = endDate

	return subscription, nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}
