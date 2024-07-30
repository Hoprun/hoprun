package nlp

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Service interface {
	NaturalLanguageToSQL(query string, dbSchema string) (string, error)
}

type service struct {
	client *openai.Client
}

func NewService(apiKey string) Service {
	return &service{
		client: openai.NewClient(apiKey),
	}
}

func (s *service) NaturalLanguageToSQL(query string, dbSchema string) (string, error) {
	prompt := fmt.Sprintf(`You are a SQL expert. Given the following database schema:

%s

Convert the following natural language query to SQL:
%s

Return only the SQL query without any markdown formatting, explanations, or additional text.`, dbSchema, query)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	sqlQuery := resp.Choices[0].Message.Content
	sqlQuery = strings.TrimSpace(sqlQuery)
	sqlQuery = strings.TrimPrefix(sqlQuery, "```sql")
	sqlQuery = strings.TrimPrefix(sqlQuery, "```")
	sqlQuery = strings.TrimSuffix(sqlQuery, "```")
	sqlQuery = strings.TrimSpace(sqlQuery)

	return sqlQuery, nil
}
