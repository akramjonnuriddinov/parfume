# 1. Prettier

```bash
pnpm exec prettier . --write
```

# 2. Ishga tushirish (dev)

```bash
# 1) PostgreSQL (docker, port 5442)
cd ../backend && docker compose up -d

# 2) API (port 8089) — migratsiya, demo ma'lumotlar va admin avtomatik yaratiladi
cd ../backend && go run .

# 3) Frontend (Vite proxy: /api va /uploads -> :8089)
cd ../frontend && pnpm dev
```

- Sayt: http://localhost:5175
- Admin: http://localhost:5175/admin — login `admin`, parol `secret`

# 3. Tuzilma

- `src/pages` — ommaviy sahifalar (bosh sahifa, katalog, mahsulot, savat, biz haqimizda, yetkazib berish, bog‘lanish)
- `src/admin` — admin panel (dashboard, buyurtmalar, atirlar (hajm va narxlar bilan), hid oilalari, xabarlar, sozlamalar)
- `src/i18n/locales` — tarjimalar: `uz.ts` (asosiy), `ru.ts`, `en.ts`
- `src/stores` — Pinia: savat (localStorage), auth, sozlamalar, kategoriyalar
- `src/api` — API klient va tiplar

# 4. Production

```bash
pnpm build            # dist/
# backend/.env: STATIC_DIR=../frontend/dist — Go server saytni ham o‘zi beradi
```
