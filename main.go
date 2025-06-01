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
	// --- Database ---
	db := mustSetupDatabase()
	defer db.Close()

	// --- Kafka Writers ---
	tweetsWriter := initKafkaWriter("tweets.created")
	fanoutWriter := initKafkaWriter("timeline.fanout")
	followWriter := initKafkaWriter("user.follow")
	defer tweetsWriter.Close()
	defer fanoutWriter.Close()
	defer followWriter.Close()

	// --- Repository and Publisher Initialization ---
	userRepo, followRepo, tweetRepo := initRepositories(db)
	tweetPub := adapters_publishers.NewKafkaTweetPublisher(tweetsWriter)
	fanoutPub := adapters_publishers.NewKafkaTimelineFanoutPublisher(fanoutWriter)
	followPub := adapters_publishers.NewKafkaFollowPublisher(followWriter)

	// --- Kafka Readers ---
	tweetCreateKafkaReader := initKafkaTweetCreateReader()
	fanoutKafkaReader := initKafkaFanoutReader()
	followKafkaReader := initKafkaFollowReader()
	defer tweetCreateKafkaReader.Close()
	defer fanoutKafkaReader.Close()
	defer followKafkaReader.Close()

	// --- Redis and Timeline Cache ---
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: "",
		DB:       0,
	})
	timelineCache := adapters_redis.NewTimelineCacheRedis(redisClient)

	// --- Start Consumers ---
	ctx := context.Background()
	go startTweetConsumer(ctx, tweetCreateKafkaReader, tweetRepo, fanoutPub, followRepo)
	go startFanoutConsumer(ctx, fanoutKafkaReader, timelineCache, followRepo)
	go startFollowConsumer(ctx, followKafkaReader, timelineCache, tweetRepo)

	// --- Services and Handlers ---
	userService, followService, tweetService := initServices(userRepo, followRepo, tweetRepo, tweetPub, followPub)
	followHandler, userHandler, tweetHandler, timelineHandler := initHandlers(userService, followService, tweetService)

	// --- HTTP Server ---
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

func initKafkaWriter(topic string) *kafka.Writer {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:29092"
	}
	return &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchSize:    1,
	}
}

func initKafkaTweetCreateReader() *kafka.Reader {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:29092"
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "tweets.created",
		GroupID: "tweet-consumer-group",
	})
}

func initKafkaFanoutReader() *kafka.Reader {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:29092"
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "timeline.fanout",
		GroupID: "fanout-consumer-group",
	})
}

func initKafkaFollowReader() *kafka.Reader {
	broker := getEnv("KAFKA_BROKER", "localhost:9092")
	groupID := getEnv("KAFKA_GROUP_ID", "timeline-follow-consumer-group")

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  groupID,
		Topic:    "user.follow",
		MinBytes: 10,   // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func initRepositories(db *sql.DB) (repoports.UserRepository, repoports.FollowRepository, repoports.TweetRepository) {
	userRepo := adapters_repositories.NewPostgreSQLUserRepository(db)
	followRepo := adapters_repositories.NewPostgreSQLFollowRepository(db)
	tweetRepo := adapters_repositories.NewPostgreSQLTweetRepository(db)
	return userRepo, followRepo, tweetRepo
}

func startTweetConsumer(ctx context.Context, reader *kafka.Reader, tweetRepo repoports.TweetRepository, fanoutPub pubports.TimelineFanoutPublisher, followRepo repoports.FollowRepository) {
	consumer := adapters_consumers.NewKafkaTweetConsumer(reader, tweetRepo, fanoutPub, followRepo)
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Error starting tweet consumer: %v", err)
	}
}

func startFanoutConsumer(ctx context.Context, reader *kafka.Reader, timelineCache repoports.TimelineCache, followRepo repoports.FollowRepository) {
	fanoutConsumer := adapters_consumers.NewKafkaTimelineFanoutConsumer(reader, timelineCache, followRepo)
	if err := fanoutConsumer.Start(ctx); err != nil {
		log.Printf("Error starting fanout consumer: %v", err)
	}
}

func startFollowConsumer(ctx context.Context, reader *kafka.Reader, timelineCache repoports.TimelineCache, tweetRepo repoports.TweetRepository) {
	followConsumer := adapters_consumers.NewKafkaFollowConsumer(reader, timelineCache, tweetRepo)
	if err := followConsumer.Start(ctx); err != nil {
		log.Printf("Error starting follow consumer: %v", err)
	}
}

func initServices(
	userRepo repoports.UserRepository,
	followRepo repoports.FollowRepository,
	tweetRepo repoports.TweetRepository,
	tweetPub pubports.TweetPublisher,
	followPub pubports.FollowPublisher,
) (*application.UserService, *application.FollowService, *application.TweetService) {
	userService := application.NewUserService(userRepo)
	followService := application.NewFollowService(userRepo, followRepo, followPub)
	tweetService := application.NewTweetService(tweetRepo, tweetPub)

	return userService, followService, tweetService
}

func initHandlers(
	userService *application.UserService,
	followService *application.FollowService,
	tweetService *application.TweetService,
) (followHandler *handlers.FollowHandler, userHandler *handlers.UserHandler, tweetHandler *handlers.TweetHandler, timelineHandler *handlers.TimelineHandler) {
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
