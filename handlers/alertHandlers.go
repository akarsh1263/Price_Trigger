package handlers

import (
    "database/sql"
    "encoding/json" // Importing the json package
    "net/http"
    "strconv" // Importing the strconv package

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "Price_Trigger/models"
)

type AlertHandler struct {
    DB          *sql.DB
    RedisClient *redis.Client
}

func NewAlertHandler(db *sql.DB, redisClient *redis.Client) *AlertHandler {
    return &AlertHandler{DB: db, RedisClient: redisClient}
}

func (h *AlertHandler) CreateAlert(c *gin.Context) {
    var alert models.Alert
    if err := c.ShouldBindJSON(&alert); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Assuming the email is set in the context by the AuthMiddleware
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    alert.Email = email.(string)
    alert.Status = "created"

    // Insert alert into database
    query := `INSERT INTO alerts (email, coin, target_price, status, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`
    err := h.DB.QueryRow(query, alert.Email, alert.Coin, alert.TargetPrice, alert.Status).Scan(&alert.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Add the created alert to Redis cache
    alertsJSON, err := h.RedisClient.Get(c, alert.Email).Result()
    if err != nil && err != redis.Nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var alerts []models.Alert
    if err == redis.Nil {
        // If there are no alerts in the cache, initialize the slice
        alerts = []models.Alert{}
    } else {
        // Unmarshal existing alerts from Redis
        if err := json.Unmarshal([]byte(alertsJSON), &alerts); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    }

    // Add the new alert to the list of alerts
    alerts = append(alerts, alert)

    // Marshal the updated list of alerts to JSON
    updatedAlertsJSON, err := json.Marshal(alerts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Update the Redis cache with the new alert
    if err := h.RedisClient.Set(c, alert.Email, updatedAlertsJSON, 0).Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, alert)
}

func (h *AlertHandler) DeleteAlert(c *gin.Context) {
    id := c.Param("id")
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    query := `DELETE FROM alerts WHERE id = $1 AND email = $2`
    result, err := h.DB.Exec(query, id, email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil || rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
        return
    }

    // Retrieve the current alerts from Redis
    alertsJSON, err := h.RedisClient.Get(c, email.(string)).Result()
    if err != nil && err != redis.Nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var alerts []models.Alert
    if err == redis.Nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "No alerts found in cache"})
        return
    } else {
        if err := json.Unmarshal([]byte(alertsJSON), &alerts); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    }

    // Convert id to int for comparison
    alertID, err := strconv.Atoi(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
        return
    }

    // Update the status of the alert to "deleted"
    alertFound := false
    for i, alert := range alerts {
        if alert.ID == alertID {
            alerts[i].Status = "deleted"
            alertFound = true
            break
        }
    }

    if !alertFound {
        c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
        return
    }

    // Marshal the updated list of alerts to JSON
    updatedAlertsJSON, err := json.Marshal(alerts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Update the Redis cache with the new alert status
    if err := h.RedisClient.Set(c, email.(string), updatedAlertsJSON, 0).Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Alert deleted"})
}


func (h *AlertHandler) GetAlerts(c *gin.Context) {
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Pagination and filtering
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    status := c.DefaultQuery("status", "")

    // Check cache first
    cacheKey := email.(string)
    cachedAlerts, err := h.RedisClient.Get(c, cacheKey).Result()
    if err == nil {
        // Deserialize cached alerts
        var alerts []models.Alert
        if err := json.Unmarshal([]byte(cachedAlerts), &alerts); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Filter alerts based on status
        if status != "" {
            var filteredAlerts []models.Alert
            for _, alert := range alerts {
                if alert.Status == status {
                    filteredAlerts = append(filteredAlerts, alert)
                }
            }
            alerts = filteredAlerts
        }

        // Paginate the alerts
        start := (page - 1) * pageSize
        end := start + pageSize
        if start > len(alerts) {
            start = len(alerts)
        }
        if end > len(alerts) {
            end = len(alerts)
        }
        paginatedAlerts := alerts[start:end]

        c.JSON(http.StatusOK, gin.H{"alerts": paginatedAlerts})
        return
    } else if err == redis.Nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "No alerts found in cache"})
        return
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
}


