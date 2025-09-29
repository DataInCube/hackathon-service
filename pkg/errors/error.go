package errors

import "github.com/DataInCube/hackathon-service/internal/models"

// HTTPError represents an error response for Swagger documentation
// @Description HTTP error response
type HTTPError struct {
	// HTTP status code
	Code int `json:"code" example:"400"`
	// Error message
	Message string `json:"message" example:"Bad request"`
}

// HackathonResponse represents a successful response for Swagger documentation
// @Description Hackathon response
type HackathonResponse struct {
	// HTTP status code
	Code int `json:"code" example:"200"`
	// Success message
	Message string `json:"message" example:"Hackathon created successfully"`
	// Data payload
	Data models.Hackathon `json:"data"`
}
