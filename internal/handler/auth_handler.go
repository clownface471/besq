package handler

import (
	"net/http"
	"pt-besq-core/internal/entity"
	"pt-besq-core/internal/repository"
	"pt-besq-core/pkg/auth" // Import helper security kita

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Repo *repository.AuthRepository
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		Repo: repository.NewAuthRepository(),
	}
}

// Register membuat user baru
func (h *AuthHandler) Register(c *gin.Context) {
	var input entity.User
	// Bind JSON (Username, Password, Role)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Hash Password (AMANKAN PASSWORD SEBELUM SIMPAN!)
	hashedPwd, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal enkripsi password"})
		return
	}
	input.PasswordHash = hashedPwd

	// 2. Set Default Role jika kosong
	if input.Role == "" {
		input.Role = "operator"
	}

	// 3. Simpan ke Database
	err = h.Repo.CreateUser(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal register (Username mungkin sudah dipakai)"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User berhasil dibuat", "username": input.Username})
}

// Login memverifikasi user dan memberi Token
func (h *AuthHandler) Login(c *gin.Context) {
	var input entity.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Cari User di Database
	user, err := h.Repo.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah"})
		return
	}

	// 2. Cek Password (Bandingkan Input vs Hash di DB)
	match := auth.CheckPasswordHash(input.Password, user.PasswordHash)
	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah"})
		return
	}

	// 3. Generate JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate token"})
		return
	}

	// 4. Kirim Token ke Client
	c.JSON(http.StatusOK, entity.LoginResponse{
		Token: token,
		Role:  user.Role,
	})
}