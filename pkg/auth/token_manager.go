package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// SECRET_KEY harusnya ditaruh di .env, tapi buat dev kita hardcode dulu gpp
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// HashPassword mengubah "rahasia123" menjadi "$2a$10$..."
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash mencocokkan input user dengan hash di database
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken membuat "Tiket Masuk" (JWT) yang berlaku 24 jam
func GenerateToken(userID int, username, role string) (string, error) {
	// Set secret default kalau di .env kosong
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("RAHASIA_DAPUR_PT_BESQ") 
	}

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Expired 1 hari
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}