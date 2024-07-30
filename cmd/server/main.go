package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cr34t1ve/hoprun/internal/api"
	"github.com/cr34t1ve/hoprun/internal/database"
	"github.com/cr34t1ve/hoprun/internal/nlp"
	"github.com/cr34t1ve/hoprun/internal/query"
)

func main() {
	// Initialize database
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres dbname=hoprun_dummy password=boondocks sslmode=disable"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	dbService := database.NewService(db)
	nlpService := nlp.NewService("sk-proj-W1ksXGClT51j5nw5EUfYT3BlbkFJW66eo9YDwU1jfAaYY7WH")
	queryService := query.NewService(dbService)

	// Initialize handler
	handler := api.NewHandler(nlpService, queryService, dbService)

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/query", handler.HandleQuery).Methods("POST")

	// Start server
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
