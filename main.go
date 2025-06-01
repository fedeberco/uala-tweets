package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	adapters_consumers "uala-tweets/internal/adapters/consumers"
	adapters_publishers "uala-tweets/internal/adapters/publishers"
	adapters_redis "uala-tweets/internal/adapters/redis"
	adapters_repositories "uala-tweets/internal/adapters/repositories"
	"uala-tweets/internal/application"
	"uala-tweets/internal/interfaces/handlers"

	pubports "uala-tweets/internal/ports/publishers"
	repoports "uala-tweets/internal/ports/repositories"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	db := mustSetupDatabase()
	defer db.Close()

	kafkaWriter := initKafkaWriter()
	defer kafkaWriter.Close()

	// --- Repositories and Publishers Initialization ---
	userRepo, followRepo, tweetRepo := initRepositories(db)
	tweetPub := adapters_publishers.NewKafkaTweetPublisher(kafkaWriter)
	fanoutPub := adapters_publishers.NewKafkaTimelineFanoutPublisher(kafkaWriter)

	// --- Consumers Initialization ---
	kafkaReader := initKafkaReader()
	fanoutKafkaReader := initKafkaFanoutReader()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: "",
		DB:       0,
	})
	timelineCache := adapters_redis.NewTimelineCacheRedis(redisClient)

	// Start consumers together
	go startTweetConsumer(kafkaReader, tweetRepo, fanoutPub)
	fanoutConsumer := adapters_consumers.NewKafkaTimelineFanoutConsumer(fanoutKafkaReader, timelineCache, followRepo)
	go func() {
		if err := fanoutConsumer.Start(context.Background()); err != nil {
			log.Printf("[ERROR] Timeline fanout consumer exited: %v", err)
		}
	}()

	tweetService := application.NewTweetService(tweetRepo, tweetPub, fanoutPub, followRepo)
	userService, followService, _ := initServices(userRepo, followRepo, tweetRepo, tweetPub)
	followHandler, userHandler, tweetHandler, timelineHandler := initHandlers(userService, followService, tweetService)

	r := setupRouter(followHandler, userHandler, tweetHandler, timelineHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initTimelineHandler() *handlers.TimelineHandler {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: "",
		DB:       0,
	})
	timelineCache := adapters_redis.NewTimelineCacheRedis(redisClient)
	timelineService := application.NewTimelineService(timelineCache)
	return handlers.NewTimelineHandler(timelineService)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustSetupDatabase() *sql.DB {
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func initKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "tweets.created",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchSize:    1,
	}
}

func initKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "tweets.created",
		GroupID: "tweet-consumer-group",
	})
}

func initKafkaFanoutReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "timeline.fanout",
		GroupID: "timeline-fanout-consumer-group",
	})
}

func initRepositories(db *sql.DB) (repoports.UserRepository, repoports.FollowRepository, repoports.TweetRepository) {
	userRepo := adapters_repositories.NewPostgreSQLUserRepository(db)
	followRepo := adapters_repositories.NewPostgreSQLFollowRepository(db)
	tweetRepo := adapters_repositories.NewPostgreSQLTweetRepository(db)
	return userRepo, followRepo, tweetRepo
}

func startTweetConsumer(reader *kafka.Reader, tweetRepo repoports.TweetRepository, fanoutPub pubports.TimelineFanoutPublisher) {
	consumer := adapters_consumers.NewKafkaTweetConsumer(reader, tweetRepo, fanoutPub)
	go func() {
		log.Println("KafkaTweetConsumer started...")
		if err := consumer.Start(context.Background()); err != nil {
			log.Printf("KafkaTweetConsumer error: %v", err)
		}
	}()
}

func initServices(userRepo repoports.UserRepository, followRepo repoports.FollowRepository, tweetRepo repoports.TweetRepository, tweetPub pubports.TweetPublisher) (*application.UserService, *application.FollowService, *application.TweetService) {
	userService := application.NewUserService(userRepo)
	followService := application.NewFollowService(userRepo, followRepo)
	tweetService := application.NewTweetService(tweetRepo, tweetPub)
	return userService, followService, tweetService
}

func initHandlers(userService *application.UserService, followService *application.FollowService, tweetService *application.TweetService) (followHandler *handlers.FollowHandler, userHandler *handlers.UserHandler, tweetHandler *handlers.TweetHandler, timelineHandler *handlers.TimelineHandler) {
	followHandler = handlers.NewFollowHandler(followService)
	userHandler = handlers.NewUserHandler(userService)
	tweetHandler = handlers.NewTweetHandler(tweetService)
	timelineHandler = initTimelineHandler()
	return
}

func setupRouter(followHandler *handlers.FollowHandler, userHandler *handlers.UserHandler, tweetHandler *handlers.TweetHandler, timelineHandler *handlers.TimelineHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.GET("/:id", userHandler.GetUser)
		userRoutes.POST("/:id/follow/:target_id", followHandler.FollowUser)
		userRoutes.POST("/:id/unfollow/:target_id", followHandler.UnfollowUser)
	}

	tweetRoutes := r.Group("/tweets")
	{
		tweetRoutes.POST("", tweetHandler.CreateTweet)
		tweetRoutes.GET("/:id", tweetHandler.GetTweet)
	}

	r.GET("/timelines/:user_id", timelineHandler.GetTimelineHandler)
	return r
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
