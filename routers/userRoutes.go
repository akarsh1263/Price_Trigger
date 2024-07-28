package routers

import (
    "github.com/gin-gonic/gin"
    "Price_Trigger/handlers"
)

func SetupUserRoutes(r *gin.Engine, userHandler *handlers.UserHandler) {
    userGroup := r.Group("/users")
    {
        userGroup.POST("/signup", userHandler.SignUp)
        userGroup.POST("/login", userHandler.Login)
    }
}