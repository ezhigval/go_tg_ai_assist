-- Исправление схемы БД для соответствия коду
-- Приводим существующую БД к новой схеме

-- Исправление таблицы users: переименовываем telegram_user_id в user_id
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'telegram_user_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'user_id'
    ) THEN
        -- Переименовываем колонку
        ALTER TABLE users RENAME COLUMN telegram_user_id TO user_id;
        -- Переименовываем индекс/constraint если есть
        ALTER TABLE users RENAME CONSTRAINT users_telegram_user_id_key TO users_user_id_key;
    END IF;
    
    -- Переименовываем time_last_message в last_seen, если нужно
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'time_last_message'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_seen'
    ) THEN
        ALTER TABLE users RENAME COLUMN time_last_message TO last_seen;
    END IF;
END $$;

-- Исправление таблицы todos: добавляем/переименовываем колонки
DO $$
BEGIN
    -- Переименовываем datetime в due_date, если нужно
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'todos' AND column_name = 'datetime'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'todos' AND column_name = 'due_date'
    ) THEN
        ALTER TABLE todos RENAME COLUMN datetime TO due_date;
        -- Меняем NOT NULL на NULL, так как due_date может быть NULL
        ALTER TABLE todos ALTER COLUMN due_date DROP NOT NULL;
    END IF;
    
    -- Добавляем status, если его нет (из is_completed)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'todos' AND column_name = 'status'
    ) THEN
        ALTER TABLE todos ADD COLUMN status TEXT DEFAULT 'pending';
        -- Заполняем status из is_completed
        UPDATE todos SET status = CASE WHEN is_completed THEN 'completed' ELSE 'pending' END;
        ALTER TABLE todos ALTER COLUMN status SET NOT NULL;
    END IF;
END $$;

