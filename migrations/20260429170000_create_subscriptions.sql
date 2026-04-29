-- +goose Up
CREATE TABLE subscriptions (
                               id BIGSERIAL PRIMARY KEY,
                               service_name TEXT NOT NULL CHECK (service_name <> ''),
                               price INTEGER NOT NULL CHECK (price > 0),
                               user_id UUID NOT NULL,
                               start_date DATE NOT NULL,
                               end_date DATE,
                               created_at TIMESTAMP NOT NULL DEFAULT now(),
                               updated_at TIMESTAMP NOT NULL DEFAULT now(),
                               CHECK (end_date IS NULL OR end_date >= start_date)
);

-- +goose Down
DROP TABLE subscriptions;