package models

type QueryInput struct {
	ProjectID     int    `json:"project_id"`
	Query         string `json:"query"`
	Visualization string `json:"visualization"`
}
