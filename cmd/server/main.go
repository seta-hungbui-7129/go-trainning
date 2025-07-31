package main

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"

	"seta-training/api/graphql/generated"
	"seta-training/api/graphql/resolvers"
	"seta-training/internal/config"
	"seta-training/internal/database"
	"seta-training/internal/handlers"
	"seta-training/internal/middleware"
	"seta-training/internal/repositories"
	"seta-training/internal/services"
	"seta-training/pkg/auth"
	"seta-training/pkg/logger"
	"seta-training/pkg/metrics"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logging
	logger.InitGlobalLogger(cfg.Logging.Level, cfg.Logging.Format, nil)
	appLogger := logger.GetLogger()

	// Initialize metrics
	appMetrics := metrics.InitGlobalMetrics()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", logger.Error(err))
	}
	defer db.Close()

	appLogger.Info("Database connection established")

	// Run migrations
	if err := db.Migrate(); err != nil {
		appLogger.Fatal("Failed to run migrations", logger.Error(err))
	}

	appLogger.Info("Database migrations completed")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiryHours)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.DB)
	teamRepo := repositories.NewTeamRepository(db.DB)
	folderRepo := repositories.NewFolderRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)

	// Initialize services
	userService := services.NewUserService(userRepo, jwtManager)
	teamService := services.NewTeamService(teamRepo, userRepo)
	folderService := services.NewFolderService(folderRepo, noteRepo)
	noteService := services.NewNoteService(noteRepo, folderRepo)
	importService := services.NewImportService(userService, appLogger)

	// Initialize handlers
	teamHandler := handlers.NewTeamHandler(teamService)
	folderHandler := handlers.NewFolderHandler(folderService)
	noteHandler := handlers.NewNoteHandler(noteService)
	assetHandler := handlers.NewAssetHandler(folderService, noteService, teamService)
	importHandler := handlers.NewImportHandler(importService, appLogger, appMetrics)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize GraphQL resolver
	resolver := &resolvers.Resolver{
		UserService: userService,
	}

	// Create GraphQL server
	gqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	// Initialize Gin router
	router := gin.Default()

	// Add metrics middleware
	router.Use(appMetrics.PrometheusMiddleware())

	// Add logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		appLogger.Info("HTTP Request",
			logger.String("method", param.Method),
			logger.String("path", param.Path),
			logger.Int("status", param.StatusCode),
			logger.Duration("latency", param.Latency),
			logger.String("client_ip", param.ClientIP),
		)
		return ""
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			appLogger.Error("Health check failed", logger.Error(err))
			appMetrics.RecordError("database", "health_check")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(appMetrics.Handler()))

	// GraphQL endpoints
	router.POST("/graphql", gin.WrapH(gqlServer))
	if cfg.GraphQL.Playground {
		router.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	}

	// REST API routes
	api := router.Group("/api/v1")
	{
		// Team management routes (require authentication)
		teams := api.Group("/teams")
		teams.Use(authMiddleware.RequireAuth())
		{
			teams.POST("", authMiddleware.RequireManager(), teamHandler.CreateTeam)
			teams.GET("/:teamId", teamHandler.GetTeam)
			teams.GET("", teamHandler.GetAllTeams)
			teams.POST("/:teamId/members", authMiddleware.RequireManager(), teamHandler.AddMember)
			teams.DELETE("/:teamId/members/:memberId", authMiddleware.RequireManager(), teamHandler.RemoveMember)
			teams.POST("/:teamId/managers", authMiddleware.RequireManager(), teamHandler.AddManager)
			teams.DELETE("/:teamId/managers/:managerId", authMiddleware.RequireManager(), teamHandler.RemoveManager)
		}

		// Folder management routes (require authentication)
		folders := api.Group("/folders")
		folders.Use(authMiddleware.RequireAuth())
		{
			folders.POST("", folderHandler.CreateFolder)
			folders.GET("/:folderId", folderHandler.GetFolder)
			folders.PUT("/:folderId", folderHandler.UpdateFolder)
			folders.DELETE("/:folderId", folderHandler.DeleteFolder)
			folders.POST("/:folderId/share", folderHandler.ShareFolder)
			folders.DELETE("/:folderId/share/:userId", folderHandler.RevokeShare)
			folders.POST("/:folderId/notes", noteHandler.CreateNote)
		}

		// Note management routes (require authentication)
		notes := api.Group("/notes")
		notes.Use(authMiddleware.RequireAuth())
		{
			notes.GET("/:noteId", noteHandler.GetNote)
			notes.PUT("/:noteId", noteHandler.UpdateNote)
			notes.DELETE("/:noteId", noteHandler.DeleteNote)
			notes.POST("/:noteId/share", noteHandler.ShareNote)
			notes.DELETE("/:noteId/share/:userId", noteHandler.RevokeShare)
		}

		// Asset viewing routes (require authentication)
		api.GET("/users/:userId/assets", authMiddleware.RequireAuth(), assetHandler.GetUserAssets)
		api.GET("/teams/:teamId/assets", authMiddleware.RequireAuth(), authMiddleware.RequireManager(), assetHandler.GetTeamAssets)

		// Import routes (require authentication and manager role)
		api.POST("/import-users", authMiddleware.RequireAuth(), authMiddleware.RequireManager(), importHandler.ImportUsers)
		api.GET("/import-users/template", authMiddleware.RequireAuth(), importHandler.GetImportTemplate)
		api.GET("/import-users/status", authMiddleware.RequireAuth(), authMiddleware.RequireManager(), importHandler.GetImportStatus)
	}

	appLogger.Info("Server starting",
		logger.String("port", cfg.Server.Port),
		logger.String("mode", cfg.Server.GinMode),
	)
	appLogger.Info("GraphQL Playground available", logger.String("url", "http://localhost:"+cfg.Server.Port+"/playground"))
	appLogger.Info("Health check available", logger.String("url", "http://localhost:"+cfg.Server.Port+"/health"))
	appLogger.Info("Metrics available", logger.String("url", "http://localhost:"+cfg.Server.Port+"/metrics"))

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		appLogger.Fatal("Failed to start server", logger.Error(err))
	}
}
