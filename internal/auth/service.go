package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/cr34t1ve/hoprun/internal/database"
	"github.com/cr34t1ve/hoprun/pkg/models"
)

type Service interface {
	RegisterUser(ctx context.Context, email, password string) (*models.User, error)
	LoginUser(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	AddProject(ctx context.Context, userID int, name string) (*models.Project, error)
	ListProjects(ctx context.Context, userID int) (*[]models.Project, error)
}

type service struct {
	dbService database.Service
	jwtKey    []byte
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func NewService(dbService database.Service) Service {
	return &service{
		dbService: dbService,
		jwtKey:    []byte(os.Getenv("JWT_SECRET")),
	}
}

func (s *service) RegisterUser(ctx context.Context, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.dbService.CreateUser(ctx, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.dbService.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *service) generateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

func (s *service) AddProject(ctx context.Context, userID int, name string) (*models.Project, error) {
	project, err := s.dbService.CreateProject(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	return project, err
}

func (s *service) ListProjects(ctx context.Context, userID int) (*[]models.Project, error) {
	projects, err := s.dbService.ListProjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	return projects, err
}
