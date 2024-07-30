package database

import (
	"context"
	"fmt"

	"github.com/cr34t1ve/hoprun/pkg/models"
	"gorm.io/gorm"
)

type Service interface {
	ExecuteRawQuery(query string) ([]map[string]interface{}, error)
	GetDatabaseSchema() (string, error)
	CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) ExecuteRawQuery(query string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Raw(query).Scan(&results).Error
	return results, err
}

func (s *service) GetDatabaseSchema() (string, error) {
	var tables []string
	err := s.db.Raw(`
	SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public'
	`).Scan(&tables).Error
	if err != nil {
		return "", err
	}

	var schema string
	for _, table := range tables {
		var columns []struct {
			ColumnName string `gorm:"column:column_name"`
			DataType   string `gorm:"column:data_type"`
		}
		err := s.db.Raw(`
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = ?
		`, table).Scan(&columns).Error
		if err != nil {
			return "", err
		}

		schema += fmt.Sprintf("Table %s:\n", table)
		for _, col := range columns {
			schema += fmt.Sprintf("  %s (%s)\n", col.ColumnName, col.DataType)
		}
		schema += "\n"
	}

	return schema, nil
}

func (s *service) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	result := s.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
