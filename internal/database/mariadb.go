package database

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"), // Di Docker ini nilainya "db"
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	
	// --- RETRY MECHANISM (SABAR MENUNGGU) ---
	// Coba connect selama 60 detik sebelum menyerah
	// Ini penting karena Database di Docker butuh waktu booting lebih lama dari Go
	maxRetries := 30
	for i := 1; i <= maxRetries; i++ {
		log.Printf("⏳ Mencoba connect ke database (%d/%d)...", i, maxRetries)
		
		DB, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			// Tes ping untuk memastikan koneksi benar-benar hidup
			if errPing := DB.Ping(); errPing == nil {
				log.Println("✅ SUKSES: Terhubung ke Database!")
				break
			}
		}

		// Jika gagal, tunggu 2 detik lalu coba lagi
		log.Printf("⚠️ Gagal connect: %v. Retrying in 2s...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalln("❌ FATAL: Gagal connect database setelah menunggu lama. Cek log container 'db'.")
	}

	// Tuning Connection Pool
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
}