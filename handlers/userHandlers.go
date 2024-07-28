package handlers

import (
    "database/sql"
    "net/http"
	"os"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/dgrijalva/jwt-go"
    "time"
    "Price_Trigger/models"
)

type UserHandler struct {
    DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
    return &UserHandler{DB: db}
}

func (h *UserHandler) SignUp(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Insert user into database
    _, err = h.DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", user.Email, string(hashedPassword))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Retrieve the user from the database
    var dbUser models.User
    err := h.DB.QueryRow("SELECT email, password FROM users WHERE email = $1", user.Email).Scan(&dbUser.Email, &dbUser.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Check password
    err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "email": dbUser.Email,
        "exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
    })

    // Sign and get the complete encoded token as a string
	jwt_secret := os.Getenv("JWT_SECRET")
    tokenString, err := token.SignedString([]byte(jwt_secret)) // Replace with a secure secret key
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}