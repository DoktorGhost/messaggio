DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_type WHERE typname = 'message_status'
        ) THEN
            -- Создаем тип ENUM, если его еще нет
            CREATE TYPE message_status AS ENUM ('pending', 'processed');
        END IF;
    END $$;

CREATE TABLE IF NOT EXISTS messages (
                                        id SERIAL PRIMARY KEY,
                                        content TEXT NOT NULL,
                                        status message_status NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);