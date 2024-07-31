package models

import "time"

type DatabaseConnection struct {
	ID         int       `json:"id"`
	ProjectID  int       `json:"projecct_id"`
	DBName     string    `json:"db_name"`
	DBUser     string    `json:"db_user"`
	DBPassword string    `json:"db_password"`
	DBHost     string    `json:"db_host"`
	DBPort     string    `json:"db_port" gorm:"default:5432"`
	CreatedAt  time.Time `json:"created_at"`
}
