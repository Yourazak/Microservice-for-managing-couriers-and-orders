-- +goose Up
ALTER TABLE couriers
ADD COLUMN IF NOT EXISTS transport_type TEXT NOT NULL DEFAULT 'on_foot';

CREATE TABLE IF NOT EXISTS delivery(
    id          BIGSERIAL PRIMARY KEY,
    courier_id  BIGINT NOT NULL REFERENCES couriers(id) ON DELETE CASCADE,
    order_id    VARCHAR(255) NOT NULL UNIQUE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deadline    TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delivery;
ALTER TABLE couriers DROP COLUMN IF EXISTS transport_type;
-- +goose StatementEnd
