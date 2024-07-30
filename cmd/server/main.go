package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cr34t1ve/hoprun/internal/api"
	"github.com/cr34t1ve/hoprun/internal/auth"
	"github.com/cr34t1ve/hoprun/internal/database"
	"github.com/cr34t1ve/hoprun/internal/nlp"
	"github.com/cr34t1ve/hoprun/internal/query"
)

func main() {
	// Initialize database
	dsn := "host=localhost user=postgres dbname=hoprun password=boondocks sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	dbService := database.NewService(db)
	nlpService := nlp.NewService(os.Getenv("JWT_SECRET"))
	queryService := query.NewService(dbService)
	authService := auth.NewService(dbService)

	// Initialize handler
	handler := api.NewHandler(nlpService, queryService, dbService, authService)

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/register", handler.Register).Methods("POST")
	r.HandleFunc("/login", handler.Login).Methods("POST")
	r.HandleFunc("/project", handler.CreateProject).Methods("POST")
	r.HandleFunc("/getproject", handler.ListUserProjects).Methods("POST")
	r.HandleFunc("/query", handler.HandleQuery).Methods("POST")

	// Start server
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
