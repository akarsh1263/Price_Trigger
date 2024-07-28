package models

import (
    "time"
)

type Alert struct {
    ID         int       `json:"id" gorm:"primary_key"`
    Email      string    `json:"email"`
    Coin       string    `json:"coin"`
    TargetPrice float64  `json:"target_price"`
    Status     string    `json:"status"`
    CreatedAt  time.Time `json:"created_at"`
}
