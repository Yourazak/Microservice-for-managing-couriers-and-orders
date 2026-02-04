-- +goose Up
-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers(status);
CREATE INDEX IF NOT EXISTS idx_couriers_phone ON couriers(phone);
CREATE INDEX IF NOT EXISTS idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_courier_id ON deliveries(courier_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries(status);
CREATE INDEX IF NOT EXISTS idx_deliveries_assigned_at ON deliveries(assigned_at);

-- +goose Down
DROP INDEX IF EXISTS idx_couriers_status;
DROP INDEX IF EXISTS idx_couriers_phone;
DROP INDEX IF EXISTS idx_deliveries_order_id;
DROP INDEX IF EXISTS idx_deliveries_courier_id;
DROP INDEX IF EXISTS idx_deliveries_status;
DROP INDEX IF EXISTS idx_deliveries_assigned_at;
