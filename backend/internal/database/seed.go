package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type seedCategory struct {
	slug, nameUz, nameRu, nameEn, descUz, descRu, descEn string
}

type seedVariant struct {
	ml    int
	kind  string
	price int64
	stock bool
}

type seedProduct struct {
	category, brand, name, gender, conc string
	descUz, descRu, descEn              string
	top, heart, base                    string
	image                               string
	featured                            bool
	variants                            []seedVariant
}

var seedCategories = []seedCategory{
	{"gulli", "Gulli", "Цветочные", "Floral",
		"Atirgul, yasmin va oq gullar", "Роза, жасмин и белые цветы", "Rose, jasmine and white flowers"},
	{"sharqona", "Sharqona", "Восточные", "Oriental",
		"Amber, za’faron, ud va ziravorlar", "Амбра, шафран, уд и пряности", "Amber, saffron, oud and spices"},
	{"yogochli", "Yog‘ochli", "Древесные", "Woody",
		"Sadr, vetiver va sandal", "Кедр, ветивер и сандал", "Cedar, vetiver and sandalwood"},
	{"yangi", "Yangi va sitrusli", "Свежие и цитрусовые", "Fresh & citrus",
		"Bergamot, dengiz va yashil notalar", "Бергамот, морские и зелёные ноты", "Bergamot, marine and green notes"},
	{"shirin", "Shirin (gurman)", "Сладкие (гурманские)", "Gourmand",
		"Vanil, karamel va pralin", "Ваниль, карамель и пралине", "Vanilla, caramel and praline"},
}

func dec(ml int, price int64) seedVariant    { return seedVariant{ml, "decant", price, true} }
func bottle(ml int, price int64) seedVariant { return seedVariant{ml, "bottle", price, true} }

var seedProducts = []seedProduct{
	{"yangi", "Dior", "Sauvage", "men", "edp",
		"Yangi va kuchli erkaklar atiri: kalabriya bergamoti, ziravorlar va ambroksan. Har kuni va kechki uchrashuvlar uchun.",
		"Свежий и мощный мужской аромат: калабрийский бергамот, пряности и амброксан. На каждый день и вечер.",
		"A fresh, powerful men’s scent: Calabrian bergamot, spices and ambroxan. For every day and evenings out.",
		"Bergamot, qalampir", "Lavanda, Sichuan qalampiri", "Ambroksan, vanil",
		"/images/perfumes/blue.svg", true,
		[]seedVariant{dec(10, 90000), bottle(60, 1450000), bottle(100, 1950000)}},
	{"gulli", "Chanel", "Coco Mademoiselle", "women", "edp",
		"Nafis va jasur: apelsin, atirgul va pachuli. Ofis va bayramlar uchun klassika.",
		"Изысканный и смелый: апельсин, роза и пачули. Классика для офиса и праздников.",
		"Elegant and bold: orange, rose and patchouli. A classic for office and celebrations.",
		"Apelsin, bergamot", "Atirgul, yasmin", "Pachuli, vetiver",
		"/images/perfumes/rose.svg", true,
		[]seedVariant{dec(10, 110000), bottle(50, 1750000), bottle(100, 2400000)}},
	{"shirin", "Lattafa", "Khamrah", "unisex", "edp",
		"Issiq va shirin sharqona atir: xurmo, pralin va dolchin. Qishki kechalar uchun ideal.",
		"Тёплый сладкий восточный аромат: финики, пралине и корица. Идеален для зимних вечеров.",
		"A warm, sweet oriental: dates, praline and cinnamon. Perfect for winter evenings.",
		"Dolchin, muskat yong‘og‘i, bergamot", "Xurmo, pralin, tuberoza", "Vanil, tonka, amber",
		"/images/perfumes/amber.svg", true,
		[]seedVariant{dec(10, 35000), bottle(100, 420000)}},
	{"yogochli", "Armaf", "Club de Nuit Intense Man", "men", "edt",
		"Ananas, qayin va mushk — hamyonbop narxda mashhur tutunli-mevali atir.",
		"Ананас, берёза и мускус — популярный дымно-фруктовый аромат по доступной цене.",
		"Pineapple, birch and musk — a popular smoky-fruity scent at an affordable price.",
		"Limon, ananas, bergamot", "Qayin, yasmin", "Mushk, ambra, vanil",
		"/images/perfumes/black.svg", false,
		[]seedVariant{dec(10, 30000), bottle(105, 390000)}},
	{"sharqona", "Maison Francis Kurkdjian", "Baccarat Rouge 540", "unisex", "edp",
		"Za’faron, yasmin va amber daraxti: havoda iz qoldiradigan, tanilgan hashamatli atir.",
		"Шафран, жасмин и амбровое дерево: роскошный узнаваемый аромат с ярким шлейфом.",
		"Saffron, jasmine and amberwood: a recognisable luxury scent with a lasting trail.",
		"Za’faron, yasmin", "Amber daraxti, ambra", "Archa smolasi, sadr",
		"/images/perfumes/gold.svg", true,
		[]seedVariant{dec(5, 160000), dec(10, 300000), bottle(70, 4200000)}},
	{"gulli", "Yves Saint Laurent", "Libre", "women", "edp",
		"Lavanda va apelsin guli: erkin va zamonaviy ayol uchun.",
		"Лаванда и цветок апельсина: для свободной и современной женщины.",
		"Lavender and orange blossom: for a free, modern woman.",
		"Lavanda, mandarin", "Apelsin guli, yasmin", "Vanil, sadr, mushk",
		"/images/perfumes/gold.svg", true,
		[]seedVariant{dec(10, 100000), bottle(50, 1650000), bottle(90, 2250000)}},
	{"sharqona", "Tom Ford", "Tobacco Vanille", "unisex", "edp",
		"Tamaki bargi, vanil va kakao — boy va iliq, kechki atir.",
		"Табачный лист, ваниль и какао — богатый и тёплый вечерний аромат.",
		"Tobacco leaf, vanilla and cacao — a rich, warm evening scent.",
		"Tamaki bargi, ziravorlar", "Vanil, kakao, tonka", "Quritilgan mevalar, yog‘och",
		"/images/perfumes/amber.svg", false,
		[]seedVariant{dec(5, 140000), bottle(50, 3900000)}},
	{"shirin", "Kayali", "Vanilla 28", "women", "edp",
		"Vanil orxideyasi va jigarrang shakar: yumshoq, shirin va jozibali.",
		"Ванильная орхидея и коричневый сахар: мягкий, сладкий и притягательный.",
		"Vanilla orchid and brown sugar: soft, sweet and alluring.",
		"Vanil orxideyasi", "Jigarrang shakar, tonka", "Amber, mushk",
		"/images/perfumes/rose.svg", true,
		[]seedVariant{dec(10, 90000), bottle(50, 1300000)}},
	{"shirin", "Al Rehab", "Choco Musk", "unisex", "oil",
		"Spirtsiz moyli atir (roll-on): shokolad, vanil va mushk. Uzoq saqlanadi, kichik va qulay.",
		"Масляные духи без спирта (роллер): шоколад, ваниль и мускус. Стойкие, маленькие и удобные.",
		"Alcohol-free oil perfume (roll-on): chocolate, vanilla and musk. Long-lasting and pocket-sized.",
		"Shokolad", "Vanil, atirgul", "Mushk, sandal",
		"/images/perfumes/oil.svg", true,
		[]seedVariant{bottle(6, 45000)}},
	{"yangi", "Jo Malone", "Wood Sage & Sea Salt", "unisex", "cologne",
		"Dengiz tuzi va shalfey: yengil, tabiiy, yozgi kunlar uchun.",
		"Морская соль и шалфей: лёгкий, естественный, для летних дней.",
		"Sea salt and sage: light and natural, for summer days.",
		"Ambrette urug‘i", "Dengiz tuzi", "Shalfey, greypfrut",
		"/images/perfumes/blue.svg", false,
		[]seedVariant{dec(10, 80000), bottle(100, 2100000)}},
	{"yangi", "Versace", "Eros", "men", "edt",
		"Yalpiz, yashil olma va vanil — yorqin va energik.",
		"Мята, зелёное яблоко и ваниль — яркий и энергичный.",
		"Mint, green apple and vanilla — bright and energetic.",
		"Yalpiz, yashil olma, limon", "Tonka, geran", "Vanil, vetiver, sadr",
		"/images/perfumes/blue.svg", true,
		[]seedVariant{dec(10, 60000), bottle(100, 1150000)}},
	{"shirin", "Lancôme", "La Vie Est Belle", "women", "edp",
		"Iris, pralin va vanil: baxtni his qildiradigan shirin gulli atir.",
		"Ирис, пралине и ваниль: сладкий цветочный аромат счастья.",
		"Iris, praline and vanilla: a sweet floral scent of happiness.",
		"Qora smorodina, nok", "Iris, yasmin", "Pralin, vanil, pachuli",
		"/images/perfumes/rose.svg", false,
		[]seedVariant{dec(10, 85000), bottle(50, 1450000), {100, "bottle", 1950000, false}}},
}

var seedSettings = map[string]string{
	"phone":         "+998 90 777 40 40",
	"email":         "info@marianaparfumes.uz",
	"telegram_url":  "https://t.me/marianaparfumes",
	"instagram_url": "https://instagram.com/marianaparfumes",
	"address_uz":    "Toshkent sh., Yunusobod tumani, Amir Temur ko‘chasi, 107",
	"address_ru":    "г. Ташкент, Юнусабадский район, ул. Амира Темура, 107",
	"address_en":    "107 Amir Temur street, Yunusabad district, Tashkent",
	"work_hours_uz": "Har kuni: 10:00–21:00",
	"work_hours_ru": "Ежедневно: 10:00–21:00",
	"work_hours_en": "Every day: 10:00–21:00",
}

// Seed fills an empty database with a demo catalog and contacts.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM categories`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ids := map[string]int64{}
	for i, c := range seedCategories {
		var id int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO categories (slug, name_uz, name_ru, name_en, description_uz, description_ru, description_en, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			c.slug, c.nameUz, c.nameRu, c.nameEn, c.descUz, c.descRu, c.descEn, i+1).Scan(&id); err != nil {
			return err
		}
		ids[c.slug] = id
	}

	for i, p := range seedProducts {
		var id int64
		// Perfume names are the same in every language.
		if err := tx.QueryRow(ctx, `
			INSERT INTO products (category_id, name_uz, name_ru, name_en, description_uz, description_ru,
				description_en, brand, gender, concentration, notes_top, notes_heart, notes_base,
				image_url, is_featured, sort_order)
			VALUES ($1, $2, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
			ids[p.category], p.name, p.descUz, p.descRu, p.descEn, p.brand, p.gender, p.conc,
			p.top, p.heart, p.base, p.image, p.featured, i+1).Scan(&id); err != nil {
			return err
		}
		for j, v := range p.variants {
			if _, err := tx.Exec(ctx, `
				INSERT INTO product_variants (product_id, volume_ml, kind, price, in_stock, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6)`, id, v.ml, v.kind, v.price, v.stock, j); err != nil {
				return err
			}
		}
	}

	for k, v := range seedSettings {
		if _, err := tx.Exec(ctx, `INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT DO NOTHING`, k, v); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
