package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// RequireRoles memastikan user memiliki salah satu dari role yang diizinkan
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil role user dari Context (diset oleh AuthMiddleware sebelumnya)
		userRole := c.GetString("role")
		if userRole == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role user tidak terdeteksi"})
			c.Abort()
			return
		}

		// 2. Cek apakah role user ada di daftar allowedRoles
		isAllowed := false
		for _, role := range allowedRoles {
			if role == userRole {
				isAllowed = true
				break
			}
		}

		// 3. Vonis
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Akses Ditolak: Anda tidak memiliki izin untuk akses ini",
				"your_role": userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}