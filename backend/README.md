# Parfume Shop API (Go + PostgreSQL)

```bash
docker compose up -d   # PostgreSQL 17 -> localhost:5442 (user/pass/db: parfume)
go run .               # http://localhost:8089
```

On start the server applies SQL migrations (`internal/database/migrations`), seeds a demo perfume
catalog (5 fragrance families, 12 perfumes with bottle and decant volumes) and contacts if the DB is empty,
and creates the admin from `ADMIN_USERNAME` / `ADMIN_PASSWORD` (default `admin` / `secret`).

Config: `.env` (see `.env.example`). Set a long random `JWT_SECRET` in production and
`STATIC_DIR=../frontend/dist` to serve the built frontend from the same server.

## Data model

- A **product** is one perfume: brand, gender (`women`/`men`/`unisex`), concentration
  (`edp`, `edt`, `parfum`, `extrait`, `cologne`, `oil`), notes (top/heart/base, comma-separated), family (category).
- Each product has **variants** — the volumes it is sold in: full `bottle`s and `decant`s ("raspiv", 5–10 ml),
  each with its own price and stock flag. Cart lines and order items are per variant.
- Order items keep a snapshot (brand, name, volume, price) so editing a product never changes past orders.

## Endpoints

Public:
- `GET  /api/categories`, `GET /api/brands`
- `GET  /api/products?category=&gender=women|men|unisex&brand=&q=&sort=newest|price_asc|price_desc&featured=1&ids=1,2&page=&limit=`
  — `gender=women` also includes unisex; `q` searches names, brand and notes
- `GET  /api/products/{id}`
- `GET  /api/settings`
- `POST /api/orders` `{full_name, phone, comment, locale, items:[{variant_id, quantity}]}` — prices are computed server-side
- `POST /api/messages` `{full_name, phone, message, product_id?}`

Admin (`Authorization: Bearer <token>`):
- `POST /api/admin/login`, `GET /api/admin/me`, `POST /api/admin/password`
- `GET /api/admin/stats`, `POST /api/admin/uploads` (multipart `file`, JPG/PNG/WEBP/GIF ≤ 5MB)
- CRUD `/api/admin/categories`, `/api/admin/products` (product body includes `variants: [{id?, volume_ml, kind, price, in_stock}]`)
- `GET/PATCH/DELETE /api/admin/orders[/{id}]` — PATCH `{status: new|processing|completed|cancelled, admin_note}`
- `GET/PATCH/DELETE /api/admin/messages[/{id}]`
- `GET/PUT /api/admin/settings`

Public POST endpoints are rate-limited (20/min per IP), login 10/min per IP.
