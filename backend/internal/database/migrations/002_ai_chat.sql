-- AI shopping assistant: stored conversations, leads and demand insights.

CREATE TABLE chat_sessions (
    id             TEXT PRIMARY KEY,
    locale         TEXT        NOT NULL DEFAULT 'uz',
    first_message  TEXT        NOT NULL DEFAULT '',
    messages_count INT         NOT NULL DEFAULT 0,
    leads_count    INT         NOT NULL DEFAULT 0,
    -- Order summary the assistant proposed and is waiting to be confirmed.
    pending_order  JSONB,
    order_id       BIGINT      REFERENCES orders (id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX chat_sessions_updated_idx ON chat_sessions (updated_at DESC);

CREATE TABLE chat_messages (
    id         BIGSERIAL PRIMARY KEY,
    session_id TEXT        NOT NULL REFERENCES chat_sessions (id) ON DELETE CASCADE,
    role       TEXT        NOT NULL CHECK (role IN ('user', 'assistant')),
    content    TEXT        NOT NULL DEFAULT '',
    meta       JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX chat_messages_session_idx ON chat_messages (session_id, id);

-- Products customers asked the assistant for that the catalog doesn't have.
CREATE TABLE chat_missing_products (
    id         BIGSERIAL PRIMARY KEY,
    session_id TEXT        REFERENCES chat_sessions (id) ON DELETE SET NULL,
    name       TEXT        NOT NULL,
    details    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX chat_missing_products_name_idx ON chat_missing_products (lower(name));

ALTER TABLE orders ADD COLUMN source TEXT NOT NULL DEFAULT 'site';

ALTER TABLE messages ADD COLUMN source TEXT NOT NULL DEFAULT 'site';
ALTER TABLE messages ADD COLUMN chat_session_id TEXT REFERENCES chat_sessions (id) ON DELETE SET NULL;
