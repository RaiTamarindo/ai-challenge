package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/feature-voting-platform/backend/internal/config"
	"github.com/feature-voting-platform/backend/internal/handlers"
	"github.com/feature-voting-platform/backend/internal/middleware"
	"github.com/feature-voting-platform/backend/internal/repository"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Feature Voting Platform API
// @version 1.0
// @description A REST API for a feature voting platform where users can propose, list, and vote on features.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@feature-voting-platform.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abc123xyz"

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	featureRepo := repository.NewFeatureRepository(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWT.Secret)
	featureHandler := handlers.NewFeatureHandler(featureRepo)
	voteHandler := handlers.NewVoteHandler(featureRepo)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Feature Voting Platform API is running",
		})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.AuthMiddleware(cfg.JWT.Secret), authHandler.GetProfile)
		}

		// Feature routes
		features := v1.Group("/features")
		{
			// Public routes (with optional auth for vote status)
			features.GET("", middleware.OptionalAuthMiddleware(cfg.JWT.Secret), featureHandler.GetFeatures)
			features.GET("/:id", middleware.OptionalAuthMiddleware(cfg.JWT.Secret), featureHandler.GetFeature)

			// Protected routes
			features.POST("", middleware.AuthMiddleware(cfg.JWT.Secret), featureHandler.CreateFeature)
			features.PUT("/:id", middleware.AuthMiddleware(cfg.JWT.Secret), featureHandler.UpdateFeature)
			features.DELETE("/:id", middleware.AuthMiddleware(cfg.JWT.Secret), featureHandler.DeleteFeature)
			features.GET("/my", middleware.AuthMiddleware(cfg.JWT.Secret), featureHandler.GetMyFeatures)

			// Voting routes
			features.POST("/:id/vote", middleware.AuthMiddleware(cfg.JWT.Secret), voteHandler.VoteForFeature)
			features.DELETE("/:id/vote", middleware.AuthMiddleware(cfg.JWT.Secret), voteHandler.RemoveVoteFromFeature)
			features.POST("/:id/toggle-vote", middleware.AuthMiddleware(cfg.JWT.Secret), voteHandler.ToggleVote)
		}

		// Vote routes
		votes := v1.Group("/votes")
		votes.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			votes.GET("/my", voteHandler.GetUserVotes)
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Swagger documentation available at: http://%s:%s/swagger/index.html", cfg.Server.Host, cfg.Server.Port)
	
	if err := r.Run(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}