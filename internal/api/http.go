package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cr34t1ve/hoprun/internal/database"
	"github.com/cr34t1ve/hoprun/internal/nlp"
	"github.com/cr34t1ve/hoprun/internal/query"
	"github.com/cr34t1ve/hoprun/pkg/models"
)

type Handler struct {
	nlpService   nlp.Service
	queryService query.Service
	dbService    database.Service
}

func NewHandler(nlpService nlp.Service, queryService query.Service, dbService database.Service) *Handler {
	return &Handler{
		nlpService:   nlpService,
		queryService: queryService,
		dbService:    dbService,
	}
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
