package entity

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Password     string    `json:"password,omitempty" db:"-"` // Input dari JSON, tidak disimpan ke DB
	PasswordHash string    `json:"-" db:"password_hash"`      // Disimpan di DB, tidak dikirim ke JSON
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Struct khusus untuk menangkap Request Login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Struct khusus untuk Response Login (Token)
type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}