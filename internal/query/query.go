package query

import (
	"github.com/cr34t1ve/hoprun/internal/database"
)

type Service interface {
	ExecuteQuery(query string) ([]map[string]interface{}, error)
	FormatResults(results []map[string]interface{}, visualization string) interface{}
}

type service struct {
	dbService database.Service
}

func NewService(dbService database.Service) Service {
	return &service{dbService: dbService}
}

func (s *service) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	return s.dbService.ExecuteRawQuery(query)
}

func (s *service) FormatResults(results []map[string]interface{}, visualization string) interface{} {
	// Implement formatting logic based on visualization type
	return results
}
