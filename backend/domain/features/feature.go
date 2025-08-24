package features

import (
	"time"
)

// Feature represents the core feature entity
type Feature struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CreatedBy       int       `json:"created_by"`
	CreatedByUser   *string   `json:"created_by_user,omitempty"`
	VoteCount       int       `json:"vote_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	HasUserVoted    bool      `json:"has_user_voted,omitempty"`
}

// CreateFeatureRequest represents the data needed to create a feature
type CreateFeatureRequest struct {
	Title       string `json:"title" binding:"required,min=5,max=255"`
	Description string `json:"description" binding:"required,min=10"`
}

// UpdateFeatureRequest represents the data needed to update a feature
type UpdateFeatureRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=5,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=10"`
}

// FeatureListResponse represents paginated feature list response
type FeatureListResponse struct {
	Features []Feature `json:"features"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}