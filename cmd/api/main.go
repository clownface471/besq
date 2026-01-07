package main

import (
	"log"
	"net/http"
	"pt-besq-core/internal/database"
	"pt-besq-core/internal/handler"
	"pt-besq-core/internal/middleware" // Middleware Security & Audit
	"pt-besq-core/internal/repository"
	"pt-besq-core/internal/websocket" // Fitur Realtime
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables (.env)
	if err := godotenv.Load("../../../.env"); err != nil {
		_ = godotenv.Load(".env")
	}

	// 2. Inisialisasi Database Connection Pool
	database.InitDB()

	// 3. Setup Gin Router
	r := gin.Default()

	// --- KONFIGURASI CORS ---
	// Mengizinkan Frontend mengakses API ini
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Ubah ke domain spesifik saat production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"}, // Content-Disposition penting buat download file!
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- SETUP WEBSOCKET HUB ---
	hub := websocket.NewHub()
	go hub.Run()

	// --- SETUP HANDLERS (Dependency Injection) ---
	authHandler := handler.NewAuthHandler()       // Auth (Login/Register)
	wfHandler := handler.NewWorkflowHandler()     // Workflow (Diagram)
	insHandler := handler.NewInstanceHandler(hub) // Instance (Input, History, Export)
	wsHandler := handler.NewWSHandler(hub)        // WebSocket
	tmplHandler := handler.NewTemplateHandler()   // Template (Menu & Form)
	dashHandler := handler.NewDashboardHandler()  // Dashboard Stats
	auditHandler := handler.NewAuditHandler()     // Audit Logs (CCTV Aktivitas) [BARU]

	// --- DEFINISI RUTE API (ENDPOINTS) ---

	// A. PUBLIC ROUTES (Tanpa Login)
	public := r.Group("/api")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
		
		// Health Check
		public.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "online",
				"system": "PT Besq Factory Core v1.3 (Enterprise)",
				"time":   time.Now(),
			})
		})
	}

	// Route WebSocket (Handshake)
	r.GET("/ws", wsHandler.HandleConnections)

	// B. PROTECTED ROUTES (Wajib Login & Bawa Token)
	protected := r.Group("/api")
	
	// 1. Pasang Satpam (Cek Token JWT)
	protected.Use(middleware.AuthMiddleware())
	
	// 2. Pasang CCTV (Rekam Aktivitas User) [BARU]
	// Middleware ini akan mencatat siapa user-nya, ngapain, dan dari IP mana
	protected.Use(middleware.AuditLogger()) 
	{
		// --- AREA UMUM (Semua User Login) ---
		
		// 1. Dashboard & Statistik
		protected.GET("/dashboard/stats", dashHandler.GetStats)

		// 2. Master Data Template
		protected.GET("/templates", func(c *gin.Context) {
			data, err := repository.GetAllTemplates()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"data": data})
		})
		protected.GET("/templates/:id/fields", tmplHandler.GetFields)

		// 3. Workflow (Read Only)
		protected.GET("/workflows", wfHandler.GetList)

		// --- C. ADMIN ONLY ROUTES ---
		// Hanya Admin yang boleh ubah struktur pabrik & lihat log
		adminOnly := protected.Group("/")
		adminOnly.Use(middleware.RequireRoles("admin"))
		{
			// Edit Diagram
			adminOnly.POST("/workflows", wfHandler.Create)
			adminOnly.PUT("/workflows/:id/layout", wfHandler.UpdateLayout)
			
			// Lihat Audit Logs (Rekaman CCTV) [BARU]
			adminOnly.GET("/audit-logs", auditHandler.GetLogs)
		}

		// --- D. WRITE ACCESS (Admin & Operator) ---
		// Operator & Admin boleh input dan kelola data produksi
		writeAccess := protected.Group("/")
		writeAccess.Use(middleware.RequireRoles("admin", "operator"))
		{
			// Input Data Baru
			writeAccess.POST("/instances", insHandler.CreateInstance)
			
			// Lihat History Data (Pagination)
			writeAccess.GET("/instances", insHandler.GetList)

			// Export Excel
			writeAccess.GET("/instances/export", insHandler.ExportExcel)
		}
	}

	// 4. Jalankan Server
	log.Println("🚀 Server running on port 8080 (Features: Full Suite + Audit Logging)...")
	r.Run(":8080")
}