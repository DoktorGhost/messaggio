-- Удаляем индекс, если он был создан
DROP INDEX IF EXISTS idx_messages_status;

-- Удаляем таблицу messages, если она была создана
DROP TABLE IF EXISTS messages;

-- Удаляем тип ENUM message_status, если он был создан
DO $$
    BEGIN
        IF EXISTS (
            SELECT 1 FROM pg_type WHERE typname = 'message_status'
        ) THEN
            DROP TYPE message_status;
        END IF;
    END $$;