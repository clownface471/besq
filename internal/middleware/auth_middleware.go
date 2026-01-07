package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware adalah Satpam yang mengecek token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil Header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Butuh token autentikasi"})
			c.Abort()
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token salah (Gunakan Bearer)"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. Validasi Token
		// (Pastikan secret key sama dengan yang di pkg/auth/token_manager.go)
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		if len(secretKey) == 0 {
			secretKey = []byte("RAHASIA_DAPUR_PT_BESQ") // Fallback key
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan metode signing-nya HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secretKey, nil
		})

		// 4. Cek validitas
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau expired"})
			c.Abort()
			return
		}

		// 5. Ekstrak data user dari token (Claims)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Simpan info user ke Context agar bisa dipakai di Handler
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token claims invalid"})
			c.Abort()
			return
		}

		// 6. Lanjut ke Handler berikutnya
		c.Next()
	}
}