package database

import (
	"clean-api/domain"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(databaseURL string) {
	if !strings.Contains(databaseURL, "?") {
		databaseURL += "?"
	} else {
		databaseURL += "&"
	}
	databaseURL += "allowNativePasswords=true"

	db, err := gorm.Open(mysql.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	log.Println("Database connected successfully")
}

func Migrate() {
	if DB == nil {
		log.Fatal("Database connection is not initialized")
	}

	err := DB.AutoMigrate(
		&domain.Branch{},
		&domain.User{},
		&domain.Donation{},
		&domain.Inventory{},
		&domain.Delivery{},
		&domain.BranchRequest{},
		&domain.MasterItem{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Database migration completed successfully")

	SeedMasterItems(DB)
}

func SeedMasterItems(db *gorm.DB) {
	items := []domain.MasterItem{
		{SKUCode: "SKU-SMB-001", Name: "Beras", Category: "Bahan Pokok & Sembako", Unit: "Kg", Description: "Beras putih kualitas medium/super"},
		{SKUCode: "SKU-SMB-002", Name: "Minyak Goreng", Category: "Bahan Pokok & Sembako", Unit: "Liter", Description: "Minyak goreng kemasan pouch/botol"},
		{SKUCode: "SKU-SMB-003", Name: "Gula Pasir", Category: "Bahan Pokok & Sembako", Unit: "Kg", Description: "Gula pasir konsumsi"},
		{SKUCode: "SKU-SMB-004", Name: "Terigu", Category: "Bahan Pokok & Sembako", Unit: "Kg", Description: "Tepung terigu protein sedang"},
		{SKUCode: "SKU-SMB-005", Name: "Kecap (600 Ml)", Category: "Bahan Pokok & Sembako", Unit: "Pouch", Description: "Kecap manis refill 600ml"},
		{SKUCode: "SKU-SMB-006", Name: "Blue Band / Margarin", Category: "Bahan Pokok & Sembako", Unit: "Sachet", Description: "Margarin sachet/serbaguna"},
		
		{SKUCode: "SKU-MKN-001", Name: "Indomie / Mi Instan", Category: "Makanan & Minuman", Unit: "Dus", Description: "Mie instan goreng/kuah per dus"},
		{SKUCode: "SKU-MKN-002", Name: "Sarden Kaleng", Category: "Makanan & Minuman", Unit: "Kaleng", Description: "Sarden konsumsi kaleng kecil/besar"},
		{SKUCode: "SKU-MKN-003", Name: "Susu Kotak UHT", Category: "Makanan & Minuman", Unit: "Kardus", Description: "Susu siap minum UHT per kardus"},
		{SKUCode: "SKU-MKN-004", Name: "Susu Kaleng", Category: "Makanan & Minuman", Unit: "Kaleng", Description: "Susu kental manis/bubuk kaleng"},
		{SKUCode: "SKU-MKN-005", Name: "Sirup", Category: "Makanan & Minuman", Unit: "Botol", Description: "Sirup rasa manis kemasan botol"},
		{SKUCode: "SKU-MKN-006", Name: "Snack / Makanan Ringan", Category: "Makanan & Minuman", Unit: "Kardus", Description: "Paket biskuit/makanan ringan kemasan kardus"},

		{SKUCode: "SKU-KBS-001", Name: "Rinso / Deterjen", Category: "Kebersihan & Sanitasi", Unit: "Kg", Description: "Deterjen bubuk cuci pakaian"},
		{SKUCode: "SKU-KBS-002", Name: "Molto (800 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Pouch", Description: "Pelembut & pewangi pakaian refill"},
		{SKUCode: "SKU-KBS-003", Name: "Sabun Cair (600 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Pouch", Description: "Sabun mandi cair refill"},
		{SKUCode: "SKU-KBS-004", Name: "Sabun Ekonomi (800 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Pouch", Description: "Sabun cuci piring/peralatan dapur"},
		{SKUCode: "SKU-KBS-005", Name: "Super Pel (800 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Pouch", Description: "Pembersih lantai refill"},
		{SKUCode: "SKU-KBS-006", Name: "Prostex (500 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Botol", Description: "Pembersih porselen/kloset botol"},
		{SKUCode: "SKU-KBS-007", Name: "Karbol (800 Ml)", Category: "Kebersihan & Sanitasi", Unit: "Pouch", Description: "Cairan desinfektan karbol lantai"},

		{SKUCode: "SKU-DIR-001", Name: "Tissu Kering (250 Lembar)", Category: "Perawatan Diri", Unit: "Pack", Description: "Tissue wajah/kering pack"},
		{SKUCode: "SKU-DIR-002", Name: "Pasta Gigi Dewasa", Category: "Perawatan Diri", Unit: "Pcs", Description: "Pasta gigi kemasan tube"},
		{SKUCode: "SKU-DIR-003", Name: "Sikat Gigi Dewasa", Category: "Perawatan Diri", Unit: "Pcs", Description: "Sikat gigi dewasa"},
		{SKUCode: "SKU-DIR-004", Name: "Cotton Bud (100 Pcs)", Category: "Perawatan Diri", Unit: "Pack", Description: "Pembersih telinga cotton bud pack"},
	}

	for _, item := range items {
		var existing domain.MasterItem
		if err := db.Where("sku_code = ?", item.SKUCode).First(&existing).Error; err != nil {
			db.Create(&item)
		}
	}
	log.Println("Seeded real-world Master Items into database successfully!")
}
