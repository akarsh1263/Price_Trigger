package main

import (
    "log"

    "github.com/gin-gonic/gin"

    "Price_Trigger/handlers"
    "Price_Trigger/routers"
    "Price_Trigger/monitor"
)

func main() {
    // Database connection
    db, err := InitDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Redis connection
    redisClient, err := InitRedis()
    if err != nil {
        log.Fatal(err)
    }
    defer redisClient.Close()

    // Router setup
    r := gin.Default()

    // Initialize handlers
    userHandler := handlers.NewUserHandler(db)
    alertHandler := handlers.NewAlertHandler(db, redisClient)

    // Setup routes
    routers.SetupUserRoutes(r, userHandler)
    routers.SetupTriggerRoutes(r, alertHandler)

    // Start WebSocket connection and price monitoring in a separate goroutine
    go monitor.MonitorPrices(db,redisClient)

    // Start server
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}