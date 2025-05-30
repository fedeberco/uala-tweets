package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"uala-tweets/internal/adapters/repositories"
	"uala-tweets/internal/application"
	"uala-tweets/internal/interfaces/http/handlers"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewPostgreSQLUserRepository(db)
	followRepo := repositories.NewPostgreSQLFollowRepository(db)

	// Initialize services
	followService := application.NewFollowService(userRepo, followRepo)

	// Initialize HTTP handlers
	followHandler := handlers.NewFollowHandler(followService)
	userHandler := handlers.NewUserHandler(userRepo)

	// Set up router
	r := gin.Default()

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// User routes
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.GET("/:id", userHandler.GetUser)
		userRoutes.POST("/:id/follow/:target_id", followHandler.FollowUser)
		userRoutes.POST("/:id/unfollow/:target_id", followHandler.UnfollowUser)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupDatabase() (*sql.DB, error) {
	// Read database connection string from environment variable
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://uala:ualapass@localhost:5432/uala_tweets?sslmode=disable"
	}

	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")
	return db, nil
}
