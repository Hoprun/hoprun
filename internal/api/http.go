package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/cr34t1ve/hoprun/internal/auth"
	"github.com/cr34t1ve/hoprun/internal/database"
	"github.com/cr34t1ve/hoprun/internal/nlp"
	"github.com/cr34t1ve/hoprun/internal/query"
	"github.com/cr34t1ve/hoprun/pkg/models"
	"gorm.io/gorm"
)

type Handler struct {
	nlpService   nlp.Service
	queryService query.Service
	dbService    database.Service
	authService  auth.Service
}

func NewHandler(nlpService nlp.Service, queryService query.Service, dbService database.Service, authService auth.Service) *Handler {
	return &Handler{
		nlpService:   nlpService,
		queryService: queryService,
		dbService:    dbService,
		authService:  authService,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.RegisterUser(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			http.Error(w, "User already exists", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.authService.LoginUser(r.Context(), input.Email, input.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *Handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var input models.QueryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbSchema, err := h.dbService.GetDatabaseSchema()
	if err != nil {
		http.Error(w, "Failed to get database schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sqlQuery, err := h.nlpService.NaturalLanguageToSQL(input.Query, dbSchema)
	if err != nil {
		http.Error(w, "Failed to generate SQL query"+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Generated SQL query: %s", sqlQuery)

	results, err := h.queryService.ExecuteQuery(sqlQuery)
	if err != nil {
		http.Error(w, "Failed to execute query"+err.Error(), http.StatusInternalServerError)
		return
	}

	formattedResults := h.queryService.FormatResults(results, input.Visualization)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formattedResults)
}
