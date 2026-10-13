CREATE TABLE admins (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fragrance families: floral, oriental, woody, fresh...
CREATE TABLE categories (
    id             BIGSERIAL PRIMARY KEY,
    slug           TEXT        NOT NULL UNIQUE,
    name_uz        TEXT        NOT NULL,
    name_ru        TEXT        NOT NULL DEFAULT '',
    name_en        TEXT        NOT NULL DEFAULT '',
    description_uz TEXT        NOT NULL DEFAULT '',
    description_ru TEXT        NOT NULL DEFAULT '',
    description_en TEXT        NOT NULL DEFAULT '',
    sort_order     INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id             BIGSERIAL PRIMARY KEY,
    category_id    BIGINT      REFERENCES categories (id) ON DELETE SET NULL,
    name_uz        TEXT        NOT NULL,
    name_ru        TEXT        NOT NULL DEFAULT '',
    name_en        TEXT        NOT NULL DEFAULT '',
    description_uz TEXT        NOT NULL DEFAULT '',
    description_ru TEXT        NOT NULL DEFAULT '',
    description_en TEXT        NOT NULL DEFAULT '',
    brand          TEXT        NOT NULL DEFAULT '',
    gender         TEXT        NOT NULL DEFAULT 'unisex' CHECK (gender IN ('women', 'men', 'unisex')),
    concentration  TEXT        NOT NULL DEFAULT 'edp'
        CHECK (concentration IN ('edp', 'edt', 'parfum', 'extrait', 'cologne', 'oil')),
    -- Comma-separated note lists, e.g. "Bergamot, Pink pepper"
    notes_top      TEXT        NOT NULL DEFAULT '',
    notes_heart    TEXT        NOT NULL DEFAULT '',
    notes_base     TEXT        NOT NULL DEFAULT '',
    image_url      TEXT        NOT NULL DEFAULT '',
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    is_featured    BOOLEAN     NOT NULL DEFAULT FALSE,
    sort_order     INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX products_category_idx ON products (category_id);
CREATE INDEX products_active_idx ON products (is_active, sort_order, id DESC);
CREATE INDEX products_brand_idx ON products (lower(brand));

-- Each product is sold in one or more volumes: full bottles and decants ("raspiv").
CREATE TABLE product_variants (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT  NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    volume_ml  INT     NOT NULL CHECK (volume_ml > 0 AND volume_ml <= 1000),
    kind       TEXT    NOT NULL DEFAULT 'bottle' CHECK (kind IN ('bottle', 'decant')),
    price      BIGINT  NOT NULL CHECK (price >= 0),
    in_stock   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT     NOT NULL DEFAULT 0
);

CREATE INDEX product_variants_product_idx ON product_variants (product_id, sort_order, volume_ml);

CREATE TABLE orders (
    id         BIGSERIAL PRIMARY KEY,
    full_name  TEXT        NOT NULL,
    phone      TEXT        NOT NULL,
    comment    TEXT        NOT NULL DEFAULT '',
    status     TEXT        NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'processing', 'completed', 'cancelled')),
    total      BIGINT      NOT NULL DEFAULT 0,
    admin_note TEXT        NOT NULL DEFAULT '',
    locale     TEXT        NOT NULL DEFAULT 'uz',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX orders_status_idx ON orders (status, created_at DESC);

-- Lines keep a snapshot of the product and volume as ordered.
CREATE TABLE order_items (
    id            BIGSERIAL PRIMARY KEY,
    order_id      BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id    BIGINT REFERENCES products (id) ON DELETE SET NULL,
    variant_id    BIGINT REFERENCES product_variants (id) ON DELETE SET NULL,
    product_name  TEXT   NOT NULL,
    product_brand TEXT   NOT NULL DEFAULT '',
    product_image TEXT   NOT NULL DEFAULT '',
    volume_ml     INT    NOT NULL,
    kind          TEXT   NOT NULL DEFAULT 'bottle',
    price         BIGINT NOT NULL,
    quantity      INT    NOT NULL CHECK (quantity > 0),
    subtotal      BIGINT NOT NULL
);

CREATE INDEX order_items_order_idx ON order_items (order_id);

CREATE TABLE messages (
    id           BIGSERIAL PRIMARY KEY,
    full_name    TEXT        NOT NULL,
    phone        TEXT        NOT NULL,
    message      TEXT        NOT NULL DEFAULT '',
    product_id   BIGINT      REFERENCES products (id) ON DELETE SET NULL,
    product_name TEXT        NOT NULL DEFAULT '',
    is_read      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX messages_created_idx ON messages (created_at DESC);

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
