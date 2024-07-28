package main 

import (
    "database/sql"
    "fmt"
    "os"

    "github.com/go-redis/redis/v8"
    _ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
        dbUser, dbPassword, dbHost, dbName)

    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }

    // Test the connection
    err = db.Ping()
    if err != nil {
        return nil, fmt.Errorf("failed to ping database: %v", err)
    }

    return db, nil
}

func InitRedis() (*redis.Client, error) {
    redisAddr := os.Getenv("REDIS_ADDR")
    redisClient := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    // Ping Redis to check connection
    _, err := redisClient.Ping(redisClient.Context()).Result()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %v", err)
    }

    return redisClient, nil
}