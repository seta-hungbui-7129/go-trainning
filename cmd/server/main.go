package main

import (
	"log"
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
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiryHours)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.DB)
	teamRepo := repositories.NewTeamRepository(db.DB)
	// folderRepo := repositories.NewFolderRepository(db.DB)
	// noteRepo := repositories.NewNoteRepository(db.DB)

	// Initialize services
	userService := services.NewUserService(userRepo, jwtManager)
	teamService := services.NewTeamService(teamRepo, userRepo)
	// folderService := services.NewFolderService(folderRepo, noteRepo)
	// noteService := services.NewNoteService(noteRepo, folderRepo)

	// Initialize handlers
	teamHandler := handlers.NewTeamHandler(teamService)

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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
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

		// TODO: Add folder and note routes
		// folders := api.Group("/folders")
		// folders.Use(authMiddleware.RequireAuth())
		// {
		//     folders.POST("", folderHandler.CreateFolder)
		//     folders.GET("/:folderId", folderHandler.GetFolder)
		//     folders.PUT("/:folderId", folderHandler.UpdateFolder)
		//     folders.DELETE("/:folderId", folderHandler.DeleteFolder)
		//     folders.POST("/:folderId/share", folderHandler.ShareFolder)
		//     folders.DELETE("/:folderId/share/:userId", folderHandler.RevokeShare)
		//     folders.POST("/:folderId/notes", noteHandler.CreateNote)
		// }

		// notes := api.Group("/notes")
		// notes.Use(authMiddleware.RequireAuth())
		// {
		//     notes.GET("/:noteId", noteHandler.GetNote)
		//     notes.PUT("/:noteId", noteHandler.UpdateNote)
		//     notes.DELETE("/:noteId", noteHandler.DeleteNote)
		//     notes.POST("/:noteId/share", noteHandler.ShareNote)
		//     notes.DELETE("/:noteId/share/:userId", noteHandler.RevokeShare)
		// }
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("GraphQL Playground available at http://localhost:%s/playground", cfg.Server.Port)
	log.Printf("Health check available at http://localhost:%s/health", cfg.Server.Port)

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
