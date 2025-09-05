package models

// HTTPError represents an error response for Swagger documentation
// @Description HTTP error response
type HTTPError struct {
	// HTTP status code
	Code    int    `json:"code" example:"400"`
	// Error message
	Message string `json:"message" example:"Bad request"`
}