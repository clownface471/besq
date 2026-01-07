package repository

import (
	"database/sql"
	"pt-besq-core/internal/database"
	"pt-besq-core/internal/entity"
)

// AuthRepository menangani database user
type AuthRepository struct{}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}

// CreateUser menyimpan user baru (Register)
func (r *AuthRepository) CreateUser(user entity.User) error {
	query := `INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`
	_, err := database.DB.Exec(query, user.Username, user.PasswordHash, user.Role)
	return err
}

// GetUserByUsername mencari data user (Login)
func (r *AuthRepository) GetUserByUsername(username string) (entity.User, error) {
	var user entity.User
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	
	err := database.DB.Get(&user, query, username)
	if err == sql.ErrNoRows {
		return user, nil // User tidak ditemukan, return kosong tanpa error
	}
	return user, err
}