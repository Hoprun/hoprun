package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cr34t1ve/hoprun/internal/auth"
	"github.com/cr34t1ve/hoprun/internal/database"
	databaseconnection "github.com/cr34t1ve/hoprun/internal/database_connection"
	"github.com/cr34t1ve/hoprun/internal/nlp"
	"github.com/cr34t1ve/hoprun/internal/query"
	"github.com/cr34t1ve/hoprun/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	nlpService         nlp.Service
	queryService       query.Service
	dbService          database.Service
	authService        auth.Service
	databaseconnection databaseconnection.Service
}

func NewHandler(nlpService nlp.Service, queryService query.Service, dbService database.Service, authService auth.Service, databaseconnection databaseconnection.Service) *Handler {
	return &Handler{
		nlpService:         nlpService,
		queryService:       queryService,
		dbService:          dbService,
		authService:        authService,
		databaseconnection: databaseconnection,
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

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int    `json:"user_id"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.authService.AddProject(r.Context(), input.UserID, input.Name)
	if err != nil {
		http.Error(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) ListUserProjects(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"user_id"`
	}
	checkDecoding(w, r.Body, &input)

	projects, err := h.authService.ListProjects(r.Context(), input.UserID)
	if err != nil {
		http.Error(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) AddConnection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  int    `json:"projecct_id"`
		DBName     string `json:"db_name"`
		DBUser     string `json:"db_user"`
		DBPassword string `json:"db_password"`
		DBHost     string `json:"db_host"`
		DBPort     string `json:"db_port"`
	}
	checkDecoding(w, r.Body, &input)

	connection, err := h.databaseconnection.AddConnection(r.Context(), input.ProjectID, input.DBName, input.DBUser, input.DBPassword, input.DBHost, input.DBPort)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			http.Error(w, "Failed to add database connection", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Failed to add connection: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(connection)
}

func (h *Handler) ListDBConns(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID int `json:"project_id"`
	}
	checkDecoding(w, r.Body, &input)

	projects, err := h.databaseconnection.ListProjectConnections(r.Context(), input.ProjectID)
	if err != nil {
		http.Error(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var input models.QueryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get database connection settings for project
	dbConn, err := h.databaseconnection.GetProjectConnection(r.Context(), input.ProjectID)
	if err != nil {
		http.Error(w, "Failed to get database schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: decrpyt db connection password

	// get db schema from db call
	// NB: Ideally this would be saved in our db but here lies the case where the database structure could
	// be changed by the user during calls. To cut out this entire process of getting the schema on each call,
	// there would be have to be a notifier from either the backend, migration tool or database (preferably)
	// anytime there is a change in the schema so a job is started to update the schema on our end and proceed
	// accordingly

	// start db connection
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", dbConn.DBHost, dbConn.DBUser, dbConn.DBName, dbConn.DBPassword)
	log.Printf("db connection gotten: %s", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		http.Error(w, "Failed to connect to user PostgreSQL connection"+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: work on this to be able to close a specific connection
	// https://stackoverflow.com/a/74299379
	// defer func () {
	// 	dbInstance, _ := db.DB()
	// 	_ = dbInstance.Close()
	// }()

	userDBService := database.NewService(db)
	userQueryService := query.NewService(userDBService)
	dbSchema, err := userDBService.GetDatabaseSchema()
	if err != nil {
		http.Error(w, "Failed to get database schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// pass schema to Natural language converter
	sqlQuery, err := h.nlpService.NaturalLanguageToSQL(input.Query, dbSchema)
	if err != nil {
		http.Error(w, "Failed to generate SQL query"+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Generated SQL query: %s", sqlQuery)

	results, err := userQueryService.ExecuteQuery(sqlQuery)
	if err != nil {
		http.Error(w, "Failed to execute query"+err.Error(), http.StatusInternalServerError)
		return
	}

	formattedResults := h.queryService.FormatResults(results, input.Visualization)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formattedResults)
}

func checkDuplicateKeyError(err error, w http.ResponseWriter, message string) {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
}

func checkForeignKeyError(err error, w http.ResponseWriter, message string) bool {
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		http.Error(w, message, http.StatusInternalServerError)
		return false
	}
	return true
}

func checkDecoding(w http.ResponseWriter, body io.ReadCloser, dec any) {
	if err := json.NewDecoder(body).Decode(&dec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
