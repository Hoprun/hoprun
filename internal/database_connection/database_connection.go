package databaseconnection

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"gorm.io/gorm"

	"github.com/cr34t1ve/hoprun/pkg/models"
)

type Service interface {
	AddConnection(ctx context.Context, projectID int, dbName, dbUser, dbPassword, dbHost, dbPort string) (*models.DatabaseConnection, error)
	ListProjectConnections(ctx context.Context, projectID int) (*[]models.DatabaseConnection, error)
	GetProjectConnection(ctx context.Context, projectID int) (*models.DatabaseConnection, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) AddConnection(ctx context.Context, projectID int, dbName, dbUser, dbPassword, dbHost, dbPort string) (*models.DatabaseConnection, error) {
	count, err := s.checkForConnectionsLength(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if count >= 1 {
		return nil, errors.New("connection count limit reached")
	}

	// hashedPassword, err := encryptPassword(dbPassword)
	// if err != nil {
	// 	return nil, err
	// }

	databaseConnection := &models.DatabaseConnection{
		ProjectID:  projectID,
		DBName:     dbName,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBHost:     dbHost,
		DBPort:     dbPort,
	}
	result := s.db.WithContext(ctx).Create(databaseConnection)
	if result.Error != nil {
		return nil, result.Error
	}
	return databaseConnection, nil
}

func (s *service) ListProjectConnections(ctx context.Context, projectID int) (*[]models.DatabaseConnection, error) {
	var connections []models.DatabaseConnection
	results := s.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&connections)
	if results.Error != nil {
		return nil, results.Error
	}
	return &connections, nil
}

func (s *service) GetProjectConnection(ctx context.Context, projectID int) (*models.DatabaseConnection, error) {
	var connections models.DatabaseConnection
	results := s.db.WithContext(ctx).Where("project_id = ?", projectID).First(&connections)
	if results.Error != nil {
		return nil, results.Error
	}

	// decryptedPassword, err := decryptPassword(connections.DBPassword)
	// if err != nil {
	// 	return nil, err
	// }

	// connections.DBPassword = decryptedPassword

	return &connections, nil
}

func (s *service) checkForConnectionsLength(ctx context.Context, projectID int) (int64, error) {
	var count int64
	c := s.db.Table("database_connections").WithContext(ctx).Where("project_id = ?", projectID).Count(&count)
	if c.Error != nil {
		return -1, c.Error
	}
	return count, nil
}

func encryptPassword(password string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY")) // 32 bytes for AES-256
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptPassword(encryptedPassword string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY")) // 32 bytes for AES-256
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText, err := base64.RawStdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", err
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
