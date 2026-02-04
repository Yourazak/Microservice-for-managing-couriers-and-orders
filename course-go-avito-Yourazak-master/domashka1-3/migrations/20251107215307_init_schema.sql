-- +goose Up
CREATE TABLE IF NOT EXISTS couriers (
                                        id          BIGSERIAL PRIMARY KEY,
                                        name        TEXT NOT NULL,
                                        phone       TEXT NOT NULL UNIQUE,
                                        status      TEXT NOT NULL CHECK (status IN ('available','busy','paused')),
                                        created_at  TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS couriers;
