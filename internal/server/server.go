package server

import (
	"net/http"
	"time"
	"tp25-api/internal/config"
	"tp25-api/internal/domain"
	"tp25-api/internal/handler"
	"tp25-api/internal/middleware"
	"tp25-api/internal/repository/mongodb"
	"tp25-api/internal/service"
	"tp25-api/lib/database"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, db *database.MongoDB) *gin.Engine {
	userRepo := mongodb.NewUserRepository(db.Database)
	zoneRepo := mongodb.NewZoneRepository(db.Database)
	sensorRepo := mongodb.NewSensorRepository(db.Database)
	settingRepo := mongodb.NewSettingRepository(db.Database)

	userService := service.NewUserService(userRepo, cfg.Auth.JWTSecret)
	zoneService := service.NewZoneService(zoneRepo)
	sensorService := service.NewSensorService(sensorRepo, zoneRepo)
	settingService := service.NewSettingService(settingRepo)

	authHandler := handler.NewAuthHandler(userService, cfg)
	userHandler := handler.NewUserHandler(userService)
	zoneHandler := handler.NewZoneHandler(zoneService)
	sensorHandler := handler.NewSensorHandler(sensorService)
	settingHandler := handler.NewSettingHandler(settingService)

	authMiddleware := middleware.NewAuthMiddleware(cfg, userService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(middleware.CORS())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	router.Static("/docs", "./docs")

	router.GET("/api-docs", func(c *gin.Context) {
		c.File("docs/swagger.html")
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware.Auth(), authHandler.Logout)
			auth.GET("/profile", authMiddleware.Auth(), authHandler.GetProfile)
			auth.PUT("/password", authMiddleware.Auth(), authHandler.SetPassword)
		}

		users := api.Group("/users")
		users.Use(authMiddleware.Auth(), authMiddleware.RequireRole(domain.RoleAdmin))
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.PUT("/:id/password", userHandler.SetUserPassword)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		zones := api.Group("/zones")
		zones.Use(authMiddleware.Auth())
		{
			zones.GET("", zoneHandler.ListZones)
			zones.POST("", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.CreateZone)
			zones.GET("/reports", zoneHandler.ReportByMetric)
			zones.GET("/:id", zoneHandler.GetZone)
			zones.PUT("/:id", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.UpdateZone)
			zones.GET("/:id/groups", zoneHandler.ListGroups)
			zones.POST("/:id/groups", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.CreateGroup)
		}

		groups := api.Group("/groups")
		groups.Use(authMiddleware.Auth())
		{
			groups.GET("/:id", zoneHandler.GetGroup)
			groups.PUT("/:id", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.UpdateGroup)
			groups.DELETE("/:id", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.DeleteGroup)
			groups.GET("/:id/boxes", zoneHandler.ListBoxes)
			groups.POST("/:id/boxes", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.CreateBox)
			groups.GET("/:id/records", sensorHandler.ListRecordsByGroup)
			groups.GET("/:id/records/latest", sensorHandler.ListRecordsLatestByGroup)
		}

		boxes := api.Group("/boxes")
		boxes.Use(authMiddleware.Auth())
		{
			boxes.GET("", zoneHandler.ListAllBoxes)
			boxes.GET("/:id", zoneHandler.GetBox)
			boxes.PUT("/:id", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.UpdateBox)
			boxes.DELETE("/:id", authMiddleware.RequireRole(domain.RoleAdmin), zoneHandler.DeleteBox)
			boxes.GET("/:id/records", sensorHandler.ListRecords)
			boxes.GET("/:id/records/export", sensorHandler.ExportRecords)
			boxes.POST("/:id/records", sensorHandler.AddRecord)
			boxes.GET("/:id/reports", sensorHandler.ReportRecords)
		}

		metrics := api.Group("/metrics")
		metrics.Use(authMiddleware.Auth())
		{
			metrics.GET("", sensorHandler.ListMetrics)
			metrics.GET("/:id", sensorHandler.GetMetric)
			metrics.POST("", authMiddleware.RequireRole(domain.RoleAdmin), sensorHandler.CreateMetric)
			metrics.PUT("/:id", authMiddleware.RequireRole(domain.RoleAdmin), sensorHandler.UpdateMetric)
			metrics.DELETE("/:id", authMiddleware.RequireRole(domain.RoleAdmin), sensorHandler.DeleteMetric)
		}

		settings := api.Group("/settings")
		settings.Use(authMiddleware.Auth(), authMiddleware.RequireRole(domain.RoleAdmin))
		{
			settings.GET("", settingHandler.ListSettings)
			settings.GET("/:id", settingHandler.GetSetting)
			settings.POST("", settingHandler.CreateSetting)
			settings.PUT("/:id", settingHandler.UpdateSetting)
			settings.DELETE("/:id", settingHandler.DeleteSetting)
		}
	}

	return router
}
