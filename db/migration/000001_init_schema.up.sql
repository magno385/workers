-- 1. Создаем ENUM-тип для статусов задач, чтобы гарантировать целостность данных
CREATE TYPE task_status AS ENUM ('pending', 'processing', 'completed', 'failed');

-- 2. Таблица для хранения прокси-серверов
CREATE TABLE proxies (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL UNIQUE,       -- Формат ip:port
    username VARCHAR(100),
    password VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,             -- Флаг работоспособности прокси
    last_used_at TIMESTAMP WITH TIME ZONE,      -- Время последнего использования
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Таблица для хранения задач на автоматизацию
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Уникальный UUID задачи
    status task_status DEFAULT 'pending',          -- Текущий статус из нашего ENUM
    attempts INT DEFAULT 0,                        -- Текущее число попыток выполнения
    max_attempts INT DEFAULT 3,                    -- Лимит попыток до перевода в статус 'failed'
    proxy_id INT REFERENCES proxies(id) ON DELETE SET NULL, -- Какой прокси выделен под задачу
    payload JSONB NOT NULL,                        -- Данные задачи в формате JSON (email, пароль, имя и т.д.)
    error_message TEXT,                            -- Текст ошибки, если задача завершилась неудачей
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска задач со статусом 'pending' (для воркеров)
CREATE INDEX idx_tasks_status_pending ON tasks(status) WHERE status = 'pending';