package main

import (
	"log"

	"github.com/feature-voting-platform/backend/adapters/auth"
	"github.com/feature-voting-platform/backend/adapters/logs"
	"github.com/feature-voting-platform/backend/adapters/postgres"
	"github.com/feature-voting-platform/backend/adapters/rest"
	"github.com/feature-voting-platform/backend/internal/config"
	"github.com/gin-gonic/gin"

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

	// Initialize logger
	logger := logs.NewJSONLogger()

	// Test our custom logger
	logger.Info("Testing custom logger on server startup")

	// Initialize database
	db, err := postgres.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	featureRepo := postgres.NewFeatureRepository(db)

	// Initialize auth services
	tokenService := auth.NewJWTService(cfg.JWT.Secret)
	passwordService := auth.NewBCryptPasswordService()

	// Initialize handlers
	authHandler := rest.NewAuthHandler(userRepo, tokenService, passwordService, logger)
	featureHandler := rest.NewFeatureHandler(featureRepo, logger)
	voteHandler := rest.NewVoteHandler(featureRepo, featureRepo, logger)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(rest.CORSMiddleware())
	r.Use(rest.LoggingMiddleware(logger))
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
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", rest.AuthMiddleware(tokenService), authHandler.GetProfile)
		}

		// Feature routes
		features := v1.Group("/features")
		{
			// Public routes (with optional auth for vote status)
			features.GET("", rest.OptionalAuthMiddleware(tokenService), featureHandler.GetFeatures)
			features.GET("/:id", rest.OptionalAuthMiddleware(tokenService), featureHandler.GetFeature)

			// Protected routes
			features.POST("", rest.AuthMiddleware(tokenService), featureHandler.CreateFeature)
			features.PUT("/:id", rest.AuthMiddleware(tokenService), featureHandler.UpdateFeature)
			features.DELETE("/:id", rest.AuthMiddleware(tokenService), featureHandler.DeleteFeature)
			features.GET("/my", rest.AuthMiddleware(tokenService), featureHandler.GetMyFeatures)

			// Voting routes
			features.POST("/:id/vote", rest.AuthMiddleware(tokenService), voteHandler.VoteForFeature)
			features.DELETE("/:id/vote", rest.AuthMiddleware(tokenService), voteHandler.RemoveVoteFromFeature)
			features.POST("/:id/toggle-vote", rest.AuthMiddleware(tokenService), voteHandler.ToggleVote)
		}

		// Vote routes
		votes := v1.Group("/votes")
		votes.Use(rest.AuthMiddleware(tokenService))
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