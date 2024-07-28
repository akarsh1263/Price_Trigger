package auth

import (
    "fmt"
    "strings"
	"os"

    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
)

var secretKey = []byte(os.Getenv("JWT_SECRET")) 

// Claims struct to define JWT claims
type Claims struct {
    Email string `json:"email"`
    jwt.StandardClaims
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // The Authorization header should be in the format: "Bearer <token>"
        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid Authorization header format"})
            c.Abort()
            return
        }

        token := bearerToken[1]
        claims, err := validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token: " + err.Error()})
            c.Abort()
            return
        }

        // Set the user ID in the context for use in subsequent handlers
        c.Set("email", claims.Email)
        c.Next()
    }
}

func validateToken(tokenString string) (*Claims, error) {
    // Parse the token
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate the signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secretKey, nil
    })

    if err != nil {
        return nil, err
    }

    // Validate the token and return the claims
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}