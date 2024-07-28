package routers

import (
    "github.com/gin-gonic/gin"
    "Price_Trigger/handlers"
    "Price_Trigger/auth"
)

func SetupTriggerRoutes(r *gin.Engine, alertHandler *handlers.AlertHandler) {
    triggerGroup := r.Group("/alerts")
    triggerGroup.Use(auth.AuthMiddleware())
    {
        triggerGroup.POST("/create", alertHandler.CreateAlert)
        triggerGroup.DELETE("/delete/:id", alertHandler.DeleteAlert)
        triggerGroup.GET("", alertHandler.GetAlerts)
    }
}

