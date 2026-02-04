-- +goose Up
-- +goose StatementBegin
INSERT INTO couriers (name, phone, status) VALUES
                                               ('Антон', '+37444111222', 'available'),
                                               ('Иван', '+79991112233', 'busy'),
                                               ('Сергей', '+79998887766', 'paused');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM couriers
WHERE phone IN ('+37444111222', '+79991112233', '+79998887766');
-- +goose StatementEnd
