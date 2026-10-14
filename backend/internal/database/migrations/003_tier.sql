-- Quality tier of the scent as the shop sells it (see the channel's pinned post):
-- premium, standart, luxcopy, tester. Empty = not specified.
ALTER TABLE products ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT ''
    CHECK (tier IN ('', 'premium', 'standart', 'luxcopy', 'tester'));

CREATE INDEX IF NOT EXISTS products_tier_idx ON products (tier);
