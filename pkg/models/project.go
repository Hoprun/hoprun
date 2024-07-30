package models

import "time"

type Project struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
