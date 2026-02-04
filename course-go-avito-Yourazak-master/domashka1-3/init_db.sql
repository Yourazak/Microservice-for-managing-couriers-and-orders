-- Создание таблиц (без goose комментариев)
CREATE TABLE IF NOT EXISTS couriers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('available','busy','paused')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Тестовые данные
INSERT INTO couriers (name, phone, status) VALUES
    ('Антон', '+37444111222', 'available'),
    ('Иван', '+79991112233', 'busy'),
    ('Сергей', '+79998887766', 'paused')
ON CONFLICT (phone) DO NOTHING;

-- Добавление transport_type
ALTER TABLE couriers
ADD COLUMN IF NOT EXISTS transport_type TEXT NOT NULL DEFAULT 'on_foot';

-- Таблица доставок
CREATE TABLE IF NOT EXISTS delivery(
    id BIGSERIAL PRIMARY KEY,
    courier_id BIGINT NOT NULL REFERENCES couriers(id) ON DELETE CASCADE,
    order_id VARCHAR(255) NOT NULL UNIQUE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deadline TIMESTAMP WITH TIME ZONE NOT NULL
);
