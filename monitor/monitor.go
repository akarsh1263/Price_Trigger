package monitor

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/url"
    "strconv"

    "github.com/gorilla/websocket"
    "github.com/go-redis/redis/v8"
    "context"
    "Price_Trigger/models"
)

var ctx = context.Background()

func MonitorPrices(db *sql.DB, redisClient *redis.Client) {
    u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@trade"}
    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer c.Close()

    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }

        var trade struct {
            Price string `json:"p"`
        }
        json.Unmarshal(message, &trade)

        price, _ := strconv.ParseFloat(trade.Price, 64)
        checkAlerts(db, redisClient, "BTC", price)
    }
}

func checkAlerts(db *sql.DB, redisClient *redis.Client, coin string, currentPrice float64) {
    rows, err := db.Query("SELECT id, email, target_price FROM alerts WHERE coin = $1 AND status = 'created'", coin)
    if err != nil {
        log.Println("Failed to fetch alerts:", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var alertID int
        var email string
        var targetPrice float64
        err := rows.Scan(&alertID, &email, &targetPrice)
        if err != nil {
            log.Println("Failed to scan alert:", err)
            continue
        }

        if (targetPrice >= currentPrice && targetPrice < currentPrice*1.001) ||
            (targetPrice <= currentPrice && targetPrice > currentPrice*0.999) {
            // Alert triggered
            _, err := db.Exec("UPDATE alerts SET status = 'triggered' WHERE id = $1", alertID)
            if err != nil {
                log.Println("Failed to update alert status:", err)
            }

            // Update Redis cache
            updateRedisCache(redisClient, email, alertID, "triggered")

            // Send email (or print to console)
            sendAlert(email, coin, targetPrice, currentPrice)
        }
    }
}

func updateRedisCache(redisClient *redis.Client, email string, alertID int, status string) {
    cacheKey := email
    cachedAlerts, err := redisClient.Get(ctx, cacheKey).Result()
    if err != nil {
        log.Println("Failed to get alerts from Redis:", err)
        return
    }

    var alerts []models.Alert
    if err := json.Unmarshal([]byte(cachedAlerts), &alerts); err != nil {
        log.Println("Failed to unmarshal alerts from Redis:", err)
        return
    }

    for i, alert := range alerts {
        if alert.ID == alertID {
            alerts[i].Status = status
            break
        }
    }

    updatedAlertsJSON, err := json.Marshal(alerts)
    if err != nil {
        log.Println("Failed to marshal updated alerts:", err)
        return
    }

    if err := redisClient.Set(ctx, cacheKey, updatedAlertsJSON, 0).Err(); err != nil {
        log.Println("Failed to update alerts in Redis:", err)
    }
}

func sendAlert(email, coin string, targetPrice, currentPrice float64) {
    // For simplicity, just print to console
    log.Printf("Alert triggered for email %s: %s reached %.2f (target: %.2f)\n", email, coin, currentPrice, targetPrice)

    // TODO: Implement email sending using Gmail SMTP
}